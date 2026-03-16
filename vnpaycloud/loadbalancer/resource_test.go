package loadbalancer

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// testLoadBalancer returns a fully populated dto.LoadBalancer for use in tests.
func testLoadBalancer() dto.LoadBalancer {
	return dto.LoadBalancer{
		ID:          "lb-001",
		Name:        "test-lb",
		Description: "a test load balancer",
		VipAddress:  "10.0.0.100",
		VipSubnetID: "subnet-001",
		Status:      "active",
		ListenerIDs: []string{"listener-001"},
		CreatedAt:   "2025-01-15T10:00:00Z",
	}
}

func TestResourceLoadBalancerCreate(t *testing.T) {
	lb := testLoadBalancer()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "POST",
			Pattern: "/v2/iac/projects/test-project-id/load-balancers",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.LoadBalancerResponse{LoadBalancer: lb}),
		},
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/load-balancers/lb-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.LoadBalancerResponse{LoadBalancer: lb}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceLoadBalancer()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":        "test-lb",
		"description": "a test load balancer",
		"subnet_id":   "subnet-001",
		"flavor":      "medium",
	})

	diags := res.CreateContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "lb-001" {
		t.Errorf("expected ID lb-001, got %s", d.Id())
	}
	if v := d.Get("name").(string); v != "test-lb" {
		t.Errorf("expected name test-lb, got %s", v)
	}
	if v := d.Get("status").(string); v != "active" {
		t.Errorf("expected status active, got %s", v)
	}
	if v := d.Get("vip_address").(string); v != "10.0.0.100" {
		t.Errorf("expected vip_address 10.0.0.100, got %s", v)
	}
}

func TestResourceLoadBalancerRead(t *testing.T) {
	lb := testLoadBalancer()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/load-balancers/lb-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.LoadBalancerResponse{LoadBalancer: lb}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceLoadBalancer()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":        "",
		"description": "",
		"subnet_id":   "",
		"flavor":      "",
	})
	d.SetId("lb-001")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if v := d.Get("name").(string); v != "test-lb" {
		t.Errorf("expected name test-lb, got %s", v)
	}
	if v := d.Get("description").(string); v != "a test load balancer" {
		t.Errorf("expected description 'a test load balancer', got %s", v)
	}
	if v := d.Get("vip_address").(string); v != "10.0.0.100" {
		t.Errorf("expected vip_address 10.0.0.100, got %s", v)
	}
	if v := d.Get("vip_subnet_id").(string); v != "subnet-001" {
		t.Errorf("expected vip_subnet_id subnet-001, got %s", v)
	}
	if v := d.Get("status").(string); v != "active" {
		t.Errorf("expected status active, got %s", v)
	}
	if v := d.Get("created_at").(string); v != "2025-01-15T10:00:00Z" {
		t.Errorf("expected created_at 2025-01-15T10:00:00Z, got %s", v)
	}
}

func TestResourceLoadBalancerRead_NotFound(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/load-balancers/lb-gone",
			Handler: testhelpers.EmptyHandler(http.StatusNotFound),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceLoadBalancer()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":        "",
		"description": "",
		"subnet_id":   "",
		"flavor":      "",
	})
	d.SetId("lb-gone")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error on 404: %v", diags)
	}

	if d.Id() != "" {
		t.Errorf("expected resource ID to be cleared after 404, got %s", d.Id())
	}
}

func TestResourceLoadBalancerDelete(t *testing.T) {
	lb := testLoadBalancer()
	deletedCalled := false

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Pattern: "/v2/iac/projects/test-project-id/load-balancers/lb-001",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case "GET":
					if deletedCalled {
						w.WriteHeader(http.StatusNotFound)
						return
					}
					testhelpers.JSONHandler(t, http.StatusOK, dto.LoadBalancerResponse{LoadBalancer: lb})(w, r)
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

	res := ResourceLoadBalancer()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":        "test-lb",
		"description": "",
		"subnet_id":   "subnet-001",
		"flavor":      "medium",
	})
	d.SetId("lb-001")

	diags := res.DeleteContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if !deletedCalled {
		t.Error("expected DELETE to have been called")
	}
}
