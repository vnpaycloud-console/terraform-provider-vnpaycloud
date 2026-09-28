package backupvault

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

func TestDataSourceBackupVaultRead_ByID(t *testing.T) {
	vault := testBackupVault()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.BackupVaultWithID(testhelpers.TestProjectID, "bv-001"),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.BackupVaultResponse{BackupVault: vault}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceBackupVault()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"id": "bv-001",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "bv-001" {
		t.Errorf("expected ID bv-001, got %s", d.Id())
	}
	if v := d.Get("name").(string); v != "test-vault" {
		t.Errorf("expected name test-vault, got %s", v)
	}
	if v := d.Get("storage_location_id").(string); v != "sl-001" {
		t.Errorf("expected storage_location_id sl-001, got %s", v)
	}
}

func TestDataSourceBackupVaultRead_ByIDNameMatches(t *testing.T) {
	vault := testBackupVault()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.BackupVaultWithID(testhelpers.TestProjectID, "bv-001"),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.BackupVaultResponse{BackupVault: vault}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceBackupVault()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"id":   "bv-001",
		"name": "test-vault",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error when id+name match: %v", diags)
	}
	if d.Id() != "bv-001" {
		t.Errorf("expected ID bv-001, got %s", d.Id())
	}
}

func TestDataSourceBackupVaultRead_ByIDNameMismatch(t *testing.T) {
	vault := testBackupVault()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.BackupVaultWithID(testhelpers.TestProjectID, "bv-001"),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.BackupVaultResponse{BackupVault: vault}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceBackupVault()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"id":   "bv-001",
		"name": "definitely-not-the-real-vault-name",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if !diags.HasError() {
		t.Fatal("expected error when id is fetched but name does not match, got none")
	}
	if !strings.Contains(diags[0].Summary, "does not match name") {
		t.Errorf("expected mismatch error, got: %s", diags[0].Summary)
	}
}

func TestDataSourceBackupVaultRead_ByName(t *testing.T) {
	vault := testBackupVault()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.BackupVaults(testhelpers.TestProjectID),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListBackupVaultsResponse{
				BackupVaults: []dto.BackupVault{vault},
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceBackupVault()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"name": "test-vault",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
	if d.Id() != "bv-001" {
		t.Errorf("expected ID bv-001, got %s", d.Id())
	}
}

func TestDataSourceBackupVaultRead_NotFound(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.BackupVaults(testhelpers.TestProjectID),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListBackupVaultsResponse{
				BackupVaults: []dto.BackupVault{},
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceBackupVault()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"name": "nonexistent",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if !diags.HasError() {
		t.Fatal("expected error for nonexistent backup vault, got none")
	}
}

func TestDataSourceBackupVaultRead_NameAmbiguous(t *testing.T) {
	v1 := testBackupVault()
	v2 := testBackupVault()
	v2.ID = "bv-002"

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.BackupVaults(testhelpers.TestProjectID),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListBackupVaultsResponse{
				BackupVaults: []dto.BackupVault{v1, v2},
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceBackupVault()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"name": "test-vault",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if !diags.HasError() {
		t.Fatal("expected error when name matches multiple vaults, got none")
	}
	if !strings.Contains(diags[0].Summary, "use id to select one") {
		t.Errorf("expected ambiguity error, got: %s", diags[0].Summary)
	}
}

func TestDataSourceBackupVaultsRead(t *testing.T) {
	v1 := testBackupVault()
	v2 := testBackupVault()
	v2.ID = "bv-002"
	v2.Name = "test-vault-2"
	v2.Type = "s3"

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.BackupVaults(testhelpers.TestProjectID),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListBackupVaultsResponse{
				BackupVaults: []dto.BackupVault{v1, v2},
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceBackupVaults()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	vaults := d.Get("backup_vaults").([]interface{})
	if len(vaults) != 2 {
		t.Fatalf("expected 2 backup vaults, got %d", len(vaults))
	}

	first := vaults[0].(map[string]interface{})
	if first["id"] != "bv-001" {
		t.Errorf("expected first vault id bv-001, got %v", first["id"])
	}
	second := vaults[1].(map[string]interface{})
	if second["type"] != "s3" {
		t.Errorf("expected second vault type s3, got %v", second["type"])
	}
}
