package customergateway

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func testCustomerGateway() dto.CustomerGateway {
	return dto.CustomerGateway{
		ID:             "cgw-001",
		Name:           "tf-cgw",
		Description:    "terraform customer gateway",
		PublicIP:       "203.0.113.10",
		VPNType:        "ROUTE_BASED",
		Status:         "active",
		RemotePrefixes: []string{"10.20.0.0/24"},
		RemoteTunnelIP: "169.254.0.2/30",
		LocalTunnelIP:  "169.254.0.1/30",
		RoutingMode:    "DYNAMIC",
		BGPConfig: &dto.BGPConfig{
			LocalAs: 65534,
			PeerAs:  65000,
			AsPath:  "65000",
		},
		CreatedAt: "2025-01-15T10:00:00Z",
		ProjectID: testhelpers.TestProjectID,
		ZoneID:    testhelpers.TestZoneID,
	}
}

func TestResourceCustomerGatewaySchema(t *testing.T) {
	res := ResourceCustomerGateway()

	if !res.Schema["vpn_type"].ForceNew {
		t.Error("expected vpn_type to be ForceNew")
	}

	bgpConfig := res.Schema["bgp_config"]
	if !bgpConfig.Optional {
		t.Error("expected bgp_config to be optional")
	}
	if bgpConfig.MaxItems != 1 {
		t.Errorf("expected bgp_config MaxItems 1, got %d", bgpConfig.MaxItems)
	}

	bgpSchema := bgpConfig.Elem.(*schema.Resource).Schema
	for _, key := range []string{"local_as", "peer_as", "as_path"} {
		if !bgpSchema[key].Required {
			t.Errorf("expected bgp_config.%s to be required", key)
		}
	}
}

func TestResourceCustomerGatewayRead(t *testing.T) {
	cg := testCustomerGateway()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.CustomerGatewayWithID(testhelpers.TestProjectID, "cgw-001"),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.CustomerGatewayResponse{CustomerGateway: cg}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceCustomerGateway()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{})
	d.SetId("cgw-001")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	assertCustomerGatewayData(t, d)
}

func TestResourceCustomerGatewayRead_ClearsBGPConfigWhenMissing(t *testing.T) {
	cg := testCustomerGateway()
	cg.BGPConfig = nil
	cg.RoutingMode = "STATIC"

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.CustomerGatewayWithID(testhelpers.TestProjectID, "cgw-001"),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.CustomerGatewayResponse{CustomerGateway: cg}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceCustomerGateway()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"bgp_config": []interface{}{
			map[string]interface{}{
				"local_as": 65534,
				"peer_as":  65000,
				"as_path":  "65000",
			},
		},
	})
	d.SetId("cgw-001")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	bgpConfig := d.Get("bgp_config").([]interface{})
	if len(bgpConfig) != 0 {
		t.Errorf("expected bgp_config to be cleared, got %v", bgpConfig)
	}
}

func TestResourceCustomerGatewayUpdate(t *testing.T) {
	cg := testCustomerGateway()
	var gotReq dto.UpdateCustomerGatewayRequest
	putCalled := false

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Pattern: client.ApiPath.CustomerGatewayWithID(testhelpers.TestProjectID, "cgw-001"),
			Handler: func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case "PUT":
					putCalled = true
					if err := json.NewDecoder(r.Body).Decode(&gotReq); err != nil {
						t.Errorf("failed to decode update request: %v", err)
					}
					w.WriteHeader(http.StatusOK)
				case "GET":
					testhelpers.JSONHandler(t, http.StatusOK, dto.CustomerGatewayResponse{CustomerGateway: cg})(w, r)
				default:
					http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				}
			},
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceCustomerGateway()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":             "tf-cgw",
		"description":      "terraform customer gateway",
		"public_ip":        "203.0.113.10",
		"vpn_type":         "ROUTE_BASED",
		"remote_prefixes":  []interface{}{"10.20.0.0/24"},
		"remote_tunnel_ip": "169.254.0.2/30",
		"local_tunnel_ip":  "169.254.0.1/30",
		"routing_mode":     "DYNAMIC",
		"bgp_config": []interface{}{
			map[string]interface{}{
				"local_as": 65534,
				"peer_as":  65000,
				"as_path":  "65000",
			},
		},
		"zone_id": testhelpers.TestZoneID,
	})
	d.SetId("cgw-001")

	diags := res.UpdateContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if !putCalled {
		t.Fatal("expected PUT to have been called")
	}
	if gotReq.Name != "tf-cgw" {
		t.Errorf("expected request name tf-cgw, got %s", gotReq.Name)
	}
	if gotReq.PublicIP != "203.0.113.10" {
		t.Errorf("expected request public_ip 203.0.113.10, got %s", gotReq.PublicIP)
	}
	if len(gotReq.RemotePrefixes) != 1 || gotReq.RemotePrefixes[0] != "10.20.0.0/24" {
		t.Errorf("expected remote prefixes [10.20.0.0/24], got %v", gotReq.RemotePrefixes)
	}
	if gotReq.BGPConfig == nil {
		t.Fatal("expected BGPConfig to be set")
	}
	if gotReq.BGPConfig.LocalAs != 65534 {
		t.Errorf("expected local_as 65534, got %d", gotReq.BGPConfig.LocalAs)
	}
	if gotReq.BGPConfig.PeerAs != 65000 {
		t.Errorf("expected peer_as 65000, got %d", gotReq.BGPConfig.PeerAs)
	}
	if gotReq.BGPConfig.AsPath != "65000" {
		t.Errorf("expected as_path 65000, got %s", gotReq.BGPConfig.AsPath)
	}
}

func TestExpandBGPConfig(t *testing.T) {
	res := ResourceCustomerGateway()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"bgp_config": []interface{}{
			map[string]interface{}{
				"local_as": 65534,
				"peer_as":  65000,
				"as_path":  "65000",
			},
		},
	})

	bgpConfig := expandBGPConfig(d)
	if bgpConfig == nil {
		t.Fatal("expected BGPConfig")
	}
	if bgpConfig.LocalAs != 65534 {
		t.Errorf("expected local_as 65534, got %d", bgpConfig.LocalAs)
	}
	if bgpConfig.PeerAs != 65000 {
		t.Errorf("expected peer_as 65000, got %d", bgpConfig.PeerAs)
	}
	if bgpConfig.AsPath != "65000" {
		t.Errorf("expected as_path 65000, got %s", bgpConfig.AsPath)
	}
}

func TestExpandBGPConfig_Empty(t *testing.T) {
	res := ResourceCustomerGateway()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{})

	if bgpConfig := expandBGPConfig(d); bgpConfig != nil {
		t.Errorf("expected nil BGPConfig, got %+v", bgpConfig)
	}
}

func assertCustomerGatewayData(t *testing.T, d *schema.ResourceData) {
	t.Helper()

	if d.Id() != "cgw-001" {
		t.Errorf("expected ID cgw-001, got %s", d.Id())
	}
	if v := d.Get("name").(string); v != "tf-cgw" {
		t.Errorf("expected name tf-cgw, got %s", v)
	}
	if v := d.Get("description").(string); v != "terraform customer gateway" {
		t.Errorf("expected description terraform customer gateway, got %s", v)
	}
	if v := d.Get("public_ip").(string); v != "203.0.113.10" {
		t.Errorf("expected public_ip 203.0.113.10, got %s", v)
	}
	if v := d.Get("vpn_type").(string); v != "ROUTE_BASED" {
		t.Errorf("expected vpn_type ROUTE_BASED, got %s", v)
	}
	if v := d.Get("status").(string); v != "active" {
		t.Errorf("expected status active, got %s", v)
	}
	if v := d.Get("routing_mode").(string); v != "DYNAMIC" {
		t.Errorf("expected routing_mode DYNAMIC, got %s", v)
	}
	if v := d.Get("created_at").(string); v != "2025-01-15T10:00:00Z" {
		t.Errorf("expected created_at 2025-01-15T10:00:00Z, got %s", v)
	}

	remotePrefixes := d.Get("remote_prefixes").(*schema.Set).List()
	if len(remotePrefixes) != 1 || remotePrefixes[0].(string) != "10.20.0.0/24" {
		t.Errorf("expected remote_prefixes [10.20.0.0/24], got %v", remotePrefixes)
	}

	bgpConfig := d.Get("bgp_config").([]interface{})
	if len(bgpConfig) != 1 {
		t.Fatalf("expected bgp_config length 1, got %d", len(bgpConfig))
	}

	bgp := bgpConfig[0].(map[string]interface{})
	if bgp["local_as"] != 65534 {
		t.Errorf("expected local_as 65534, got %v", bgp["local_as"])
	}
	if bgp["peer_as"] != 65000 {
		t.Errorf("expected peer_as 65000, got %v", bgp["peer_as"])
	}
	if bgp["as_path"] != "65000" {
		t.Errorf("expected as_path 65000, got %v", bgp["as_path"])
	}
}
