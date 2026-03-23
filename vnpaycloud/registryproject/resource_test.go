package registryproject

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// testRegistryProject returns a fully populated dto.RegistryProject for use in tests.
func testRegistryProject() dto.RegistryProject {
	return dto.RegistryProject{
		ID:           "reg-001",
		Name:         "my-registry",
		IsPublic:     false,
		StorageLimit: "10737418240",
		StorageUsed:  0,
		RepoCount:    0,
		Status:       "active",
		CreatedAt:    "2025-06-01T10:00:00Z",
	}
}

func TestResourceRegistryProjectCreate(t *testing.T) {
	reg := testRegistryProject()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Pattern: "/v2/iac/projects/test-project-id/registries",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case "POST":
					testhelpers.JSONHandler(t, http.StatusOK, dto.RegistryProjectResponse{Registry: reg})(w, r)
				default:
					http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				}
			},
		},
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/registries/reg-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.RegistryProjectResponse{Registry: reg}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceRegistryProject()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":          "my-registry",
		"is_public":     false,
		"storage_limit": "10737418240",
	})

	diags := res.CreateContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "reg-001" {
		t.Errorf("expected ID reg-001, got %s", d.Id())
	}
	if v := d.Get("name").(string); v != "my-registry" {
		t.Errorf("expected name my-registry, got %s", v)
	}
	if v := d.Get("status").(string); v != "active" {
		t.Errorf("expected status active, got %s", v)
	}
	if v := d.Get("is_public").(bool); v {
		t.Error("expected is_public false, got true")
	}
	if v := d.Get("storage_limit").(string); v != "10737418240" {
		t.Errorf("expected storage_limit 10737418240, got %s", v)
	}
	if v := d.Get("created_at").(string); v != "2025-06-01T10:00:00Z" {
		t.Errorf("expected created_at 2025-06-01T10:00:00Z, got %s", v)
	}
}

func TestResourceRegistryProjectRead(t *testing.T) {
	reg := testRegistryProject()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/registries/reg-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.RegistryProjectResponse{Registry: reg}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceRegistryProject()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":          "",
		"is_public":     false,
		"storage_limit": "",
	})
	d.SetId("reg-001")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if v := d.Get("name").(string); v != "my-registry" {
		t.Errorf("expected name my-registry, got %s", v)
	}
	if v := d.Get("status").(string); v != "active" {
		t.Errorf("expected status active, got %s", v)
	}
	if v := d.Get("repo_count").(int); v != 0 {
		t.Errorf("expected repo_count 0, got %d", v)
	}
	if v := d.Get("storage_used").(int); v != 0 {
		t.Errorf("expected storage_used 0, got %d", v)
	}
	if v := d.Get("created_at").(string); v != "2025-06-01T10:00:00Z" {
		t.Errorf("expected created_at 2025-06-01T10:00:00Z, got %s", v)
	}
}

func TestResourceRegistryProjectRead_NotFound(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/registries/reg-gone",
			Handler: testhelpers.EmptyHandler(http.StatusNotFound),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceRegistryProject()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":          "",
		"is_public":     false,
		"storage_limit": "",
	})
	d.SetId("reg-gone")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error on 404: %v", diags)
	}

	if d.Id() != "" {
		t.Errorf("expected resource ID to be cleared after 404, got %s", d.Id())
	}
}

func TestResourceRegistryProjectDelete(t *testing.T) {
	reg := testRegistryProject()
	deletedCalled := false

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Pattern: "/v2/iac/projects/test-project-id/registries/reg-001",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case "GET":
					if deletedCalled {
						w.WriteHeader(http.StatusNotFound)
						return
					}
					testhelpers.JSONHandler(t, http.StatusOK, dto.RegistryProjectResponse{Registry: reg})(w, r)
				case "DELETE":
					deletedCalled = true
					w.WriteHeader(http.StatusAccepted)
				default:
					http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				}
			},
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceRegistryProject()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":          "my-registry",
		"is_public":     false,
		"storage_limit": "10737418240",
	})
	d.SetId("reg-001")

	diags := res.DeleteContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if !deletedCalled {
		t.Error("expected DELETE to have been called")
	}
}
