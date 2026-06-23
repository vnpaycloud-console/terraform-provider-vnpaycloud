package vpnpublicip

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func testVPNPublicIP() dto.VPNPublicIP {
	return dto.VPNPublicIP{
		ID:          "vpn-public-ip-001",
		Name:        "test-vpn-public-ip",
		Description: "a test vpn public ip",
		FloatingIP:  "203.0.113.10",
		Status:      "active",
		CreatedAt:   "2025-01-15T10:00:00Z",
		ProjectID:   testhelpers.TestProjectID,
		ZoneID:      testhelpers.TestZoneID,
	}
}

func TestResourceVPNPublicIPCreate(t *testing.T) {
	vpnPublicIP := testVPNPublicIP()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "POST",
			Pattern: "/v2/iac/projects/test-project-id/vpn-public-ips",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.VPNPublicIPResponse{VPNPublicIP: vpnPublicIP}),
		},
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/vpn-public-ips/vpn-public-ip-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.VPNPublicIPResponse{VPNPublicIP: vpnPublicIP}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceVPNPublicIP()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":        "test-vpn-public-ip",
		"description": "a test vpn public ip",
		"zone_id":     testhelpers.TestZoneID,
	})

	diags := res.CreateContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	assertVPNPublicIPResourceData(t, d, vpnPublicIP)
}

func TestResourceVPNPublicIPRead(t *testing.T) {
	vpnPublicIP := testVPNPublicIP()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/vpn-public-ips/vpn-public-ip-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.VPNPublicIPResponse{VPNPublicIP: vpnPublicIP}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceVPNPublicIP()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{})
	d.SetId("vpn-public-ip-001")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	assertVPNPublicIPResourceData(t, d, vpnPublicIP)
}

func TestResourceVPNPublicIPUpdate(t *testing.T) {
	vpnPublicIP := testVPNPublicIP()
	vpnPublicIP.Name = "updated-vpn-public-ip"
	vpnPublicIP.Description = "updated test vpn public ip"
	updateCalled := false

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Pattern: "/v2/iac/projects/test-project-id/vpn-public-ips/vpn-public-ip-001",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case "PUT":
					updateCalled = true
					testhelpers.JSONHandler(t, http.StatusOK, dto.VPNPublicIPResponse{VPNPublicIP: vpnPublicIP})(w, r)
				case "GET":
					testhelpers.JSONHandler(t, http.StatusOK, dto.VPNPublicIPResponse{VPNPublicIP: vpnPublicIP})(w, r)
				default:
					http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				}
			},
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceVPNPublicIP()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":        "updated-vpn-public-ip",
		"description": "updated test vpn public ip",
		"zone_id":     testhelpers.TestZoneID,
	})
	d.SetId("vpn-public-ip-001")

	diags := res.UpdateContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if !updateCalled {
		t.Error("expected PUT to have been called")
	}
	assertVPNPublicIPResourceData(t, d, vpnPublicIP)
}

func TestResourceVPNPublicIPRead_NotFound(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/vpn-public-ips/vpn-gone",
			Handler: testhelpers.EmptyHandler(http.StatusNotFound),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceVPNPublicIP()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{})
	d.SetId("vpn-gone")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error on 404: %v", diags)
	}

	if d.Id() != "" {
		t.Errorf("expected resource ID to be cleared after 404, got %s", d.Id())
	}
}

func TestResourceVPNPublicIPDelete(t *testing.T) {
	vpnPublicIP := testVPNPublicIP()
	deletedCalled := false

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Pattern: "/v2/iac/projects/test-project-id/vpn-public-ips/vpn-public-ip-001",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case "GET":
					if deletedCalled {
						w.WriteHeader(http.StatusNotFound)
						return
					}
					testhelpers.JSONHandler(t, http.StatusOK, dto.VPNPublicIPResponse{VPNPublicIP: vpnPublicIP})(w, r)
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

	res := ResourceVPNPublicIP()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":    "test-vpn-public-ip",
		"zone_id": testhelpers.TestZoneID,
	})
	d.SetId("vpn-public-ip-001")

	diags := res.DeleteContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if !deletedCalled {
		t.Error("expected DELETE to have been called")
	}
}

func assertVPNPublicIPResourceData(t *testing.T, d *schema.ResourceData, vpnPublicIP dto.VPNPublicIP) {
	t.Helper()

	if d.Id() != vpnPublicIP.ID {
		t.Errorf("expected ID %s, got %s", vpnPublicIP.ID, d.Id())
	}
	if v := d.Get("name").(string); v != vpnPublicIP.Name {
		t.Errorf("expected name %s, got %s", vpnPublicIP.Name, v)
	}
	if v := d.Get("description").(string); v != vpnPublicIP.Description {
		t.Errorf("expected description %s, got %s", vpnPublicIP.Description, v)
	}
	if v := d.Get("address").(string); v != vpnPublicIP.FloatingIP {
		t.Errorf("expected address %s, got %s", vpnPublicIP.FloatingIP, v)
	}
	if v := d.Get("status").(string); v != vpnPublicIP.Status {
		t.Errorf("expected status %s, got %s", vpnPublicIP.Status, v)
	}
	if v := d.Get("created_at").(string); v != vpnPublicIP.CreatedAt {
		t.Errorf("expected created_at %s, got %s", vpnPublicIP.CreatedAt, v)
	}
}
