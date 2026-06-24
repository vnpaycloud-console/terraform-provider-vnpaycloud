package networkacl

import (
	"context"
	"net/http"
	"sync/atomic"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func testNetworkACL() dto.NetworkACL {
	return dto.NetworkACL{
		ID:          "nacl-001",
		Name:        "test-acl",
		Description: "a test network acl",
		VpcID:       "vpc-001",
		SubnetIDs:   []string{},
		TotalRules:  0,
		Status:      "active",
		CreatedAt:   "2025-01-15T10:00:00Z",
		ProjectID:   testhelpers.TestProjectID,
		ZoneID:      testhelpers.TestZoneID,
	}
}

func TestResourceNetworkACLCreate(t *testing.T) {
	acl := testNetworkACL()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "POST",
			Pattern: "/v2/iac/projects/test-project-id/network-acls",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.NetworkACLResponse{NetworkACL: acl}),
		},
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/network-acls/nacl-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.NetworkACLResponse{NetworkACL: acl}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceNetworkACL()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":        "test-acl",
		"vpc_id":      "vpc-001",
		"description": "a test network acl",
	})

	diags := res.CreateContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
	if d.Id() != "nacl-001" {
		t.Errorf("expected ID nacl-001, got %s", d.Id())
	}
	if v := d.Get("vpc_id").(string); v != "vpc-001" {
		t.Errorf("expected vpc_id vpc-001, got %s", v)
	}
}

func TestResourceNetworkACLCreate_WithSubnetMapping(t *testing.T) {
	acl := testNetworkACL()
	mapped := acl
	mapped.SubnetIDs = []string{"subnet-001"}

	var mapCalled bool

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "POST",
			Pattern: "/v2/iac/projects/test-project-id/network-acls",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.NetworkACLResponse{NetworkACL: acl}),
		},
		{
			Method:  "PUT",
			Pattern: "/v2/iac/projects/test-project-id/network-acls/nacl-001/networks/subnet-001",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				mapCalled = true
				w.WriteHeader(http.StatusOK)
			},
		},
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/network-acls/nacl-001",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				resp := acl
				if mapCalled {
					resp = mapped
				}
				testhelpers.JSONHandler(t, http.StatusOK, dto.NetworkACLResponse{NetworkACL: resp})(w, r)
			},
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceNetworkACL()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":       "test-acl",
		"vpc_id":     "vpc-001",
		"subnet_ids": []interface{}{"subnet-001"},
	})

	diags := res.CreateContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
	if !mapCalled {
		t.Fatal("expected PUT map subnet to have been called")
	}
	if d.Get("subnet_ids").(*schema.Set).Len() != 1 {
		t.Errorf("expected 1 subnet_id, got %d", d.Get("subnet_ids").(*schema.Set).Len())
	}
}

func TestResourceNetworkACLRead(t *testing.T) {
	acl := testNetworkACL()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/network-acls/nacl-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.NetworkACLResponse{NetworkACL: acl}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceNetworkACL()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{})
	d.SetId("nacl-001")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
	if v := d.Get("name").(string); v != "test-acl" {
		t.Errorf("expected name test-acl, got %s", v)
	}
}

func TestResourceNetworkACLRead_NotFound(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/network-acls/nacl-gone",
			Handler: testhelpers.EmptyHandler(http.StatusNotFound),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceNetworkACL()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{})
	d.SetId("nacl-gone")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error on 404: %v", diags)
	}
	if d.Id() != "" {
		t.Errorf("expected ID cleared after 404, got %s", d.Id())
	}
}

func TestResourceNetworkACLDelete(t *testing.T) {
	acl := testNetworkACL()
	deleteCalled := false
	var getCalls int32

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Pattern: "/v2/iac/projects/test-project-id/network-acls/nacl-001",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case http.MethodDelete:
					deleteCalled = true
					w.WriteHeader(http.StatusNoContent)
				case http.MethodGet:
					// First GET (pre-delete read) returns the ACL; later polls return 404.
					if atomic.AddInt32(&getCalls, 1) == 1 {
						testhelpers.JSONHandler(t, http.StatusOK, dto.NetworkACLResponse{NetworkACL: acl})(w, r)
						return
					}
					w.WriteHeader(http.StatusNotFound)
				default:
					http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				}
			},
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceNetworkACL()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{})
	d.SetId("nacl-001")

	diags := res.DeleteContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
	if !deleteCalled {
		t.Error("expected DELETE to have been called")
	}
}
