package pool

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// testPool returns a fully populated dto.Pool for use in tests.
func testPool() dto.Pool {
	return dto.Pool{
		ID:          "pool-001",
		Name:        "test-pool",
		ListenerID:  "listener-001",
		LBAlgorithm: "ROUND_ROBIN",
		Protocol:    "HTTP",
		Members: []dto.PoolMember{
			{
				ID:           "member-001",
				Name:         "member-1",
				Address:      "10.0.0.10",
				ProtocolPort: 8080,
				Weight:       1,
				Status:       "active",
			},
		},
		Status:    "active",
		CreatedAt: "2025-01-15T10:00:00Z",
	}
}

func TestResourcePoolCreate(t *testing.T) {
	p := testPool()
	// The create response returns pool without members (API creates pool first,
	// then members are added via PUT update).
	createPool := dto.Pool{
		ID:          "pool-001",
		Name:        "test-pool",
		ListenerID:  "listener-001",
		LBAlgorithm: "ROUND_ROBIN",
		Protocol:    "HTTP",
		Members:     nil,
		Status:      "active",
		CreatedAt:   "2025-01-15T10:00:00Z",
	}

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "POST",
			Pattern: "/v2/iac/projects/test-project-id/pools",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.PoolResponse{Pool: createPool}),
		},
		{
			Pattern: "/v2/iac/projects/test-project-id/pools/pool-001",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case "GET":
					testhelpers.JSONHandler(t, http.StatusOK, dto.PoolResponse{Pool: p})(w, r)
				case "PUT":
					// Members added via PUT after pool creation
					w.WriteHeader(http.StatusOK)
				default:
					http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				}
			},
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourcePool()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":         "test-pool",
		"listener_id":  "listener-001",
		"lb_algorithm": "ROUND_ROBIN",
		"protocol":     "HTTP",
		"member": []interface{}{
			map[string]interface{}{
				"address":       "10.0.0.10",
				"protocol_port": 8080,
				"weight":        1,
			},
		},
	})

	diags := res.CreateContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "pool-001" {
		t.Errorf("expected ID pool-001, got %s", d.Id())
	}
	if v := d.Get("name").(string); v != "test-pool" {
		t.Errorf("expected name test-pool, got %s", v)
	}
	if v := d.Get("status").(string); v != "active" {
		t.Errorf("expected status active, got %s", v)
	}
	if v := d.Get("lb_algorithm").(string); v != "ROUND_ROBIN" {
		t.Errorf("expected lb_algorithm ROUND_ROBIN, got %s", v)
	}
}

func TestResourcePoolCreate_NoMembers(t *testing.T) {
	p := dto.Pool{
		ID:          "pool-002",
		Name:        "test-pool-empty",
		ListenerID:  "listener-001",
		LBAlgorithm: "ROUND_ROBIN",
		Protocol:    "TCP",
		Members:     nil,
		Status:      "active",
		CreatedAt:   "2025-01-15T10:00:00Z",
	}

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "POST",
			Pattern: "/v2/iac/projects/test-project-id/pools",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.PoolResponse{Pool: p}),
		},
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/pools/pool-002",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.PoolResponse{Pool: p}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourcePool()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":         "test-pool-empty",
		"listener_id":  "listener-001",
		"lb_algorithm": "ROUND_ROBIN",
		"protocol":     "TCP",
		"member":       []interface{}{},
	})

	diags := res.CreateContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "pool-002" {
		t.Errorf("expected ID pool-002, got %s", d.Id())
	}
}

func TestResourcePoolRead(t *testing.T) {
	p := testPool()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/pools/pool-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.PoolResponse{Pool: p}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourcePool()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":         "",
		"listener_id":  "",
		"lb_algorithm": "ROUND_ROBIN",
		"protocol":     "HTTP",
		"member":       []interface{}{},
	})
	d.SetId("pool-001")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
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

func TestResourcePoolRead_NotFound(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/pools/pool-gone",
			Handler: testhelpers.EmptyHandler(http.StatusNotFound),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourcePool()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":         "",
		"listener_id":  "",
		"lb_algorithm": "ROUND_ROBIN",
		"protocol":     "HTTP",
		"member":       []interface{}{},
	})
	d.SetId("pool-gone")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error on 404: %v", diags)
	}

	if d.Id() != "" {
		t.Errorf("expected resource ID to be cleared after 404, got %s", d.Id())
	}
}

func TestResourcePoolDelete(t *testing.T) {
	p := testPool()
	deletedCalled := false

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Pattern: "/v2/iac/projects/test-project-id/pools/pool-001",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case "GET":
					if deletedCalled {
						w.WriteHeader(http.StatusNotFound)
						return
					}
					testhelpers.JSONHandler(t, http.StatusOK, dto.PoolResponse{Pool: p})(w, r)
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

	res := ResourcePool()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":         "test-pool",
		"listener_id":  "listener-001",
		"lb_algorithm": "ROUND_ROBIN",
		"protocol":     "HTTP",
		"member":       []interface{}{},
	})
	d.SetId("pool-001")

	diags := res.DeleteContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if !deletedCalled {
		t.Error("expected DELETE to have been called")
	}
}
