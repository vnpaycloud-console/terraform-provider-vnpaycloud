package privategateway

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// testPrivateGateway returns a fully populated dto.PrivateGateway for use in tests.
func testPrivateGateway() dto.PrivateGateway {
	return dto.PrivateGateway{
		ID:             "pgw-001",
		Name:           "test-private-gateway",
		Description:    "a test private gateway",
		LoadBalancerID: "lb-001",
		SubnetID:       "subnet-001",
		FlavorID:       "flavor-001",
		Status:         "active",
		CreatedAt:      "2025-01-15T10:00:00Z",
		ProjectID:      testhelpers.TestProjectID,
		ZoneID:         testhelpers.TestZoneID,
	}
}

func TestResourcePrivateGatewayCreate(t *testing.T) {
	pgw := testPrivateGateway()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "POST",
			Pattern: "/v2/iac/projects/test-project-id/private-gateways",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.PrivateGatewayResponse{PrivateGateway: pgw}),
		},
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/private-gateways/pgw-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.PrivateGatewayResponse{PrivateGateway: pgw}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourcePrivateGateway()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":        "test-private-gateway",
		"description": "a test private gateway",
	})

	diags := res.CreateContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "pgw-001" {
		t.Errorf("expected ID pgw-001, got %s", d.Id())
	}
	if v := d.Get("name").(string); v != "test-private-gateway" {
		t.Errorf("expected name test-private-gateway, got %s", v)
	}
	if v := d.Get("description").(string); v != "a test private gateway" {
		t.Errorf("expected description 'a test private gateway', got %s", v)
	}
	if v := d.Get("load_balancer_id").(string); v != "lb-001" {
		t.Errorf("expected load_balancer_id lb-001, got %s", v)
	}
	if v := d.Get("subnet_id").(string); v != "subnet-001" {
		t.Errorf("expected subnet_id subnet-001, got %s", v)
	}
	if v := d.Get("flavor_id").(string); v != "flavor-001" {
		t.Errorf("expected flavor_id flavor-001, got %s", v)
	}
	if v := d.Get("status").(string); v != "active" {
		t.Errorf("expected status active, got %s", v)
	}
	if v := d.Get("created_at").(string); v != "2025-01-15T10:00:00Z" {
		t.Errorf("expected created_at 2025-01-15T10:00:00Z, got %s", v)
	}
}

func TestResourcePrivateGatewayRead(t *testing.T) {
	pgw := testPrivateGateway()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/private-gateways/pgw-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.PrivateGatewayResponse{PrivateGateway: pgw}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourcePrivateGateway()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":        "",
		"description": "",
	})
	d.SetId("pgw-001")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if v := d.Get("name").(string); v != "test-private-gateway" {
		t.Errorf("expected name test-private-gateway, got %s", v)
	}
	if v := d.Get("description").(string); v != "a test private gateway" {
		t.Errorf("expected description 'a test private gateway', got %s", v)
	}
	if v := d.Get("load_balancer_id").(string); v != "lb-001" {
		t.Errorf("expected load_balancer_id lb-001, got %s", v)
	}
	if v := d.Get("subnet_id").(string); v != "subnet-001" {
		t.Errorf("expected subnet_id subnet-001, got %s", v)
	}
	if v := d.Get("flavor_id").(string); v != "flavor-001" {
		t.Errorf("expected flavor_id flavor-001, got %s", v)
	}
	if v := d.Get("status").(string); v != "active" {
		t.Errorf("expected status active, got %s", v)
	}
	if v := d.Get("created_at").(string); v != "2025-01-15T10:00:00Z" {
		t.Errorf("expected created_at 2025-01-15T10:00:00Z, got %s", v)
	}
}

func TestResourcePrivateGatewayRead_NotFound(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/private-gateways/pgw-gone",
			Handler: testhelpers.EmptyHandler(http.StatusNotFound),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourcePrivateGateway()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":        "",
		"description": "",
	})
	d.SetId("pgw-gone")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error on 404: %v", diags)
	}

	if d.Id() != "" {
		t.Errorf("expected resource ID to be cleared after 404, got %s", d.Id())
	}
}

func TestResourcePrivateGatewayDelete(t *testing.T) {
	pgw := testPrivateGateway()
	deletedCalled := false

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Pattern: "/v2/iac/projects/test-project-id/private-gateways/pgw-001",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case "GET":
					if deletedCalled {
						w.WriteHeader(http.StatusNotFound)
						return
					}
					testhelpers.JSONHandler(t, http.StatusOK, dto.PrivateGatewayResponse{PrivateGateway: pgw})(w, r)
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

	res := ResourcePrivateGateway()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":        "test-private-gateway",
		"description": "",
	})
	d.SetId("pgw-001")

	diags := res.DeleteContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if !deletedCalled {
		t.Error("expected DELETE to have been called")
	}
}
