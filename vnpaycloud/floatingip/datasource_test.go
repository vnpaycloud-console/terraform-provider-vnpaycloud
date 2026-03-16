package floatingip

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestDataSourceFloatingIPRead_ByID(t *testing.T) {
	fip := testFloatingIP()
	fip.PortID = "port-001"
	fip.InstanceID = "inst-001"
	fip.InstanceName = "my-instance"

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/floating-ips/fip-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.FloatingIPResponse{FloatingIP: fip}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceFloatingIP()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"id": "fip-001",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "fip-001" {
		t.Errorf("expected ID fip-001, got %s", d.Id())
	}
	if v := d.Get("address").(string); v != "203.0.113.10" {
		t.Errorf("expected address 203.0.113.10, got %s", v)
	}
	if v := d.Get("status").(string); v != "active" {
		t.Errorf("expected status active, got %s", v)
	}
	if v := d.Get("port_id").(string); v != "port-001" {
		t.Errorf("expected port_id port-001, got %s", v)
	}
	if v := d.Get("instance_id").(string); v != "inst-001" {
		t.Errorf("expected instance_id inst-001, got %s", v)
	}
	if v := d.Get("instance_name").(string); v != "my-instance" {
		t.Errorf("expected instance_name my-instance, got %s", v)
	}
	if v := d.Get("created_at").(string); v != "2025-01-15T10:00:00Z" {
		t.Errorf("expected created_at 2025-01-15T10:00:00Z, got %s", v)
	}
}

func TestDataSourceFloatingIPsRead(t *testing.T) {
	fip1 := testFloatingIP()
	fip2 := dto.FloatingIP{
		ID:           "fip-002",
		Address:      "203.0.113.20",
		Status:       "active",
		PortID:       "port-002",
		InstanceID:   "inst-002",
		InstanceName: "second-instance",
		CreatedAt:    "2025-01-16T12:00:00Z",
		ProjectID:    testhelpers.TestProjectID,
		ZoneID:       testhelpers.TestZoneID,
	}

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/floating-ips",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListFloatingIPsResponse{
				FloatingIPs: []dto.FloatingIP{fip1, fip2},
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceFloatingIPs()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	floatingIPs := d.Get("floating_ips").([]interface{})
	if len(floatingIPs) != 2 {
		t.Fatalf("expected 2 floating IPs, got %d", len(floatingIPs))
	}

	first := floatingIPs[0].(map[string]interface{})
	if first["id"] != "fip-001" {
		t.Errorf("expected first floating IP id fip-001, got %v", first["id"])
	}
	if first["address"] != "203.0.113.10" {
		t.Errorf("expected first floating IP address 203.0.113.10, got %v", first["address"])
	}

	second := floatingIPs[1].(map[string]interface{})
	if second["id"] != "fip-002" {
		t.Errorf("expected second floating IP id fip-002, got %v", second["id"])
	}
	if second["address"] != "203.0.113.20" {
		t.Errorf("expected second floating IP address 203.0.113.20, got %v", second["address"])
	}
}
