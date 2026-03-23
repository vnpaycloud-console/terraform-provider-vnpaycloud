package securitygrouprule

import (
	"context"
	"net/http"
	"sync/atomic"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"
	"terraform-provider-vnpaycloud/vnpaycloud/securitygroup"

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

// TestResourceSecurityGroupRuleCreate_Error verifies that a POST error
// is returned to the caller.
func TestResourceSecurityGroupRuleCreate_Error(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "POST",
			Pattern: "/v2/iac/projects/test-project-id/security-group-rules",
			Handler: testhelpers.EmptyHandler(http.StatusInternalServerError),
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
	if !diags.HasError() {
		t.Fatal("expected error for 500 response, got nil")
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

// TestCreateSecurityGroupAndRulesTogether simulates the Terraform flow of
// creating a security group and multiple security group rules simultaneously.
// This mirrors a real Terraform config like:
//
//	resource "vnpaycloud_security_group" "web" { name = "web-sg" }
//	resource "vnpaycloud_security_group_rule" "http"  { security_group_id = vnpaycloud_security_group.web.id ... }
//	resource "vnpaycloud_security_group_rule" "https" { security_group_id = vnpaycloud_security_group.web.id ... }
func TestCreateSecurityGroupAndRulesTogether(t *testing.T) {
	// --- fixtures ---
	sg := dto.SecurityGroup{
		ID:          "sg-web-001",
		Name:        "web-sg",
		Description: "web security group",
		Status:      "active",
		Rules:       []dto.SecurityGroupRule{},
		CreatedAt:   "2025-06-01T10:00:00Z",
		ProjectID:   testhelpers.TestProjectID,
		ZoneID:      testhelpers.TestZoneID,
	}

	httpRule := dto.SecurityGroupRule{
		ID:              "sgr-http-001",
		SecurityGroupID: "sg-web-001",
		Direction:       "ingress",
		Protocol:        "tcp",
		EtherType:       "IPv4",
		PortRangeMin:    80,
		PortRangeMax:    80,
		RemoteIPPrefix:  "0.0.0.0/0",
		ProjectID:       testhelpers.TestProjectID,
		ZoneID:          testhelpers.TestZoneID,
	}

	httpsRule := dto.SecurityGroupRule{
		ID:              "sgr-https-001",
		SecurityGroupID: "sg-web-001",
		Direction:       "ingress",
		Protocol:        "tcp",
		EtherType:       "IPv4",
		PortRangeMin:    443,
		PortRangeMax:    443,
		RemoteIPPrefix:  "0.0.0.0/0",
		ProjectID:       testhelpers.TestProjectID,
		ZoneID:          testhelpers.TestZoneID,
	}

	sshRule := dto.SecurityGroupRule{
		ID:              "sgr-ssh-001",
		SecurityGroupID: "sg-web-001",
		Direction:       "ingress",
		Protocol:        "tcp",
		EtherType:       "IPv4",
		PortRangeMin:    22,
		PortRangeMax:    22,
		RemoteIPPrefix:  "10.0.0.0/8",
		ProjectID:       testhelpers.TestProjectID,
		ZoneID:          testhelpers.TestZoneID,
	}

	// SG after all rules are attached
	sgWithRules := sg
	sgWithRules.Rules = []dto.SecurityGroupRule{httpRule, httpsRule, sshRule}

	// Track SG create polling
	var sgGetCalls int32

	// Track rule creation order
	var ruleCreateCalls int32

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		// --- Security Group endpoints ---
		{
			Method:  "POST",
			Pattern: "/v2/iac/projects/test-project-id/security-groups",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				creating := sg
				creating.Status = "creating"
				testhelpers.JSONHandler(t, http.StatusOK, dto.SecurityGroupResponse{SecurityGroup: creating})(w, r)
			},
		},
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/security-groups/sg-web-001",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				n := atomic.AddInt32(&sgGetCalls, 1)
				resp := sgWithRules
				if n <= 1 {
					resp = sg
					resp.Status = "creating"
				}
				testhelpers.JSONHandler(t, http.StatusOK, dto.SecurityGroupResponse{SecurityGroup: resp})(w, r)
			},
		},
		// --- Security Group Rule endpoints ---
		{
			Method:  "POST",
			Pattern: "/v2/iac/projects/test-project-id/security-group-rules",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				n := atomic.AddInt32(&ruleCreateCalls, 1)
				var rule dto.SecurityGroupRule
				switch n {
				case 1:
					rule = httpRule
				case 2:
					rule = httpsRule
				case 3:
					rule = sshRule
				}
				testhelpers.JSONHandler(t, http.StatusOK, dto.SecurityGroupRuleResponse{Rule: rule})(w, r)
			},
		},
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/security-group-rules/sgr-http-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.SecurityGroupRuleResponse{Rule: httpRule}),
		},
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/security-group-rules/sgr-https-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.SecurityGroupRuleResponse{Rule: httpsRule}),
		},
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/security-group-rules/sgr-ssh-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.SecurityGroupRuleResponse{Rule: sshRule}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	// ============================
	// Step 1: Create Security Group
	// ============================
	sgRes := securitygroup.ResourceSecurityGroup()
	sgData := schema.TestResourceDataRaw(t, sgRes.Schema, map[string]interface{}{
		"name":        "web-sg",
		"description": "web security group",
	})

	diags := sgRes.CreateContext(context.Background(), sgData, cfg)
	if diags.HasError() {
		t.Fatalf("failed to create security group: %v", diags)
	}

	if sgData.Id() != "sg-web-001" {
		t.Fatalf("expected SG ID sg-web-001, got %s", sgData.Id())
	}

	// Verify polling happened
	if n := atomic.LoadInt32(&sgGetCalls); n < 2 {
		t.Errorf("expected at least 2 GET polls for SG, got %d", n)
	}

	// ================================================
	// Step 2: Create Security Group Rules (3 rules)
	// ================================================
	type ruleTestCase struct {
		name           string
		direction      string
		protocol       string
		portMin        int
		portMax        int
		remoteIPPrefix string
		expectedID     string
	}

	ruleCases := []ruleTestCase{
		{"HTTP rule", "ingress", "tcp", 80, 80, "0.0.0.0/0", "sgr-http-001"},
		{"HTTPS rule", "ingress", "tcp", 443, 443, "0.0.0.0/0", "sgr-https-001"},
		{"SSH rule", "ingress", "tcp", 22, 22, "10.0.0.0/8", "sgr-ssh-001"},
	}

	ruleRes := ResourceSecurityGroupRule()

	for _, rc := range ruleCases {
		t.Run(rc.name, func(t *testing.T) {
			ruleData := schema.TestResourceDataRaw(t, ruleRes.Schema, map[string]interface{}{
				"security_group_id": sgData.Id(),
				"direction":         rc.direction,
				"protocol":          rc.protocol,
				"ethertype":         "IPv4",
				"port_range_min":    rc.portMin,
				"port_range_max":    rc.portMax,
				"remote_ip_prefix":  rc.remoteIPPrefix,
				"remote_group_id":   "",
			})

			diags := ruleRes.CreateContext(context.Background(), ruleData, cfg)
			if diags.HasError() {
				t.Fatalf("failed to create %s: %v", rc.name, diags)
			}

			if ruleData.Id() != rc.expectedID {
				t.Errorf("expected rule ID %s, got %s", rc.expectedID, ruleData.Id())
			}
			if v := ruleData.Get("security_group_id").(string); v != "sg-web-001" {
				t.Errorf("expected security_group_id sg-web-001, got %s", v)
			}
			if v := ruleData.Get("port_range_min").(int); v != rc.portMin {
				t.Errorf("expected port_range_min %d, got %d", rc.portMin, v)
			}
			if v := ruleData.Get("port_range_max").(int); v != rc.portMax {
				t.Errorf("expected port_range_max %d, got %d", rc.portMax, v)
			}
		})
	}

	// Verify all 3 rules were created
	if n := atomic.LoadInt32(&ruleCreateCalls); n != 3 {
		t.Errorf("expected 3 rule POST calls, got %d", n)
	}

	// ================================================
	// Step 3: Read back SG and verify rules embedded
	// ================================================
	sgReadData := schema.TestResourceDataRaw(t, sgRes.Schema, map[string]interface{}{
		"name":        "",
		"description": "",
	})
	sgReadData.SetId("sg-web-001")

	diags = sgRes.ReadContext(context.Background(), sgReadData, cfg)
	if diags.HasError() {
		t.Fatalf("failed to read security group: %v", diags)
	}

	rules := sgReadData.Get("rules").([]interface{})
	if len(rules) != 3 {
		t.Fatalf("expected 3 rules embedded in SG, got %d", len(rules))
	}

	// Verify each rule is present
	expectedPorts := map[string]int{
		"sgr-http-001":  80,
		"sgr-https-001": 443,
		"sgr-ssh-001":   22,
	}
	for _, r := range rules {
		rule := r.(map[string]interface{})
		id := rule["id"].(string)
		expectedPort, ok := expectedPorts[id]
		if !ok {
			t.Errorf("unexpected rule ID: %s", id)
			continue
		}
		if rule["port_range_min"] != expectedPort {
			t.Errorf("rule %s: expected port_range_min %d, got %v", id, expectedPort, rule["port_range_min"])
		}
		if rule["security_group_id"] != "sg-web-001" {
			t.Errorf("rule %s: expected security_group_id sg-web-001, got %v", id, rule["security_group_id"])
		}
	}
}

func TestResourceSecurityGroupRuleDelete(t *testing.T) {
	deletedCalled := false

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "DELETE",
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
