package backupserver

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

func TestDataSourceBackupServerRead_ByID(t *testing.T) {
	server := testBackupServer()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.BackupServerWithID(testhelpers.TestProjectID, "bs-001"),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.BackupServerResponse{BackupServer: server}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceBackupServer()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{"id": "bs-001"})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
	if d.Id() != "bs-001" {
		t.Errorf("expected ID bs-001, got %s", d.Id())
	}
	if v := d.Get("backup_policy_id").(string); v != "bps-001" {
		t.Errorf("expected backup_policy_id bps-001, got %s", v)
	}
}

func TestDataSourceBackupServerRead_ByIDNameMismatch(t *testing.T) {
	server := testBackupServer()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.BackupServerWithID(testhelpers.TestProjectID, "bs-001"),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.BackupServerResponse{BackupServer: server}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceBackupServer()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"id":   "bs-001",
		"name": "other-backup",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if !diags.HasError() {
		t.Fatal("expected an error when id and name refer to different records")
	}
	if !strings.Contains(diags[0].Summary, "does not match name") {
		t.Errorf("unexpected error: %v", diags[0].Summary)
	}
}

func TestDataSourceBackupServerRead_ByName(t *testing.T) {
	server := testBackupServer()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.BackupServers(testhelpers.TestProjectID),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListBackupServersResponse{
				BackupServers: []dto.BackupServer{server},
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceBackupServer()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{"name": "test-server-backup"})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
	if d.Id() != "bs-001" {
		t.Errorf("expected ID bs-001, got %s", d.Id())
	}
}

func TestDataSourceBackupServerRead_ByNameNotFound(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.BackupServers(testhelpers.TestProjectID),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListBackupServersResponse{}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceBackupServer()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{"name": "missing"})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if !diags.HasError() {
		t.Fatal("expected an error when no record matches the name")
	}
}

func TestDataSourceBackupServersRead(t *testing.T) {
	first := testBackupServer()
	second := testBackupServer()
	second.ID = "bs-002"
	second.Name = "second-backup"
	second.Purpose = "on_demand"

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.BackupServers(testhelpers.TestProjectID),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListBackupServersResponse{
				BackupServers: []dto.BackupServer{first, second},
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceBackupServers()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	servers := d.Get("backup_servers").([]interface{})
	if len(servers) != 2 {
		t.Fatalf("expected 2 backup servers, got %d", len(servers))
	}
	if v := d.Get("backup_servers.1.name").(string); v != "second-backup" {
		t.Errorf("expected second record name second-backup, got %s", v)
	}
	if v := d.Get("backup_servers.0.volume_ids.#").(int); v != 2 {
		t.Errorf("expected 2 volume_ids on first record, got %d", v)
	}
}

func TestDataSourceBackupServersPurposeValidation(t *testing.T) {
	ds := DataSourceBackupServers()

	if _, errs := ds.Schema["purpose"].ValidateFunc("on_demand", "purpose"); len(errs) > 0 {
		t.Errorf("purpose=on_demand should be accepted: %v", errs)
	}
	for _, invalid := range []string{"disaster", "bogus"} {
		if _, errs := ds.Schema["purpose"].ValidateFunc(invalid, "purpose"); len(errs) == 0 {
			t.Errorf("purpose=%s should be rejected", invalid)
		}
	}
}
