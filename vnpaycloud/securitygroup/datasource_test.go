package securitygroup

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestDataSourceSecurityGroupRead_ByID(t *testing.T) {
	sg := testSecurityGroup()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/security-groups/sg-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.SecurityGroupResponse{SecurityGroup: sg}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceSecurityGroup()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"id": "sg-001",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
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
	if v := d.Get("status").(string); v != "active" {
		t.Errorf("expected status active, got %s", v)
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
}

func TestDataSourceSecurityGroupRead_ByName(t *testing.T) {
	sg := testSecurityGroup()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method: "GET",
			Pattern: "/v2/iac/projects/test-project-id/security-groups",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListSecurityGroupsResponse{
				SecurityGroups: []dto.SecurityGroup{sg},
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceSecurityGroup()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"name": "test-sg",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "sg-001" {
		t.Errorf("expected ID sg-001, got %s", d.Id())
	}
	if v := d.Get("name").(string); v != "test-sg" {
		t.Errorf("expected name test-sg, got %s", v)
	}
}

func TestDataSourceSecurityGroupRead_NotFound(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method: "GET",
			Pattern: "/v2/iac/projects/test-project-id/security-groups",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListSecurityGroupsResponse{
				SecurityGroups: []dto.SecurityGroup{},
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceSecurityGroup()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"name": "nonexistent",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if !diags.HasError() {
		t.Fatal("expected error for nonexistent security group, got none")
	}
}

func TestDataSourceSecurityGroupsRead(t *testing.T) {
	sg1 := testSecurityGroup()
	sg2 := dto.SecurityGroup{
		ID:          "sg-002",
		Name:        "test-sg-2",
		Description: "second security group",
		Status:      "active",
		Rules:       []dto.SecurityGroupRule{},
		CreatedAt:   "2025-01-16T12:00:00Z",
		ProjectID:   testhelpers.TestProjectID,
		ZoneID:      testhelpers.TestZoneID,
	}

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method: "GET",
			Pattern: "/v2/iac/projects/test-project-id/security-groups",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListSecurityGroupsResponse{
				SecurityGroups: []dto.SecurityGroup{sg1, sg2},
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceSecurityGroups()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	sgs := d.Get("security_groups").([]interface{})
	if len(sgs) != 2 {
		t.Fatalf("expected 2 security groups, got %d", len(sgs))
	}

	first := sgs[0].(map[string]interface{})
	if first["id"] != "sg-001" {
		t.Errorf("expected first sg id sg-001, got %v", first["id"])
	}
	if first["name"] != "test-sg" {
		t.Errorf("expected first sg name test-sg, got %v", first["name"])
	}

	second := sgs[1].(map[string]interface{})
	if second["id"] != "sg-002" {
		t.Errorf("expected second sg id sg-002, got %v", second["id"])
	}
	if second["name"] != "test-sg-2" {
		t.Errorf("expected second sg name test-sg-2, got %v", second["name"])
	}
}
