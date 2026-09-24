package backupvault

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func testBackupVault() dto.BackupVault {
	return dto.BackupVault{
		ID:                "bv-001",
		Name:              "test-vault",
		Purpose:           "backup_server",
		Type:              "ceph",
		ZoneID:            testhelpers.TestZoneID,
		StorageLocationID: "sl-001",
		Description:       "a test backup vault",
		DiskUsed:          10,
		Quota:             100,
		CommittedQuota:    50,
		ObjectLock:        false,
		Status:            "active",
		CreatedAt:         "2026-01-15T10:00:00Z",
		ProjectID:         testhelpers.TestProjectID,
	}
}

func TestBackupVaultDecodeInt64AsString(t *testing.T) {
	body := `{"backupVault":{"id":"bv-001","type":"ceph","diskUsed":"12","quota":"0","committedQuota":"9223372036854775807","lockTimeNumber":7}}`

	var resp dto.BackupVaultResponse
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	v := resp.BackupVault
	if v.DiskUsed != 12 || v.Quota != 0 || v.CommittedQuota != math.MaxInt64 || v.LockTimeNumber != 7 {
		t.Errorf("unexpected values: disk_used=%d quota=%d committed_quota=%d lock_time_number=%d",
			v.DiskUsed, v.Quota, v.CommittedQuota, v.LockTimeNumber)
	}
}

func TestResourceBackupVaultSchema(t *testing.T) {
	res := ResourceBackupVault()

	for _, key := range []string{"purpose", "type"} {
		if !res.Schema[key].Required || !res.Schema[key].ForceNew {
			t.Errorf("expected %s to be Required+ForceNew", key)
		}
	}
	if !res.Schema["name"].Required {
		t.Error("expected name to be Required")
	}
	if res.Schema["name"].ForceNew {
		t.Error("expected name to be updatable (not ForceNew)")
	}
	for _, key := range []string{"zone_id", "storage_location_id", "disk_used", "quota", "committed_quota", "status", "created_at",
		"object_lock", "lock_time_unit", "lock_time_number"} {
		if !res.Schema[key].Computed || res.Schema[key].Optional {
			t.Errorf("expected %s to be Computed only", key)
		}
	}
}

func TestResourceBackupVaultSchemaValidation(t *testing.T) {
	res := ResourceBackupVault()

	tests := []struct {
		field   string
		value   interface{}
		wantErr bool
	}{
		{"purpose", "backup_server", false},
		{"purpose", "workload_cluster", false},
		{"purpose", "bogus", true},
		{"type", "ceph", false},
		{"type", "s3", false},
		{"type", "bogus", true},
		{"name", "ab", true},
		{"name", "valid-name.1_x", false},
		{"name", "bad name!", true},
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

func TestResourceBackupVaultRead(t *testing.T) {
	vault := testBackupVault()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.BackupVaultWithID(testhelpers.TestProjectID, "bv-001"),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.BackupVaultResponse{BackupVault: vault}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceBackupVault()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{})
	d.SetId("bv-001")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if v := d.Get("name").(string); v != "test-vault" {
		t.Errorf("expected name test-vault, got %s", v)
	}
	if v := d.Get("purpose").(string); v != "backup_server" {
		t.Errorf("expected purpose backup_server, got %s", v)
	}
	if v := d.Get("type").(string); v != "ceph" {
		t.Errorf("expected type ceph, got %s", v)
	}
	if v := d.Get("quota").(int); v != 100 {
		t.Errorf("expected quota 100, got %d", v)
	}
	if v := d.Get("status").(string); v != "active" {
		t.Errorf("expected status active, got %s", v)
	}
}

func TestResourceBackupVaultReadUnlimitedCommittedQuota(t *testing.T) {
	vault := testBackupVault()
	vault.CommittedQuota = math.MaxInt64

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.BackupVaultWithID(testhelpers.TestProjectID, "bv-001"),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.BackupVaultResponse{BackupVault: vault}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceBackupVault()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{})
	d.SetId("bv-001")

	if diags := res.ReadContext(context.Background(), d, cfg); diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
	if v := d.Get("committed_quota").(int); v != -1 {
		t.Errorf("expected committed_quota -1 for unlimited, got %d", v)
	}
}

func TestResourceBackupVaultReadNotFound(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.BackupVaultWithID(testhelpers.TestProjectID, "bv-404"),
			Handler: testhelpers.JSONHandler(t, http.StatusNotFound, map[string]string{"message": "not found"}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceBackupVault()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{})
	d.SetId("bv-404")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("expected 404 to clear the ID, got error: %v", diags)
	}
	if d.Id() != "" {
		t.Errorf("expected ID to be cleared on 404, got %s", d.Id())
	}
}
