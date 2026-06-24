package vpnconnection

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func testVPNConnection() dto.VPNConnection {
	return dto.VPNConnection{
		ID:                "vpn-conn-001",
		Name:              "test-vpn-connection",
		Description:       "a test vpn connection",
		VPNGatewayID:      "vpn-gw-001",
		CustomerGatewayID: "customer-gw-001",
		VPNType:           "ROUTE_BASED",
		Status:            "active",
		VPNPublicIPID:     "vpn-public-ip-001",
		IKEProfileConfig: &dto.IKEProfileConfig{
			IKEVersion:     "IKE_V2",
			IKELifetime:    28800,
			IKECloseAction: "START",
			IKEDH:          "GROUP_14",
			IKEEncryption:  "AES128_GCM96",
			IKEHash:        "SHA256",
			IKEPRF:         "SHA1",
			IKEDPDAction:   "RESTART",
			IKEDPDInterval: 30,
			IKEDPDTimeout:  120,
			IKEV2Reauth:    false,
		},
		IPSecProfileConfig: &dto.IPSecProfileConfig{
			IPSecLifetime:        3600,
			IPSecPFS:             "GROUP_14",
			IPSecEncryption:      "AES256",
			IPSecHash:            "SHA256",
			IPSecDisableRekey:    false,
			IPSecLifetimeBytes:   0,
			IPSecLifetimePackets: 0,
		},
		RouteBaseConfig: &dto.RouteBaseConfig{
			VTIMSS: 1360,
		},
		ConnectionBGPConfig: &dto.ConnectionBGPConfig{
			BGPKeepalive: 60,
			BGPHoldtime:  180,
		},
		CreatedAt: "2025-01-15T10:00:00Z",
		ProjectID: testhelpers.TestProjectID,
		ZoneID:    testhelpers.TestZoneID,
	}
}

func TestResourceVPNConnectionCreate(t *testing.T) {
	vpnConnection := testVPNConnection()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "POST",
			Pattern: "/v2/iac/projects/test-project-id/vpn-connections",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.VPNConnectionResponse{VPNConnection: vpnConnection}),
		},
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/vpn-connections/vpn-conn-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.VPNConnectionResponse{VPNConnection: vpnConnection}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceVPNConnection()
	d := schema.TestResourceDataRaw(t, res.Schema, testVPNConnectionResourceData())

	diags := res.CreateContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	assertVPNConnectionResourceData(t, d, vpnConnection)
}

func TestResourceVPNConnectionRead(t *testing.T) {
	vpnConnection := testVPNConnection()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/vpn-connections/vpn-conn-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.VPNConnectionResponse{VPNConnection: vpnConnection}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceVPNConnection()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{})
	d.SetId("vpn-conn-001")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	assertVPNConnectionResourceData(t, d, vpnConnection)
}

func TestResourceVPNConnectionRead_NotFound(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/vpn-connections/vpn-gone",
			Handler: testhelpers.EmptyHandler(http.StatusNotFound),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceVPNConnection()
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

func TestResourceVPNConnectionDelete(t *testing.T) {
	vpnConnection := testVPNConnection()
	deletedCalled := false

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Pattern: "/v2/iac/projects/test-project-id/vpn-connections/vpn-conn-001",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case "GET":
					if deletedCalled {
						w.WriteHeader(http.StatusNotFound)
						return
					}
					testhelpers.JSONHandler(t, http.StatusOK, dto.VPNConnectionResponse{VPNConnection: vpnConnection})(w, r)
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

	res := ResourceVPNConnection()
	d := schema.TestResourceDataRaw(t, res.Schema, testVPNConnectionResourceData())
	d.SetId("vpn-conn-001")

	diags := res.DeleteContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if !deletedCalled {
		t.Error("expected DELETE to have been called")
	}
}

func testVPNConnectionResourceData() map[string]interface{} {
	return map[string]interface{}{
		"name":                "test-vpn-connection",
		"description":         "a test vpn connection",
		"vpn_gateway_id":      "vpn-gw-001",
		"customer_gateway_id": "customer-gw-001",
		"vpn_type":            "ROUTE_BASED",
		"vpn_public_ip_id":    "vpn-public-ip-001",
		"ipsec_auth_config": []interface{}{
			map[string]interface{}{
				"pre_shared_key": "test-secret-key",
			},
		},
		"ike_profile_config": []interface{}{
			map[string]interface{}{
				"ike_version":      "IKE_V2",
				"ike_lifetime":     28800,
				"ike_close_action": "START",
				"ike_dh":           "GROUP_14",
				"ike_encryption":   "AES128_GCM96",
				"ike_hash":         "SHA256",
				"ike_prf":          "SHA1",
				"ike_dpd_action":   "RESTART",
				"ike_dpd_interval": 30,
				"ike_dpd_timeout":  120,
				"ikev2_reauth":     false,
			},
		},
		"ipsec_profile_config": []interface{}{
			map[string]interface{}{
				"ipsec_lifetime":         3600,
				"ipsec_pfs":              "GROUP_14",
				"ipsec_encryption":       "AES256",
				"ipsec_hash":             "SHA256",
				"ipsec_disable_rekey":    false,
				"ipsec_lifetime_bytes":   0,
				"ipsec_lifetime_packets": 0,
			},
		},
		"route_base_config": []interface{}{
			map[string]interface{}{
				"vti_mss": 1360,
			},
		},
		"connection_bgp_config": []interface{}{
			map[string]interface{}{
				"bgp_keepalive": 60,
				"bgp_holdtime":  180,
			},
		},
		"zone_id": testhelpers.TestZoneID,
	}
}

func assertVPNConnectionResourceData(t *testing.T, d *schema.ResourceData, vpnConnection dto.VPNConnection) {
	t.Helper()

	if d.Id() != vpnConnection.ID {
		t.Errorf("expected ID %s, got %s", vpnConnection.ID, d.Id())
	}
	if v := d.Get("name").(string); v != vpnConnection.Name {
		t.Errorf("expected name %s, got %s", vpnConnection.Name, v)
	}
	if v := d.Get("description").(string); v != vpnConnection.Description {
		t.Errorf("expected description %s, got %s", vpnConnection.Description, v)
	}
	if v := d.Get("vpn_gateway_id").(string); v != vpnConnection.VPNGatewayID {
		t.Errorf("expected vpn_gateway_id %s, got %s", vpnConnection.VPNGatewayID, v)
	}
	if v := d.Get("customer_gateway_id").(string); v != vpnConnection.CustomerGatewayID {
		t.Errorf("expected customer_gateway_id %s, got %s", vpnConnection.CustomerGatewayID, v)
	}
	if v := d.Get("vpn_type").(string); v != vpnConnection.VPNType {
		t.Errorf("expected vpn_type %s, got %s", vpnConnection.VPNType, v)
	}
	if v := d.Get("vpn_public_ip_id").(string); v != vpnConnection.VPNPublicIPID {
		t.Errorf("expected vpn_public_ip_id %s, got %s", vpnConnection.VPNPublicIPID, v)
	}
	if v := d.Get("status").(string); v != vpnConnection.Status {
		t.Errorf("expected status %s, got %s", vpnConnection.Status, v)
	}
	if v := d.Get("created_at").(string); v != vpnConnection.CreatedAt {
		t.Errorf("expected created_at %s, got %s", vpnConnection.CreatedAt, v)
	}

	ikeProfileConfig := d.Get("ike_profile_config").([]interface{})
	if len(ikeProfileConfig) != 1 {
		t.Fatalf("expected one ike_profile_config, got %d", len(ikeProfileConfig))
	}
	ike := ikeProfileConfig[0].(map[string]interface{})
	if ike["ike_version"] != vpnConnection.IKEProfileConfig.IKEVersion {
		t.Errorf("expected ike_version %s, got %v", vpnConnection.IKEProfileConfig.IKEVersion, ike["ike_version"])
	}
	if ike["ike_dh"] != vpnConnection.IKEProfileConfig.IKEDH {
		t.Errorf("expected ike_dh %s, got %v", vpnConnection.IKEProfileConfig.IKEDH, ike["ike_dh"])
	}

	ipsecProfileConfig := d.Get("ipsec_profile_config").([]interface{})
	if len(ipsecProfileConfig) != 1 {
		t.Fatalf("expected one ipsec_profile_config, got %d", len(ipsecProfileConfig))
	}
	ipsec := ipsecProfileConfig[0].(map[string]interface{})
	if ipsec["ipsec_lifetime"] != vpnConnection.IPSecProfileConfig.IPSecLifetime {
		t.Errorf("expected ipsec_lifetime %d, got %v", vpnConnection.IPSecProfileConfig.IPSecLifetime, ipsec["ipsec_lifetime"])
	}
	if ipsec["ipsec_lifetime_bytes"] != int(vpnConnection.IPSecProfileConfig.IPSecLifetimeBytes) {
		t.Errorf("expected ipsec_lifetime_bytes %d, got %v", vpnConnection.IPSecProfileConfig.IPSecLifetimeBytes, ipsec["ipsec_lifetime_bytes"])
	}

	routeBaseConfig := d.Get("route_base_config").([]interface{})
	if len(routeBaseConfig) != 1 {
		t.Fatalf("expected one route_base_config, got %d", len(routeBaseConfig))
	}
	routeBase := routeBaseConfig[0].(map[string]interface{})
	if routeBase["vti_mss"] != vpnConnection.RouteBaseConfig.VTIMSS {
		t.Errorf("expected vti_mss %d, got %v", vpnConnection.RouteBaseConfig.VTIMSS, routeBase["vti_mss"])
	}

	connectionBGPConfig := d.Get("connection_bgp_config").([]interface{})
	if len(connectionBGPConfig) != 1 {
		t.Fatalf("expected one connection_bgp_config, got %d", len(connectionBGPConfig))
	}
	bgp := connectionBGPConfig[0].(map[string]interface{})
	if bgp["bgp_keepalive"] != vpnConnection.ConnectionBGPConfig.BGPKeepalive {
		t.Errorf("expected bgp_keepalive %d, got %v", vpnConnection.ConnectionBGPConfig.BGPKeepalive, bgp["bgp_keepalive"])
	}
	if bgp["bgp_holdtime"] != vpnConnection.ConnectionBGPConfig.BGPHoldtime {
		t.Errorf("expected bgp_holdtime %d, got %v", vpnConnection.ConnectionBGPConfig.BGPHoldtime, bgp["bgp_holdtime"])
	}
}
