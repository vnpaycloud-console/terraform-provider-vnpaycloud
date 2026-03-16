package loadbalancer

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestDataSourceLoadBalancerRead_ByID(t *testing.T) {
	lb := testLoadBalancer()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/load-balancers/lb-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.LoadBalancerResponse{LoadBalancer: lb}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceLoadBalancer()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"id": "lb-001",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "lb-001" {
		t.Errorf("expected ID lb-001, got %s", d.Id())
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

func TestDataSourceLoadBalancersRead(t *testing.T) {
	lb1 := testLoadBalancer()
	lb2 := dto.LoadBalancer{
		ID:          "lb-002",
		Name:        "test-lb-2",
		Description: "second lb",
		VipAddress:  "10.0.0.101",
		VipSubnetID: "subnet-002",
		Status:      "active",
		ListenerIDs: []string{},
		CreatedAt:   "2025-01-16T12:00:00Z",
	}

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method: "GET",
			Pattern: "/v2/iac/projects/test-project-id/load-balancers",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListLoadBalancersResponse{
				LoadBalancers: []dto.LoadBalancer{lb1, lb2},
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceLoadBalancers()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	lbs := d.Get("load_balancers").([]interface{})
	if len(lbs) != 2 {
		t.Fatalf("expected 2 load balancers, got %d", len(lbs))
	}

	first := lbs[0].(map[string]interface{})
	if first["id"] != "lb-001" {
		t.Errorf("expected first lb id lb-001, got %v", first["id"])
	}
	if first["name"] != "test-lb" {
		t.Errorf("expected first lb name test-lb, got %v", first["name"])
	}

	second := lbs[1].(map[string]interface{})
	if second["id"] != "lb-002" {
		t.Errorf("expected second lb id lb-002, got %v", second["id"])
	}
	if second["name"] != "test-lb-2" {
		t.Errorf("expected second lb name test-lb-2, got %v", second["name"])
	}
}
