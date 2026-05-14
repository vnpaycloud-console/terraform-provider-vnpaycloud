package databaseredissentinel

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestDataSourceRedisSentinelInstanceRead(t *testing.T) {
	inst := testRedisSentinelInstance()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/database/redis-sentinel-instances/rs-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.RedisSentinelInstanceResponse{RedisSentinelInstance: inst}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceDatabaseRedisSentinelInstance()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"id": "rs-001",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "rs-001" {
		t.Errorf("expected ID rs-001, got %s", d.Id())
	}
	if v := d.Get("name").(string); v != "test-redis-sentinel" {
		t.Errorf("expected name test-redis-sentinel, got %s", v)
	}
	if v := d.Get("version").(string); v != "7.4.1" {
		t.Errorf("expected version 7.4.1, got %s", v)
	}
	if v := d.Get("sentinel_name").(string); v != "mysentinel" {
		t.Errorf("expected sentinel_name mysentinel, got %s", v)
	}
	if v := d.Get("sentinel_replica").(int); v != 3 {
		t.Errorf("expected sentinel_replica 3, got %d", v)
	}
	if v := d.Get("primary_ip").(string); v != "10.0.0.50" {
		t.Errorf("expected primary_ip 10.0.0.50, got %s", v)
	}
	if v := d.Get("status").(string); v != "active" {
		t.Errorf("expected status active, got %s", v)
	}
}

func TestDataSourceRedisSentinelInstancesRead(t *testing.T) {
	inst1 := testRedisSentinelInstance()
	inst2 := dto.RedisSentinelInstance{
		ID:              "rs-002",
		Name:            "prod-redis-sentinel",
		Version:         "7.2.6",
		Replica:         3,
		SentinelReplica: 5,
		PrimaryIP:       "10.0.0.60",
		PrimaryPort:     6379,
		Status:          "active",
		CreatedAt:       "2025-02-01T08:00:00Z",
	}

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/database/redis-sentinel-instances",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListRedisSentinelInstancesResponse{
				RedisSentinelInstances: []dto.RedisSentinelInstance{inst1, inst2},
				Total:                  2,
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceDatabaseRedisSentinelInstances()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	instances := d.Get("redis_sentinel_instances").([]interface{})
	if len(instances) != 2 {
		t.Fatalf("expected 2 instances, got %d", len(instances))
	}

	first := instances[0].(map[string]interface{})
	if first["id"] != "rs-001" {
		t.Errorf("expected first id rs-001, got %v", first["id"])
	}
	if first["name"] != "test-redis-sentinel" {
		t.Errorf("expected first name test-redis-sentinel, got %v", first["name"])
	}
	if first["sentinel_replica"] != 3 {
		t.Errorf("expected first sentinel_replica 3, got %v", first["sentinel_replica"])
	}

	second := instances[1].(map[string]interface{})
	if second["id"] != "rs-002" {
		t.Errorf("expected second id rs-002, got %v", second["id"])
	}
	if second["name"] != "prod-redis-sentinel" {
		t.Errorf("expected second name prod-redis-sentinel, got %v", second["name"])
	}
}

func TestDataSourceRedisSentinelInstanceRead_Error(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/database/redis-sentinel-instances/rs-err",
			Handler: testhelpers.EmptyHandler(http.StatusInternalServerError),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceDatabaseRedisSentinelInstance()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"id": "rs-err",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if !diags.HasError() {
		t.Fatal("expected error on 500 response, got none")
	}
}

func TestDataSourceRedisSentinelInstancesRead_Error(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/database/redis-sentinel-instances",
			Handler: testhelpers.EmptyHandler(http.StatusInternalServerError),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceDatabaseRedisSentinelInstances()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if !diags.HasError() {
		t.Fatal("expected error on 500 response, got none")
	}
}

func TestDataSourceRedisSentinelInstancesRead_Empty(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/database/redis-sentinel-instances",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListRedisSentinelInstancesResponse{
				RedisSentinelInstances: []dto.RedisSentinelInstance{},
				Total:                  0,
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceDatabaseRedisSentinelInstances()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	instances := d.Get("redis_sentinel_instances").([]interface{})
	if len(instances) != 0 {
		t.Errorf("expected 0 instances, got %d", len(instances))
	}
}
