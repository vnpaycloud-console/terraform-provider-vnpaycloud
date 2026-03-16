package workergroup

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// testWorkerGroup returns a fully populated dto.WorkerGroup for use in tests.
func testWorkerGroup() dto.WorkerGroup {
	return dto.WorkerGroup{
		ID:          "wg-001",
		Name:        "test-worker-group",
		ClusterID:   "cluster-001",
		Flavor:      "v1.medium",
		NumWorkers:  3,
		MinWorkers:  1,
		MaxWorkers:  5,
		AutoScaling: true,
		Status:      "active",
		CreatedAt:   "2025-01-15T10:00:00Z",
	}
}

func TestResourceWorkerGroupCreate(t *testing.T) {
	wg := testWorkerGroup()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "POST",
			Pattern: "/v2/iac/projects/test-project-id/clusters/cluster-001/worker-groups",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.WorkerGroupResponse{WorkerGroup: wg}),
		},
		{
			Pattern: "/v2/iac/projects/test-project-id/clusters/cluster-001/worker-groups/wg-001",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" {
					http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
					return
				}
				testhelpers.JSONHandler(t, http.StatusOK, dto.WorkerGroupResponse{WorkerGroup: wg})(w, r)
			},
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceWorkerGroup()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"cluster_id":   "cluster-001",
		"name":         "test-worker-group",
		"flavor":       "v1.medium",
		"num_workers":  3,
		"auto_scaling": true,
		"min_workers":  1,
		"max_workers":  5,
		"volume_type":  "",
		"volume_size":  0,
		"ssh_key_id":   "",
	})

	diags := res.CreateContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "wg-001" {
		t.Errorf("expected ID wg-001, got %s", d.Id())
	}
	if v := d.Get("name").(string); v != "test-worker-group" {
		t.Errorf("expected name test-worker-group, got %s", v)
	}
	if v := d.Get("status").(string); v != "active" {
		t.Errorf("expected status active, got %s", v)
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
}

func TestResourceWorkerGroupRead(t *testing.T) {
	wg := testWorkerGroup()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/clusters/cluster-001/worker-groups/wg-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.WorkerGroupResponse{WorkerGroup: wg}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceWorkerGroup()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"cluster_id":   "cluster-001",
		"name":         "",
		"flavor":       "",
		"num_workers":  1,
		"auto_scaling": false,
		"min_workers":  0,
		"max_workers":  0,
		"volume_type":  "",
		"volume_size":  0,
		"ssh_key_id":   "",
	})
	d.SetId("wg-001")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
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

func TestResourceWorkerGroupRead_NotFound(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/clusters/cluster-001/worker-groups/wg-gone",
			Handler: testhelpers.EmptyHandler(http.StatusNotFound),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceWorkerGroup()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"cluster_id":   "cluster-001",
		"name":         "",
		"flavor":       "",
		"num_workers":  1,
		"auto_scaling": false,
		"min_workers":  0,
		"max_workers":  0,
		"volume_type":  "",
		"volume_size":  0,
		"ssh_key_id":   "",
	})
	d.SetId("wg-gone")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error on 404: %v", diags)
	}

	if d.Id() != "" {
		t.Errorf("expected resource ID to be cleared after 404, got %s", d.Id())
	}
}

func TestResourceWorkerGroupDelete(t *testing.T) {
	wg := testWorkerGroup()
	deletedCalled := false

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Pattern: "/v2/iac/projects/test-project-id/clusters/cluster-001/worker-groups/wg-001",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case "GET":
					if deletedCalled {
						w.WriteHeader(http.StatusNotFound)
						return
					}
					testhelpers.JSONHandler(t, http.StatusOK, dto.WorkerGroupResponse{WorkerGroup: wg})(w, r)
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

	res := ResourceWorkerGroup()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"cluster_id":   "cluster-001",
		"name":         "test-worker-group",
		"flavor":       "v1.medium",
		"num_workers":  3,
		"auto_scaling": false,
		"min_workers":  0,
		"max_workers":  0,
		"volume_type":  "",
		"volume_size":  0,
		"ssh_key_id":   "",
	})
	d.SetId("wg-001")

	diags := res.DeleteContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if !deletedCalled {
		t.Error("expected DELETE to have been called")
	}
}
