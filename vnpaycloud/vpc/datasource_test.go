package vpc

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestDataSourceVpcRead_ByID(t *testing.T) {
	vpc := testVPC()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/vpcs/vpc-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.VPCResponse{VPC: vpc}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceVpc()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"id": "vpc-001",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "vpc-001" {
		t.Errorf("expected ID vpc-001, got %s", d.Id())
	}
	if v := d.Get("name").(string); v != "test-vpc" {
		t.Errorf("expected name test-vpc, got %s", v)
	}
	if v := d.Get("description").(string); v != "a test vpc" {
		t.Errorf("expected description 'a test vpc', got %s", v)
	}
	if v := d.Get("cidr").(string); v != "10.0.0.0/16" {
		t.Errorf("expected cidr 10.0.0.0/16, got %s", v)
	}
	if v := d.Get("status").(string); v != "active" {
		t.Errorf("expected status active, got %s", v)
	}
	if v := d.Get("enable_snat").(bool); v {
		t.Error("expected enable_snat false, got true")
	}
	if v := d.Get("created_at").(string); v != "2025-01-15T10:00:00Z" {
		t.Errorf("expected created_at 2025-01-15T10:00:00Z, got %s", v)
	}
}

func TestDataSourceVpcRead_ByName(t *testing.T) {
	vpc := testVPC()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/vpcs",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListVPCsResponse{
				VPCs: []dto.VPC{vpc},
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceVpc()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"name": "test-vpc",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "vpc-001" {
		t.Errorf("expected ID vpc-001, got %s", d.Id())
	}
	if v := d.Get("name").(string); v != "test-vpc" {
		t.Errorf("expected name test-vpc, got %s", v)
	}
}

func TestDataSourceVpcRead_NotFound(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/vpcs",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListVPCsResponse{
				VPCs: []dto.VPC{},
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceVpc()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"name": "nonexistent",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if !diags.HasError() {
		t.Fatal("expected error for nonexistent vpc, got none")
	}
}

func TestDataSourceVpcsRead(t *testing.T) {
	vpc1 := testVPC()
	vpc2 := dto.VPC{
		ID:          "vpc-002",
		Name:        "test-vpc-2",
		Description: "second vpc",
		CIDR:        "10.1.0.0/16",
		Status:      "active",
		EnableSnat:  true,
		SnatAddress: "203.0.113.10",
		SubnetIDs:   []string{"subnet-002"},
		CreatedAt:   "2025-01-16T12:00:00Z",
		ProjectID:   testhelpers.TestProjectID,
		ZoneID:      testhelpers.TestZoneID,
	}

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/vpcs",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListVPCsResponse{
				VPCs: []dto.VPC{vpc1, vpc2},
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceVpcs()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	vpcs := d.Get("vpcs").([]interface{})
	if len(vpcs) != 2 {
		t.Fatalf("expected 2 vpcs, got %d", len(vpcs))
	}

	first := vpcs[0].(map[string]interface{})
	if first["id"] != "vpc-001" {
		t.Errorf("expected first vpc id vpc-001, got %v", first["id"])
	}
	if first["name"] != "test-vpc" {
		t.Errorf("expected first vpc name test-vpc, got %v", first["name"])
	}

	second := vpcs[1].(map[string]interface{})
	if second["id"] != "vpc-002" {
		t.Errorf("expected second vpc id vpc-002, got %v", second["id"])
	}
	if second["name"] != "test-vpc-2" {
		t.Errorf("expected second vpc name test-vpc-2, got %v", second["name"])
	}
}
