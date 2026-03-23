package servergroup

import (
	"context"
	"net/http"
	"sync/atomic"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestResourceServerGroupCreate(t *testing.T) {
	createResp := dto.ServerGroupResponse{
		ServerGroup: dto.ServerGroup{
			ID:        "sg-123",
			Name:      "my-server-group",
			Policy:    "anti-affinity",
			MemberIDs: []string{},
			CreatedAt: "2025-01-15T10:00:00Z",
			ProjectID: testhelpers.TestProjectID,
		},
	}

	// After create, the state poller calls GET to check status.
	// Then create calls resourceServerGroupRead which also calls GET.
	// Return the active server group for all GET calls.
	readResp := dto.ServerGroupResponse{
		ServerGroup: dto.ServerGroup{
			ID:        "sg-123",
			Name:      "my-server-group",
			Policy:    "anti-affinity",
			MemberIDs: []string{},
			CreatedAt: "2025-01-15T10:00:00Z",
			ProjectID: testhelpers.TestProjectID,
		},
	}

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "POST",
			Pattern: client.ApiPath.ServerGroups(testhelpers.TestProjectID),
			Handler: testhelpers.JSONHandler(t, http.StatusCreated, createResp),
		},
		{
			Method:  "GET",
			Pattern: client.ApiPath.ServerGroupWithID(testhelpers.TestProjectID, "sg-123"),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, readResp),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceServerGroup()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":   "my-server-group",
		"policy": "anti-affinity",
	})

	diags := res.CreateContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "sg-123" {
		t.Errorf("expected ID 'sg-123', got '%s'", d.Id())
	}
	if got := d.Get("name").(string); got != "my-server-group" {
		t.Errorf("expected name 'my-server-group', got '%s'", got)
	}
	if got := d.Get("policy").(string); got != "anti-affinity" {
		t.Errorf("expected policy 'anti-affinity', got '%s'", got)
	}
	if got := d.Get("created_at").(string); got != "2025-01-15T10:00:00Z" {
		t.Errorf("expected created_at '2025-01-15T10:00:00Z', got '%s'", got)
	}
}

func TestResourceServerGroupRead(t *testing.T) {
	readResp := dto.ServerGroupResponse{
		ServerGroup: dto.ServerGroup{
			ID:        "sg-123",
			Name:      "my-server-group",
			Policy:    "affinity",
			MemberIDs: []string{"instance-1", "instance-2"},
			CreatedAt: "2025-01-15T10:00:00Z",
			ProjectID: testhelpers.TestProjectID,
		},
	}

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.ServerGroupWithID(testhelpers.TestProjectID, "sg-123"),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, readResp),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceServerGroup()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":   "my-server-group",
		"policy": "affinity",
	})
	d.SetId("sg-123")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "sg-123" {
		t.Errorf("expected ID 'sg-123', got '%s'", d.Id())
	}
	if got := d.Get("name").(string); got != "my-server-group" {
		t.Errorf("expected name 'my-server-group', got '%s'", got)
	}
	if got := d.Get("policy").(string); got != "affinity" {
		t.Errorf("expected policy 'affinity', got '%s'", got)
	}

	memberIDs := d.Get("member_ids").([]interface{})
	if len(memberIDs) != 2 {
		t.Fatalf("expected 2 member_ids, got %d", len(memberIDs))
	}
	if memberIDs[0].(string) != "instance-1" {
		t.Errorf("expected first member_id 'instance-1', got '%s'", memberIDs[0])
	}
	if memberIDs[1].(string) != "instance-2" {
		t.Errorf("expected second member_id 'instance-2', got '%s'", memberIDs[1])
	}
}

func TestResourceServerGroupRead_NotFound(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.ServerGroupWithID(testhelpers.TestProjectID, "sg-gone"),
			Handler: testhelpers.EmptyHandler(http.StatusNotFound),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceServerGroup()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":   "gone-group",
		"policy": "anti-affinity",
	})
	d.SetId("sg-gone")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "" {
		t.Errorf("expected resource ID to be cleared on 404, got '%s'", d.Id())
	}
}

func TestResourceServerGroupDelete(t *testing.T) {
	readResp := dto.ServerGroupResponse{
		ServerGroup: dto.ServerGroup{
			ID:        "sg-123",
			Name:      "my-server-group",
			Policy:    "anti-affinity",
			MemberIDs: []string{},
			CreatedAt: "2025-01-15T10:00:00Z",
			ProjectID: testhelpers.TestProjectID,
		},
	}

	// Delete flow:
	// 1. GET to confirm resource exists
	// 2. DELETE the resource
	// 3. State poller: GET until 404 (returns "deleted" state)
	//
	// Use an atomic counter so the first GET returns the resource,
	// the DELETE succeeds, and subsequent GETs return 404.
	//
	// Because http.ServeMux does not allow the same pattern to be
	// registered twice (even with different methods), we use a single
	// handler that dispatches on the HTTP method.
	var getCount atomic.Int32

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			// Method intentionally left empty so both GET and DELETE match.
			Pattern: client.ApiPath.ServerGroupWithID(testhelpers.TestProjectID, "sg-123"),
			Handler: func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case http.MethodGet:
					n := getCount.Add(1)
					if n == 1 {
						// First GET: resource exists (pre-delete check)
						testhelpers.JSONHandler(t, http.StatusOK, readResp)(w, r)
					} else {
						// Subsequent GETs: resource deleted (state poller)
						testhelpers.EmptyHandler(http.StatusNotFound)(w, r)
					}
				case http.MethodDelete:
					testhelpers.EmptyHandler(http.StatusNoContent)(w, r)
				default:
					http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				}
			},
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceServerGroup()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":   "my-server-group",
		"policy": "anti-affinity",
	})
	d.SetId("sg-123")

	diags := res.DeleteContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
}
