package networkaclrule

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func testNetworkACLRule() dto.NetworkACLRule {
	return dto.NetworkACLRule{
		ID:          "nacr-001",
		NaclID:      "nacl-001",
		Name:        "allow-https",
		Priority:    100,
		Type:        "HTTPS",
		Action:      "allow",
		PortStart:   443,
		PortEnd:     443,
		Source:      "0.0.0.0/0",
		Destination: "10.0.1.0/24",
		Status:      "active",
	}
}

func TestResourceNetworkACLRuleCreate(t *testing.T) {
	rule := testNetworkACLRule()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "POST",
			Pattern: "/v2/iac/projects/test-project-id/network-acl-rules",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.NetworkACLRuleResponse{Rule: rule}),
		},
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/network-acl-rules/nacr-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.NetworkACLRuleResponse{Rule: rule}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceNetworkACLRule()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"nacl_id":  "nacl-001",
		"name":     "allow-https",
		"priority": 100,
		"type":        "HTTPS",
		"action":      "allow",
		"source":      "0.0.0.0/0",
		"destination": "10.0.1.0/24",
	})

	diags := res.CreateContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
	if d.Id() != "nacr-001" {
		t.Errorf("expected ID nacr-001, got %s", d.Id())
	}
	if v := d.Get("type").(string); v != "HTTPS" {
		t.Errorf("expected type HTTPS, got %s", v)
	}
	if v := d.Get("port_start").(int); v != 443 {
		t.Errorf("expected port_start 443 (computed from preset), got %d", v)
	}
}

func TestResourceNetworkACLRuleCreate_CustomTCPRequiresPorts(t *testing.T) {
	// validateNetworkACLRuleConfig should reject CUSTOM_TCP without ports before any API call.
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceNetworkACLRule()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"nacl_id":  "nacl-001",
		"name":     "bad-custom",
		"priority": 100,
		"type":     "CUSTOM_TCP",
		"action":   "allow",
	})

	diags := res.CreateContext(context.Background(), d, cfg)
	if !diags.HasError() {
		t.Fatal("expected error for CUSTOM_TCP without ports, got nil")
	}
}

func TestResourceNetworkACLRuleRead(t *testing.T) {
	rule := testNetworkACLRule()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/network-acl-rules/nacr-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.NetworkACLRuleResponse{Rule: rule}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceNetworkACLRule()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{})
	d.SetId("nacr-001")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
	if v := d.Get("name").(string); v != "allow-https" {
		t.Errorf("expected name allow-https, got %s", v)
	}
	if v := d.Get("nacl_id").(string); v != "nacl-001" {
		t.Errorf("expected nacl_id nacl-001, got %s", v)
	}
}

func TestResourceNetworkACLRuleRead_NotFound(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/network-acl-rules/nacr-gone",
			Handler: testhelpers.EmptyHandler(http.StatusNotFound),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceNetworkACLRule()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{})
	d.SetId("nacr-gone")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error on 404: %v", diags)
	}
	if d.Id() != "" {
		t.Errorf("expected ID cleared after 404, got %s", d.Id())
	}
}

func TestResourceNetworkACLRuleDelete(t *testing.T) {
	deleteCalled := false

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "DELETE",
			Pattern: "/v2/iac/projects/test-project-id/network-acl-rules/nacr-001",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				deleteCalled = true
				w.WriteHeader(http.StatusNoContent)
			},
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceNetworkACLRule()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{})
	d.SetId("nacr-001")

	diags := res.DeleteContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
	if !deleteCalled {
		t.Error("expected DELETE to have been called")
	}
}
