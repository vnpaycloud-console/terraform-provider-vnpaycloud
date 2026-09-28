package backuppolicyserver

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestDataSourceBackupPolicyServerRead_ByID(t *testing.T) {
	policy := testBackupPolicyServer()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.BackupPolicyServerWithID(testhelpers.TestProjectID, "bps-001"),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.BackupPolicyServerResponse{BackupPolicyServer: policy}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceBackupPolicyServer()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"id": "bps-001",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "bps-001" {
		t.Errorf("expected ID bps-001, got %s", d.Id())
	}
	if v := d.Get("name").(string); v != "test-policy" {
		t.Errorf("expected name test-policy, got %s", v)
	}
	if v := d.Get("weekly.0.day_of_week").(int); v != 1 {
		t.Errorf("expected weekly.day_of_week 1, got %d", v)
	}
}

func TestDataSourceBackupPolicyServerRead_ByIDNameMismatch(t *testing.T) {
	policy := testBackupPolicyServer()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.BackupPolicyServerWithID(testhelpers.TestProjectID, "bps-001"),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.BackupPolicyServerResponse{BackupPolicyServer: policy}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceBackupPolicyServer()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"id":   "bps-001",
		"name": "other-policy",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if !diags.HasError() {
		t.Fatal("expected an error when id and name refer to different policies")
	}
	if !strings.Contains(diags[0].Summary, "does not match name") {
		t.Errorf("unexpected error: %v", diags[0].Summary)
	}
}

func TestDataSourceBackupPolicyServerRead_ByName(t *testing.T) {
	policy := testBackupPolicyServer()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.BackupPolicyServers(testhelpers.TestProjectID),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListBackupPolicyServersResponse{
				BackupPolicyServers: []dto.BackupPolicyServer{policy},
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceBackupPolicyServer()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"name": "test-policy",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
	if d.Id() != "bps-001" {
		t.Errorf("expected ID bps-001, got %s", d.Id())
	}
}

func TestDataSourceBackupPolicyServerRead_ByNameNotFound(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.BackupPolicyServers(testhelpers.TestProjectID),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListBackupPolicyServersResponse{}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceBackupPolicyServer()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"name": "missing-policy",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if !diags.HasError() {
		t.Fatal("expected an error when no policy matches the name")
	}
}

func TestDataSourceBackupPolicyServersRead(t *testing.T) {
	first := testBackupPolicyServer()
	second := testBackupPolicyServer()
	second.ID = "bps-002"
	second.Name = "second-policy"
	second.SecondTierEnabled = false
	second.SecondTierWeekly = nil
	second.SecondTierYearly = nil

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.BackupPolicyServers(testhelpers.TestProjectID),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListBackupPolicyServersResponse{
				BackupPolicyServers: []dto.BackupPolicyServer{first, second},
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceBackupPolicyServers()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	policies := d.Get("backup_policy_servers").([]interface{})
	if len(policies) != 2 {
		t.Fatalf("expected 2 policies, got %d", len(policies))
	}
	if v := d.Get("backup_policy_servers.1.name").(string); v != "second-policy" {
		t.Errorf("expected second policy name second-policy, got %s", v)
	}
	if v := d.Get("backup_policy_servers.1.second_tier_yearly.#").(int); v != 0 {
		t.Errorf("expected second policy to have no second_tier_yearly block, got %d", v)
	}
}
