package backupkubernetesrestorepoint

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func testRestorePoint() dto.BackupKubernetesRestorePoint {
	return dto.BackupKubernetesRestorePoint{
		ID:                "rp-001",
		Name:              "duytq-private-20260826t1700z07",
		ClusterID:         "clu-001",
		ClusterName:       "duytq-private",
		ClusterBackupID:   "cbk-001",
		PVCBackupPolicyID: "pol-001",
		ProjectID:         testhelpers.TestProjectID,
		ZoneID:            testhelpers.TestZoneID,
		BackupVaultIDs:    []string{"bv-001"},
		VeleroBackupName:  "velero-rp-001",
		BackupPoint:       "2026-08-26T17:00:00Z",
		KubernetesVersion: "1.28.2",
		Status:            "active",
	}
}

func TestResourceRestorePointSchema(t *testing.T) {
	res := ResourceBackupKubernetesRestorePoint()

	if !res.Schema["restore_point_id"].Required || !res.Schema["restore_point_id"].ForceNew {
		t.Error("expected restore_point_id to be Required+ForceNew")
	}
	if res.UpdateContext != nil {
		t.Error("expected no UpdateContext (restore point is not updatable)")
	}
	for _, key := range []string{"name", "cluster_id", "cluster_backup_id", "pvc_backup_policy_id", "zone_id", "status", "backup_point"} {
		if !res.Schema[key].Computed {
			t.Errorf("expected %s to be Computed", key)
		}
	}
}

func TestResourceRestorePointCreate(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.BackupKubernetesRestorePointWithID(testhelpers.TestProjectID, "rp-001"),
			Handler: testhelpers.JSONHandler(t, http.StatusOK,
				dto.BackupKubernetesRestorePointResponse{RestorePoint: testRestorePoint()}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceBackupKubernetesRestorePoint()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"restore_point_id": "rp-001",
	})

	diags := res.CreateContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
	if d.Id() != "rp-001" {
		t.Errorf("expected id rp-001, got %s", d.Id())
	}
	if v := d.Get("cluster_name").(string); v != "duytq-private" {
		t.Errorf("expected cluster_name duytq-private, got %s", v)
	}
	if v := d.Get("status").(string); v != "active" {
		t.Errorf("expected status active, got %s", v)
	}
}

func TestResourceRestorePointReadNotFound(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.BackupKubernetesRestorePointWithID(testhelpers.TestProjectID, "rp-404"),
			Handler: testhelpers.JSONHandler(t, http.StatusNotFound, map[string]string{"message": "not found"}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceBackupKubernetesRestorePoint()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{})
	d.SetId("rp-404")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("expected 404 to clear the ID, got error: %v", diags)
	}
	if d.Id() != "" {
		t.Errorf("expected ID cleared on 404, got %s", d.Id())
	}
}

func TestResourceRestorePointDelete(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			// DELETE accepts; the follow-up GET during the delete-wait returns
			// 404, the synthetic "deleted" terminal state the wait targets.
			Pattern: client.ApiPath.BackupKubernetesRestorePointWithID(testhelpers.TestProjectID, "rp-001"),
			Handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodDelete {
					w.WriteHeader(http.StatusOK)
					return
				}
				w.WriteHeader(http.StatusNotFound)
			},
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceBackupKubernetesRestorePoint()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{})
	d.SetId("rp-001")

	diags := res.DeleteContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
}
