package internetgatewayacl

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestDataSourceInternetGatewayACLRead(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: blacklistPath,
			Handler: testhelpers.JSONHandler(t, http.StatusOK, testBlacklist("1.2.3.0/24", "8.8.8.8/32")),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceInternetGatewayACL()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"internet_gateway_id": "igw-001",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "igw-001" {
		t.Errorf("expected ID igw-001, got %s", d.Id())
	}
	if v := d.Get("total_match").(int); v != 2 {
		t.Errorf("expected total_match 2, got %d", v)
	}
	if match := d.Get("match").(*schema.Set); !match.Contains("8.8.8.8/32") {
		t.Errorf("expected match to contain 8.8.8.8/32, got %v", match.List())
	}
	if v := d.Get("zone_id").(string); v != testhelpers.TestZoneID {
		t.Errorf("expected zone_id %s, got %s", testhelpers.TestZoneID, v)
	}
}

func TestDataSourceInternetGatewayACLReadEmpty(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: blacklistPath,
			Handler: testhelpers.JSONHandler(t, http.StatusOK, testBlacklist()),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceInternetGatewayACL()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"internet_gateway_id": "igw-001",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if v := d.Get("total_match").(int); v != 0 {
		t.Errorf("expected total_match 0, got %d", v)
	}
	if v := d.Get("match").(*schema.Set).Len(); v != 0 {
		t.Errorf("expected empty match, got %d entries", v)
	}
}
