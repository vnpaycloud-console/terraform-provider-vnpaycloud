package registryproject

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestDataSourceRegistryProjectRead_ByID(t *testing.T) {
	reg := testRegistryProject()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/registries/reg-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.RegistryProjectResponse{Registry: reg}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := DataSourceRegistryProject()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"id": "reg-001",
	})

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "reg-001" {
		t.Errorf("expected ID reg-001, got %s", d.Id())
	}
	if v := d.Get("name").(string); v != "my-registry" {
		t.Errorf("expected name my-registry, got %s", v)
	}
	if v := d.Get("is_public").(bool); v {
		t.Error("expected is_public false, got true")
	}
	if v := d.Get("storage_limit").(int); v != 10737418240 {
		t.Errorf("expected storage_limit 10737418240, got %d", v)
	}
	if v := d.Get("storage_used").(int); v != 0 {
		t.Errorf("expected storage_used 0, got %d", v)
	}
	if v := d.Get("repo_count").(int); v != 0 {
		t.Errorf("expected repo_count 0, got %d", v)
	}
	if v := d.Get("status").(string); v != "active" {
		t.Errorf("expected status active, got %s", v)
	}
	if v := d.Get("created_at").(string); v != "2025-06-01T10:00:00Z" {
		t.Errorf("expected created_at 2025-06-01T10:00:00Z, got %s", v)
	}
}

func TestDataSourceRegistryProjectsRead(t *testing.T) {
	listResp := dto.ListRegistryProjectsResponse{
		Registries: []dto.RegistryProject{
			{
				ID:           "reg-001",
				Name:         "my-registry",
				IsPublic:     false,
				StorageLimit: 10737418240,
				StorageUsed:  0,
				RepoCount:    0,
				Status:       "active",
				CreatedAt:    "2025-06-01T10:00:00Z",
			},
			{
				ID:           "reg-002",
				Name:         "public-registry",
				IsPublic:     true,
				StorageLimit: 5368709120,
				StorageUsed:  1073741824,
				RepoCount:    3,
				Status:       "active",
				CreatedAt:    "2025-06-02T12:00:00Z",
			},
		},
	}

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/registries",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, listResp),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := DataSourceRegistryProjects()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{})

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != testhelpers.TestProjectID {
		t.Errorf("expected ID %s, got %s", testhelpers.TestProjectID, d.Id())
	}

	registries := d.Get("registries").([]interface{})
	if len(registries) != 2 {
		t.Fatalf("expected 2 registries, got %d", len(registries))
	}

	first := registries[0].(map[string]interface{})
	if first["id"] != "reg-001" {
		t.Errorf("expected first registry id 'reg-001', got '%s'", first["id"])
	}
	if first["name"] != "my-registry" {
		t.Errorf("expected first registry name 'my-registry', got '%s'", first["name"])
	}
	if first["is_public"] != false {
		t.Error("expected first registry is_public false, got true")
	}
	if first["status"] != "active" {
		t.Errorf("expected first registry status 'active', got '%s'", first["status"])
	}

	second := registries[1].(map[string]interface{})
	if second["id"] != "reg-002" {
		t.Errorf("expected second registry id 'reg-002', got '%s'", second["id"])
	}
	if second["name"] != "public-registry" {
		t.Errorf("expected second registry name 'public-registry', got '%s'", second["name"])
	}
	if second["is_public"] != true {
		t.Error("expected second registry is_public true, got false")
	}
	// repo_count comes from int32 in DTO, Terraform schema stores as int.
	if fmt.Sprintf("%v", second["repo_count"]) != "3" {
		t.Errorf("expected second registry repo_count 3, got %v", second["repo_count"])
	}
}
