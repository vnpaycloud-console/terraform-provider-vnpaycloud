package internetgateway

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestDataSourceInternetGatewayRead_ByID(t *testing.T) {
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

	ds := DataSourceInternetGateway()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"id": "igw-001",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "igw-001" {
		t.Errorf("expected ID igw-001, got %s", d.Id())
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

func TestDataSourceInternetGatewaysRead(t *testing.T) {
	igw1 := testInternetGateway()
	igw2 := dto.InternetGateway{
		ID:          "igw-002",
		Name:        "test-igw-2",
		Description: "second internet gateway",
		VPCID:       "vpc-002",
		Status:      "active",
		CreatedAt:   "2025-01-16T12:00:00Z",
		ProjectID:   testhelpers.TestProjectID,
		ZoneID:      testhelpers.TestZoneID,
	}

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/internet-gateways",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListInternetGatewaysResponse{
				InternetGateways: []dto.InternetGateway{igw1, igw2},
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceInternetGateways()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	internetGateways := d.Get("internet_gateways").([]interface{})
	if len(internetGateways) != 2 {
		t.Fatalf("expected 2 internet gateways, got %d", len(internetGateways))
	}

	first := internetGateways[0].(map[string]interface{})
	if first["id"] != "igw-001" {
		t.Errorf("expected first internet gateway id igw-001, got %v", first["id"])
	}
	if first["name"] != "test-igw" {
		t.Errorf("expected first internet gateway name test-igw, got %v", first["name"])
	}

	second := internetGateways[1].(map[string]interface{})
	if second["id"] != "igw-002" {
		t.Errorf("expected second internet gateway id igw-002, got %v", second["id"])
	}
	if second["name"] != "test-igw-2" {
		t.Errorf("expected second internet gateway name test-igw-2, got %v", second["name"])
	}
}
