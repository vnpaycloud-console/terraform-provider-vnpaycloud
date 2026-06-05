package robotaccount

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// testRobotAccount returns a fully populated dto.RobotAccount for use in tests.
func testRobotAccount() dto.RobotAccount {
	return dto.RobotAccount{
		ID:   "robot-001",
		Name: "robot$my-registry+robot-test",
		Permissions: []dto.RobotAccountPermission{
			{
				RegistryID: "reg-001",
				Actions:    []string{"push", "pull"},
			},
			{
				RegistryID: "reg-002",
				Actions:    []string{"pull"},
			},
		},
		ExpiresAt: "2026-06-01T10:00:00Z",
		Enabled:   true,
		CreatedAt: "2025-06-01T10:00:00Z",
	}
}

func TestResourceRobotAccountCreate(t *testing.T) {
	robot := testRobotAccount()
	secret := "RegistrySecretXYZ123"

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "POST",
			Pattern: "/v2/iac/projects/test-project-id/robot-accounts",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.RobotAccountResponse{
				RobotAccount: robot,
				Secret:       secret,
			}),
		},
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/robot-accounts/robot-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.RobotAccountResponse{
				RobotAccount: robot,
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceRobotAccount()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name": "robot-test",
		"permission": []interface{}{
			map[string]interface{}{
				"registry_id": "reg-001",
				"actions":     []interface{}{"push", "pull"},
			},
			map[string]interface{}{
				"registry_id": "reg-002",
				"actions":     []interface{}{"pull"},
			},
		},
		"expires_in_days": 365,
	})

	diags := res.CreateContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "robot-001" {
		t.Errorf("expected ID robot-001, got %s", d.Id())
	}
	if v := d.Get("secret").(string); v != secret {
		t.Errorf("expected secret %s, got %s", secret, v)
	}
	if v := d.Get("enabled").(bool); !v {
		t.Error("expected enabled true, got false")
	}
	if v := d.Get("expires_at").(string); v != "2026-06-01T10:00:00Z" {
		t.Errorf("expected expires_at 2026-06-01T10:00:00Z, got %s", v)
	}
	if v := d.Get("created_at").(string); v != "2025-06-01T10:00:00Z" {
		t.Errorf("expected created_at 2025-06-01T10:00:00Z, got %s", v)
	}

	// Verify permissions
	perms := d.Get("permission").([]interface{})
	if len(perms) != 2 {
		t.Fatalf("expected 2 permission blocks, got %d", len(perms))
	}
	perm0 := perms[0].(map[string]interface{})
	if perm0["registry_id"].(string) != "reg-001" {
		t.Errorf("expected first permission registry_id reg-001, got %s", perm0["registry_id"])
	}
	actions0 := perm0["actions"].([]interface{})
	if len(actions0) != 2 {
		t.Fatalf("expected 2 actions in first permission, got %d", len(actions0))
	}
}

func TestResourceRobotAccountRead(t *testing.T) {
	robot := testRobotAccount()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/robot-accounts/robot-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.RobotAccountResponse{
				RobotAccount: robot,
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceRobotAccount()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name": "robot-test",
		"permission": []interface{}{
			map[string]interface{}{
				"registry_id": "reg-001",
				"actions":     []interface{}{},
			},
		},
		"expires_in_days": 0,
	})
	d.SetId("robot-001")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if v := d.Get("enabled").(bool); !v {
		t.Error("expected enabled true, got false")
	}
	if v := d.Get("expires_at").(string); v != "2026-06-01T10:00:00Z" {
		t.Errorf("expected expires_at 2026-06-01T10:00:00Z, got %s", v)
	}
	if v := d.Get("created_at").(string); v != "2025-06-01T10:00:00Z" {
		t.Errorf("expected created_at 2025-06-01T10:00:00Z, got %s", v)
	}

	perms := d.Get("permission").([]interface{})
	if len(perms) != 2 {
		t.Fatalf("expected 2 permission blocks, got %d", len(perms))
	}
	perm0 := perms[0].(map[string]interface{})
	if perm0["registry_id"].(string) != "reg-001" {
		t.Errorf("expected first permission registry_id reg-001, got %s", perm0["registry_id"])
	}
	actions0 := perm0["actions"].([]interface{})
	if len(actions0) != 2 {
		t.Fatalf("expected 2 actions in first permission, got %d", len(actions0))
	}
	if actions0[0].(string) != "push" {
		t.Errorf("expected first action push, got %s", actions0[0])
	}
	if actions0[1].(string) != "pull" {
		t.Errorf("expected second action pull, got %s", actions0[1])
	}
}

func TestResourceRobotAccountRead_NotFound(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/robot-accounts/robot-gone",
			Handler: testhelpers.EmptyHandler(http.StatusNotFound),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceRobotAccount()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name": "",
		"permission": []interface{}{
			map[string]interface{}{
				"registry_id": "reg-001",
				"actions":     []interface{}{},
			},
		},
		"expires_in_days": 0,
	})
	d.SetId("robot-gone")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error on 404: %v", diags)
	}

	if d.Id() != "" {
		t.Errorf("expected resource ID to be cleared after 404, got %s", d.Id())
	}
}

func TestResourceRobotAccountDelete(t *testing.T) {
	robot := testRobotAccount()
	deletedCalled := false

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Pattern: "/v2/iac/projects/test-project-id/robot-accounts/robot-001",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case "GET":
					if deletedCalled {
						w.WriteHeader(http.StatusNotFound)
						return
					}
					testhelpers.JSONHandler(t, http.StatusOK, dto.RobotAccountResponse{
						RobotAccount: robot,
					})(w, r)
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

	res := ResourceRobotAccount()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name": "robot-test",
		"permission": []interface{}{
			map[string]interface{}{
				"registry_id": "reg-001",
				"actions":     []interface{}{"push", "pull"},
			},
		},
		"expires_in_days": 365,
	})
	d.SetId("robot-001")

	diags := res.DeleteContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if !deletedCalled {
		t.Error("expected DELETE to have been called")
	}
}
