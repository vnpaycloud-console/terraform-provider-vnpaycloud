package backuppolicykubernetes

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

func TestDataSourceBackupPolicyKubernetesRead_ByID(t *testing.T) {
	policy := testBackupPolicyKubernetes()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.BackupPolicyKubernetesWithID(testhelpers.TestProjectID, "bpk-001"),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.BackupPolicyKubernetesResponse{BackupPolicyKubernetes: policy}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceBackupPolicyKubernetes()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"id": "bpk-001",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "bpk-001" {
		t.Errorf("expected ID bpk-001, got %s", d.Id())
	}
	if v := d.Get("name").(string); v != "test-k8s-policy" {
		t.Errorf("expected name test-k8s-policy, got %s", v)
	}
	if v := d.Get("monthly.0.day_type").(string); v != "fixed_day" {
		t.Errorf("expected monthly.day_type fixed_day, got %s", v)
	}
}

func TestDataSourceBackupPolicyKubernetesRead_ByIDNameMismatch(t *testing.T) {
	policy := testBackupPolicyKubernetes()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.BackupPolicyKubernetesWithID(testhelpers.TestProjectID, "bpk-001"),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.BackupPolicyKubernetesResponse{BackupPolicyKubernetes: policy}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceBackupPolicyKubernetes()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"id":   "bpk-001",
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

func TestDataSourceBackupPolicyKubernetesRead_ByName(t *testing.T) {
	policy := testBackupPolicyKubernetes()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.BackupPolicyKuberneteses(testhelpers.TestProjectID),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListBackupPolicyKubernetesResponse{
				BackupPolicyKubernetes: []dto.BackupPolicyKubernetes{policy},
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceBackupPolicyKubernetes()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"name": "test-k8s-policy",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
	if d.Id() != "bpk-001" {
		t.Errorf("expected ID bpk-001, got %s", d.Id())
	}
}

func TestDataSourceBackupPolicyKubernetesRead_ByNameNotFound(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.BackupPolicyKuberneteses(testhelpers.TestProjectID),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListBackupPolicyKubernetesResponse{}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceBackupPolicyKubernetes()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"name": "missing-policy",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if !diags.HasError() {
		t.Fatal("expected an error when no policy matches the name")
	}
}

func TestDataSourceBackupPolicyKubernetesListRead(t *testing.T) {
	first := testBackupPolicyKubernetes()
	second := testBackupPolicyKubernetes()
	second.ID = "bpk-002"
	second.Name = "second-k8s-policy"
	second.Monthly = nil

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.BackupPolicyKuberneteses(testhelpers.TestProjectID),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListBackupPolicyKubernetesResponse{
				BackupPolicyKubernetes: []dto.BackupPolicyKubernetes{first, second},
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceBackupPolicyKubernetesList()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	policies := d.Get("backup_policy_kubernetes").([]interface{})
	if len(policies) != 2 {
		t.Fatalf("expected 2 policies, got %d", len(policies))
	}
	if v := d.Get("backup_policy_kubernetes.1.name").(string); v != "second-k8s-policy" {
		t.Errorf("expected second policy name second-k8s-policy, got %s", v)
	}
	if v := d.Get("backup_policy_kubernetes.1.monthly.#").(int); v != 0 {
		t.Errorf("expected second policy to have no monthly block, got %d", v)
	}
}
