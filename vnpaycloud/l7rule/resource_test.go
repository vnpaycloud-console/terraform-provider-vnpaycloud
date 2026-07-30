package l7rule

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestResourceL7RuleSchema(t *testing.T) {
	res := ResourceL7Rule()

	if !res.Schema["l7policy_id"].Required || !res.Schema["l7policy_id"].ForceNew {
		t.Error("expected l7policy_id to be Required+ForceNew")
	}
	for _, key := range []string{"rule_type", "compare_type", "value"} {
		if !res.Schema[key].Required {
			t.Errorf("expected %s to be required", key)
		}
	}
}

func TestResourceL7RuleRead(t *testing.T) {
	rule := dto.L7Rule{
		ID:          "l7rule-001",
		L7PolicyID:  "l7pol-001",
		RuleType:    "COOKIE",
		CompareType: "EQUAL_TO",
		Value:       "sessionid",
		Key:         "cookie_name",
		Invert:      true,
		Status:      "ACTIVE",
	}

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.L7RuleWithID(testhelpers.TestProjectID, "l7pol-001", "l7rule-001"),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.L7RuleResponse{L7Rule: rule}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceL7Rule()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"l7policy_id": "l7pol-001",
	})
	d.SetId("l7rule-001")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if v := d.Get("rule_type").(string); v != "COOKIE" {
		t.Errorf("expected rule_type COOKIE, got %s", v)
	}
	if v := d.Get("compare_type").(string); v != "EQUAL_TO" {
		t.Errorf("expected compare_type EQUAL_TO, got %s", v)
	}
	if v := d.Get("value").(string); v != "sessionid" {
		t.Errorf("expected value sessionid, got %s", v)
	}
	if v := d.Get("key").(string); v != "cookie_name" {
		t.Errorf("expected key cookie_name, got %s", v)
	}
	if v := d.Get("invert").(bool); !v {
		t.Error("expected invert true")
	}
}
