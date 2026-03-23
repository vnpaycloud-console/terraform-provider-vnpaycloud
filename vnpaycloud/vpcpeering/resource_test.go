package vpcpeering

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// testPeeringConnection returns a fully populated dto.PeeringConnection for use in tests.
func testPeeringConnection() dto.PeeringConnection {
	return dto.PeeringConnection{
		ID:            "peer-001",
		Name:          "test-peering",
		Status:        "active",
		PeeringStatus: "established",
		SrcVpcID:      "vpc-src-001",
		SrcVpcCIDR:    "10.0.0.0/16",
		DestVpcID:     "vpc-dest-001",
		DestVpcCIDR:   "10.1.0.0/16",
		CreatedAt:     "2025-01-15T10:00:00Z",
	}
}

// testReversePeeringConnection returns the reverse direction peering for bidirectional tests.
func testReversePeeringConnection() dto.PeeringConnection {
	return dto.PeeringConnection{
		ID:            "peer-002",
		Name:          "test-peering-reverse",
		Status:        "active",
		PeeringStatus: "established",
		SrcVpcID:      "vpc-dest-001",
		SrcVpcCIDR:    "10.1.0.0/16",
		DestVpcID:     "vpc-src-001",
		DestVpcCIDR:   "10.0.0.0/16",
		CreatedAt:     "2025-01-15T10:00:00Z",
	}
}

func TestResourceVPCPeeringCreate(t *testing.T) {
	primary := testPeeringConnection()
	reverse := testReversePeeringConnection()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Pattern: "/v2/iac/peering-connections",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case "POST":
					testhelpers.JSONHandler(t, http.StatusOK, dto.ListPeeringConnectionsResponse{
						PeeringConnections: []dto.PeeringConnection{primary, reverse},
					})(w, r)
				case "GET":
					testhelpers.JSONHandler(t, http.StatusOK, dto.ListPeeringConnectionsResponse{
						PeeringConnections: []dto.PeeringConnection{primary, reverse},
					})(w, r)
				default:
					http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				}
			},
		},
		{
			Method:  "GET",
			Pattern: "/v2/iac/peering-connections/peer-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.PeeringConnectionResponse{PeeringConnection: primary}),
		},
		{
			Method:  "GET",
			Pattern: "/v2/iac/peering-connections/peer-002",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.PeeringConnectionResponse{PeeringConnection: reverse}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceVPCPeering()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":        "",
		"src_vpc_id":  "vpc-src-001",
		"dest_vpc_id": "vpc-dest-001",
		"description": "test peering connection",
	})

	diags := res.CreateContext(context.Background(), d, cfg)
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

func TestResourceVPCPeeringRead(t *testing.T) {
	peering := testPeeringConnection()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/peering-connections/peer-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.PeeringConnectionResponse{PeeringConnection: peering}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceVPCPeering()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":        "",
		"src_vpc_id":  "",
		"dest_vpc_id": "",
		"description": "",
	})
	d.SetId("peer-001")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
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

func TestResourceVPCPeeringRead_NotFound(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/peering-connections/peer-gone",
			Handler: testhelpers.EmptyHandler(http.StatusNotFound),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceVPCPeering()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":        "",
		"src_vpc_id":  "",
		"dest_vpc_id": "",
		"description": "",
	})
	d.SetId("peer-gone")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error on 404: %v", diags)
	}

	if d.Id() != "" {
		t.Errorf("expected resource ID to be cleared after 404, got %s", d.Id())
	}
}

func TestResourceVPCPeeringDelete(t *testing.T) {
	peering := testPeeringConnection()
	primaryDeleteCalled := false
	reverseDeleteCalled := false

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Pattern: "/v2/iac/peering-connections/peer-001",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case "GET":
					if primaryDeleteCalled {
						w.WriteHeader(http.StatusNotFound)
						return
					}
					testhelpers.JSONHandler(t, http.StatusOK, dto.PeeringConnectionResponse{PeeringConnection: peering})(w, r)
				case "DELETE":
					primaryDeleteCalled = true
					w.WriteHeader(http.StatusAccepted)
				default:
					http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				}
			},
		},
		{
			Pattern: "/v2/iac/peering-connections/peer-002",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				reverse := testReversePeeringConnection()
				switch r.Method {
				case "GET":
					if reverseDeleteCalled {
						w.WriteHeader(http.StatusNotFound)
						return
					}
					testhelpers.JSONHandler(t, http.StatusOK, dto.PeeringConnectionResponse{PeeringConnection: reverse})(w, r)
				case "DELETE":
					reverseDeleteCalled = true
					w.WriteHeader(http.StatusAccepted)
				default:
					http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				}
			},
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceVPCPeering()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":        "test-peering",
		"src_vpc_id":  "vpc-src-001",
		"dest_vpc_id": "vpc-dest-001",
		"description": "",
	})
	d.SetId("peer-001")
	d.Set("reverse_peering_id", "peer-002")

	diags := res.DeleteContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if !primaryDeleteCalled {
		t.Error("expected primary DELETE to have been called")
	}
	if !reverseDeleteCalled {
		t.Error("expected reverse DELETE to have been called")
	}
}
