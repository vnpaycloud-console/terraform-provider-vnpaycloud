package backupserver

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func testBackupServer() dto.BackupServer {
	return dto.BackupServer{
		ID:             "bs-001",
		Name:           "test-server-backup",
		ServerID:       "srv-001",
		VolumeIDs:      []string{"vol-001", "vol-002"},
		BackupPolicyID: "bps-001",
		Purpose:        "on_demand",
		ZoneID:         testhelpers.TestZoneID,
		IsCompliant:    "true",
		Status:         "active",
		CreatedAt:      "2026-01-15T10:00:00Z",
		ProjectID:      testhelpers.TestProjectID,
	}
}

func TestResourceBackupServerSchema(t *testing.T) {
	res := ResourceBackupServer()

	for _, key := range []string{"server_id"} {
		if !res.Schema[key].Required || !res.Schema[key].ForceNew {
			t.Errorf("expected %s to be Required+ForceNew", key)
		}
	}
	for _, key := range []string{"volume_ids", "backup_policy_id"} {
		if !res.Schema[key].Required {
			t.Errorf("expected %s to be Required", key)
		}
		if res.Schema[key].ForceNew {
			t.Errorf("expected %s to be updatable (not ForceNew)", key)
		}
	}
	if res.Schema["volume_ids"].MinItems != 1 {
		t.Error("expected volume_ids to require at least one element")
	}
	for _, key := range []string{"name", "purpose", "zone_id", "is_compliant", "compliance_msg", "status", "created_at"} {
		if !res.Schema[key].Computed {
			t.Errorf("expected %s to be Computed", key)
		}
	}
}

func TestResourceBackupServerSchemaValidation(t *testing.T) {
	res := ResourceBackupServer()

	tests := []struct {
		field   string
		value   interface{}
		wantErr bool
	}{
		{"server_id", "srv-001", false},
		{"server_id", "ab", true},
		{"server_id", "has space", true},
		{"backup_policy_id", "bps-001", false},
		{"backup_policy_id", "bad!", true},
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

	volumeElem := res.Schema["volume_ids"].Elem.(*schema.Schema)
	if _, errs := volumeElem.ValidateFunc("vol-001", "volume_ids"); len(errs) > 0 {
		t.Errorf("volume_ids element vol-001 should be accepted: %v", errs)
	}
	if _, errs := volumeElem.ValidateFunc("ab", "volume_ids"); len(errs) == 0 {
		t.Error("volume_ids element ab should be rejected (min 3 chars)")
	}
}

func TestResourceBackupServerRead(t *testing.T) {
	server := testBackupServer()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.BackupServerWithID(testhelpers.TestProjectID, "bs-001"),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.BackupServerResponse{BackupServer: server}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceBackupServer()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{})
	d.SetId("bs-001")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if v := d.Get("name").(string); v != "test-server-backup" {
		t.Errorf("expected name test-server-backup, got %s", v)
	}
	if v := d.Get("server_id").(string); v != "srv-001" {
		t.Errorf("expected server_id srv-001, got %s", v)
	}
	if v := d.Get("volume_ids.#").(int); v != 2 {
		t.Errorf("expected 2 volume_ids, got %d", v)
	}
	if v := d.Get("volume_ids.1").(string); v != "vol-002" {
		t.Errorf("expected volume_ids.1 vol-002, got %s", v)
	}
	if v := d.Get("purpose").(string); v != "on_demand" {
		t.Errorf("expected type on_demand, got %s", v)
	}
	if v := d.Get("status").(string); v != "active" {
		t.Errorf("expected status active, got %s", v)
	}
}

func TestResourceBackupServerReadNotFound(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.BackupServerWithID(testhelpers.TestProjectID, "bs-404"),
			Handler: testhelpers.JSONHandler(t, http.StatusNotFound, map[string]string{"message": "not found"}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceBackupServer()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{})
	d.SetId("bs-404")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("expected 404 to clear the ID, got error: %v", diags)
	}
	if d.Id() != "" {
		t.Errorf("expected ID to be cleared on 404, got %s", d.Id())
	}
}
