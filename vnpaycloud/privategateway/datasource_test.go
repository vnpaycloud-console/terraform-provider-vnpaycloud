package privategateway

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

func TestDataSourcePrivateGatewayRead_ByID(t *testing.T) {
	pgw := testPrivateGateway()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.PrivateGatewayWithID(testhelpers.TestProjectID, "pgw-001"),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.PrivateGatewayResponse{PrivateGateway: pgw}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourcePrivateGateway()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"id": "pgw-001",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "pgw-001" {
		t.Errorf("expected ID pgw-001, got %s", d.Id())
	}
	if v := d.Get("name").(string); v != "test-private-gateway" {
		t.Errorf("expected name test-private-gateway, got %s", v)
	}
	if v := d.Get("load_balancer_id").(string); v != "lb-001" {
		t.Errorf("expected load_balancer_id lb-001, got %s", v)
	}
	if v := d.Get("subnet_id").(string); v != "subnet-001" {
		t.Errorf("expected subnet_id subnet-001, got %s", v)
	}
}

func TestDataSourcePrivateGatewayRead_ByIDNameMatches(t *testing.T) {
	pgw := testPrivateGateway()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.PrivateGatewayWithID(testhelpers.TestProjectID, "pgw-001"),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.PrivateGatewayResponse{PrivateGateway: pgw}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourcePrivateGateway()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"id":   "pgw-001",
		"name": "test-private-gateway",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error when id+name match: %v", diags)
	}
	if d.Id() != "pgw-001" {
		t.Errorf("expected ID pgw-001, got %s", d.Id())
	}
}

func TestDataSourcePrivateGatewayRead_ByIDNameMismatch(t *testing.T) {
	pgw := testPrivateGateway()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.PrivateGatewayWithID(testhelpers.TestProjectID, "pgw-001"),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.PrivateGatewayResponse{PrivateGateway: pgw}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourcePrivateGateway()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"id":   "pgw-001",
		"name": "definitely-not-the-real-pgw-name",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if !diags.HasError() {
		t.Fatal("expected error when id is fetched but name does not match, got none")
	}
	if !strings.Contains(diags[0].Summary, "does not match name") {
		t.Errorf("expected mismatch error, got: %s", diags[0].Summary)
	}
}

func TestDataSourcePrivateGatewayRead_ByName(t *testing.T) {
	pgw := testPrivateGateway()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.PrivateGateways(testhelpers.TestProjectID),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListPrivateGatewaysResponse{
				PrivateGateways: []dto.PrivateGateway{pgw},
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourcePrivateGateway()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"name": "test-private-gateway",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
	if d.Id() != "pgw-001" {
		t.Errorf("expected ID pgw-001, got %s", d.Id())
	}
}

func TestDataSourcePrivateGatewayRead_NotFound(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.PrivateGateways(testhelpers.TestProjectID),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListPrivateGatewaysResponse{
				PrivateGateways: []dto.PrivateGateway{},
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourcePrivateGateway()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"name": "nonexistent",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if !diags.HasError() {
		t.Fatal("expected error for nonexistent private gateway, got none")
	}
}
