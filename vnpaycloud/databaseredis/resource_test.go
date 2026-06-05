package databaseredis

import (
	"context"
	"net/http"
	"sync/atomic"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func testRedisInstance() dto.RedisInstance {
	return dto.RedisInstance{
		ID:                 "redis-001",
		Name:               "test-redis",
		Description:        "test redis instance",
		FlavorDatabaseID:   "flavor-002",
		Version:            "7.4.1",
		VolumeType:         "SSD",
		VolumeSize:         20,
		Mode:               "standalone",
		PrimaryIP:          "10.0.0.30",
		PrimaryPort:        6379,
		Replica:            1,
		Purpose:            "caching",
		IsAutoExpandVolume: false,
		EnableTls:          false,
		ZoneID:             "test-zone-id",
		Status:             "active",
		CreatedAt:          "2025-01-15T10:00:00Z",
	}
}

func redisSchemaRaw() map[string]interface{} {
	return map[string]interface{}{
		"name":                  "test-redis",
		"description":          "test redis instance",
		"flavor_database_id":   "flavor-002",
		"version":              "7.4.1",
		"volume_type":          "SSD",
		"volume_size":          20,
		"replica":              1,
		"purpose":              "caching",
		"enable_tls":           false,
		"certificate_id":       "",
		"is_auto_expand_volume": false,
		"usage_threshold":      0,
		"scale_percent":        0,
	}
}

func TestResourceRedisInstanceCreate(t *testing.T) {
	inst := testRedisInstance()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "POST",
			Pattern: "/v2/iac/projects/test-project-id/database/redis-instances",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.RedisInstanceResponse{RedisInstance: inst}),
		},
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/database/redis-instances/redis-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.RedisInstanceResponse{RedisInstance: inst}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceDatabaseRedisInstance()
	d := schema.TestResourceDataRaw(t, res.Schema, redisSchemaRaw())

	diags := res.CreateContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "redis-001" {
		t.Errorf("expected ID redis-001, got %s", d.Id())
	}
	if v := d.Get("name").(string); v != "test-redis" {
		t.Errorf("expected name test-redis, got %s", v)
	}
	if v := d.Get("status").(string); v != "active" {
		t.Errorf("expected status active, got %s", v)
	}
	if v := d.Get("primary_ip").(string); v != "10.0.0.30" {
		t.Errorf("expected primary_ip 10.0.0.30, got %s", v)
	}
	if v := d.Get("primary_port").(int); v != 6379 {
		t.Errorf("expected primary_port 6379, got %d", v)
	}
}

func TestResourceRedisInstanceRead(t *testing.T) {
	inst := testRedisInstance()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/database/redis-instances/redis-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.RedisInstanceResponse{RedisInstance: inst}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceDatabaseRedisInstance()
	d := schema.TestResourceDataRaw(t, res.Schema, redisSchemaRaw())
	d.SetId("redis-001")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if v := d.Get("name").(string); v != "test-redis" {
		t.Errorf("expected name test-redis, got %s", v)
	}
	if v := d.Get("flavor_database_id").(string); v != "flavor-002" {
		t.Errorf("expected flavor_database_id flavor-002, got %s", v)
	}
	if v := d.Get("version").(string); v != "7.4.1" {
		t.Errorf("expected version 7.4.1, got %s", v)
	}
	if v := d.Get("volume_size").(int); v != 20 {
		t.Errorf("expected volume_size 20, got %d", v)
	}
	if v := d.Get("primary_ip").(string); v != "10.0.0.30" {
		t.Errorf("expected primary_ip 10.0.0.30, got %s", v)
	}
	if v := d.Get("status").(string); v != "active" {
		t.Errorf("expected status active, got %s", v)
	}
	if v := d.Get("created_at").(string); v != "2025-01-15T10:00:00Z" {
		t.Errorf("expected created_at 2025-01-15T10:00:00Z, got %s", v)
	}
}

func TestResourceRedisInstanceRead_NotFound(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/database/redis-instances/redis-gone",
			Handler: testhelpers.EmptyHandler(http.StatusNotFound),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceDatabaseRedisInstance()
	d := schema.TestResourceDataRaw(t, res.Schema, redisSchemaRaw())
	d.SetId("redis-gone")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error on 404: %v", diags)
	}

	if d.Id() != "" {
		t.Errorf("expected resource ID to be cleared after 404, got %s", d.Id())
	}
}

func TestResourceRedisInstanceDelete(t *testing.T) {
	inst := testRedisInstance()
	deletedCalled := false

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Pattern: "/v2/iac/projects/test-project-id/database/redis-instances/redis-001",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case "GET":
					if deletedCalled {
						w.WriteHeader(http.StatusNotFound)
						return
					}
					testhelpers.JSONHandler(t, http.StatusOK, dto.RedisInstanceResponse{RedisInstance: inst})(w, r)
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

	res := ResourceDatabaseRedisInstance()
	d := schema.TestResourceDataRaw(t, res.Schema, redisSchemaRaw())
	d.SetId("redis-001")

	diags := res.DeleteContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if !deletedCalled {
		t.Error("expected DELETE to have been called")
	}
}

func TestResourceRedisInstanceCreate_StateTransition(t *testing.T) {
	inst := testRedisInstance()

	var getCalls int32
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "POST",
			Pattern: "/v2/iac/projects/test-project-id/database/redis-instances",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.RedisInstanceResponse{RedisInstance: inst}),
		},
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/database/redis-instances/redis-001",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				n := atomic.AddInt32(&getCalls, 1)
				resp := inst
				if n <= 2 {
					resp.Status = "creating"
				}
				testhelpers.JSONHandler(t, http.StatusOK, dto.RedisInstanceResponse{RedisInstance: resp})(w, r)
			},
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceDatabaseRedisInstance()
	d := schema.TestResourceDataRaw(t, res.Schema, redisSchemaRaw())

	diags := res.CreateContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "redis-001" {
		t.Errorf("expected ID redis-001, got %s", d.Id())
	}
	if v := d.Get("status").(string); v != "active" {
		t.Errorf("expected status active, got %s", v)
	}
	if n := atomic.LoadInt32(&getCalls); n < 3 {
		t.Errorf("expected at least 3 GET calls for state polling, got %d", n)
	}
}

func TestResourceRedisInstanceCreate_Error(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "POST",
			Pattern: "/v2/iac/projects/test-project-id/database/redis-instances",
			Handler: testhelpers.EmptyHandler(http.StatusInternalServerError),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceDatabaseRedisInstance()
	d := schema.TestResourceDataRaw(t, res.Schema, redisSchemaRaw())

	diags := res.CreateContext(context.Background(), d, cfg)
	if !diags.HasError() {
		t.Fatal("expected error on 500 response, got none")
	}
}

func TestResourceRedisInstanceCreate_WithAutoExpand(t *testing.T) {
	inst := testRedisInstance()
	instWithAutoExpand := inst
	instWithAutoExpand.IsAutoExpandVolume = true
	instWithAutoExpand.UsageThreshold = 80
	instWithAutoExpand.ScalePercent = 20

	autoExpandCalled := false
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "POST",
			Pattern: "/v2/iac/projects/test-project-id/database/redis-instances/redis-001/enable-auto-expand-volume",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				autoExpandCalled = true
				w.WriteHeader(http.StatusOK)
			},
		},
		{
			Method:  "POST",
			Pattern: "/v2/iac/projects/test-project-id/database/redis-instances",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.RedisInstanceResponse{RedisInstance: inst}),
		},
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/database/redis-instances/redis-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.RedisInstanceResponse{RedisInstance: instWithAutoExpand}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceDatabaseRedisInstance()
	raw := redisSchemaRaw()
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
}

func TestResourceRedisInstanceCreate_WithTLS(t *testing.T) {
	inst := testRedisInstance()
	inst.EnableTls = true
	inst.CertificateID = "cert-001"

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "POST",
			Pattern: "/v2/iac/projects/test-project-id/database/redis-instances",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.RedisInstanceResponse{RedisInstance: inst}),
		},
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/database/redis-instances/redis-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.RedisInstanceResponse{RedisInstance: inst}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceDatabaseRedisInstance()
	raw := redisSchemaRaw()
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

func TestResourceRedisInstanceDelete_StateTransition(t *testing.T) {
	inst := testRedisInstance()
	deletedCalled := false

	var getCalls int32
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Pattern: "/v2/iac/projects/test-project-id/database/redis-instances/redis-001",
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
					testhelpers.JSONHandler(t, http.StatusOK, dto.RedisInstanceResponse{RedisInstance: deleting})(w, r)
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

	res := ResourceDatabaseRedisInstance()
	d := schema.TestResourceDataRaw(t, res.Schema, redisSchemaRaw())
	d.SetId("redis-001")

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

func TestResourceRedisInstanceRead_Error(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/database/redis-instances/redis-err",
			Handler: testhelpers.EmptyHandler(http.StatusInternalServerError),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceDatabaseRedisInstance()
	d := schema.TestResourceDataRaw(t, res.Schema, redisSchemaRaw())
	d.SetId("redis-err")

	diags := res.ReadContext(context.Background(), d, cfg)
	if !diags.HasError() {
		t.Fatal("expected error on 500 response, got none")
	}
	if d.Id() == "" {
		t.Error("expected resource ID to be preserved on non-404 error")
	}
}
