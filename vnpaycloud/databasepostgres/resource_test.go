package databasepostgres

import (
	"context"
	"net/http"
	"sync/atomic"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func testPostgresInstance() dto.PostgresInstance {
	return dto.PostgresInstance{
		ID:                 "pg-001",
		Name:               "test-postgres",
		Description:        "test postgres instance",
		FlavorDatabaseID:   "flavor-001",
		Version:            "17.5",
		VolumeType:         "SSD",
		VolumeSize:         50,
		Mode:               "standalone",
		PrimaryIP:          "10.0.0.10",
		PrimaryPort:        5432,
		StandbyIP:          "",
		StandbyPort:        0,
		Replica:            1,
		Purpose:            "testing",
		IsAutoExpandVolume: false,
		EnableTls:          false,
		ZoneID:             "test-zone-id",
		Status:             "active",
		CreatedAt:          "2025-01-15T10:00:00Z",
	}
}

func postgresSchemaRaw() map[string]interface{} {
	return map[string]interface{}{
		"name":                  "test-postgres",
		"description":           "test postgres instance",
		"flavor_database_id":    "flavor-001",
		"version":               "17.5",
		"volume_type":           "SSD",
		"volume_size":           50,
		"mode":                  "standalone",
		"replica":               1,
		"purpose":               "testing",
		"enable_tls":            false,
		"certificate_id":        "",
		"tls_mode":              "",
		"is_auto_expand_volume": false,
		"usage_threshold":       0,
		"scale_percent":         0,
	}
}

func TestResourcePostgresInstanceCreate(t *testing.T) {
	inst := testPostgresInstance()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "POST",
			Pattern: "/v2/iac/projects/test-project-id/database/postgres-instances",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.PostgresInstanceResponse{PostgresInstance: inst}),
		},
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/database/postgres-instances/pg-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.PostgresInstanceResponse{PostgresInstance: inst}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceDatabasePostgresInstance()
	d := schema.TestResourceDataRaw(t, res.Schema, postgresSchemaRaw())

	diags := res.CreateContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "pg-001" {
		t.Errorf("expected ID pg-001, got %s", d.Id())
	}
	if v := d.Get("name").(string); v != "test-postgres" {
		t.Errorf("expected name test-postgres, got %s", v)
	}
	if v := d.Get("status").(string); v != "active" {
		t.Errorf("expected status active, got %s", v)
	}
	if v := d.Get("primary_ip").(string); v != "10.0.0.10" {
		t.Errorf("expected primary_ip 10.0.0.10, got %s", v)
	}
	if v := d.Get("primary_port").(int); v != 5432 {
		t.Errorf("expected primary_port 5432, got %d", v)
	}
}

func TestResourcePostgresInstanceRead(t *testing.T) {
	inst := testPostgresInstance()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/database/postgres-instances/pg-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.PostgresInstanceResponse{PostgresInstance: inst}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceDatabasePostgresInstance()
	d := schema.TestResourceDataRaw(t, res.Schema, postgresSchemaRaw())
	d.SetId("pg-001")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if v := d.Get("name").(string); v != "test-postgres" {
		t.Errorf("expected name test-postgres, got %s", v)
	}
	if v := d.Get("flavor_database_id").(string); v != "flavor-001" {
		t.Errorf("expected flavor_database_id flavor-001, got %s", v)
	}
	if v := d.Get("version").(string); v != "17.5" {
		t.Errorf("expected version 17.5, got %s", v)
	}
	if v := d.Get("volume_type").(string); v != "SSD" {
		t.Errorf("expected volume_type SSD, got %s", v)
	}
	if v := d.Get("volume_size").(int); v != 50 {
		t.Errorf("expected volume_size 50, got %d", v)
	}
	if v := d.Get("mode").(string); v != "standalone" {
		t.Errorf("expected mode standalone, got %s", v)
	}
	if v := d.Get("replica").(int); v != 1 {
		t.Errorf("expected replica 1, got %d", v)
	}
	if v := d.Get("primary_ip").(string); v != "10.0.0.10" {
		t.Errorf("expected primary_ip 10.0.0.10, got %s", v)
	}
	if v := d.Get("status").(string); v != "active" {
		t.Errorf("expected status active, got %s", v)
	}
	if v := d.Get("created_at").(string); v != "2025-01-15T10:00:00Z" {
		t.Errorf("expected created_at 2025-01-15T10:00:00Z, got %s", v)
	}
}

func TestResourcePostgresInstanceRead_NotFound(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/database/postgres-instances/pg-gone",
			Handler: testhelpers.EmptyHandler(http.StatusNotFound),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceDatabasePostgresInstance()
	d := schema.TestResourceDataRaw(t, res.Schema, postgresSchemaRaw())
	d.SetId("pg-gone")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error on 404: %v", diags)
	}

	if d.Id() != "" {
		t.Errorf("expected resource ID to be cleared after 404, got %s", d.Id())
	}
}

func TestResourcePostgresInstanceDelete(t *testing.T) {
	inst := testPostgresInstance()
	deletedCalled := false

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Pattern: "/v2/iac/projects/test-project-id/database/postgres-instances/pg-001",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case "GET":
					if deletedCalled {
						w.WriteHeader(http.StatusNotFound)
						return
					}
					testhelpers.JSONHandler(t, http.StatusOK, dto.PostgresInstanceResponse{PostgresInstance: inst})(w, r)
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

	res := ResourceDatabasePostgresInstance()
	d := schema.TestResourceDataRaw(t, res.Schema, postgresSchemaRaw())
	d.SetId("pg-001")

	diags := res.DeleteContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if !deletedCalled {
		t.Error("expected DELETE to have been called")
	}
}

func TestResourcePostgresInstanceCreate_StateTransition(t *testing.T) {
	inst := testPostgresInstance()

	var getCalls int32
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "POST",
			Pattern: "/v2/iac/projects/test-project-id/database/postgres-instances",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.PostgresInstanceResponse{PostgresInstance: inst}),
		},
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/database/postgres-instances/pg-001",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				n := atomic.AddInt32(&getCalls, 1)
				resp := inst
				if n <= 2 {
					resp.Status = "creating"
				}
				testhelpers.JSONHandler(t, http.StatusOK, dto.PostgresInstanceResponse{PostgresInstance: resp})(w, r)
			},
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceDatabasePostgresInstance()
	d := schema.TestResourceDataRaw(t, res.Schema, postgresSchemaRaw())

	diags := res.CreateContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "pg-001" {
		t.Errorf("expected ID pg-001, got %s", d.Id())
	}
	if v := d.Get("status").(string); v != "active" {
		t.Errorf("expected status active, got %s", v)
	}
	if n := atomic.LoadInt32(&getCalls); n < 3 {
		t.Errorf("expected at least 3 GET calls for state polling, got %d", n)
	}
}

func TestResourcePostgresInstanceCreate_Error(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "POST",
			Pattern: "/v2/iac/projects/test-project-id/database/postgres-instances",
			Handler: testhelpers.EmptyHandler(http.StatusInternalServerError),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceDatabasePostgresInstance()
	d := schema.TestResourceDataRaw(t, res.Schema, postgresSchemaRaw())

	diags := res.CreateContext(context.Background(), d, cfg)
	if !diags.HasError() {
		t.Fatal("expected error on 500 response, got none")
	}
}

func TestResourcePostgresInstanceCreate_WithAutoExpand(t *testing.T) {
	inst := testPostgresInstance()
	instWithAutoExpand := inst
	instWithAutoExpand.IsAutoExpandVolume = true
	instWithAutoExpand.UsageThreshold = 80
	instWithAutoExpand.ScalePercent = 20

	autoExpandCalled := false
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "POST",
			Pattern: "/v2/iac/projects/test-project-id/database/postgres-instances/pg-001/enable-auto-expand-volume",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				autoExpandCalled = true
				w.WriteHeader(http.StatusOK)
			},
		},
		{
			Method:  "POST",
			Pattern: "/v2/iac/projects/test-project-id/database/postgres-instances",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.PostgresInstanceResponse{PostgresInstance: inst}),
		},
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/database/postgres-instances/pg-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.PostgresInstanceResponse{PostgresInstance: instWithAutoExpand}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceDatabasePostgresInstance()
	raw := postgresSchemaRaw()
	raw["is_auto_expand_volume"] = true
	raw["usage_threshold"] = 80
	raw["scale_percent"] = 20
	d := schema.TestResourceDataRaw(t, res.Schema, raw)

	diags := res.CreateContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if !autoExpandCalled {
		t.Error("expected auto-expand POST to have been called")
	}
	if v := d.Get("is_auto_expand_volume").(bool); !v {
		t.Errorf("expected is_auto_expand_volume true, got false")
	}
	if v := d.Get("usage_threshold").(int); v != 80 {
		t.Errorf("expected usage_threshold 80, got %d", v)
	}
	if v := d.Get("scale_percent").(int); v != 20 {
		t.Errorf("expected scale_percent 20, got %d", v)
	}
}

func TestResourcePostgresInstanceCreate_WithTLS(t *testing.T) {
	inst := testPostgresInstance()
	inst.EnableTls = true
	inst.CertificateID = "cert-001"
	inst.TlsMode = "require"

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "POST",
			Pattern: "/v2/iac/projects/test-project-id/database/postgres-instances",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.PostgresInstanceResponse{PostgresInstance: inst}),
		},
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/database/postgres-instances/pg-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.PostgresInstanceResponse{PostgresInstance: inst}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceDatabasePostgresInstance()
	raw := postgresSchemaRaw()
	raw["enable_tls"] = true
	raw["certificate_id"] = "cert-001"
	raw["tls_mode"] = "require"
	d := schema.TestResourceDataRaw(t, res.Schema, raw)

	diags := res.CreateContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if v := d.Get("enable_tls").(bool); !v {
		t.Errorf("expected enable_tls true, got false")
	}
	if v := d.Get("certificate_id").(string); v != "cert-001" {
		t.Errorf("expected certificate_id cert-001, got %s", v)
	}
	if v := d.Get("tls_mode").(string); v != "require" {
		t.Errorf("expected tls_mode require, got %s", v)
	}
}

func TestResourcePostgresInstanceDelete_StateTransition(t *testing.T) {
	inst := testPostgresInstance()
	deletedCalled := false

	var getCalls int32
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Pattern: "/v2/iac/projects/test-project-id/database/postgres-instances/pg-001",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case "GET":
					n := atomic.AddInt32(&getCalls, 1)
					if n >= 3 {
						w.WriteHeader(http.StatusNotFound)
						return
					}
					deleting := inst
					deleting.Status = "deleting"
					testhelpers.JSONHandler(t, http.StatusOK, dto.PostgresInstanceResponse{PostgresInstance: deleting})(w, r)
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

	res := ResourceDatabasePostgresInstance()
	d := schema.TestResourceDataRaw(t, res.Schema, postgresSchemaRaw())
	d.SetId("pg-001")

	diags := res.DeleteContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if !deletedCalled {
		t.Error("expected DELETE to have been called")
	}
	if n := atomic.LoadInt32(&getCalls); n < 3 {
		t.Errorf("expected at least 3 GET calls for delete polling, got %d", n)
	}
}

func TestResourcePostgresInstanceRead_Error(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/database/postgres-instances/pg-err",
			Handler: testhelpers.EmptyHandler(http.StatusInternalServerError),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceDatabasePostgresInstance()
	d := schema.TestResourceDataRaw(t, res.Schema, postgresSchemaRaw())
	d.SetId("pg-err")

	diags := res.ReadContext(context.Background(), d, cfg)
	if !diags.HasError() {
		t.Fatal("expected error on 500 response, got none")
	}
	if d.Id() == "" {
		t.Error("expected resource ID to be preserved on non-404 error")
	}
}
