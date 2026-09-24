package backuppolicykubernetes

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testBackupPolicyKubernetes() dto.BackupPolicyKubernetes {
	return dto.BackupPolicyKubernetes{
		ID:                 "bpk-001",
		Name:               "test-k8s-policy",
		Description:        "a test pvc backup policy",
		AutoApplyNewVolume: true,
		IsAuto:             true,
		StartHour:          3,
		RunPriority:        1,
		Purpose:            "on_demand",
		BackupVaultIDs:     []string{"bv-k8s-001"},
		Daily:              &dto.BackupPolicyKubernetesDaily{Retentions: 7},
		Weekly:             &dto.BackupPolicyKubernetesWeekly{Retentions: 4, DayOfWeek: 0},
		Monthly: &dto.BackupPolicyKubernetesMonthly{
			Retentions: 3,
			DayType:    "fixed_day",
			DayOfMonth: 15,
		},
		Status:    "active",
		CreatedAt: "2026-01-15T10:00:00Z",
		ProjectID: testhelpers.TestProjectID,
	}
}

func TestResourceBackupPolicyKubernetesSchema(t *testing.T) {
	res := ResourceBackupPolicyKubernetes()

	for _, key := range []string{"name", "backup_vault_ids"} {
		if !res.Schema[key].Required || !res.Schema[key].ForceNew {
			t.Errorf("expected %s to be Required+ForceNew", key)
		}
	}
	if p := res.Schema["purpose"]; !p.Computed || p.Required || p.Optional {
		t.Error("expected purpose to be Computed-only")
	}
	if !res.Schema["daily"].Optional || res.Schema["daily"].Required {
		t.Error("expected daily to be Optional (k8s policy needs at least one schedule, not daily specifically)")
	}
	if sh := res.Schema["start_hour"]; !sh.Optional || !sh.Computed || sh.ForceNew {
		t.Error("expected start_hour to be Optional+Computed and updatable")
	}
	if rp := res.Schema["run_priority"]; !rp.Optional || !rp.Computed {
		t.Error("expected run_priority to be Optional+Computed")
	}
	for _, key := range []string{"status", "created_at"} {
		if !res.Schema[key].Computed {
			t.Errorf("expected %s to be Computed", key)
		}
	}
}

func TestResourceBackupPolicyKubernetesSchemaValidation(t *testing.T) {
	res := ResourceBackupPolicyKubernetes()

	tests := []struct {
		field   string
		value   interface{}
		wantErr bool
	}{
		{"start_hour", 0, false},
		{"start_hour", 23, false},
		{"start_hour", 24, true},
		{"start_hour", -1, true},
		{"name", "ab", true},
		{"name", "valid-name_1.2", false},
		{"name", "bad name", true},
		{"name", "bad!name", true},
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

func TestResourceBackupPolicyKubernetesNestedValidation(t *testing.T) {
	res := ResourceBackupPolicyKubernetes()

	daily := res.Schema["daily"].Elem.(*schema.Resource).Schema
	if _, errs := daily["retentions"].ValidateFunc(6, "retentions"); len(errs) == 0 {
		t.Error("daily.retentions=6 should be rejected (min 7)")
	}
	if _, errs := daily["retentions"].ValidateFunc(7, "retentions"); len(errs) > 0 {
		t.Errorf("daily.retentions=7 should be accepted: %v", errs)
	}

	weekly := res.Schema["weekly"].Elem.(*schema.Resource).Schema
	if _, errs := weekly["day_of_week"].ValidateFunc(0, "day_of_week"); len(errs) > 0 {
		t.Errorf("weekly.day_of_week=0 should be accepted (0-6): %v", errs)
	}
	if _, errs := weekly["day_of_week"].ValidateFunc(7, "day_of_week"); len(errs) == 0 {
		t.Error("weekly.day_of_week=7 should be rejected (0-6)")
	}

	monthly := res.Schema["monthly"].Elem.(*schema.Resource).Schema
	if _, errs := monthly["day_type"].ValidateFunc("sixth_week", "day_type"); len(errs) == 0 {
		t.Error("monthly.day_type=sixth_week should be rejected")
	}
	if _, errs := monthly["day_type"].ValidateFunc("fixed_day", "day_type"); len(errs) > 0 {
		t.Errorf("monthly.day_type=fixed_day should be accepted: %v", errs)
	}
}

func TestResourceBackupPolicyKubernetesMonthlyDayTypeValidation(t *testing.T) {
	res := ResourceBackupPolicyKubernetes()

	cases := []struct {
		name    string
		monthly map[string]interface{}
		wantErr bool
	}{
		{"fixed_day without day_of_month", map[string]interface{}{"retentions": 3, "day_type": "fixed_day"}, true},
		{"fixed_day with day_of_month", map[string]interface{}{"retentions": 3, "day_type": "fixed_day", "day_of_month": 15}, false},
		{"fixed_day with stray day_of_week", map[string]interface{}{"retentions": 3, "day_type": "fixed_day", "day_of_month": 15, "day_of_week": 2}, true},
		{"second_week with day_of_week", map[string]interface{}{"retentions": 3, "day_type": "second_week", "day_of_week": 1}, false},
		{"second_week default day_of_week=0", map[string]interface{}{"retentions": 3, "day_type": "second_week"}, false},
		{"second_week with stray day_of_month", map[string]interface{}{"retentions": 3, "day_type": "second_week", "day_of_month": 10}, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			raw := map[string]interface{}{
				"name":             "test-k8s-policy",
				"backup_vault_ids": []interface{}{"bv-k8s-001"},
				"monthly":          []interface{}{tc.monthly},
			}
			_, err := res.Diff(context.Background(), nil, terraform.NewResourceConfigRaw(raw), nil)
			if tc.wantErr && err == nil {
				t.Fatal("expected a validation error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected validation error: %v", err)
			}
		})
	}
}

func TestResourceBackupPolicyKubernetesCreate(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "POST",
			Pattern: client.ApiPath.BackupPolicyKuberneteses(testhelpers.TestProjectID),
			Handler: testhelpers.JSONHandler(t, http.StatusOK,
				dto.BackupPolicyKubernetesResponse{BackupPolicyKubernetes: testBackupPolicyKubernetes()}),
		},
		{
			Method:  "GET",
			Pattern: client.ApiPath.BackupPolicyKubernetesWithID(testhelpers.TestProjectID, "bpk-001"),
			Handler: testhelpers.JSONHandler(t, http.StatusOK,
				dto.BackupPolicyKubernetesResponse{BackupPolicyKubernetes: testBackupPolicyKubernetes()}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceBackupPolicyKubernetes()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":             "test-k8s-policy",
		"start_hour":       3,
		"backup_vault_ids": []interface{}{"bv-k8s-001"},
		"daily":            []interface{}{map[string]interface{}{"retentions": 7}},
	})

	diags := res.CreateContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "bpk-001" {
		t.Errorf("expected id bpk-001, got %s", d.Id())
	}
	if v := d.Get("purpose").(string); v != "on_demand" {
		t.Errorf("expected purpose on_demand in state, got %s", v)
	}
	if v := d.Get("auto_apply_new_volume").(bool); !v {
		t.Error("expected auto_apply_new_volume true")
	}
	if v := d.Get("daily.0.retentions").(int); v != 7 {
		t.Errorf("expected daily.retentions 7, got %d", v)
	}
}

func TestResourceBackupPolicyKubernetesRead(t *testing.T) {
	policy := testBackupPolicyKubernetes()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.BackupPolicyKubernetesWithID(testhelpers.TestProjectID, "bpk-001"),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.BackupPolicyKubernetesResponse{BackupPolicyKubernetes: policy}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceBackupPolicyKubernetes()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{})
	d.SetId("bpk-001")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if v := d.Get("name").(string); v != "test-k8s-policy" {
		t.Errorf("expected name test-k8s-policy, got %s", v)
	}
	if v := d.Get("start_hour").(int); v != 3 {
		t.Errorf("expected start_hour 3, got %d", v)
	}
	if v := d.Get("run_priority").(int); v != 1 {
		t.Errorf("expected run_priority 1, got %d", v)
	}
	if v := d.Get("monthly.0.day_type").(string); v != "fixed_day" {
		t.Errorf("expected monthly.day_type fixed_day, got %s", v)
	}
	if v := d.Get("weekly.0.day_of_week").(int); v != 0 {
		t.Errorf("expected weekly.day_of_week 0, got %d", v)
	}
	if v := d.Get("status").(string); v != "active" {
		t.Errorf("expected status active, got %s", v)
	}
}

func TestResourceBackupPolicyKubernetesReadNotFound(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.BackupPolicyKubernetesWithID(testhelpers.TestProjectID, "bpk-404"),
			Handler: testhelpers.JSONHandler(t, http.StatusNotFound, map[string]string{"message": "not found"}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceBackupPolicyKubernetes()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{})
	d.SetId("bpk-404")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("expected 404 to clear the ID, got error: %v", diags)
	}
	if d.Id() != "" {
		t.Errorf("expected ID to be cleared on 404, got %s", d.Id())
	}
}
