package internetgateway

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// testInternetGateway returns a fully populated dto.InternetGateway for use in tests.
func testInternetGateway() dto.InternetGateway {
	return dto.InternetGateway{
		ID:          "igw-001",
		Name:        "test-igw",
		Description: "a test internet gateway",
		VPCID:       "",
		Status:      "active",
		CreatedAt:   "2025-01-15T10:00:00Z",
		ProjectID:   testhelpers.TestProjectID,
		ZoneID:      testhelpers.TestZoneID,
	}
}

func TestResourceInternetGatewayCreate(t *testing.T) {
	igw := testInternetGateway()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "POST",
			Pattern: "/v2/iac/projects/test-project-id/internet-gateways",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.InternetGatewayResponse{InternetGateway: igw}),
		},
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/internet-gateways/igw-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.InternetGatewayResponse{InternetGateway: igw}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceInternetGateway()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":        "test-igw",
		"description": "a test internet gateway",
		"vpc_id":      "",
	})

	diags := res.CreateContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "igw-001" {
		t.Errorf("expected ID igw-001, got %s", d.Id())
	}
	if v := d.Get("name").(string); v != "test-igw" {
		t.Errorf("expected name test-igw, got %s", v)
	}
	if v := d.Get("status").(string); v != "active" {
		t.Errorf("expected status active, got %s", v)
	}
}

func TestResourceInternetGatewayRead(t *testing.T) {
	igw := testInternetGateway()
	igw.VPCID = "vpc-001"

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/internet-gateways/igw-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.InternetGatewayResponse{InternetGateway: igw}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceInternetGateway()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":        "",
		"description": "",
		"vpc_id":      "",
	})
	d.SetId("igw-001")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if v := d.Get("name").(string); v != "test-igw" {
		t.Errorf("expected name test-igw, got %s", v)
	}
	if v := d.Get("description").(string); v != "a test internet gateway" {
		t.Errorf("expected description 'a test internet gateway', got %s", v)
	}
	if v := d.Get("vpc_id").(string); v != "vpc-001" {
		t.Errorf("expected vpc_id vpc-001, got %s", v)
	}
	if v := d.Get("status").(string); v != "active" {
		t.Errorf("expected status active, got %s", v)
	}
	if v := d.Get("created_at").(string); v != "2025-01-15T10:00:00Z" {
		t.Errorf("expected created_at 2025-01-15T10:00:00Z, got %s", v)
	}
	if v := d.Get("zone_id").(string); v != testhelpers.TestZoneID {
		t.Errorf("expected zone_id %s, got %s", testhelpers.TestZoneID, v)
	}
}

func TestResourceInternetGatewayRead_NotFound(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/internet-gateways/igw-gone",
			Handler: testhelpers.EmptyHandler(http.StatusNotFound),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceInternetGateway()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":        "",
		"description": "",
		"vpc_id":      "",
	})
	d.SetId("igw-gone")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error on 404: %v", diags)
	}

	if d.Id() != "" {
		t.Errorf("expected resource ID to be cleared after 404, got %s", d.Id())
	}
}

func TestResourceInternetGatewayDelete(t *testing.T) {
	igw := testInternetGateway()
	deletedCalled := false

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Pattern: "/v2/iac/projects/test-project-id/internet-gateways/igw-001",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case "GET":
					if deletedCalled {
						w.WriteHeader(http.StatusNotFound)
						return
					}
					testhelpers.JSONHandler(t, http.StatusOK, dto.InternetGatewayResponse{InternetGateway: igw})(w, r)
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

	res := ResourceInternetGateway()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":        "test-igw",
		"description": "",
		"vpc_id":      "",
	})
	d.SetId("igw-001")

	diags := res.DeleteContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if !deletedCalled {
		t.Error("expected DELETE to have been called")
	}
}
