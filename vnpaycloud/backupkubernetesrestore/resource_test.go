package backupkubernetesrestore

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestResourceRestoreSchema(t *testing.T) {
	res := ResourceBackupKubernetesRestore()

	for _, key := range []string{"restore_point_id", "dest_cluster_id"} {
		if !res.Schema[key].Required || !res.Schema[key].ForceNew {
			t.Errorf("expected %s to be Required+ForceNew", key)
		}
	}
	for _, key := range []string{"restore_type", "transform_lb_to_cluster_ip", "keep_node_port_numbers", "restore_persistent_volumes", "include_cluster_scoped_resource"} {
		if !res.Schema[key].ForceNew {
			t.Errorf("expected %s to be ForceNew", key)
		}
	}
	if res.Schema["restore_persistent_volumes"].Default != true {
		t.Error("expected restore_persistent_volumes default true")
	}
	if res.Schema["restore_type"].Default != "all_resources" {
		t.Error("expected restore_type default all_resources")
	}
}

func TestResourceRestoreStatusIsComputed(t *testing.T) {
	res := ResourceBackupKubernetesRestore()
	s, ok := res.Schema["status"]
	if !ok {
		t.Fatal("expected a status attribute")
	}
	if !s.Computed || s.Optional || s.Required {
		t.Error("expected status to be Computed only")
	}
}

func TestResourceRestoreTypeValidation(t *testing.T) {
	res := ResourceBackupKubernetesRestore()
	if _, errs := res.Schema["restore_type"].ValidateFunc("all_resources", "restore_type"); len(errs) > 0 {
		t.Errorf("all_resources should be valid: %v", errs)
	}
	if _, errs := res.Schema["restore_type"].ValidateFunc("bogus", "restore_type"); len(errs) == 0 {
		t.Error("bogus restore_type should be rejected")
	}
}

func TestResourceRestoreCreate(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "POST",
			Pattern: client.ApiPath.BackupKubernetesRestores(testhelpers.TestProjectID),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.BackupKubernetesRestoreResponse{
				DestClusterID:  "clu-001",
				RestorePointID: "rp-001",
			}),
		},
		{
			Method:  "GET",
			Pattern: client.ApiPath.ClusterWithID(testhelpers.TestProjectID, "clu-001"),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.K8sClusterResponse{
				Cluster: dto.K8sCluster{ID: "clu-001", RestoreStatus: "success"},
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceBackupKubernetesRestore()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"restore_point_id": "rp-001",
		"dest_cluster_id":  "clu-001",
		"restore_type":     "all_resources",
	})

	diags := res.CreateContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
	if d.Id() != "clu-001:rp-001" {
		t.Errorf("expected ID clu-001:rp-001, got %s", d.Id())
	}
	if v := d.Get("status").(string); v != "success" {
		t.Errorf("expected the finished restore status to be recorded as success, got %q", v)
	}
}
