package healthmonitor

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// testHealthMonitor returns a fully populated dto.HealthMonitor for use in tests.
func testHealthMonitor() dto.HealthMonitor {
	return dto.HealthMonitor{
		ID:            "hm-001",
		PoolID:        "pool-001",
		Type:          "HTTP",
		Delay:         5,
		Timeout:       10,
		MaxRetries:    3,
		HTTPMethod:    "GET",
		URLPath:       "/health",
		ExpectedCodes: "200",
		Status:        "active",
	}
}

func TestResourceHealthMonitorCreate(t *testing.T) {
	hm := testHealthMonitor()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "POST",
			Pattern: "/v2/iac/projects/test-project-id/health-monitors",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.HealthMonitorResponse{HealthMonitor: hm}),
		},
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/health-monitors/hm-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.HealthMonitorResponse{HealthMonitor: hm}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceHealthMonitor()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"pool_id":        "pool-001",
		"type":           "HTTP",
		"delay":          5,
		"timeout":        10,
		"max_retries":    3,
		"http_method":    "GET",
		"url_path":       "/health",
		"expected_codes": "200",
	})

	diags := res.CreateContext(context.Background(), d, cfg)
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
	if v := d.Get("status").(string); v != "active" {
		t.Errorf("expected status active, got %s", v)
	}
}

func TestResourceHealthMonitorCreate_TCPMinimal(t *testing.T) {
	hm := dto.HealthMonitor{
		ID:         "hm-002",
		PoolID:     "pool-002",
		Type:       "TCP",
		Delay:      10,
		Timeout:    5,
		MaxRetries: 2,
		Status:     "active",
	}

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "POST",
			Pattern: "/v2/iac/projects/test-project-id/health-monitors",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.HealthMonitorResponse{HealthMonitor: hm}),
		},
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/health-monitors/hm-002",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.HealthMonitorResponse{HealthMonitor: hm}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceHealthMonitor()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"pool_id":        "pool-002",
		"type":           "TCP",
		"delay":          10,
		"timeout":        5,
		"max_retries":    2,
		"http_method":    "",
		"url_path":       "",
		"expected_codes": "",
	})

	diags := res.CreateContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "hm-002" {
		t.Errorf("expected ID hm-002, got %s", d.Id())
	}
	if v := d.Get("type").(string); v != "TCP" {
		t.Errorf("expected type TCP, got %s", v)
	}
}

func TestResourceHealthMonitorRead(t *testing.T) {
	hm := testHealthMonitor()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/health-monitors/hm-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.HealthMonitorResponse{HealthMonitor: hm}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceHealthMonitor()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"pool_id":        "",
		"type":           "HTTP",
		"delay":          0,
		"timeout":        0,
		"max_retries":    0,
		"http_method":    "",
		"url_path":       "",
		"expected_codes": "",
	})
	d.SetId("hm-001")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
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

func TestResourceHealthMonitorRead_NotFound(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/health-monitors/hm-gone",
			Handler: testhelpers.EmptyHandler(http.StatusNotFound),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceHealthMonitor()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"pool_id":        "",
		"type":           "HTTP",
		"delay":          0,
		"timeout":        0,
		"max_retries":    0,
		"http_method":    "",
		"url_path":       "",
		"expected_codes": "",
	})
	d.SetId("hm-gone")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error on 404: %v", diags)
	}

	if d.Id() != "" {
		t.Errorf("expected resource ID to be cleared after 404, got %s", d.Id())
	}
}

func TestResourceHealthMonitorDelete(t *testing.T) {
	hm := testHealthMonitor()
	deletedCalled := false

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Pattern: "/v2/iac/projects/test-project-id/health-monitors/hm-001",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case "GET":
					if deletedCalled {
						w.WriteHeader(http.StatusNotFound)
						return
					}
					testhelpers.JSONHandler(t, http.StatusOK, dto.HealthMonitorResponse{HealthMonitor: hm})(w, r)
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

	res := ResourceHealthMonitor()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"pool_id":        "pool-001",
		"type":           "HTTP",
		"delay":          5,
		"timeout":        10,
		"max_retries":    3,
		"http_method":    "GET",
		"url_path":       "/health",
		"expected_codes": "200",
	})
	d.SetId("hm-001")

	diags := res.DeleteContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if !deletedCalled {
		t.Error("expected DELETE to have been called")
	}
}
