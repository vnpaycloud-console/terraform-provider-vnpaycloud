package vpngateway

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestDataSourceVPNGatewayRead_ByID(t *testing.T) {
	gw := testVPNGateway()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/vpn-gateways/vpn-gw-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.VPNGatewayResponse{VPNGateway: gw}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceVPNGateway()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"id": "vpn-gw-001",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	assertVPNGatewayResourceData(t, d, gw)
}

func TestDataSourceVPNGatewayRead_ByName(t *testing.T) {
	gw := testVPNGateway()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/vpn-gateways",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListVPNGatewaysResponse{
				VPNGateways: []dto.VPNGateway{gw},
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceVPNGateway()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"name": "test-vpn-gateway",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	assertVPNGatewayResourceData(t, d, gw)
}

func TestDataSourceVPNGatewayRead_NotFound(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/vpn-gateways",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListVPNGatewaysResponse{
				VPNGateways: []dto.VPNGateway{},
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceVPNGateway()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"name": "missing-vpn-gateway",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if !diags.HasError() {
		t.Fatal("expected error for nonexistent vpn gateway, got none")
	}
}

func TestDataSourceVPNGatewaysRead(t *testing.T) {
	gw1 := testVPNGateway()
	gw2 := testVPNGateway()
	gw2.ID = "vpn-gw-002"
	gw2.Name = "test-vpn-gateway-2"
	gw2.Description = "second test vpn gateway"
	gw2.VPNType = "POLICY_BASED"
	gw2.AttachedVPCIDs = []string{"vpc-002"}

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/vpn-gateways",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListVPNGatewaysResponse{
				VPNGateways: []dto.VPNGateway{gw1, gw2},
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceVPNGateways()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	vpnGateways := d.Get("vpn_gateways").([]interface{})
	if len(vpnGateways) != 2 {
		t.Fatalf("expected 2 vpn gateways, got %d", len(vpnGateways))
	}

	first := vpnGateways[0].(map[string]interface{})
	if first["id"] != "vpn-gw-001" {
		t.Errorf("expected first vpn gateway id vpn-gw-001, got %v", first["id"])
	}
	if first["name"] != "test-vpn-gateway" {
		t.Errorf("expected first vpn gateway name test-vpn-gateway, got %v", first["name"])
	}

	second := vpnGateways[1].(map[string]interface{})
	if second["id"] != "vpn-gw-002" {
		t.Errorf("expected second vpn gateway id vpn-gw-002, got %v", second["id"])
	}
	if second["vpn_type"] != "POLICY_BASED" {
		t.Errorf("expected second vpn gateway vpn_type POLICY_BASED, got %v", second["vpn_type"])
	}
}
