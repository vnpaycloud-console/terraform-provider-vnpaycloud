package pool

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestDataSourcePoolRead_ByID(t *testing.T) {
	p := testPool()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/pools/pool-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.PoolResponse{Pool: p}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourcePool()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"id": "pool-001",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "pool-001" {
		t.Errorf("expected ID pool-001, got %s", d.Id())
	}
	if v := d.Get("name").(string); v != "test-pool" {
		t.Errorf("expected name test-pool, got %s", v)
	}
	if v := d.Get("listener_id").(string); v != "listener-001" {
		t.Errorf("expected listener_id listener-001, got %s", v)
	}
	if v := d.Get("lb_algorithm").(string); v != "ROUND_ROBIN" {
		t.Errorf("expected lb_algorithm ROUND_ROBIN, got %s", v)
	}
	if v := d.Get("protocol").(string); v != "HTTP" {
		t.Errorf("expected protocol HTTP, got %s", v)
	}
	if v := d.Get("status").(string); v != "active" {
		t.Errorf("expected status active, got %s", v)
	}
	if v := d.Get("created_at").(string); v != "2025-01-15T10:00:00Z" {
		t.Errorf("expected created_at 2025-01-15T10:00:00Z, got %s", v)
	}

	members := d.Get("member").([]interface{})
	if len(members) != 1 {
		t.Fatalf("expected 1 member, got %d", len(members))
	}
	m := members[0].(map[string]interface{})
	if m["address"] != "10.0.0.10" {
		t.Errorf("expected member address 10.0.0.10, got %v", m["address"])
	}
	if m["protocol_port"] != 8080 {
		t.Errorf("expected member protocol_port 8080, got %v", m["protocol_port"])
	}
	if m["weight"] != 1 {
		t.Errorf("expected member weight 1, got %v", m["weight"])
	}
}
