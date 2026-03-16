package routetable

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// testRouteTable returns a fully populated dto.RouteTable for use in tests.
func testRouteTable() dto.RouteTable {
	return dto.RouteTable{
		ID:         "rt-001",
		VpcID:      "vpc-001",
		DestCIDR:   "10.1.0.0/16",
		TargetID:   "pgw-001",
		TargetType: "private_gateway",
		TargetName: "my-private-gateway",
		Name:       "test-route",
		Status:     "active",
		CreatedAt:  "2025-01-15T10:00:00Z",
		ProjectID:  testhelpers.TestProjectID,
		ZoneID:     testhelpers.TestZoneID,
	}
}

func TestResourceRouteTableCreate(t *testing.T) {
	rt := testRouteTable()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Pattern: "/v2/iac/projects/test-project-id/route-tables",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case "POST":
					testhelpers.JSONHandler(t, http.StatusOK, dto.RouteTableResponse{RouteTable: rt})(w, r)
				case "GET":
					testhelpers.JSONHandler(t, http.StatusOK, dto.ListRouteTablesResponse{
						RouteTables: []dto.RouteTable{rt},
					})(w, r)
				default:
					http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				}
			},
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceRouteTable()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"vpc_id":      "vpc-001",
		"dest_cidr":   "10.1.0.0/16",
		"target_id":   "pgw-001",
		"target_type": "private_gateway",
	})

	diags := res.CreateContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "rt-001" {
		t.Errorf("expected ID rt-001, got %s", d.Id())
	}
	if v := d.Get("vpc_id").(string); v != "vpc-001" {
		t.Errorf("expected vpc_id vpc-001, got %s", v)
	}
	if v := d.Get("dest_cidr").(string); v != "10.1.0.0/16" {
		t.Errorf("expected dest_cidr 10.1.0.0/16, got %s", v)
	}
	if v := d.Get("target_id").(string); v != "pgw-001" {
		t.Errorf("expected target_id pgw-001, got %s", v)
	}
	if v := d.Get("target_type").(string); v != "private_gateway" {
		t.Errorf("expected target_type private_gateway, got %s", v)
	}
	if v := d.Get("name").(string); v != "test-route" {
		t.Errorf("expected name test-route, got %s", v)
	}
	if v := d.Get("target_name").(string); v != "my-private-gateway" {
		t.Errorf("expected target_name my-private-gateway, got %s", v)
	}
	if v := d.Get("status").(string); v != "active" {
		t.Errorf("expected status active, got %s", v)
	}
	if v := d.Get("created_at").(string); v != "2025-01-15T10:00:00Z" {
		t.Errorf("expected created_at 2025-01-15T10:00:00Z, got %s", v)
	}
}

func TestResourceRouteTableRead(t *testing.T) {
	rt := testRouteTable()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			// Read uses list endpoint and searches by ID
			Method: "GET",
			Pattern: "/v2/iac/projects/test-project-id/route-tables",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListRouteTablesResponse{
				RouteTables: []dto.RouteTable{rt},
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceRouteTable()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"vpc_id":      "",
		"dest_cidr":   "",
		"target_id":   "",
		"target_type": "",
	})
	d.SetId("rt-001")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if v := d.Get("vpc_id").(string); v != "vpc-001" {
		t.Errorf("expected vpc_id vpc-001, got %s", v)
	}
	if v := d.Get("dest_cidr").(string); v != "10.1.0.0/16" {
		t.Errorf("expected dest_cidr 10.1.0.0/16, got %s", v)
	}
	if v := d.Get("target_id").(string); v != "pgw-001" {
		t.Errorf("expected target_id pgw-001, got %s", v)
	}
	if v := d.Get("target_type").(string); v != "private_gateway" {
		t.Errorf("expected target_type private_gateway, got %s", v)
	}
	if v := d.Get("name").(string); v != "test-route" {
		t.Errorf("expected name test-route, got %s", v)
	}
	if v := d.Get("target_name").(string); v != "my-private-gateway" {
		t.Errorf("expected target_name my-private-gateway, got %s", v)
	}
	if v := d.Get("status").(string); v != "active" {
		t.Errorf("expected status active, got %s", v)
	}
	if v := d.Get("created_at").(string); v != "2025-01-15T10:00:00Z" {
		t.Errorf("expected created_at 2025-01-15T10:00:00Z, got %s", v)
	}
}

func TestResourceRouteTableRead_NotFound(t *testing.T) {
	// Read uses list endpoint; route table not found when list returns no matching ID
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method: "GET",
			Pattern: "/v2/iac/projects/test-project-id/route-tables",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListRouteTablesResponse{
				RouteTables: []dto.RouteTable{},
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceRouteTable()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"vpc_id":      "",
		"dest_cidr":   "",
		"target_id":   "",
		"target_type": "",
	})
	d.SetId("rt-gone")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error on not found: %v", diags)
	}

	if d.Id() != "" {
		t.Errorf("expected resource ID to be cleared when not found in list, got %s", d.Id())
	}
}

func TestResourceRouteTableDelete(t *testing.T) {
	deletedCalled := false

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method: "DELETE",
			Pattern: "/v2/iac/projects/test-project-id/route-tables/rt-001",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				deletedCalled = true
				w.WriteHeader(http.StatusAccepted)
			},
		},
		{
			// State refresh uses list endpoint to check deletion status
			Method: "GET",
			Pattern: "/v2/iac/projects/test-project-id/route-tables",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListRouteTablesResponse{
				RouteTables: []dto.RouteTable{},
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceRouteTable()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"vpc_id":      "vpc-001",
		"dest_cidr":   "10.1.0.0/16",
		"target_id":   "pgw-001",
		"target_type": "private_gateway",
	})
	d.SetId("rt-001")

	diags := res.DeleteContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if !deletedCalled {
		t.Error("expected DELETE to have been called")
	}
}
