package vpcpeering

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestDataSourceVPCPeeringRead_ByID(t *testing.T) {
	peering := testPeeringConnection()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/peering-connections/peer-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.PeeringConnectionResponse{PeeringConnection: peering}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceVPCPeering()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"id": "peer-001",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "peer-001" {
		t.Errorf("expected ID peer-001, got %s", d.Id())
	}
	if v := d.Get("name").(string); v != "test-peering" {
		t.Errorf("expected name test-peering, got %s", v)
	}
	if v := d.Get("src_vpc_id").(string); v != "vpc-src-001" {
		t.Errorf("expected src_vpc_id vpc-src-001, got %s", v)
	}
	if v := d.Get("dest_vpc_id").(string); v != "vpc-dest-001" {
		t.Errorf("expected dest_vpc_id vpc-dest-001, got %s", v)
	}
	if v := d.Get("status").(string); v != "active" {
		t.Errorf("expected status active, got %s", v)
	}
	if v := d.Get("peering_status").(string); v != "established" {
		t.Errorf("expected peering_status established, got %s", v)
	}
	if v := d.Get("src_vpc_cidr").(string); v != "10.0.0.0/16" {
		t.Errorf("expected src_vpc_cidr 10.0.0.0/16, got %s", v)
	}
	if v := d.Get("dest_vpc_cidr").(string); v != "10.1.0.0/16" {
		t.Errorf("expected dest_vpc_cidr 10.1.0.0/16, got %s", v)
	}
	if v := d.Get("created_at").(string); v != "2025-01-15T10:00:00Z" {
		t.Errorf("expected created_at 2025-01-15T10:00:00Z, got %s", v)
	}
}

func TestDataSourceVPCPeeringRead_ByName(t *testing.T) {
	peering := testPeeringConnection()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method: "GET",
			Pattern: "/v2/iac/peering-connections",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListPeeringConnectionsResponse{
				PeeringConnections: []dto.PeeringConnection{peering},
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceVPCPeering()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"name": "test-peering",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "peer-001" {
		t.Errorf("expected ID peer-001, got %s", d.Id())
	}
	if v := d.Get("name").(string); v != "test-peering" {
		t.Errorf("expected name test-peering, got %s", v)
	}
}

func TestDataSourceVPCPeeringRead_NotFound(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method: "GET",
			Pattern: "/v2/iac/peering-connections",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListPeeringConnectionsResponse{
				PeeringConnections: []dto.PeeringConnection{},
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceVPCPeering()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"name": "nonexistent",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if !diags.HasError() {
		t.Fatal("expected error for nonexistent peering, got none")
	}
}

func TestDataSourceVPCPeeringsRead(t *testing.T) {
	peering1 := testPeeringConnection()
	peering2 := dto.PeeringConnection{
		ID:            "peer-002",
		Name:          "test-peering-2",
		Status:        "active",
		PeeringStatus: "established",
		SrcVpcID:      "vpc-src-002",
		SrcVpcCIDR:    "10.2.0.0/16",
		DestVpcID:     "vpc-dest-002",
		DestVpcCIDR:   "10.3.0.0/16",
		CreatedAt:     "2025-01-16T12:00:00Z",
	}

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method: "GET",
			Pattern: "/v2/iac/peering-connections",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListPeeringConnectionsResponse{
				PeeringConnections: []dto.PeeringConnection{peering1, peering2},
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceVPCPeerings()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	peerings := d.Get("vpc_peerings").([]interface{})
	if len(peerings) != 2 {
		t.Fatalf("expected 2 peerings, got %d", len(peerings))
	}

	first := peerings[0].(map[string]interface{})
	if first["id"] != "peer-001" {
		t.Errorf("expected first peering id peer-001, got %v", first["id"])
	}
	if first["name"] != "test-peering" {
		t.Errorf("expected first peering name test-peering, got %v", first["name"])
	}
	if first["src_vpc_id"] != "vpc-src-001" {
		t.Errorf("expected first peering src_vpc_id vpc-src-001, got %v", first["src_vpc_id"])
	}
	if first["dest_vpc_id"] != "vpc-dest-001" {
		t.Errorf("expected first peering dest_vpc_id vpc-dest-001, got %v", first["dest_vpc_id"])
	}
	if first["status"] != "active" {
		t.Errorf("expected first peering status active, got %v", first["status"])
	}
	if first["peering_status"] != "established" {
		t.Errorf("expected first peering peering_status established, got %v", first["peering_status"])
	}

	second := peerings[1].(map[string]interface{})
	if second["id"] != "peer-002" {
		t.Errorf("expected second peering id peer-002, got %v", second["id"])
	}
	if second["name"] != "test-peering-2" {
		t.Errorf("expected second peering name test-peering-2, got %v", second["name"])
	}
}
