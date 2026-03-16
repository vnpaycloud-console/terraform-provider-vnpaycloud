package vpc

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// testVPC returns a fully populated dto.VPC for use in tests.
func testVPC() dto.VPC {
	return dto.VPC{
		ID:          "vpc-001",
		Name:        "test-vpc",
		Description: "a test vpc",
		CIDR:        "10.0.0.0/16",
		Status:      "active",
		EnableSnat:  false,
		SnatAddress: "",
		SubnetIDs:   []string{"subnet-001"},
		CreatedAt:   "2025-01-15T10:00:00Z",
		ProjectID:   testhelpers.TestProjectID,
		ZoneID:      testhelpers.TestZoneID,
	}
}

func TestResourceVpcCreate(t *testing.T) {
	vpc := testVPC()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "POST",
			Pattern: "/v2/iac/projects/test-project-id/vpcs",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.VPCResponse{VPC: vpc}),
		},
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/vpcs/vpc-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.VPCResponse{VPC: vpc}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceVpc()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":        "test-vpc",
		"description": "a test vpc",
		"cidr":        "10.0.0.0/16",
		"enable_snat": false,
	})

	diags := res.CreateContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "vpc-001" {
		t.Errorf("expected ID vpc-001, got %s", d.Id())
	}
	if v := d.Get("name").(string); v != "test-vpc" {
		t.Errorf("expected name test-vpc, got %s", v)
	}
	if v := d.Get("status").(string); v != "active" {
		t.Errorf("expected status active, got %s", v)
	}
	if v := d.Get("cidr").(string); v != "10.0.0.0/16" {
		t.Errorf("expected cidr 10.0.0.0/16, got %s", v)
	}
	if v := d.Get("created_at").(string); v != "2025-01-15T10:00:00Z" {
		t.Errorf("expected created_at 2025-01-15T10:00:00Z, got %s", v)
	}
}

func TestResourceVpcCreate_WithSNAT(t *testing.T) {
	vpc := testVPC()
	vpc.EnableSnat = true
	vpc.SnatAddress = "203.0.113.10"

	snatCalled := false

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "POST",
			Pattern: "/v2/iac/projects/test-project-id/vpcs",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.VPCResponse{VPC: vpc}),
		},
		{
			Method: "PUT",
			Pattern: "/v2/iac/projects/test-project-id/vpcs/vpc-001/snat",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				snatCalled = true
				w.WriteHeader(http.StatusOK)
			},
		},
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/vpcs/vpc-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.VPCResponse{VPC: vpc}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceVpc()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":        "test-vpc",
		"description": "a test vpc",
		"cidr":        "10.0.0.0/16",
		"enable_snat": true,
	})

	diags := res.CreateContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if !snatCalled {
		t.Error("expected SNAT PUT to have been called")
	}
	if v := d.Get("enable_snat").(bool); !v {
		t.Error("expected enable_snat true, got false")
	}
	if v := d.Get("snat_address").(string); v != "203.0.113.10" {
		t.Errorf("expected snat_address 203.0.113.10, got %s", v)
	}
}

func TestResourceVpcRead(t *testing.T) {
	vpc := testVPC()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/vpcs/vpc-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.VPCResponse{VPC: vpc}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceVpc()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":        "",
		"description": "",
		"cidr":        "",
		"enable_snat": false,
	})
	d.SetId("vpc-001")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
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

func TestResourceVpcRead_NotFound(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/vpcs/vpc-gone",
			Handler: testhelpers.EmptyHandler(http.StatusNotFound),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceVpc()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":        "",
		"description": "",
		"cidr":        "",
		"enable_snat": false,
	})
	d.SetId("vpc-gone")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error on 404: %v", diags)
	}

	if d.Id() != "" {
		t.Errorf("expected resource ID to be cleared after 404, got %s", d.Id())
	}
}

func TestResourceVpcDelete(t *testing.T) {
	vpc := testVPC()
	deletedCalled := false

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Pattern: "/v2/iac/projects/test-project-id/vpcs/vpc-001",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case "GET":
					if deletedCalled {
						w.WriteHeader(http.StatusNotFound)
						return
					}
					testhelpers.JSONHandler(t, http.StatusOK, dto.VPCResponse{VPC: vpc})(w, r)
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

	res := ResourceVpc()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":        "test-vpc",
		"description": "",
		"cidr":        "10.0.0.0/16",
		"enable_snat": false,
	})
	d.SetId("vpc-001")

	diags := res.DeleteContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if !deletedCalled {
		t.Error("expected DELETE to have been called")
	}
}
