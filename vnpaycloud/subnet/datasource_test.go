package subnet

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestDataSourceSubnetRead_ByID(t *testing.T) {
	sub := testSubnet()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/subnets/subnet-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.SubnetResponse{Subnet: sub}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceSubnet()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"id": "subnet-001",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "subnet-001" {
		t.Errorf("expected ID subnet-001, got %s", d.Id())
	}
	if v := d.Get("name").(string); v != "test-subnet" {
		t.Errorf("expected name test-subnet, got %s", v)
	}
	if v := d.Get("vpc_id").(string); v != "vpc-001" {
		t.Errorf("expected vpc_id vpc-001, got %s", v)
	}
	if v := d.Get("cidr").(string); v != "10.0.1.0/24" {
		t.Errorf("expected cidr 10.0.1.0/24, got %s", v)
	}
	if v := d.Get("gateway_ip").(string); v != "10.0.1.1" {
		t.Errorf("expected gateway_ip 10.0.1.1, got %s", v)
	}
	if v := d.Get("enable_dhcp").(bool); !v {
		t.Error("expected enable_dhcp true, got false")
	}
	if v := d.Get("enable_snat").(bool); v {
		t.Error("expected enable_snat false, got true")
	}
	if v := d.Get("status").(string); v != "active" {
		t.Errorf("expected status active, got %s", v)
	}
	if v := d.Get("created_at").(string); v != "2025-01-15T10:00:00Z" {
		t.Errorf("expected created_at 2025-01-15T10:00:00Z, got %s", v)
	}
}

func TestDataSourceSubnetRead_ByName(t *testing.T) {
	sub := testSubnet()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/subnets",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListSubnetsResponse{
				Subnets: []dto.Subnet{sub},
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceSubnet()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"name": "test-subnet",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "subnet-001" {
		t.Errorf("expected ID subnet-001, got %s", d.Id())
	}
	if v := d.Get("name").(string); v != "test-subnet" {
		t.Errorf("expected name test-subnet, got %s", v)
	}
}

func TestDataSourceSubnetRead_NotFound(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/subnets",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListSubnetsResponse{
				Subnets: []dto.Subnet{},
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceSubnet()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"name": "nonexistent",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if !diags.HasError() {
		t.Fatal("expected error for nonexistent subnet, got none")
	}
}

func TestDataSourceSubnetsRead(t *testing.T) {
	sub1 := testSubnet()
	sub2 := dto.Subnet{
		ID:           "subnet-002",
		Name:         "test-subnet-2",
		VpcID:        "vpc-001",
		CIDR:         "10.0.2.0/24",
		GatewayIP:    "10.0.2.1",
		EnableDHCP:   true,
		EnableSnat:   true,
		ExternalIpID: "fip-001",
		UsedByK8S:    true,
		Status:       "active",
		CreatedAt:    "2025-01-16T12:00:00Z",
		ProjectID:    testhelpers.TestProjectID,
		ZoneID:       testhelpers.TestZoneID,
	}

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/subnets",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListSubnetsResponse{
				Subnets: []dto.Subnet{sub1, sub2},
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceSubnets()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	subnets := d.Get("subnets").([]interface{})
	if len(subnets) != 2 {
		t.Fatalf("expected 2 subnets, got %d", len(subnets))
	}

	first := subnets[0].(map[string]interface{})
	if first["id"] != "subnet-001" {
		t.Errorf("expected first subnet id subnet-001, got %v", first["id"])
	}
	if first["name"] != "test-subnet" {
		t.Errorf("expected first subnet name test-subnet, got %v", first["name"])
	}

	second := subnets[1].(map[string]interface{})
	if second["id"] != "subnet-002" {
		t.Errorf("expected second subnet id subnet-002, got %v", second["id"])
	}
	if second["name"] != "test-subnet-2" {
		t.Errorf("expected second subnet name test-subnet-2, got %v", second["name"])
	}
}
