package databaseredis

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestDataSourceRedisInstanceRead(t *testing.T) {
	inst := testRedisInstance()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/database/redis-instances/redis-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.RedisInstanceResponse{RedisInstance: inst}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceDatabaseRedisInstance()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"id": "redis-001",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "redis-001" {
		t.Errorf("expected ID redis-001, got %s", d.Id())
	}
	if v := d.Get("name").(string); v != "test-redis" {
		t.Errorf("expected name test-redis, got %s", v)
	}
	if v := d.Get("version").(string); v != "7.4.1" {
		t.Errorf("expected version 7.4.1, got %s", v)
	}
	if v := d.Get("primary_ip").(string); v != "10.0.0.30" {
		t.Errorf("expected primary_ip 10.0.0.30, got %s", v)
	}
	if v := d.Get("status").(string); v != "active" {
		t.Errorf("expected status active, got %s", v)
	}
}

func TestDataSourceRedisInstancesRead(t *testing.T) {
	inst1 := testRedisInstance()
	inst2 := dto.RedisInstance{
		ID:          "redis-002",
		Name:        "prod-redis",
		Version:     "7.2.6",
		Replica:     1,
		PrimaryIP:   "10.0.0.40",
		PrimaryPort: 6379,
		Status:      "active",
		CreatedAt:   "2025-02-01T08:00:00Z",
	}

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/database/redis-instances",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListRedisInstancesResponse{
				RedisInstances: []dto.RedisInstance{inst1, inst2},
				Total:          2,
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceDatabaseRedisInstances()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	instances := d.Get("redis_instances").([]interface{})
	if len(instances) != 2 {
		t.Fatalf("expected 2 instances, got %d", len(instances))
	}

	first := instances[0].(map[string]interface{})
	if first["id"] != "redis-001" {
		t.Errorf("expected first id redis-001, got %v", first["id"])
	}
	if first["name"] != "test-redis" {
		t.Errorf("expected first name test-redis, got %v", first["name"])
	}

	second := instances[1].(map[string]interface{})
	if second["id"] != "redis-002" {
		t.Errorf("expected second id redis-002, got %v", second["id"])
	}
	if second["name"] != "prod-redis" {
		t.Errorf("expected second name prod-redis, got %v", second["name"])
	}
}

func TestDataSourceRedisInstanceRead_Error(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/database/redis-instances/redis-err",
			Handler: testhelpers.EmptyHandler(http.StatusInternalServerError),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceDatabaseRedisInstance()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"id": "redis-err",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if !diags.HasError() {
		t.Fatal("expected error on 500 response, got none")
	}
}

func TestDataSourceRedisInstancesRead_Error(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/database/redis-instances",
			Handler: testhelpers.EmptyHandler(http.StatusInternalServerError),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceDatabaseRedisInstances()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if !diags.HasError() {
		t.Fatal("expected error on 500 response, got none")
	}
}

func TestDataSourceRedisInstancesRead_Empty(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/database/redis-instances",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListRedisInstancesResponse{
				RedisInstances: []dto.RedisInstance{},
				Total:          0,
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceDatabaseRedisInstances()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	instances := d.Get("redis_instances").([]interface{})
	if len(instances) != 0 {
		t.Errorf("expected 0 instances, got %d", len(instances))
	}
}
