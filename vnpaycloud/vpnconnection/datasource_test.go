package vpnconnection

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestDataSourceVPNConnectionRead_ByID(t *testing.T) {
	vpnConnection := testVPNConnection()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/vpn-connections/vpn-conn-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.VPNConnectionResponse{VPNConnection: vpnConnection}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceVPNConnection()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"id": "vpn-conn-001",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	assertVPNConnectionResourceData(t, d, vpnConnection)
}

func TestDataSourceVPNConnectionRead_ByName(t *testing.T) {
	vpnConnection := testVPNConnection()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/vpn-connections",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListVPNConnectionsResponse{
				VPNConnections: []dto.VPNConnection{vpnConnection},
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceVPNConnection()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"name": "test-vpn-connection",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	assertVPNConnectionResourceData(t, d, vpnConnection)
}

func TestDataSourceVPNConnectionRead_NotFound(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/vpn-connections",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListVPNConnectionsResponse{
				VPNConnections: []dto.VPNConnection{},
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceVPNConnection()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"name": "missing-vpn-connection",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if !diags.HasError() {
		t.Fatal("expected error for nonexistent vpn connection, got none")
	}
}

func TestDataSourceVPNConnectionsRead(t *testing.T) {
	vpnConnection1 := testVPNConnection()
	vpnConnection2 := testVPNConnection()
	vpnConnection2.ID = "vpn-conn-002"
	vpnConnection2.Name = "test-vpn-connection-2"
	vpnConnection2.VPNType = "POLICY_BASED"

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/vpn-connections",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListVPNConnectionsResponse{
				VPNConnections: []dto.VPNConnection{vpnConnection1, vpnConnection2},
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceVPNConnections()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	vpnConnections := d.Get("vpn_connections").([]interface{})
	if len(vpnConnections) != 2 {
		t.Fatalf("expected 2 vpn connections, got %d", len(vpnConnections))
	}

	first := vpnConnections[0].(map[string]interface{})
	if first["id"] != "vpn-conn-001" {
		t.Errorf("expected first vpn connection id vpn-conn-001, got %v", first["id"])
	}
	if first["vpn_type"] != "ROUTE_BASED" {
		t.Errorf("expected first vpn connection vpn_type ROUTE_BASED, got %v", first["vpn_type"])
	}

	second := vpnConnections[1].(map[string]interface{})
	if second["id"] != "vpn-conn-002" {
		t.Errorf("expected second vpn connection id vpn-conn-002, got %v", second["id"])
	}
	if second["vpn_type"] != "POLICY_BASED" {
		t.Errorf("expected second vpn connection vpn_type POLICY_BASED, got %v", second["vpn_type"])
	}
}
