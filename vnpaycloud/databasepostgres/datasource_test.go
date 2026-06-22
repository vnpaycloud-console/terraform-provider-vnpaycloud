package databasepostgres

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestDataSourcePostgresInstanceRead(t *testing.T) {
	inst := testPostgresInstance()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/database/postgres-instances/pg-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.PostgresInstanceResponse{PostgresInstance: inst}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceDatabasePostgresInstance()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"id": "pg-001",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "pg-001" {
		t.Errorf("expected ID pg-001, got %s", d.Id())
	}
	if v := d.Get("name").(string); v != "test-postgres" {
		t.Errorf("expected name test-postgres, got %s", v)
	}
	if v := d.Get("version").(string); v != "17.5" {
		t.Errorf("expected version 17.5, got %s", v)
	}
	if v := d.Get("mode").(string); v != "standalone" {
		t.Errorf("expected mode standalone, got %s", v)
	}
	if v := d.Get("primary_ip").(string); v != "10.0.0.10" {
		t.Errorf("expected primary_ip 10.0.0.10, got %s", v)
	}
	if v := d.Get("status").(string); v != "active" {
		t.Errorf("expected status active, got %s", v)
	}
}

func TestDataSourcePostgresInstancesRead(t *testing.T) {
	inst1 := testPostgresInstance()
	inst2 := dto.PostgresInstance{
		ID:          "pg-002",
		Name:        "prod-postgres",
		Version:     "16.9",
		Mode:        "cluster",
		Replica:     3,
		PrimaryIP:   "10.0.0.20",
		PrimaryPort: 5432,
		Status:      "active",
		CreatedAt:   "2025-02-01T08:00:00Z",
	}

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/database/postgres-instances",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListPostgresInstancesResponse{
				PostgresInstances: []dto.PostgresInstance{inst1, inst2},
				Total:             2,
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceDatabasePostgresInstances()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	instances := d.Get("postgres_instances").([]interface{})
	if len(instances) != 2 {
		t.Fatalf("expected 2 instances, got %d", len(instances))
	}

	first := instances[0].(map[string]interface{})
	if first["id"] != "pg-001" {
		t.Errorf("expected first id pg-001, got %v", first["id"])
	}
	if first["name"] != "test-postgres" {
		t.Errorf("expected first name test-postgres, got %v", first["name"])
	}

	second := instances[1].(map[string]interface{})
	if second["id"] != "pg-002" {
		t.Errorf("expected second id pg-002, got %v", second["id"])
	}
	if second["name"] != "prod-postgres" {
		t.Errorf("expected second name prod-postgres, got %v", second["name"])
	}
	if second["mode"] != "cluster" {
		t.Errorf("expected second mode cluster, got %v", second["mode"])
	}
}

func TestDataSourcePostgresInstanceRead_Error(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/database/postgres-instances/pg-err",
			Handler: testhelpers.EmptyHandler(http.StatusInternalServerError),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceDatabasePostgresInstance()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"id": "pg-err",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if !diags.HasError() {
		t.Fatal("expected error on 500 response, got none")
	}
}

func TestDataSourcePostgresInstancesRead_Error(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/database/postgres-instances",
			Handler: testhelpers.EmptyHandler(http.StatusInternalServerError),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceDatabasePostgresInstances()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if !diags.HasError() {
		t.Fatal("expected error on 500 response, got none")
	}
}

func TestDataSourcePostgresInstancesRead_Empty(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/database/postgres-instances",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListPostgresInstancesResponse{
				PostgresInstances: []dto.PostgresInstance{},
				Total:             0,
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceDatabasePostgresInstances()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	instances := d.Get("postgres_instances").([]interface{})
	if len(instances) != 0 {
		t.Errorf("expected 0 instances, got %d", len(instances))
	}
}
