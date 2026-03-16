package securitygrouprule

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// testSecurityGroupRule returns a fully populated dto.SecurityGroupRule for use in tests.
func testSecurityGroupRule() dto.SecurityGroupRule {
	return dto.SecurityGroupRule{
		ID:              "sgr-001",
		SecurityGroupID: "sg-001",
		Direction:       "ingress",
		Protocol:        "tcp",
		EtherType:       "IPv4",
		PortRangeMin:    443,
		PortRangeMax:    443,
		RemoteIPPrefix:  "0.0.0.0/0",
		RemoteGroupID:   "",
		ProjectID:       testhelpers.TestProjectID,
		ZoneID:          testhelpers.TestZoneID,
	}
}

func TestResourceSecurityGroupRuleCreate(t *testing.T) {
	rule := testSecurityGroupRule()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "POST",
			Pattern: "/v2/iac/projects/test-project-id/security-group-rules",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.SecurityGroupRuleResponse{Rule: rule}),
		},
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/security-group-rules/sgr-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.SecurityGroupRuleResponse{Rule: rule}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceSecurityGroupRule()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"security_group_id": "sg-001",
		"direction":         "ingress",
		"protocol":          "tcp",
		"ethertype":         "IPv4",
		"port_range_min":    443,
		"port_range_max":    443,
		"remote_ip_prefix":  "0.0.0.0/0",
		"remote_group_id":   "",
	})

	diags := res.CreateContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "sgr-001" {
		t.Errorf("expected ID sgr-001, got %s", d.Id())
	}
	if v := d.Get("security_group_id").(string); v != "sg-001" {
		t.Errorf("expected security_group_id sg-001, got %s", v)
	}
	if v := d.Get("direction").(string); v != "ingress" {
		t.Errorf("expected direction ingress, got %s", v)
	}
	if v := d.Get("protocol").(string); v != "tcp" {
		t.Errorf("expected protocol tcp, got %s", v)
	}
	if v := d.Get("ethertype").(string); v != "IPv4" {
		t.Errorf("expected ethertype IPv4, got %s", v)
	}
	if v := d.Get("port_range_min").(int); v != 443 {
		t.Errorf("expected port_range_min 443, got %d", v)
	}
	if v := d.Get("port_range_max").(int); v != 443 {
		t.Errorf("expected port_range_max 443, got %d", v)
	}
	if v := d.Get("remote_ip_prefix").(string); v != "0.0.0.0/0" {
		t.Errorf("expected remote_ip_prefix 0.0.0.0/0, got %s", v)
	}
}

func TestResourceSecurityGroupRuleRead(t *testing.T) {
	rule := testSecurityGroupRule()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/security-group-rules/sgr-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.SecurityGroupRuleResponse{Rule: rule}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceSecurityGroupRule()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"security_group_id": "",
		"direction":         "",
		"protocol":          "",
		"ethertype":         "IPv4",
		"port_range_min":    0,
		"port_range_max":    0,
		"remote_ip_prefix":  "",
		"remote_group_id":   "",
	})
	d.SetId("sgr-001")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if v := d.Get("security_group_id").(string); v != "sg-001" {
		t.Errorf("expected security_group_id sg-001, got %s", v)
	}
	if v := d.Get("direction").(string); v != "ingress" {
		t.Errorf("expected direction ingress, got %s", v)
	}
	if v := d.Get("protocol").(string); v != "tcp" {
		t.Errorf("expected protocol tcp, got %s", v)
	}
	if v := d.Get("ethertype").(string); v != "IPv4" {
		t.Errorf("expected ethertype IPv4, got %s", v)
	}
	if v := d.Get("port_range_min").(int); v != 443 {
		t.Errorf("expected port_range_min 443, got %d", v)
	}
	if v := d.Get("port_range_max").(int); v != 443 {
		t.Errorf("expected port_range_max 443, got %d", v)
	}
	if v := d.Get("remote_ip_prefix").(string); v != "0.0.0.0/0" {
		t.Errorf("expected remote_ip_prefix 0.0.0.0/0, got %s", v)
	}
	if v := d.Get("remote_group_id").(string); v != "" {
		t.Errorf("expected remote_group_id empty, got %s", v)
	}
}

func TestResourceSecurityGroupRuleRead_NotFound(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/security-group-rules/sgr-gone",
			Handler: testhelpers.EmptyHandler(http.StatusNotFound),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceSecurityGroupRule()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"security_group_id": "",
		"direction":         "",
		"protocol":          "",
		"ethertype":         "IPv4",
		"port_range_min":    0,
		"port_range_max":    0,
		"remote_ip_prefix":  "",
		"remote_group_id":   "",
	})
	d.SetId("sgr-gone")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error on 404: %v", diags)
	}

	if d.Id() != "" {
		t.Errorf("expected resource ID to be cleared after 404, got %s", d.Id())
	}
}

func TestResourceSecurityGroupRuleDelete(t *testing.T) {
	deletedCalled := false

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method: "DELETE",
			Pattern: "/v2/iac/projects/test-project-id/security-group-rules/sgr-001",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				deletedCalled = true
				w.WriteHeader(http.StatusNoContent)
			},
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceSecurityGroupRule()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"security_group_id": "sg-001",
		"direction":         "ingress",
		"protocol":          "tcp",
		"ethertype":         "IPv4",
		"port_range_min":    443,
		"port_range_max":    443,
		"remote_ip_prefix":  "0.0.0.0/0",
		"remote_group_id":   "",
	})
	d.SetId("sgr-001")

	diags := res.DeleteContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if !deletedCalled {
		t.Error("expected DELETE to have been called")
	}
}
