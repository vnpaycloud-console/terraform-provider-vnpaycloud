package l7policy

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestResourceL7PolicySchema(t *testing.T) {
	res := ResourceL7Policy()

	if !res.Schema["name"].Required {
		t.Error("expected name to be required")
	}
	if res.Schema["name"].Computed {
		t.Error("name should not be Computed (it is not auto-generated)")
	}
	if !res.Schema["listener_id"].Required || !res.Schema["listener_id"].ForceNew {
		t.Error("expected listener_id to be Required+ForceNew")
	}
	if !res.Schema["action"].Required {
		t.Error("expected action to be required")
	}
}

func TestResourceL7PolicyRead(t *testing.T) {
	pol := dto.L7Policy{
		ID:             "l7pol-001",
		Name:           "reject-policy",
		ListenerID:     "lsnr-001",
		Action:         "REJECT",
		Position:       1,
		Description:    "terraform policy",
		RedirectPoolID: "",
		RedirectURL:    "",
		Status:         "ACTIVE",
	}

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.L7PolicyWithID(testhelpers.TestProjectID, "l7pol-001"),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.L7PolicyResponse{L7Policy: pol}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceL7Policy()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{})
	d.SetId("l7pol-001")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if v := d.Get("name").(string); v != "reject-policy" {
		t.Errorf("expected name reject-policy, got %s", v)
	}
	if v := d.Get("listener_id").(string); v != "lsnr-001" {
		t.Errorf("expected listener_id lsnr-001, got %s", v)
	}
	if v := d.Get("action").(string); v != "REJECT" {
		t.Errorf("expected action REJECT, got %s", v)
	}
	if v := d.Get("position").(int); v != 1 {
		t.Errorf("expected position 1, got %d", v)
	}
}
