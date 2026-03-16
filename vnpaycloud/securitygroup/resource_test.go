package securitygroup

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// testSecurityGroup returns a fully populated dto.SecurityGroup for use in tests.
func testSecurityGroup() dto.SecurityGroup {
	return dto.SecurityGroup{
		ID:          "sg-001",
		Name:        "test-sg",
		Description: "a test security group",
		Status:      "active",
		Rules: []dto.SecurityGroupRule{
			{
				ID:              "sgr-001",
				SecurityGroupID: "sg-001",
				Direction:       "ingress",
				Protocol:        "tcp",
				EtherType:       "IPv4",
				PortRangeMin:    80,
				PortRangeMax:    80,
				RemoteIPPrefix:  "0.0.0.0/0",
				RemoteGroupID:   "",
			},
		},
		CreatedAt: "2025-01-15T10:00:00Z",
		ProjectID: testhelpers.TestProjectID,
		ZoneID:    testhelpers.TestZoneID,
	}
}

func TestResourceSecurityGroupCreate(t *testing.T) {
	sg := testSecurityGroup()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "POST",
			Pattern: "/v2/iac/projects/test-project-id/security-groups",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.SecurityGroupResponse{SecurityGroup: sg}),
		},
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/security-groups/sg-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.SecurityGroupResponse{SecurityGroup: sg}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceSecurityGroup()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":        "test-sg",
		"description": "a test security group",
	})

	diags := res.CreateContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "sg-001" {
		t.Errorf("expected ID sg-001, got %s", d.Id())
	}
	if v := d.Get("name").(string); v != "test-sg" {
		t.Errorf("expected name test-sg, got %s", v)
	}
	if v := d.Get("description").(string); v != "a test security group" {
		t.Errorf("expected description 'a test security group', got %s", v)
	}
	if v := d.Get("created_at").(string); v != "2025-01-15T10:00:00Z" {
		t.Errorf("expected created_at 2025-01-15T10:00:00Z, got %s", v)
	}

	rules := d.Get("rules").([]interface{})
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
	rule := rules[0].(map[string]interface{})
	if rule["id"] != "sgr-001" {
		t.Errorf("expected rule id sgr-001, got %v", rule["id"])
	}
	if rule["direction"] != "ingress" {
		t.Errorf("expected rule direction ingress, got %v", rule["direction"])
	}
	if rule["protocol"] != "tcp" {
		t.Errorf("expected rule protocol tcp, got %v", rule["protocol"])
	}
	if rule["port_range_min"] != 80 {
		t.Errorf("expected rule port_range_min 80, got %v", rule["port_range_min"])
	}
	if rule["port_range_max"] != 80 {
		t.Errorf("expected rule port_range_max 80, got %v", rule["port_range_max"])
	}
}

func TestResourceSecurityGroupRead(t *testing.T) {
	sg := testSecurityGroup()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/security-groups/sg-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.SecurityGroupResponse{SecurityGroup: sg}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceSecurityGroup()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":        "",
		"description": "",
	})
	d.SetId("sg-001")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if v := d.Get("name").(string); v != "test-sg" {
		t.Errorf("expected name test-sg, got %s", v)
	}
	if v := d.Get("description").(string); v != "a test security group" {
		t.Errorf("expected description 'a test security group', got %s", v)
	}
	if v := d.Get("created_at").(string); v != "2025-01-15T10:00:00Z" {
		t.Errorf("expected created_at 2025-01-15T10:00:00Z, got %s", v)
	}

	rules := d.Get("rules").([]interface{})
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
	rule := rules[0].(map[string]interface{})
	if rule["id"] != "sgr-001" {
		t.Errorf("expected rule id sgr-001, got %v", rule["id"])
	}
	if rule["security_group_id"] != "sg-001" {
		t.Errorf("expected rule security_group_id sg-001, got %v", rule["security_group_id"])
	}
	if rule["direction"] != "ingress" {
		t.Errorf("expected rule direction ingress, got %v", rule["direction"])
	}
	if rule["ethertype"] != "IPv4" {
		t.Errorf("expected rule ethertype IPv4, got %v", rule["ethertype"])
	}
	if rule["remote_ip_prefix"] != "0.0.0.0/0" {
		t.Errorf("expected rule remote_ip_prefix 0.0.0.0/0, got %v", rule["remote_ip_prefix"])
	}
}

func TestResourceSecurityGroupRead_NotFound(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/security-groups/sg-gone",
			Handler: testhelpers.EmptyHandler(http.StatusNotFound),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceSecurityGroup()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":        "",
		"description": "",
	})
	d.SetId("sg-gone")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error on 404: %v", diags)
	}

	if d.Id() != "" {
		t.Errorf("expected resource ID to be cleared after 404, got %s", d.Id())
	}
}

func TestResourceSecurityGroupDelete(t *testing.T) {
	deletedCalled := false

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "DELETE",
			Pattern: "/v2/iac/projects/test-project-id/security-groups/sg-001",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				deletedCalled = true
				w.WriteHeader(http.StatusNoContent)
			},
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceSecurityGroup()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":        "test-sg",
		"description": "",
	})
	d.SetId("sg-001")

	diags := res.DeleteContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if !deletedCalled {
		t.Error("expected DELETE to have been called")
	}
}
