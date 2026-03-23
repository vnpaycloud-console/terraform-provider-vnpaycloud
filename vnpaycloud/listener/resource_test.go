package listener

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// testListener returns a fully populated dto.Listener for use in tests.
func testListener() dto.Listener {
	return dto.Listener{
		ID:             "listener-001",
		Name:           "test-listener",
		LoadBalancerID: "lb-001",
		Protocol:       "HTTP",
		ProtocolPort:   80,
		DefaultPoolID:  "pool-001",
		Status:         "active",
		CreatedAt:      "2025-01-15T10:00:00Z",
	}
}

func TestResourceListenerCreate(t *testing.T) {
	lis := testListener()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "POST",
			Pattern: "/v2/iac/projects/test-project-id/listeners",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListenerResponse{Listener: lis}),
		},
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/listeners/listener-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListenerResponse{Listener: lis}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceListener()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":             "test-listener",
		"load_balancer_id": "lb-001",
		"protocol":         "HTTP",
		"protocol_port":    80,
		"default_pool_id":  "pool-001",
	})

	diags := res.CreateContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "listener-001" {
		t.Errorf("expected ID listener-001, got %s", d.Id())
	}
	if v := d.Get("name").(string); v != "test-listener" {
		t.Errorf("expected name test-listener, got %s", v)
	}
	if v := d.Get("status").(string); v != "active" {
		t.Errorf("expected status active, got %s", v)
	}
	if v := d.Get("protocol").(string); v != "HTTP" {
		t.Errorf("expected protocol HTTP, got %s", v)
	}
	if v := d.Get("protocol_port").(int); v != 80 {
		t.Errorf("expected protocol_port 80, got %d", v)
	}
}

func TestResourceListenerRead(t *testing.T) {
	lis := testListener()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/listeners/listener-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListenerResponse{Listener: lis}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceListener()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":             "",
		"load_balancer_id": "",
		"protocol":         "HTTP",
		"protocol_port":    0,
		"default_pool_id":  "",
	})
	d.SetId("listener-001")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
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

func TestResourceListenerRead_NotFound(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/listeners/listener-gone",
			Handler: testhelpers.EmptyHandler(http.StatusNotFound),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceListener()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":             "",
		"load_balancer_id": "",
		"protocol":         "HTTP",
		"protocol_port":    0,
		"default_pool_id":  "",
	})
	d.SetId("listener-gone")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error on 404: %v", diags)
	}

	if d.Id() != "" {
		t.Errorf("expected resource ID to be cleared after 404, got %s", d.Id())
	}
}

func TestResourceListenerDelete(t *testing.T) {
	lis := testListener()
	deletedCalled := false

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Pattern: "/v2/iac/projects/test-project-id/listeners/listener-001",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case "GET":
					if deletedCalled {
						w.WriteHeader(http.StatusNotFound)
						return
					}
					testhelpers.JSONHandler(t, http.StatusOK, dto.ListenerResponse{Listener: lis})(w, r)
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

	res := ResourceListener()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":             "test-listener",
		"load_balancer_id": "lb-001",
		"protocol":         "HTTP",
		"protocol_port":    80,
		"default_pool_id":  "",
	})
	d.SetId("listener-001")

	diags := res.DeleteContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if !deletedCalled {
		t.Error("expected DELETE to have been called")
	}
}
