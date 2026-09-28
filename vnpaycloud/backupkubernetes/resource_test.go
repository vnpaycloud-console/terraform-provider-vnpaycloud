package backupkubernetes

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func testBackupKubernetes() dto.BackupKubernetes {
	return dto.BackupKubernetes{
		ID:                 "cbk-001",
		Name:               "prod-cluster",
		Description:        "full cluster backup",
		ProtectedClusterID: "clu-001",
		PVCBackupPolicyID:  "kpol-001",
		ClusterBackupType:  "on_demand",
		ZoneID:             testhelpers.TestZoneID,
		ProjectID:          testhelpers.TestProjectID,
		CustomerUsername:   "duytq1",
		Status:             "active",
		CreatedAt:          "2026-08-26T10:00:00Z",
	}
}

func TestResourceBackupKubernetesSchema(t *testing.T) {
	res := ResourceBackupKubernetes()

	for _, key := range []string{"protected_cluster_id", "pvc_backup_policy_id"} {
		if !res.Schema[key].Required || !res.Schema[key].ForceNew {
			t.Errorf("expected %s to be Required+ForceNew", key)
		}
	}
	if res.Schema["description"].Required || res.Schema["description"].ForceNew {
		t.Error("expected description to be Optional and updatable")
	}
	for _, key := range []string{"name", "cluster_backup_type", "zone_id", "project_id", "customer_username", "status", "created_at"} {
		if !res.Schema[key].Computed {
			t.Errorf("expected %s to be Computed", key)
		}
	}
}

func TestResourceBackupKubernetesSchemaValidation(t *testing.T) {
	res := ResourceBackupKubernetes()

	tests := []struct {
		field   string
		value   interface{}
		wantErr bool
	}{
		{"protected_cluster_id", "clu-001", false},
		{"protected_cluster_id", "ab", true},
		{"protected_cluster_id", "has space", true},
		{"pvc_backup_policy_id", "kpol-001", false},
		{"pvc_backup_policy_id", "bad!", true},
	}

	for _, tt := range tests {
		_, errs := res.Schema[tt.field].ValidateFunc(tt.value, tt.field)
		if tt.wantErr && len(errs) == 0 {
			t.Errorf("%s=%v: expected validation error, got none", tt.field, tt.value)
		}
		if !tt.wantErr && len(errs) > 0 {
			t.Errorf("%s=%v: unexpected validation error: %v", tt.field, tt.value, errs)
		}
	}
}

func TestResourceBackupKubernetesCreate(t *testing.T) {
	backup := testBackupKubernetes()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "POST",
			Pattern: client.ApiPath.BackupKuberneteses(testhelpers.TestProjectID),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.BackupKubernetesResponse{BackupKubernetes: backup}),
		},
		{
			Method:  "GET",
			Pattern: client.ApiPath.BackupKubernetesWithID(testhelpers.TestProjectID, "cbk-001"),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.BackupKubernetesResponse{BackupKubernetes: backup}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceBackupKubernetes()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"protected_cluster_id": "clu-001",
		"pvc_backup_policy_id": "kpol-001",
		"description":          "full cluster backup",
	})

	diags := res.CreateContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
	if d.Id() != "cbk-001" {
		t.Errorf("expected ID cbk-001, got %s", d.Id())
	}
	if v := d.Get("status").(string); v != "active" {
		t.Errorf("expected status active, got %s", v)
	}
	if v := d.Get("cluster_backup_type").(string); v != "on_demand" {
		t.Errorf("expected cluster_backup_type on_demand, got %s", v)
	}
}

func TestResourceBackupKubernetesRead(t *testing.T) {
	backup := testBackupKubernetes()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.BackupKubernetesWithID(testhelpers.TestProjectID, "cbk-001"),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.BackupKubernetesResponse{BackupKubernetes: backup}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceBackupKubernetes()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{})
	d.SetId("cbk-001")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
	if v := d.Get("protected_cluster_id").(string); v != "clu-001" {
		t.Errorf("expected protected_cluster_id clu-001, got %s", v)
	}
	if v := d.Get("pvc_backup_policy_id").(string); v != "kpol-001" {
		t.Errorf("expected pvc_backup_policy_id kpol-001, got %s", v)
	}
	if v := d.Get("name").(string); v != "prod-cluster" {
		t.Errorf("expected name prod-cluster, got %s", v)
	}
}

func TestResourceBackupKubernetesReadNotFound(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.BackupKubernetesWithID(testhelpers.TestProjectID, "cbk-404"),
			Handler: testhelpers.JSONHandler(t, http.StatusNotFound, map[string]string{"message": "not found"}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceBackupKubernetes()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{})
	d.SetId("cbk-404")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("expected 404 to clear the ID, got error: %v", diags)
	}
	if d.Id() != "" {
		t.Errorf("expected ID to be cleared on 404, got %s", d.Id())
	}
}
