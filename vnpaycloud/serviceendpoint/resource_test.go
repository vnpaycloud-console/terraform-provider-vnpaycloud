package serviceendpoint

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestResourceServiceEndpointSchema(t *testing.T) {
	res := ResourceServiceEndpoint()

	for _, key := range []string{"provider_id", "service_id", "service_gateway_id", "port"} {
		if !res.Schema[key].Required || !res.Schema[key].ForceNew {
			t.Errorf("expected %s to be Required+ForceNew", key)
		}
	}
	if !res.Schema["name"].Required {
		t.Error("expected name to be required")
	}
}

func TestResourceServiceEndpointRead(t *testing.T) {
	se := dto.ServiceEndpoint{
		ID:               "se-001",
		Name:             "tf-endpoint",
		Description:      "terraform endpoint",
		ProviderID:       "prov-001",
		ServiceID:        "svc-001",
		ServiceGatewayID: "sgw-001",
		Port:             443,
		AllowedCIDRs:     []string{"10.0.0.0/24"},
		Status:           "active",
		CreatedAt:        "2026-01-15T10:00:00Z",
	}

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.ServiceEndpointWithID(testhelpers.TestProjectID, "se-001"),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ServiceEndpointResponse{ServiceEndpoint: se}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceServiceEndpoint()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{})
	d.SetId("se-001")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if v := d.Get("name").(string); v != "tf-endpoint" {
		t.Errorf("expected name tf-endpoint, got %s", v)
	}
	if v := d.Get("service_gateway_id").(string); v != "sgw-001" {
		t.Errorf("expected service_gateway_id sgw-001, got %s", v)
	}
	if v := d.Get("port").(int); v != 443 {
		t.Errorf("expected port 443, got %d", v)
	}
	cidrs := d.Get("allowed_cidrs").([]interface{})
	if len(cidrs) != 1 || cidrs[0].(string) != "10.0.0.0/24" {
		t.Errorf("expected allowed_cidrs [10.0.0.0/24], got %v", cidrs)
	}
}
