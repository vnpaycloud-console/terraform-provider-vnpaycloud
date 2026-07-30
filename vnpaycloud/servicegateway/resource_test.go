package servicegateway

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestResourceServiceGatewaySchema(t *testing.T) {
	res := ResourceServiceGateway()

	if !res.Schema["subnet_id"].Required || !res.Schema["subnet_id"].ForceNew {
		t.Error("expected subnet_id to be Required+ForceNew")
	}
	for _, key := range []string{"name", "flavor_id"} {
		if !res.Schema[key].Required {
			t.Errorf("expected %s to be required", key)
		}
	}
}

func TestResourceServiceGatewayRead(t *testing.T) {
	sg := dto.ServiceGateway{
		ID:             "sgw-001",
		Name:           "tf-sgw",
		Description:    "terraform service gateway",
		VPCID:          "vpc-001",
		SubnetID:       "subnet-001",
		FlavorID:       "t2-small",
		LoadBalancerID: "lb-001",
		VipAddress:     "10.0.0.5",
		AllowedICMP:    true,
		Status:         "active",
		CreatedAt:      "2026-01-15T10:00:00Z",
	}

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.ServiceGatewayWithID(testhelpers.TestProjectID, "sgw-001"),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ServiceGatewayResponse{ServiceGateway: sg}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceServiceGateway()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{})
	d.SetId("sgw-001")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if v := d.Get("name").(string); v != "tf-sgw" {
		t.Errorf("expected name tf-sgw, got %s", v)
	}
	if v := d.Get("subnet_id").(string); v != "subnet-001" {
		t.Errorf("expected subnet_id subnet-001, got %s", v)
	}
	if v := d.Get("flavor_id").(string); v != "t2-small" {
		t.Errorf("expected flavor_id t2-small, got %s", v)
	}
	if v := d.Get("allowed_icmp").(bool); !v {
		t.Error("expected allowed_icmp true")
	}
	if v := d.Get("load_balancer_id").(string); v != "lb-001" {
		t.Errorf("expected load_balancer_id lb-001, got %s", v)
	}
}
