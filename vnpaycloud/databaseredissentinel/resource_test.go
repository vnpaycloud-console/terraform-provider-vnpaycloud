package databaseredissentinel

import (
	"context"
	"net/http"
	"sync/atomic"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func testRedisSentinelInstance() dto.RedisSentinelInstance {
	return dto.RedisSentinelInstance{
		ID:                       "rs-001",
		Name:                     "test-redis-sentinel",
		Description:              "test redis sentinel",
		FlavorDatabaseID:         "flavor-003",
		Version:                  "7.4.1",
		VolumeType:               "SSD",
		VolumeSize:               30,
		PrimaryIP:                "10.0.0.50",
		PrimaryPort:              6379,
		StandbyIP:                "10.0.0.51",
		StandbyPort:              6379,
		Replica:                  2,
		Purpose:                  "ha-caching",
		SentinelName:             "mysentinel",
		SentinelReplica:          3,
		SentinelFlavorDatabaseID: "flavor-004",
		SentinelVolumeSize:       10,
		IsAutoExpandVolume:       false,
		EnableTls:                false,
		ZoneID:                   "test-zone-id",
		Status:                   "active",
		CreatedAt:                "2025-01-15T10:00:00Z",
	}
}

func redisSentinelSchemaRaw() map[string]interface{} {
	return map[string]interface{}{
		"name":                        "test-redis-sentinel",
		"description":                 "test redis sentinel",
		"flavor_database_id":          "flavor-003",
		"version":                     "7.4.1",
		"volume_type":                 "SSD",
		"volume_size":                 30,
		"replica":                     2,
		"purpose":                     "ha-caching",
		"sentinel_name":               "mysentinel",
		"sentinel_replica":            3,
		"sentinel_flavor_database_id": "flavor-004",
		"sentinel_volume_size":        10,
		"enable_tls":                  false,
		"certificate_id":              "",
		"is_auto_expand_volume":       false,
		"usage_threshold":             0,
		"scale_percent":               0,
	}
}

func TestResourceRedisSentinelInstanceCreate(t *testing.T) {
	inst := testRedisSentinelInstance()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "POST",
			Pattern: "/v2/iac/projects/test-project-id/database/redis-sentinel-instances",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.RedisSentinelInstanceResponse{RedisSentinelInstance: inst}),
		},
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/database/redis-sentinel-instances/rs-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.RedisSentinelInstanceResponse{RedisSentinelInstance: inst}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceDatabaseRedisSentinelInstance()
	d := schema.TestResourceDataRaw(t, res.Schema, redisSentinelSchemaRaw())

	diags := res.CreateContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "rs-001" {
		t.Errorf("expected ID rs-001, got %s", d.Id())
	}
	if v := d.Get("name").(string); v != "test-redis-sentinel" {
		t.Errorf("expected name test-redis-sentinel, got %s", v)
	}
	if v := d.Get("status").(string); v != "active" {
		t.Errorf("expected status active, got %s", v)
	}
	if v := d.Get("primary_ip").(string); v != "10.0.0.50" {
		t.Errorf("expected primary_ip 10.0.0.50, got %s", v)
	}
	if v := d.Get("standby_ip").(string); v != "10.0.0.51" {
		t.Errorf("expected standby_ip 10.0.0.51, got %s", v)
	}
	if v := d.Get("sentinel_replica").(int); v != 3 {
		t.Errorf("expected sentinel_replica 3, got %d", v)
	}
}

func TestResourceRedisSentinelInstanceRead(t *testing.T) {
	inst := testRedisSentinelInstance()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/database/redis-sentinel-instances/rs-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.RedisSentinelInstanceResponse{RedisSentinelInstance: inst}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceDatabaseRedisSentinelInstance()
	d := schema.TestResourceDataRaw(t, res.Schema, redisSentinelSchemaRaw())
	d.SetId("rs-001")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if v := d.Get("name").(string); v != "test-redis-sentinel" {
		t.Errorf("expected name test-redis-sentinel, got %s", v)
	}
	if v := d.Get("flavor_database_id").(string); v != "flavor-003" {
		t.Errorf("expected flavor_database_id flavor-003, got %s", v)
	}
	if v := d.Get("version").(string); v != "7.4.1" {
		t.Errorf("expected version 7.4.1, got %s", v)
	}
	if v := d.Get("volume_size").(int); v != 30 {
		t.Errorf("expected volume_size 30, got %d", v)
	}
	if v := d.Get("replica").(int); v != 2 {
		t.Errorf("expected replica 2, got %d", v)
	}
	if v := d.Get("sentinel_name").(string); v != "mysentinel" {
		t.Errorf("expected sentinel_name mysentinel, got %s", v)
	}
	if v := d.Get("sentinel_replica").(int); v != 3 {
		t.Errorf("expected sentinel_replica 3, got %d", v)
	}
	if v := d.Get("sentinel_flavor_database_id").(string); v != "flavor-004" {
		t.Errorf("expected sentinel_flavor_database_id flavor-004, got %s", v)
	}
	if v := d.Get("sentinel_volume_size").(int); v != 10 {
		t.Errorf("expected sentinel_volume_size 10, got %d", v)
	}
	if v := d.Get("primary_ip").(string); v != "10.0.0.50" {
		t.Errorf("expected primary_ip 10.0.0.50, got %s", v)
	}
	if v := d.Get("standby_ip").(string); v != "10.0.0.51" {
		t.Errorf("expected standby_ip 10.0.0.51, got %s", v)
	}
	if v := d.Get("status").(string); v != "active" {
		t.Errorf("expected status active, got %s", v)
	}
	if v := d.Get("created_at").(string); v != "2025-01-15T10:00:00Z" {
		t.Errorf("expected created_at 2025-01-15T10:00:00Z, got %s", v)
	}
}

func TestResourceRedisSentinelInstanceRead_NotFound(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/database/redis-sentinel-instances/rs-gone",
			Handler: testhelpers.EmptyHandler(http.StatusNotFound),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceDatabaseRedisSentinelInstance()
	d := schema.TestResourceDataRaw(t, res.Schema, redisSentinelSchemaRaw())
	d.SetId("rs-gone")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error on 404: %v", diags)
	}

	if d.Id() != "" {
		t.Errorf("expected resource ID to be cleared after 404, got %s", d.Id())
	}
}

func TestResourceRedisSentinelInstanceDelete(t *testing.T) {
	inst := testRedisSentinelInstance()
	deletedCalled := false

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Pattern: "/v2/iac/projects/test-project-id/database/redis-sentinel-instances/rs-001",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case "GET":
					if deletedCalled {
						w.WriteHeader(http.StatusNotFound)
						return
					}
					testhelpers.JSONHandler(t, http.StatusOK, dto.RedisSentinelInstanceResponse{RedisSentinelInstance: inst})(w, r)
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

	res := ResourceDatabaseRedisSentinelInstance()
	d := schema.TestResourceDataRaw(t, res.Schema, redisSentinelSchemaRaw())
	d.SetId("rs-001")

	diags := res.DeleteContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if !deletedCalled {
		t.Error("expected DELETE to have been called")
	}
}

func TestResourceRedisSentinelInstanceCreate_StateTransition(t *testing.T) {
	inst := testRedisSentinelInstance()

	var getCalls int32
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "POST",
			Pattern: "/v2/iac/projects/test-project-id/database/redis-sentinel-instances",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.RedisSentinelInstanceResponse{RedisSentinelInstance: inst}),
		},
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/database/redis-sentinel-instances/rs-001",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				n := atomic.AddInt32(&getCalls, 1)
				resp := inst
				if n <= 2 {
					resp.Status = "creating"
				}
				testhelpers.JSONHandler(t, http.StatusOK, dto.RedisSentinelInstanceResponse{RedisSentinelInstance: resp})(w, r)
			},
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceDatabaseRedisSentinelInstance()
	d := schema.TestResourceDataRaw(t, res.Schema, redisSentinelSchemaRaw())

	diags := res.CreateContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "rs-001" {
		t.Errorf("expected ID rs-001, got %s", d.Id())
	}
	if v := d.Get("status").(string); v != "active" {
		t.Errorf("expected status active, got %s", v)
	}
	if n := atomic.LoadInt32(&getCalls); n < 3 {
		t.Errorf("expected at least 3 GET calls for state polling, got %d", n)
	}
}

func TestResourceRedisSentinelInstanceCreate_Error(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "POST",
			Pattern: "/v2/iac/projects/test-project-id/database/redis-sentinel-instances",
			Handler: testhelpers.EmptyHandler(http.StatusInternalServerError),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceDatabaseRedisSentinelInstance()
	d := schema.TestResourceDataRaw(t, res.Schema, redisSentinelSchemaRaw())

	diags := res.CreateContext(context.Background(), d, cfg)
	if !diags.HasError() {
		t.Fatal("expected error on 500 response, got none")
	}
}

func TestResourceRedisSentinelInstanceCreate_WithAutoExpand(t *testing.T) {
	inst := testRedisSentinelInstance()
	instWithAutoExpand := inst
	instWithAutoExpand.IsAutoExpandVolume = true
	instWithAutoExpand.UsageThreshold = 70
	instWithAutoExpand.ScalePercent = 30

	autoExpandCalled := false
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "POST",
			Pattern: "/v2/iac/projects/test-project-id/database/redis-sentinel-instances/rs-001/enable-auto-expand-volume",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				autoExpandCalled = true
				w.WriteHeader(http.StatusOK)
			},
		},
		{
			Method:  "POST",
			Pattern: "/v2/iac/projects/test-project-id/database/redis-sentinel-instances",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.RedisSentinelInstanceResponse{RedisSentinelInstance: inst}),
		},
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/database/redis-sentinel-instances/rs-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.RedisSentinelInstanceResponse{RedisSentinelInstance: instWithAutoExpand}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceDatabaseRedisSentinelInstance()
	raw := redisSentinelSchemaRaw()
	raw["is_auto_expand_volume"] = true
	raw["usage_threshold"] = 70
	raw["scale_percent"] = 30
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
}

func TestResourceRedisSentinelInstanceCreate_WithTLS(t *testing.T) {
	inst := testRedisSentinelInstance()
	inst.EnableTls = true
	inst.CertificateID = "cert-001"

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "POST",
			Pattern: "/v2/iac/projects/test-project-id/database/redis-sentinel-instances",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.RedisSentinelInstanceResponse{RedisSentinelInstance: inst}),
		},
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/database/redis-sentinel-instances/rs-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.RedisSentinelInstanceResponse{RedisSentinelInstance: inst}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceDatabaseRedisSentinelInstance()
	raw := redisSentinelSchemaRaw()
	raw["enable_tls"] = true
	raw["certificate_id"] = "cert-001"
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
}

func TestResourceRedisSentinelInstanceDelete_StateTransition(t *testing.T) {
	inst := testRedisSentinelInstance()
	deletedCalled := false

	var getCalls int32
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Pattern: "/v2/iac/projects/test-project-id/database/redis-sentinel-instances/rs-001",
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
					testhelpers.JSONHandler(t, http.StatusOK, dto.RedisSentinelInstanceResponse{RedisSentinelInstance: deleting})(w, r)
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

	res := ResourceDatabaseRedisSentinelInstance()
	d := schema.TestResourceDataRaw(t, res.Schema, redisSentinelSchemaRaw())
	d.SetId("rs-001")

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

func TestResourceRedisSentinelInstanceRead_Error(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/database/redis-sentinel-instances/rs-err",
			Handler: testhelpers.EmptyHandler(http.StatusInternalServerError),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceDatabaseRedisSentinelInstance()
	d := schema.TestResourceDataRaw(t, res.Schema, redisSentinelSchemaRaw())
	d.SetId("rs-err")

	diags := res.ReadContext(context.Background(), d, cfg)
	if !diags.HasError() {
		t.Fatal("expected error on 500 response, got none")
	}
	if d.Id() == "" {
		t.Error("expected resource ID to be preserved on non-404 error")
	}
}
