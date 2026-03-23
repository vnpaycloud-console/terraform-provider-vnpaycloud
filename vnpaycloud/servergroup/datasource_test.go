package servergroup

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestDataSourceServerGroupRead_ByID(t *testing.T) {
	sgResp := dto.ServerGroupResponse{
		ServerGroup: dto.ServerGroup{
			ID:        "sg-ds-1",
			Name:      "my-server-group",
			Policy:    "anti-affinity",
			MemberIDs: []string{"inst-1"},
			CreatedAt: "2025-01-15T10:00:00Z",
			ProjectID: testhelpers.TestProjectID,
		},
	}

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.ServerGroupWithID(testhelpers.TestProjectID, "sg-ds-1"),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, sgResp),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := DataSourceServerGroup()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"id": "sg-ds-1",
	})

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "sg-ds-1" {
		t.Errorf("expected ID 'sg-ds-1', got '%s'", d.Id())
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

	memberIDs := d.Get("member_ids").([]interface{})
	if len(memberIDs) != 1 {
		t.Fatalf("expected 1 member_id, got %d", len(memberIDs))
	}
	if memberIDs[0].(string) != "inst-1" {
		t.Errorf("expected member_id 'inst-1', got '%s'", memberIDs[0])
	}
}

func TestDataSourceServerGroupRead_ByName(t *testing.T) {
	listResp := dto.ListServerGroupsResponse{
		ServerGroups: []dto.ServerGroup{
			{
				ID:        "sg-aaa",
				Name:      "group-alpha",
				Policy:    "affinity",
				MemberIDs: []string{},
				CreatedAt: "2025-01-10T08:00:00Z",
				ProjectID: testhelpers.TestProjectID,
			},
			{
				ID:        "sg-bbb",
				Name:      "group-beta",
				Policy:    "anti-affinity",
				MemberIDs: []string{"inst-x"},
				CreatedAt: "2025-01-12T09:00:00Z",
				ProjectID: testhelpers.TestProjectID,
			},
		},
	}

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.ServerGroups(testhelpers.TestProjectID),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, listResp),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := DataSourceServerGroup()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name": "group-beta",
	})

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "sg-bbb" {
		t.Errorf("expected ID 'sg-bbb', got '%s'", d.Id())
	}
	if got := d.Get("name").(string); got != "group-beta" {
		t.Errorf("expected name 'group-beta', got '%s'", got)
	}
	if got := d.Get("policy").(string); got != "anti-affinity" {
		t.Errorf("expected policy 'anti-affinity', got '%s'", got)
	}
}

func TestDataSourceServerGroupsRead(t *testing.T) {
	listResp := dto.ListServerGroupsResponse{
		ServerGroups: []dto.ServerGroup{
			{
				ID:        "sg-1",
				Name:      "group-one",
				Policy:    "affinity",
				MemberIDs: []string{"inst-a", "inst-b"},
				CreatedAt: "2025-01-10T08:00:00Z",
				ProjectID: testhelpers.TestProjectID,
			},
			{
				ID:        "sg-2",
				Name:      "group-two",
				Policy:    "anti-affinity",
				MemberIDs: []string{},
				CreatedAt: "2025-01-12T09:00:00Z",
				ProjectID: testhelpers.TestProjectID,
			},
		},
	}

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.ServerGroups(testhelpers.TestProjectID),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, listResp),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := DataSourceServerGroups()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{})

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	expectedID := "server-groups-" + testhelpers.TestProjectID
	if d.Id() != expectedID {
		t.Errorf("expected ID '%s', got '%s'", expectedID, d.Id())
	}

	serverGroups := d.Get("server_groups").([]interface{})
	if len(serverGroups) != 2 {
		t.Fatalf("expected 2 server_groups, got %d", len(serverGroups))
	}

	first := serverGroups[0].(map[string]interface{})
	if first["id"] != "sg-1" {
		t.Errorf("expected first server_group id 'sg-1', got '%s'", first["id"])
	}
	if first["name"] != "group-one" {
		t.Errorf("expected first server_group name 'group-one', got '%s'", first["name"])
	}
	if first["policy"] != "affinity" {
		t.Errorf("expected first server_group policy 'affinity', got '%s'", first["policy"])
	}

	second := serverGroups[1].(map[string]interface{})
	if second["id"] != "sg-2" {
		t.Errorf("expected second server_group id 'sg-2', got '%s'", second["id"])
	}
	if second["name"] != "group-two" {
		t.Errorf("expected second server_group name 'group-two', got '%s'", second["name"])
	}
}
