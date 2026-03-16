package healthmonitor

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestDataSourceHealthMonitorRead_ByID(t *testing.T) {
	hm := testHealthMonitor()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/health-monitors/hm-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.HealthMonitorResponse{HealthMonitor: hm}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceHealthMonitor()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"id": "hm-001",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "hm-001" {
		t.Errorf("expected ID hm-001, got %s", d.Id())
	}
	if v := d.Get("pool_id").(string); v != "pool-001" {
		t.Errorf("expected pool_id pool-001, got %s", v)
	}
	if v := d.Get("type").(string); v != "HTTP" {
		t.Errorf("expected type HTTP, got %s", v)
	}
	if v := d.Get("delay").(int); v != 5 {
		t.Errorf("expected delay 5, got %d", v)
	}
	if v := d.Get("timeout").(int); v != 10 {
		t.Errorf("expected timeout 10, got %d", v)
	}
	if v := d.Get("max_retries").(int); v != 3 {
		t.Errorf("expected max_retries 3, got %d", v)
	}
	if v := d.Get("http_method").(string); v != "GET" {
		t.Errorf("expected http_method GET, got %s", v)
	}
	if v := d.Get("url_path").(string); v != "/health" {
		t.Errorf("expected url_path /health, got %s", v)
	}
	if v := d.Get("expected_codes").(string); v != "200" {
		t.Errorf("expected expected_codes 200, got %s", v)
	}
	if v := d.Get("status").(string); v != "active" {
		t.Errorf("expected status active, got %s", v)
	}
}
