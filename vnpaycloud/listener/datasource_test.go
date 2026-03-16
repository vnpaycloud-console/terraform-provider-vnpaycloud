package listener

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestDataSourceListenerRead_ByID(t *testing.T) {
	lis := testListener()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/listeners/listener-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListenerResponse{Listener: lis}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceListener()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"id": "listener-001",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "listener-001" {
		t.Errorf("expected ID listener-001, got %s", d.Id())
	}
	if v := d.Get("name").(string); v != "test-listener" {
		t.Errorf("expected name test-listener, got %s", v)
	}
	if v := d.Get("load_balancer_id").(string); v != "lb-001" {
		t.Errorf("expected load_balancer_id lb-001, got %s", v)
	}
	if v := d.Get("protocol").(string); v != "HTTP" {
		t.Errorf("expected protocol HTTP, got %s", v)
	}
	if v := d.Get("protocol_port").(int); v != 80 {
		t.Errorf("expected protocol_port 80, got %d", v)
	}
	if v := d.Get("default_pool_id").(string); v != "pool-001" {
		t.Errorf("expected default_pool_id pool-001, got %s", v)
	}
	if v := d.Get("status").(string); v != "active" {
		t.Errorf("expected status active, got %s", v)
	}
	if v := d.Get("created_at").(string); v != "2025-01-15T10:00:00Z" {
		t.Errorf("expected created_at 2025-01-15T10:00:00Z, got %s", v)
	}
}
