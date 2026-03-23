package floatingip

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// testFloatingIP returns a fully populated dto.FloatingIP for use in tests.
func testFloatingIP() dto.FloatingIP {
	return dto.FloatingIP{
		ID:           "fip-001",
		Address:      "203.0.113.10",
		Status:       "active",
		PortID:       "",
		VpcID:        "",
		InstanceID:   "",
		InstanceName: "",
		CreatedAt:    "2025-01-15T10:00:00Z",
		ProjectID:    testhelpers.TestProjectID,
		ZoneID:       testhelpers.TestZoneID,
	}
}

func TestResourceFloatingIPCreate(t *testing.T) {
	fip := testFloatingIP()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "POST",
			Pattern: "/v2/iac/projects/test-project-id/floating-ips",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.FloatingIPResponse{FloatingIP: fip}),
		},
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/floating-ips/fip-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.FloatingIPResponse{FloatingIP: fip}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceFloatingIP()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"port_id": "",
		"vpc_id":  "",
	})

	diags := res.CreateContext(context.Background(), d, cfg)
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
}

func TestResourceFloatingIPRead(t *testing.T) {
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

	res := ResourceFloatingIP()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"port_id": "",
		"vpc_id":  "",
	})
	d.SetId("fip-001")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
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

func TestResourceFloatingIPRead_NotFound(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/floating-ips/fip-gone",
			Handler: testhelpers.EmptyHandler(http.StatusNotFound),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceFloatingIP()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"port_id": "",
		"vpc_id":  "",
	})
	d.SetId("fip-gone")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error on 404: %v", diags)
	}

	if d.Id() != "" {
		t.Errorf("expected resource ID to be cleared after 404, got %s", d.Id())
	}
}

func TestResourceFloatingIPDelete(t *testing.T) {
	fip := testFloatingIP()
	deletedCalled := false

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Pattern: "/v2/iac/projects/test-project-id/floating-ips/fip-001",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case "GET":
					if deletedCalled {
						w.WriteHeader(http.StatusNotFound)
						return
					}
					testhelpers.JSONHandler(t, http.StatusOK, dto.FloatingIPResponse{FloatingIP: fip})(w, r)
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

	res := ResourceFloatingIP()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"port_id": "",
		"vpc_id":  "",
	})
	d.SetId("fip-001")

	diags := res.DeleteContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if !deletedCalled {
		t.Error("expected DELETE to have been called")
	}
}
