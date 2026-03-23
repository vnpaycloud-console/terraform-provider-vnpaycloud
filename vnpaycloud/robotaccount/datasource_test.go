package robotaccount

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestDataSourceRobotAccountRead_ByID(t *testing.T) {
	robot := testRobotAccount()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/registries/reg-001/robot-accounts/robot-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.RobotAccountResponse{
				RobotAccount: robot,
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := DataSourceRobotAccount()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"id":          "robot-001",
		"registry_id": "reg-001",
	})

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "robot-001" {
		t.Errorf("expected ID robot-001, got %s", d.Id())
	}
	if v := d.Get("name").(string); v != "robot$my-registry+robot-test" {
		t.Errorf("expected name 'robot$my-registry+robot-test', got %s", v)
	}
	if v := d.Get("registry_id").(string); v != "reg-001" {
		t.Errorf("expected registry_id reg-001, got %s", v)
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

	perms := d.Get("permissions").([]interface{})
	if len(perms) != 2 {
		t.Fatalf("expected 2 permissions, got %d", len(perms))
	}
	if perms[0].(string) != "push" {
		t.Errorf("expected first permission push, got %s", perms[0])
	}
	if perms[1].(string) != "pull" {
		t.Errorf("expected second permission pull, got %s", perms[1])
	}
}
