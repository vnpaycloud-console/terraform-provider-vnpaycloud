package subnet

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// testSubnet returns a fully populated dto.Subnet for use in tests.
func testSubnet() dto.Subnet {
	return dto.Subnet{
		ID:           "subnet-001",
		Name:         "test-subnet",
		VpcID:        "vpc-001",
		CIDR:         "10.0.1.0/24",
		GatewayIP:    "10.0.1.1",
		EnableDHCP:   true,
		EnableSnat:   false,
		ExternalIpID: "",
		UsedByK8S:    false,
		Status:       "active",
		CreatedAt:    "2025-01-15T10:00:00Z",
		ProjectID:    testhelpers.TestProjectID,
		ZoneID:       testhelpers.TestZoneID,
	}
}

func TestResourceSubnetCreate(t *testing.T) {
	sub := testSubnet()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "POST",
			Pattern: "/v2/iac/projects/test-project-id/subnets",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.SubnetResponse{Subnet: sub}),
		},
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/subnets/subnet-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.SubnetResponse{Subnet: sub}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceSubnet()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":        "test-subnet",
		"vpc_id":      "vpc-001",
		"cidr":        "10.0.1.0/24",
		"gateway_ip":  "10.0.1.1",
		"enable_dhcp": true,
		"used_by_k8s": false,
	})

	diags := res.CreateContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "subnet-001" {
		t.Errorf("expected ID subnet-001, got %s", d.Id())
	}
	if v := d.Get("name").(string); v != "test-subnet" {
		t.Errorf("expected name test-subnet, got %s", v)
	}
	if v := d.Get("vpc_id").(string); v != "vpc-001" {
		t.Errorf("expected vpc_id vpc-001, got %s", v)
	}
	if v := d.Get("cidr").(string); v != "10.0.1.0/24" {
		t.Errorf("expected cidr 10.0.1.0/24, got %s", v)
	}
	if v := d.Get("status").(string); v != "active" {
		t.Errorf("expected status active, got %s", v)
	}
	if v := d.Get("gateway_ip").(string); v != "10.0.1.1" {
		t.Errorf("expected gateway_ip 10.0.1.1, got %s", v)
	}
	if v := d.Get("created_at").(string); v != "2025-01-15T10:00:00Z" {
		t.Errorf("expected created_at 2025-01-15T10:00:00Z, got %s", v)
	}
}

func TestResourceSubnetRead(t *testing.T) {
	sub := testSubnet()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/subnets/subnet-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.SubnetResponse{Subnet: sub}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceSubnet()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":        "",
		"vpc_id":      "",
		"cidr":        "",
		"gateway_ip":  "",
		"enable_dhcp": true,
		"used_by_k8s": false,
	})
	d.SetId("subnet-001")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if v := d.Get("name").(string); v != "test-subnet" {
		t.Errorf("expected name test-subnet, got %s", v)
	}
	if v := d.Get("vpc_id").(string); v != "vpc-001" {
		t.Errorf("expected vpc_id vpc-001, got %s", v)
	}
	if v := d.Get("cidr").(string); v != "10.0.1.0/24" {
		t.Errorf("expected cidr 10.0.1.0/24, got %s", v)
	}
	if v := d.Get("gateway_ip").(string); v != "10.0.1.1" {
		t.Errorf("expected gateway_ip 10.0.1.1, got %s", v)
	}
	if v := d.Get("enable_dhcp").(bool); !v {
		t.Error("expected enable_dhcp true, got false")
	}
	if v := d.Get("used_by_k8s").(bool); v {
		t.Error("expected used_by_k8s false, got true")
	}
	if v := d.Get("status").(string); v != "active" {
		t.Errorf("expected status active, got %s", v)
	}
	if v := d.Get("created_at").(string); v != "2025-01-15T10:00:00Z" {
		t.Errorf("expected created_at 2025-01-15T10:00:00Z, got %s", v)
	}
}

func TestResourceSubnetRead_NotFound(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/subnets/subnet-gone",
			Handler: testhelpers.EmptyHandler(http.StatusNotFound),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceSubnet()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":        "",
		"vpc_id":      "",
		"cidr":        "",
		"gateway_ip":  "",
		"enable_dhcp": true,
		"used_by_k8s": false,
	})
	d.SetId("subnet-gone")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error on 404: %v", diags)
	}

	if d.Id() != "" {
		t.Errorf("expected resource ID to be cleared after 404, got %s", d.Id())
	}
}

func TestResourceSubnetDelete(t *testing.T) {
	sub := testSubnet()
	deletedCalled := false

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Pattern: "/v2/iac/projects/test-project-id/subnets/subnet-001",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case "GET":
					if deletedCalled {
						w.WriteHeader(http.StatusNotFound)
						return
					}
					testhelpers.JSONHandler(t, http.StatusOK, dto.SubnetResponse{Subnet: sub})(w, r)
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

	res := ResourceSubnet()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":        "test-subnet",
		"vpc_id":      "vpc-001",
		"cidr":        "10.0.1.0/24",
		"gateway_ip":  "",
		"enable_dhcp": true,
		"used_by_k8s": false,
	})
	d.SetId("subnet-001")

	diags := res.DeleteContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if !deletedCalled {
		t.Error("expected DELETE to have been called")
	}
}
