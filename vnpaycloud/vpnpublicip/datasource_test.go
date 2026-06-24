package vpnpublicip

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestDataSourceVPNPublicIPRead_ByID(t *testing.T) {
	vpnPublicIP := testVPNPublicIP()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/vpn-public-ips/vpn-public-ip-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.VPNPublicIPResponse{VPNPublicIP: vpnPublicIP}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceVPNPublicIP()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"id": "vpn-public-ip-001",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	assertVPNPublicIPResourceData(t, d, vpnPublicIP)
}

func TestDataSourceVPNPublicIPRead_ByName(t *testing.T) {
	vpnPublicIP := testVPNPublicIP()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/vpn-public-ips",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListVPNPublicIPsResponse{
				VPNPublicIPs: []dto.VPNPublicIP{vpnPublicIP},
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceVPNPublicIP()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"name": "test-vpn-public-ip",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	assertVPNPublicIPResourceData(t, d, vpnPublicIP)
}

func TestDataSourceVPNPublicIPRead_NotFound(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/vpn-public-ips",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListVPNPublicIPsResponse{
				VPNPublicIPs: []dto.VPNPublicIP{},
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceVPNPublicIP()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"name": "missing-vpn-public-ip",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if !diags.HasError() {
		t.Fatal("expected error for nonexistent vpn public ip, got none")
	}
}

func TestDataSourceVPNPublicIPsRead(t *testing.T) {
	vpnPublicIP1 := testVPNPublicIP()
	vpnPublicIP2 := testVPNPublicIP()
	vpnPublicIP2.ID = "vpn-public-ip-002"
	vpnPublicIP2.Name = "test-vpn-public-ip-2"
	vpnPublicIP2.FloatingIP = "203.0.113.11"

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/vpn-public-ips",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListVPNPublicIPsResponse{
				VPNPublicIPs: []dto.VPNPublicIP{vpnPublicIP1, vpnPublicIP2},
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceVPNPublicIPs()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	vpnPublicIPs := d.Get("vpn_public_ips").([]interface{})
	if len(vpnPublicIPs) != 2 {
		t.Fatalf("expected 2 vpn public ips, got %d", len(vpnPublicIPs))
	}

	first := vpnPublicIPs[0].(map[string]interface{})
	if first["id"] != "vpn-public-ip-001" {
		t.Errorf("expected first vpn public ip id vpn-public-ip-001, got %v", first["id"])
	}
	if first["address"] != "203.0.113.10" {
		t.Errorf("expected first vpn public ip address 203.0.113.10, got %v", first["address"])
	}

	second := vpnPublicIPs[1].(map[string]interface{})
	if second["id"] != "vpn-public-ip-002" {
		t.Errorf("expected second vpn public ip id vpn-public-ip-002, got %v", second["id"])
	}
	if second["address"] != "203.0.113.11" {
		t.Errorf("expected second vpn public ip address 203.0.113.11, got %v", second["address"])
	}
}
