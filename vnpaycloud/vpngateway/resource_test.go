package vpngateway

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func testVPNGateway() dto.VPNGateway {
	return dto.VPNGateway{
		ID:             "vpn-gw-001",
		Name:           "test-vpn-gateway",
		Description:    "a test vpn gateway",
		VPNType:        "ROUTE_BASED",
		Status:         "active",
		AttachedVPCIDs: []string{"vpc-001"},
		CreatedAt:      "2025-01-15T10:00:00Z",
		ProjectID:      testhelpers.TestProjectID,
		ZoneID:         testhelpers.TestZoneID,
	}
}

func TestResourceVPNGatewayCreate(t *testing.T) {
	gw := testVPNGateway()

	var gotReq dto.CreateVPNGatewayRequest
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "POST",
			Pattern: client.ApiPath.VPNGateways(testhelpers.TestProjectID),
			Handler: func(w http.ResponseWriter, r *http.Request) {
				if err := json.NewDecoder(r.Body).Decode(&gotReq); err != nil {
					t.Errorf("failed to decode create request: %v", err)
				}
				testhelpers.JSONHandler(t, http.StatusOK, dto.VPNGatewayResponse{VPNGateway: gw})(w, r)
			},
		},
		{
			Method:  "GET",
			Pattern: client.ApiPath.VPNGatewayWithID(testhelpers.TestProjectID, "vpn-gw-001"),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.VPNGatewayResponse{VPNGateway: gw}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceVPNGateway()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":        "test-vpn-gateway",
		"description": "a test vpn gateway",
		"vpn_type":    "ROUTE_BASED",
		"zone_id":     testhelpers.TestZoneID,
	})

	diags := res.CreateContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if gotReq.Name != "test-vpn-gateway" {
		t.Errorf("expected request name test-vpn-gateway, got %s", gotReq.Name)
	}
	if gotReq.Description != "a test vpn gateway" {
		t.Errorf("expected request description a test vpn gateway, got %s", gotReq.Description)
	}
	if gotReq.VPNType != "ROUTE_BASED" {
		t.Errorf("expected request vpn_type ROUTE_BASED, got %s", gotReq.VPNType)
	}
	assertVPNGatewayResourceData(t, d, gw)
}

func TestResourceVPNGatewayRead(t *testing.T) {
	gw := testVPNGateway()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.VPNGatewayWithID(testhelpers.TestProjectID, "vpn-gw-001"),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.VPNGatewayResponse{VPNGateway: gw}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceVPNGateway()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{})
	d.SetId("vpn-gw-001")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	assertVPNGatewayResourceData(t, d, gw)
}

func TestResourceVPNGatewayRead_NotFound(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.VPNGatewayWithID(testhelpers.TestProjectID, "vpn-gone"),
			Handler: testhelpers.EmptyHandler(http.StatusNotFound),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceVPNGateway()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{})
	d.SetId("vpn-gone")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error on 404: %v", diags)
	}
	if d.Id() != "" {
		t.Errorf("expected resource ID to be cleared after 404, got %s", d.Id())
	}
}

func TestResourceVPNGatewayUpdate(t *testing.T) {
	gw := testVPNGateway()
	gw.Name = "test-vpn-gateway-updated"
	gw.Description = "updated test vpn gateway"

	var gotReq dto.UpdateVPNGatewayRequest
	putCalled := false

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Pattern: client.ApiPath.VPNGatewayWithID(testhelpers.TestProjectID, "vpn-gw-001"),
			Handler: func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case "PUT":
					putCalled = true
					if err := json.NewDecoder(r.Body).Decode(&gotReq); err != nil {
						t.Errorf("failed to decode update request: %v", err)
					}
					w.WriteHeader(http.StatusOK)
				case "GET":
					testhelpers.JSONHandler(t, http.StatusOK, dto.VPNGatewayResponse{VPNGateway: gw})(w, r)
				default:
					http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				}
			},
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceVPNGateway()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":        "test-vpn-gateway-updated",
		"description": "updated test vpn gateway",
		"vpn_type":    "ROUTE_BASED",
		"zone_id":     testhelpers.TestZoneID,
	})
	d.SetId("vpn-gw-001")

	diags := res.UpdateContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if !putCalled {
		t.Fatal("expected PUT to have been called")
	}
	if gotReq.Name != "test-vpn-gateway-updated" {
		t.Errorf("expected request name test-vpn-gateway-updated, got %s", gotReq.Name)
	}
	if gotReq.Description != "updated test vpn gateway" {
		t.Errorf("expected request description updated test vpn gateway, got %s", gotReq.Description)
	}
	assertVPNGatewayResourceData(t, d, gw)
}

func TestResourceVPNGatewayDelete(t *testing.T) {
	gw := testVPNGateway()
	deletedCalled := false

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Pattern: client.ApiPath.VPNGatewayWithID(testhelpers.TestProjectID, "vpn-gw-001"),
			Handler: func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case "GET":
					if deletedCalled {
						w.WriteHeader(http.StatusNotFound)
						return
					}
					testhelpers.JSONHandler(t, http.StatusOK, dto.VPNGatewayResponse{VPNGateway: gw})(w, r)
				case "DELETE":
					deletedCalled = true
					w.WriteHeader(http.StatusAccepted)
				default:
					http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				}
			},
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceVPNGateway()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":        "test-vpn-gateway",
		"description": "a test vpn gateway",
		"vpn_type":    "ROUTE_BASED",
		"zone_id":     testhelpers.TestZoneID,
	})
	d.SetId("vpn-gw-001")

	diags := res.DeleteContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if !deletedCalled {
		t.Error("expected DELETE to have been called")
	}
}

func assertVPNGatewayResourceData(t *testing.T, d *schema.ResourceData, gw dto.VPNGateway) {
	t.Helper()

	if d.Id() != gw.ID {
		t.Errorf("expected ID %s, got %s", gw.ID, d.Id())
	}
	if v := d.Get("name").(string); v != gw.Name {
		t.Errorf("expected name %s, got %s", gw.Name, v)
	}
	if v := d.Get("description").(string); v != gw.Description {
		t.Errorf("expected description %s, got %s", gw.Description, v)
	}
	if v := d.Get("vpn_type").(string); v != gw.VPNType {
		t.Errorf("expected vpn_type %s, got %s", gw.VPNType, v)
	}
	if v := d.Get("status").(string); v != gw.Status {
		t.Errorf("expected status %s, got %s", gw.Status, v)
	}
	if v := d.Get("created_at").(string); v != gw.CreatedAt {
		t.Errorf("expected created_at %s, got %s", gw.CreatedAt, v)
	}

	attachedVPCIDs := d.Get("attached_vpc_ids").([]interface{})
	if len(attachedVPCIDs) != len(gw.AttachedVPCIDs) {
		t.Fatalf("expected %d attached_vpc_ids, got %d", len(gw.AttachedVPCIDs), len(attachedVPCIDs))
	}
	for i, vpcID := range gw.AttachedVPCIDs {
		if attachedVPCIDs[i].(string) != vpcID {
			t.Errorf("expected attached_vpc_ids[%d] %s, got %s", i, vpcID, attachedVPCIDs[i].(string))
		}
	}
}
