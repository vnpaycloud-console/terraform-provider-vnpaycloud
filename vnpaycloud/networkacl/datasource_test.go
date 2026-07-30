package networkacl

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestDataSourceNetworkACLRead_ByID(t *testing.T) {
	acl := testNetworkACL()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.NetworkACLWithID(testhelpers.TestProjectID, "nacl-001"),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.NetworkACLResponse{NetworkACL: acl}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceNetworkACL()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"id": "nacl-001",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "nacl-001" {
		t.Errorf("expected ID nacl-001, got %s", d.Id())
	}
	if v := d.Get("name").(string); v != "test-acl" {
		t.Errorf("expected name test-acl, got %s", v)
	}
	if v := d.Get("vpc_id").(string); v != "vpc-001" {
		t.Errorf("expected vpc_id vpc-001, got %s", v)
	}
}

func TestDataSourceNetworkACLRead_ByIDSelectorsMatch(t *testing.T) {
	acl := testNetworkACL()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.NetworkACLWithID(testhelpers.TestProjectID, "nacl-001"),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.NetworkACLResponse{NetworkACL: acl}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceNetworkACL()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"id":     "nacl-001",
		"name":   "test-acl",
		"vpc_id": "vpc-001",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error when id+name+vpc_id all match: %v", diags)
	}
	if d.Id() != "nacl-001" {
		t.Errorf("expected ID nacl-001, got %s", d.Id())
	}
}

func TestDataSourceNetworkACLRead_ByIDNameMismatch(t *testing.T) {
	acl := testNetworkACL()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.NetworkACLWithID(testhelpers.TestProjectID, "nacl-001"),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.NetworkACLResponse{NetworkACL: acl}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceNetworkACL()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"id":   "nacl-001",
		"name": "definitely-not-the-real-acl-name",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if !diags.HasError() {
		t.Fatal("expected error when id is fetched but name does not match, got none")
	}
	if !strings.Contains(diags[0].Summary, "does not match name") {
		t.Errorf("expected mismatch error, got: %s", diags[0].Summary)
	}
}

func TestDataSourceNetworkACLRead_ByIDVPCIDMismatch(t *testing.T) {
	acl := testNetworkACL()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.NetworkACLWithID(testhelpers.TestProjectID, "nacl-001"),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.NetworkACLResponse{NetworkACL: acl}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceNetworkACL()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"id":     "nacl-001",
		"vpc_id": "vpc-999-wrong",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if !diags.HasError() {
		t.Fatal("expected error when id is fetched but vpc_id does not match, got none")
	}
	if !strings.Contains(diags[0].Summary, "does not match vpc_id") {
		t.Errorf("expected mismatch error, got: %s", diags[0].Summary)
	}
}

func TestDataSourceNetworkACLRead_ByName(t *testing.T) {
	acl := testNetworkACL()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.NetworkACLs(testhelpers.TestProjectID),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListNetworkACLsResponse{
				NetworkACLs: []dto.NetworkACL{acl},
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceNetworkACL()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"name": "test-acl",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
	if d.Id() != "nacl-001" {
		t.Errorf("expected ID nacl-001, got %s", d.Id())
	}
}

func TestDataSourceNetworkACLRead_NotFound(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.NetworkACLs(testhelpers.TestProjectID),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListNetworkACLsResponse{
				NetworkACLs: []dto.NetworkACL{},
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceNetworkACL()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"name": "nonexistent",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if !diags.HasError() {
		t.Fatal("expected error for nonexistent network acl, got none")
	}
}
