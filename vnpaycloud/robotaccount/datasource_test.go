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
			Pattern: "/v2/iac/projects/test-project-id/robot-accounts/robot-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.RobotAccountResponse{
				RobotAccount: robot,
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := DataSourceRobotAccount()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"id": "robot-001",
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
	if actions0[0].(string) != "repository:push" {
		t.Errorf("expected first action repository:push, got %s", actions0[0])
	}
	if actions0[1].(string) != "repository:pull" {
		t.Errorf("expected second action repository:pull, got %s", actions0[1])
	}
}
