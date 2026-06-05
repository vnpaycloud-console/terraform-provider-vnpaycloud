package databaseflavor

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func testFlavorDatabases() []dto.FlavorDatabase {
	return []dto.FlavorDatabase{
		{
			ID:       "flavor-001",
			Class:    "standard",
			Ratio:    "1:2",
			Name:     "db.standard.small",
			CpuReq:   1,
			MemReq:   2048,
			CpuLimit: 2,
			MemLimit: 4096,
		},
		{
			ID:       "flavor-002",
			Class:    "high_mem",
			Ratio:    "1:4",
			Name:     "db.highmem.medium",
			CpuReq:   2,
			MemReq:   8192,
			CpuLimit: 4,
			MemLimit: 16384,
		},
	}
}

func TestDataSourceDatabaseFlavorRead(t *testing.T) {
	flavors := testFlavorDatabases()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/database/flavor-databases",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListFlavorDatabasesResponse{
				FlavorDatabases: flavors,
				Total:           2,
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceDatabaseFlavor()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"id": "flavor-001",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "flavor-001" {
		t.Errorf("expected ID flavor-001, got %s", d.Id())
	}
	if v := d.Get("name").(string); v != "db.standard.small" {
		t.Errorf("expected name db.standard.small, got %s", v)
	}
	if v := d.Get("class").(string); v != "standard" {
		t.Errorf("expected class standard, got %s", v)
	}
	if v := d.Get("ratio").(string); v != "1:2" {
		t.Errorf("expected ratio 1:2, got %s", v)
	}
	if v := d.Get("cpu_req").(int); v != 1 {
		t.Errorf("expected cpu_req 1, got %d", v)
	}
	if v := d.Get("mem_req").(int); v != 2048 {
		t.Errorf("expected mem_req 2048, got %d", v)
	}
}

func TestDataSourceDatabaseFlavorRead_NotFound(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/database/flavor-databases",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListFlavorDatabasesResponse{
				FlavorDatabases: testFlavorDatabases(),
				Total:           2,
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceDatabaseFlavor()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"id": "nonexistent",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if !diags.HasError() {
		t.Fatal("expected error for nonexistent flavor, got none")
	}
}

func TestDataSourceDatabaseFlavorsRead(t *testing.T) {
	flavors := testFlavorDatabases()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/database/flavor-databases",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListFlavorDatabasesResponse{
				FlavorDatabases: flavors,
				Total:           2,
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceDatabaseFlavors()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	result := d.Get("flavors").([]interface{})
	if len(result) != 2 {
		t.Fatalf("expected 2 flavors, got %d", len(result))
	}

	first := result[0].(map[string]interface{})
	if first["id"] != "flavor-001" {
		t.Errorf("expected first id flavor-001, got %v", first["id"])
	}
	if first["name"] != "db.standard.small" {
		t.Errorf("expected first name db.standard.small, got %v", first["name"])
	}
	if first["class"] != "standard" {
		t.Errorf("expected first class standard, got %v", first["class"])
	}

	second := result[1].(map[string]interface{})
	if second["id"] != "flavor-002" {
		t.Errorf("expected second id flavor-002, got %v", second["id"])
	}
	if second["name"] != "db.highmem.medium" {
		t.Errorf("expected second name db.highmem.medium, got %v", second["name"])
	}
}

func TestDataSourceDatabaseFlavorRead_APIError(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/database/flavor-databases",
			Handler: testhelpers.EmptyHandler(http.StatusInternalServerError),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceDatabaseFlavor()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"id": "flavor-001",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if !diags.HasError() {
		t.Fatal("expected error on 500 response, got none")
	}
}

func TestDataSourceDatabaseFlavorsRead_Error(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/database/flavor-databases",
			Handler: testhelpers.EmptyHandler(http.StatusInternalServerError),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceDatabaseFlavors()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if !diags.HasError() {
		t.Fatal("expected error on 500 response, got none")
	}
}

func TestDataSourceDatabaseFlavorsRead_Empty(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/database/flavor-databases",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListFlavorDatabasesResponse{
				FlavorDatabases: []dto.FlavorDatabase{},
				Total:           0,
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceDatabaseFlavors()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	result := d.Get("flavors").([]interface{})
	if len(result) != 0 {
		t.Errorf("expected 0 flavors, got %d", len(result))
	}
}

func TestDataSourceDatabaseFlavorRead_SecondItem(t *testing.T) {
	flavors := testFlavorDatabases()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/database/flavor-databases",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListFlavorDatabasesResponse{
				FlavorDatabases: flavors,
				Total:           2,
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceDatabaseFlavor()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"id": "flavor-002",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "flavor-002" {
		t.Errorf("expected ID flavor-002, got %s", d.Id())
	}
	if v := d.Get("name").(string); v != "db.highmem.medium" {
		t.Errorf("expected name db.highmem.medium, got %s", v)
	}
	if v := d.Get("class").(string); v != "high_mem" {
		t.Errorf("expected class high_mem, got %s", v)
	}
	if v := d.Get("cpu_req").(int); v != 2 {
		t.Errorf("expected cpu_req 2, got %d", v)
	}
	if v := d.Get("mem_req").(int); v != 8192 {
		t.Errorf("expected mem_req 8192, got %d", v)
	}
	if v := d.Get("cpu_limit").(int); v != 4 {
		t.Errorf("expected cpu_limit 4, got %d", v)
	}
	if v := d.Get("mem_limit").(int); v != 16384 {
		t.Errorf("expected mem_limit 16384, got %d", v)
	}
}
