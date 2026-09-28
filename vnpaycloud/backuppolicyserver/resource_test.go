package backuppolicyserver

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestResourceBackupPolicyServerMonthlyDayTypeValidation(t *testing.T) {
	res := ResourceBackupPolicyServer()

	cases := []struct {
		name    string
		block   string // "monthly" or "second_tier_monthly"
		m       map[string]interface{}
		wantErr bool
	}{
		{"monthly day_of_month missing dom", "monthly", map[string]interface{}{"retentions": 3, "day_type": "day_of_month"}, true},
		{"monthly day_of_month with dom", "monthly", map[string]interface{}{"retentions": 3, "day_type": "day_of_month", "day_of_month": 15}, false},
		{"monthly day_of_month stray dow", "monthly", map[string]interface{}{"retentions": 3, "day_type": "day_of_month", "day_of_month": 15, "day_of_week": 3}, true},
		{"monthly week type missing dow (tier1 1-7)", "monthly", map[string]interface{}{"retentions": 3, "day_type": "day_of_2nd_week"}, true},
		{"monthly week type with dow", "monthly", map[string]interface{}{"retentions": 3, "day_type": "day_of_2nd_week", "day_of_week": 3}, false},
		{"tier2 day_of_month missing dom", "second_tier_monthly", map[string]interface{}{"retentions": 2, "day_type": "day_of_month"}, true},
		{"tier2 week type dow=0 ok", "second_tier_monthly", map[string]interface{}{"retentions": 2, "day_type": "day_of_2nd_week"}, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			raw := map[string]interface{}{
				"name":          "tf-bps",
				"start_hour":    3,
				"resource_type": "vm",
				"daily":         []interface{}{map[string]interface{}{"retentions": 7}},
				tc.block:        []interface{}{tc.m},
			}
			if tc.block == "second_tier_monthly" {
				raw["second_tier_enabled"] = true
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

func testBackupPolicyServer() dto.BackupPolicyServer {
	return dto.BackupPolicyServer{
		ID:                   "bps-001",
		Name:                 "test-policy",
		Description:          "a test backup policy",
		StartHour:            2,
		ResourceType:         "vm",
		Purpose:              "on_demand",
		IsAuto:               true,
		IsAutoApplyForVolume: false,
		BackupVaultIDs:       []string{"bv-001"},
		Daily:                &dto.BackupPolicyServerDaily{Retentions: 7},
		Weekly:               &dto.BackupPolicyServerWeekly{Retentions: 4, DayOfWeek: 1},
		Monthly: &dto.BackupPolicyServerMonthly{
			Retentions: 3,
			DayType:    "day_of_month",
			DayOfMonth: 15,
		},
		SecondTierEnabled: true,
		SecondTierWeekly:  &dto.BackupPolicyServerSecondTierWeekly{Retentions: 2, DayOfWeek: 0},
		SecondTierYearly:  &dto.BackupPolicyServerSecondTierYearly{Retentions: 1, Month: 12, DayOfMonth: 31},
		ProtectedServers:  5,
		ComplianceState:   "compliant",
		Status:            "active",
		CreatedAt:         "2026-01-15T10:00:00Z",
		ProjectID:         testhelpers.TestProjectID,
	}
}

func TestResourceBackupPolicyServerSchema(t *testing.T) {
	res := ResourceBackupPolicyServer()

	for _, key := range []string{"name", "resource_type"} {
		if !res.Schema[key].Required || !res.Schema[key].ForceNew {
			t.Errorf("expected %s to be Required+ForceNew", key)
		}
	}
	if p := res.Schema["purpose"]; !p.Computed || p.Required || p.Optional {
		t.Error("expected purpose to be Computed-only")
	}
	if !res.Schema["daily"].Required {
		t.Error("expected daily to be Required")
	}
	if !res.Schema["start_hour"].Required || res.Schema["start_hour"].ForceNew {
		t.Error("expected start_hour to be Required and updatable")
	}
	for _, key := range []string{"protected_servers", "compliance_state", "status", "created_at"} {
		if !res.Schema[key].Computed {
			t.Errorf("expected %s to be Computed", key)
		}
	}
}

func TestResourceBackupPolicyServerSchemaValidation(t *testing.T) {
	res := ResourceBackupPolicyServer()

	tests := []struct {
		field   string
		value   interface{}
		wantErr bool
	}{
		{"resource_type", "vm", false},
		{"resource_type", "volume", false},
		{"resource_type", "bogus", true},
		{"start_hour", 0, false},
		{"start_hour", 23, false},
		{"start_hour", 24, true},
		{"start_hour", -1, true},
		{"name", "ab", true},
		{"name", "valid name-1_x.y", false},
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

func TestResourceBackupPolicyServerNestedValidation(t *testing.T) {
	res := ResourceBackupPolicyServer()

	daily := res.Schema["daily"].Elem.(*schema.Resource).Schema
	if _, errs := daily["retentions"].ValidateFunc(6, "retentions"); len(errs) == 0 {
		t.Error("daily.retentions=6 should be rejected (min 7)")
	}
	if _, errs := daily["retentions"].ValidateFunc(7, "retentions"); len(errs) > 0 {
		t.Errorf("daily.retentions=7 should be accepted: %v", errs)
	}

	monthly := res.Schema["monthly"].Elem.(*schema.Resource).Schema
	if _, errs := monthly["day_type"].ValidateFunc("day_of_6th_week", "day_type"); len(errs) == 0 {
		t.Error("monthly.day_type=day_of_6th_week should be rejected")
	}
	if _, errs := monthly["day_of_week"].ValidateFunc(0, "day_of_week"); len(errs) == 0 {
		t.Error("monthly.day_of_week=0 should be rejected (first tier is 1-7)")
	}

	stWeekly := res.Schema["second_tier_weekly"].Elem.(*schema.Resource).Schema
	if _, errs := stWeekly["day_of_week"].ValidateFunc(0, "day_of_week"); len(errs) > 0 {
		t.Errorf("second_tier_weekly.day_of_week=0 should be accepted (second tier is 0-6): %v", errs)
	}
	if _, errs := stWeekly["day_of_week"].ValidateFunc(7, "day_of_week"); len(errs) == 0 {
		t.Error("second_tier_weekly.day_of_week=7 should be rejected (second tier is 0-6)")
	}
}

func TestResourceBackupPolicyServerCreateAlwaysSendsOnDemand(t *testing.T) {
	var gotPurpose string

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "POST",
			Pattern: client.ApiPath.BackupPolicyServers(testhelpers.TestProjectID),
			Handler: func(w http.ResponseWriter, r *http.Request) {
				var body dto.CreateBackupPolicyServerRequest
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatalf("failed to decode create body: %v", err)
				}
				gotPurpose = body.Purpose
				testhelpers.JSONHandler(t, http.StatusOK,
					dto.BackupPolicyServerResponse{BackupPolicyServer: testBackupPolicyServer()})(w, r)
			},
		},
		{
			Method:  "GET",
			Pattern: client.ApiPath.BackupPolicyServerWithID(testhelpers.TestProjectID, "bps-001"),
			Handler: testhelpers.JSONHandler(t, http.StatusOK,
				dto.BackupPolicyServerResponse{BackupPolicyServer: testBackupPolicyServer()}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceBackupPolicyServer()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":          "test-policy",
		"start_hour":    2,
		"resource_type": "vm",
		"daily":         []interface{}{map[string]interface{}{"retentions": 7}},
	})

	diags := res.CreateContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if gotPurpose != "on_demand" {
		t.Errorf("expected create to send purpose on_demand, got %q", gotPurpose)
	}
	if v := d.Get("purpose").(string); v != "on_demand" {
		t.Errorf("expected purpose on_demand in state, got %s", v)
	}
}

func TestResourceBackupPolicyServerRead(t *testing.T) {
	policy := testBackupPolicyServer()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.BackupPolicyServerWithID(testhelpers.TestProjectID, "bps-001"),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.BackupPolicyServerResponse{BackupPolicyServer: policy}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceBackupPolicyServer()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{})
	d.SetId("bps-001")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if v := d.Get("name").(string); v != "test-policy" {
		t.Errorf("expected name test-policy, got %s", v)
	}
	if v := d.Get("resource_type").(string); v != "vm" {
		t.Errorf("expected resource_type vm, got %s", v)
	}
	if v := d.Get("start_hour").(int); v != 2 {
		t.Errorf("expected start_hour 2, got %d", v)
	}
	if v := d.Get("daily.0.retentions").(int); v != 7 {
		t.Errorf("expected daily.retentions 7, got %d", v)
	}
	if v := d.Get("monthly.0.day_type").(string); v != "day_of_month" {
		t.Errorf("expected monthly.day_type day_of_month, got %s", v)
	}
	if v := d.Get("second_tier_yearly.0.month").(int); v != 12 {
		t.Errorf("expected second_tier_yearly.month 12, got %d", v)
	}
	if v := d.Get("status").(string); v != "active" {
		t.Errorf("expected status active, got %s", v)
	}
}

func TestResourceBackupPolicyServerReadNotFound(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.BackupPolicyServerWithID(testhelpers.TestProjectID, "bps-404"),
			Handler: testhelpers.JSONHandler(t, http.StatusNotFound, map[string]string{"message": "not found"}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceBackupPolicyServer()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{})
	d.SetId("bps-404")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("expected 404 to clear the ID, got error: %v", diags)
	}
	if d.Id() != "" {
		t.Errorf("expected ID to be cleared on 404, got %s", d.Id())
	}
}
