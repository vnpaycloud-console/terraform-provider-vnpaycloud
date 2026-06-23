package customergateway

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestDataSourceCustomerGatewayRead_ByID(t *testing.T) {
	cg := testCustomerGateway()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.CustomerGatewayWithID(testhelpers.TestProjectID, "cgw-001"),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.CustomerGatewayResponse{CustomerGateway: cg}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceCustomerGateway()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"id": "cgw-001",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	assertCustomerGatewayData(t, d)
}

func TestDataSourceCustomerGatewayRead_ByName(t *testing.T) {
	cg := testCustomerGateway()
	otherCG := testCustomerGateway()
	otherCG.ID = "cgw-002"
	otherCG.Name = "other-cgw"

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.CustomerGateways(testhelpers.TestProjectID),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListCustomerGatewaysResponse{
				CustomerGateways: []dto.CustomerGateway{otherCG, cg},
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceCustomerGateway()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"name": "tf-cgw",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "cgw-001" {
		t.Errorf("expected ID cgw-001, got %s", d.Id())
	}
}

func TestDataSourceCustomerGatewayRead_NotFoundByName(t *testing.T) {
	cg := testCustomerGateway()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.CustomerGateways(testhelpers.TestProjectID),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListCustomerGatewaysResponse{
				CustomerGateways: []dto.CustomerGateway{cg},
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceCustomerGateway()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"name": "missing-cgw",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if !diags.HasError() {
		t.Fatal("expected error for missing customer gateway")
	}
}

func TestDataSourceCustomerGatewaysRead(t *testing.T) {
	cg := testCustomerGateway()
	routeStaticCG := dto.CustomerGateway{
		ID:             "cgw-002",
		Name:           "route-static-cgw",
		Description:    "route based static customer gateway",
		PublicIP:       "203.0.113.11",
		VPNType:        "ROUTE_BASED",
		Status:         "active",
		RemotePrefixes: []string{"10.30.0.0/24"},
		RemoteTunnelIP: "169.254.1.2/30",
		LocalTunnelIP:  "169.254.1.1/30",
		RoutingMode:    "STATIC",
		CreatedAt:      "2025-01-16T10:00:00Z",
		ProjectID:      testhelpers.TestProjectID,
		ZoneID:         testhelpers.TestZoneID,
	}

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.CustomerGateways(testhelpers.TestProjectID),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListCustomerGatewaysResponse{
				CustomerGateways: []dto.CustomerGateway{cg, routeStaticCG},
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceCustomerGateways()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	expectedID := "customer-gateways-" + testhelpers.TestProjectID
	if d.Id() != expectedID {
		t.Errorf("expected ID %s, got %s", expectedID, d.Id())
	}

	customerGateways := d.Get("customer_gateways").([]interface{})
	if len(customerGateways) != 2 {
		t.Fatalf("expected 2 customer gateways, got %d", len(customerGateways))
	}

	first := customerGateways[0].(map[string]interface{})
	if first["id"] != "cgw-001" {
		t.Errorf("expected first customer gateway id cgw-001, got %v", first["id"])
	}

	bgpConfig := first["bgp_config"].([]interface{})
	if len(bgpConfig) != 1 {
		t.Fatalf("expected first customer gateway bgp_config length 1, got %d", len(bgpConfig))
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

	second := customerGateways[1].(map[string]interface{})
	secondBGPConfig := second["bgp_config"].([]interface{})
	if len(secondBGPConfig) != 0 {
		t.Errorf("expected empty bgp_config for static customer gateway, got %v", secondBGPConfig)
	}
}
