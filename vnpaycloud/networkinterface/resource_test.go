package networkinterface

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// testNetworkInterface returns a fully populated dto.NetworkInterface for use in tests.
func testNetworkInterface() dto.NetworkInterface {
	return dto.NetworkInterface{
		ID:                  "nic-001",
		Name:                "test-nic",
		NetworkID:           "net-001",
		SubnetID:            "subnet-001",
		IPAddress:           "10.0.0.5",
		MACAddress:          "fa:16:3e:aa:bb:cc",
		Status:              "active",
		SecurityGroups:      []string{"sg-001", "sg-002"},
		PortSecurityEnabled: true,
		NetworkType:         "vxlan",
		Description:         "a test network interface",
		CreatedAt:           "2025-01-15T10:00:00Z",
		ProjectID:           testhelpers.TestProjectID,
		ZoneID:              testhelpers.TestZoneID,
	}
}

func TestEmptySecurityGroupsConfig(t *testing.T) {
	cases := []struct {
		name string
		raw  cty.Value
		want bool
	}{
		{
			name: "explicit empty security groups",
			raw: cty.ObjectVal(map[string]cty.Value{
				"security_groups": cty.SetValEmpty(cty.String),
			}),
			want: true,
		},
		{
			name: "omitted security groups",
			raw: cty.ObjectVal(map[string]cty.Value{
				"security_groups": cty.NullVal(cty.Set(cty.String)),
			}),
			want: false,
		},
		{
			name: "non-empty security groups",
			raw: cty.ObjectVal(map[string]cty.Value{
				"security_groups": cty.SetVal([]cty.Value{cty.StringVal("sg-001")}),
			}),
			want: false,
		},
		{
			name: "unknown security groups",
			raw: cty.ObjectVal(map[string]cty.Value{
				"security_groups": cty.UnknownVal(cty.Set(cty.String)),
			}),
			want: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := emptySecurityGroupsConfig(tc.raw)
			if got != tc.want {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

func TestInvalidNetworkInterfaceSecurityGroupsConfig(t *testing.T) {
	cases := []struct {
		name              string
		raw               cty.Value
		securityGroupsLen int
		want              bool
	}{
		{
			name: "port security false with security groups",
			raw: cty.ObjectVal(map[string]cty.Value{
				"port_security_enabled": cty.False,
				"security_groups":       cty.SetVal([]cty.Value{cty.StringVal("sg-001")}),
			}),
			securityGroupsLen: 1,
			want:              true,
		},
		{
			name: "port security true with security groups",
			raw: cty.ObjectVal(map[string]cty.Value{
				"port_security_enabled": cty.True,
				"security_groups":       cty.SetVal([]cty.Value{cty.StringVal("sg-001")}),
			}),
			securityGroupsLen: 1,
			want:              false,
		},
		{
			name: "port security false without configured security groups",
			raw: cty.ObjectVal(map[string]cty.Value{
				"port_security_enabled": cty.False,
				"security_groups":       cty.NullVal(cty.Set(cty.String)),
			}),
			securityGroupsLen: 0,
			want:              false,
		},
		{
			name: "computed port security with security groups",
			raw: cty.ObjectVal(map[string]cty.Value{
				"port_security_enabled": cty.NullVal(cty.Bool),
				"security_groups":       cty.SetVal([]cty.Value{cty.StringVal("sg-001")}),
			}),
			securityGroupsLen: 1,
			want:              false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := invalidNetworkInterfaceSecurityGroupsConfig(tc.raw, tc.securityGroupsLen)
			if got != tc.want {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

func TestResourceNetworkInterfaceCreate(t *testing.T) {
	ni := testNetworkInterface()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "POST",
			Pattern: "/v2/iac/projects/test-project-id/network-interfaces",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.NetworkInterfaceResponse{NetworkInterface: ni}),
		},
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/network-interfaces/nic-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.NetworkInterfaceResponse{NetworkInterface: ni}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceNetworkInterface()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":        "test-nic",
		"subnet_id":   "subnet-001",
		"ip_address":  "10.0.0.5",
		"description": "a test network interface",
	})

	diags := res.CreateContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "nic-001" {
		t.Errorf("expected ID nic-001, got %s", d.Id())
	}
	if v := d.Get("name").(string); v != "test-nic" {
		t.Errorf("expected name test-nic, got %s", v)
	}
	if v := d.Get("status").(string); v != "active" {
		t.Errorf("expected status active, got %s", v)
	}
	if v := d.Get("mac_address").(string); v != "fa:16:3e:aa:bb:cc" {
		t.Errorf("expected mac_address fa:16:3e:aa:bb:cc, got %s", v)
	}
}

func TestResourceNetworkInterfaceRead(t *testing.T) {
	ni := testNetworkInterface()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/network-interfaces/nic-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.NetworkInterfaceResponse{NetworkInterface: ni}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceNetworkInterface()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":        "",
		"subnet_id":   "",
		"ip_address":  "",
		"description": "",
	})
	d.SetId("nic-001")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if v := d.Get("name").(string); v != "test-nic" {
		t.Errorf("expected name test-nic, got %s", v)
	}
	if v := d.Get("network_id").(string); v != "net-001" {
		t.Errorf("expected network_id net-001, got %s", v)
	}
	if v := d.Get("subnet_id").(string); v != "subnet-001" {
		t.Errorf("expected subnet_id subnet-001, got %s", v)
	}
	if v := d.Get("ip_address").(string); v != "10.0.0.5" {
		t.Errorf("expected ip_address 10.0.0.5, got %s", v)
	}
	if v := d.Get("mac_address").(string); v != "fa:16:3e:aa:bb:cc" {
		t.Errorf("expected mac_address fa:16:3e:aa:bb:cc, got %s", v)
	}
	if v := d.Get("status").(string); v != "active" {
		t.Errorf("expected status active, got %s", v)
	}
	sgs := d.Get("security_groups").(*schema.Set)
	if sgs.Len() != 2 {
		t.Fatalf("expected 2 security groups, got %d", sgs.Len())
	}
	if !sgs.Contains("sg-001") || !sgs.Contains("sg-002") {
		t.Errorf("expected security groups sg-001 and sg-002, got %#v", sgs.List())
	}
	if v := d.Get("port_security_enabled").(bool); !v {
		t.Error("expected port_security_enabled true, got false")
	}
	if v := d.Get("network_type").(string); v != "vxlan" {
		t.Errorf("expected network_type vxlan, got %s", v)
	}
	if v := d.Get("description").(string); v != "a test network interface" {
		t.Errorf("expected description 'a test network interface', got %s", v)
	}
	if v := d.Get("created_at").(string); v != "2025-01-15T10:00:00Z" {
		t.Errorf("expected created_at 2025-01-15T10:00:00Z, got %s", v)
	}
}

func TestResourceNetworkInterfaceCreate_WithPostCreateOptions(t *testing.T) {
	ni := testNetworkInterface()
	var gotPairs dto.UpdateNetworkInterfaceAllowedAddressPairsRequest
	var gotSecurityGroups dto.UpdateNetworkInterfaceSecurityGroupsRequest

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "POST",
			Pattern: "/v2/iac/projects/test-project-id/network-interfaces",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.NetworkInterfaceResponse{NetworkInterface: ni}),
		},
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/network-interfaces/nic-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.NetworkInterfaceResponse{NetworkInterface: ni}),
		},
		{
			Method:  "PUT",
			Pattern: "/v2/iac/projects/test-project-id/network-interfaces/nic-001/allowed-address-pairs",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				if err := json.NewDecoder(r.Body).Decode(&gotPairs); err != nil {
					t.Fatalf("failed to decode allowed address pairs request: %v", err)
				}
				w.WriteHeader(http.StatusOK)
			},
		},
		{
			Method:  "PUT",
			Pattern: "/v2/iac/projects/test-project-id/network-interfaces/nic-001/security-groups",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				if err := json.NewDecoder(r.Body).Decode(&gotSecurityGroups); err != nil {
					t.Fatalf("failed to decode security groups request: %v", err)
				}
				w.WriteHeader(http.StatusOK)
			},
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceNetworkInterface()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":      "test-nic",
		"subnet_id": "subnet-001",
		"allowed_address_pairs": []interface{}{
			map[string]interface{}{
				"ip_address": "10.0.0.100",
			},
		},
		"security_groups": []interface{}{"sg-003", "sg-004"},
	})

	diags := res.CreateContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if len(gotPairs.AllowedAddressPairs) != 1 {
		t.Fatalf("expected 1 allowed address pair, got %d", len(gotPairs.AllowedAddressPairs))
	}
	if gotPairs.AllowedAddressPairs[0].IPAddress != "10.0.0.100" {
		t.Errorf("expected allowed address pair IP 10.0.0.100, got %s", gotPairs.AllowedAddressPairs[0].IPAddress)
	}
	if gotPairs.AllowedAddressPairs[0].MACAddress != ni.MACAddress {
		t.Errorf("expected default MAC %s, got %s", ni.MACAddress, gotPairs.AllowedAddressPairs[0].MACAddress)
	}

	if len(gotSecurityGroups.SecurityGroupIDs) != 2 {
		t.Fatalf("expected 2 security groups, got %d", len(gotSecurityGroups.SecurityGroupIDs))
	}
}

func TestResourceNetworkInterfaceCreate_APIError(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "POST",
			Pattern: "/v2/iac/projects/test-project-id/network-interfaces",
			Handler: testhelpers.EmptyHandler(http.StatusBadRequest),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceNetworkInterface()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":      "bad-nic",
		"subnet_id": "subnet-001",
	})

	diags := res.CreateContext(context.Background(), d, cfg)
	if !diags.HasError() {
		t.Fatal("expected error for create API failure, got none")
	}
}

func TestResourceNetworkInterfaceRead_NotFound(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/network-interfaces/nic-gone",
			Handler: testhelpers.EmptyHandler(http.StatusNotFound),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceNetworkInterface()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":        "",
		"subnet_id":   "",
		"ip_address":  "",
		"description": "",
	})
	d.SetId("nic-gone")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error on 404: %v", diags)
	}

	if d.Id() != "" {
		t.Errorf("expected resource ID to be cleared after 404, got %s", d.Id())
	}
}

func TestResourceNetworkInterfaceDelete(t *testing.T) {
	ni := testNetworkInterface()
	deletedCalled := false

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Pattern: "/v2/iac/projects/test-project-id/network-interfaces/nic-001",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case "GET":
					if deletedCalled {
						w.WriteHeader(http.StatusNotFound)
						return
					}
					testhelpers.JSONHandler(t, http.StatusOK, dto.NetworkInterfaceResponse{NetworkInterface: ni})(w, r)
				case "DELETE":
					deletedCalled = true
					w.WriteHeader(http.StatusAccepted)
				default:
					http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				}
			},
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceNetworkInterface()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":        "test-nic",
		"subnet_id":   "subnet-001",
		"ip_address":  "",
		"description": "",
	})
	d.SetId("nic-001")

	diags := res.DeleteContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if !deletedCalled {
		t.Error("expected DELETE to have been called")
	}
}
