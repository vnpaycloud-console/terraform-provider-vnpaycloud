package workergroup

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestDataSourceWorkerGroupRead_ByID(t *testing.T) {
	wg := testWorkerGroup()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/clusters/cluster-001/worker-groups/wg-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.WorkerGroupResponse{WorkerGroup: wg}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceWorkerGroup()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"id":         "wg-001",
		"cluster_id": "cluster-001",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "wg-001" {
		t.Errorf("expected ID wg-001, got %s", d.Id())
	}
	if v := d.Get("cluster_id").(string); v != "cluster-001" {
		t.Errorf("expected cluster_id cluster-001, got %s", v)
	}
	if v := d.Get("name").(string); v != "test-worker-group" {
		t.Errorf("expected name test-worker-group, got %s", v)
	}
	if v := d.Get("flavor").(string); v != "v1.medium" {
		t.Errorf("expected flavor v1.medium, got %s", v)
	}
	if v := d.Get("num_workers").(int); v != 3 {
		t.Errorf("expected num_workers 3, got %d", v)
	}
	if v := d.Get("auto_scaling").(bool); !v {
		t.Error("expected auto_scaling true, got false")
	}
	if v := d.Get("min_workers").(int); v != 1 {
		t.Errorf("expected min_workers 1, got %d", v)
	}
	if v := d.Get("max_workers").(int); v != 5 {
		t.Errorf("expected max_workers 5, got %d", v)
	}
	if v := d.Get("status").(string); v != "active" {
		t.Errorf("expected status active, got %s", v)
	}
	if v := d.Get("created_at").(string); v != "2025-01-15T10:00:00Z" {
		t.Errorf("expected created_at 2025-01-15T10:00:00Z, got %s", v)
	}
}

func TestDataSourceWorkerGroupsRead(t *testing.T) {
	wg1 := testWorkerGroup()
	wg2 := dto.WorkerGroup{
		ID:          "wg-002",
		Name:        "gpu-workers",
		ClusterID:   "cluster-001",
		Flavor:      "v1.gpu",
		NumWorkers:  2,
		MinWorkers:  1,
		MaxWorkers:  4,
		AutoScaling: false,
		Status:      "active",
		CreatedAt:   "2025-02-01T08:00:00Z",
	}

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method: "GET",
			Pattern: "/v2/iac/projects/test-project-id/clusters/cluster-001/worker-groups",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListWorkerGroupsResponse{
				WorkerGroups: []dto.WorkerGroup{wg1, wg2},
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceWorkerGroups()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"cluster_id": "cluster-001",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	workerGroups := d.Get("worker_groups").([]interface{})
	if len(workerGroups) != 2 {
		t.Fatalf("expected 2 worker groups, got %d", len(workerGroups))
	}

	first := workerGroups[0].(map[string]interface{})
	if first["id"] != "wg-001" {
		t.Errorf("expected first worker group id wg-001, got %v", first["id"])
	}
	if first["name"] != "test-worker-group" {
		t.Errorf("expected first worker group name test-worker-group, got %v", first["name"])
	}
	if first["flavor"] != "v1.medium" {
		t.Errorf("expected first worker group flavor v1.medium, got %v", first["flavor"])
	}
	if first["num_workers"] != 3 {
		t.Errorf("expected first worker group num_workers 3, got %v", first["num_workers"])
	}
	if first["auto_scaling"] != true {
		t.Errorf("expected first worker group auto_scaling true, got %v", first["auto_scaling"])
	}
	if first["status"] != "active" {
		t.Errorf("expected first worker group status active, got %v", first["status"])
	}

	second := workerGroups[1].(map[string]interface{})
	if second["id"] != "wg-002" {
		t.Errorf("expected second worker group id wg-002, got %v", second["id"])
	}
	if second["name"] != "gpu-workers" {
		t.Errorf("expected second worker group name gpu-workers, got %v", second["name"])
	}
	if second["flavor"] != "v1.gpu" {
		t.Errorf("expected second worker group flavor v1.gpu, got %v", second["flavor"])
	}
	if second["num_workers"] != 2 {
		t.Errorf("expected second worker group num_workers 2, got %v", second["num_workers"])
	}
}
