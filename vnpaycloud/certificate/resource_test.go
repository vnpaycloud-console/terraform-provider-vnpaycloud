package certificate

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func testCertificate() dto.Certificate {
	return dto.Certificate{
		ID:              "cert-001",
		Name:            "tf-cert",
		CertType:        "CT_SELF_SIGNED",
		DomainName:      "example.com",
		Description:     "terraform certificate",
		Expiration:      "2026-01-15T10:00:00Z",
		Status:          "active",
		ZoneID:          testhelpers.TestZoneID,
		LoadBalancerIDs: []string{"lb-001"},
	}
}

func TestResourceCertificateSchema(t *testing.T) {
	res := ResourceCertificate()

	for _, key := range []string{"type", "name"} {
		if !res.Schema[key].Required {
			t.Errorf("expected %s to be required", key)
		}
		if !res.Schema[key].ForceNew {
			t.Errorf("expected %s to be ForceNew", key)
		}
	}

	if res.UpdateContext != nil {
		t.Error("certificate has no updatable fields; UpdateContext should be nil")
	}

	for _, key := range []string{"private_key", "certificate_body", "intermediate_ca"} {
		if res.Schema[key] == nil || !res.Schema[key].ForceNew {
			t.Errorf("expected %s to be a ForceNew field", key)
		}
	}
}

func TestResourceCertificateRead(t *testing.T) {
	c := testCertificate()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.CertificateWithID(testhelpers.TestProjectID, "cert-001"),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.CertificateResponse{Certificate: c}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceCertificate()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{})
	d.SetId("cert-001")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if v := d.Get("name").(string); v != "tf-cert" {
		t.Errorf("expected name tf-cert, got %s", v)
	}
	if v := d.Get("type").(string); v != typeSelfSigned {
		t.Errorf("expected type %s, got %s", typeSelfSigned, v)
	}
	if v := d.Get("cert_type").(string); v != "CT_SELF_SIGNED" {
		t.Errorf("expected cert_type CT_SELF_SIGNED, got %s", v)
	}
	if v := d.Get("status").(string); v != "active" {
		t.Errorf("expected status active, got %s", v)
	}
	if v := d.Get("expires_at").(string); v != "2026-01-15T10:00:00Z" {
		t.Errorf("expected expires_at 2026-01-15T10:00:00Z, got %s", v)
	}
	lbs := d.Get("load_balancer_ids").([]interface{})
	if len(lbs) != 1 || lbs[0].(string) != "lb-001" {
		t.Errorf("expected load_balancer_ids [lb-001], got %v", lbs)
	}
}

func TestResourceCertificateRead_NotFound(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.CertificateWithID(testhelpers.TestProjectID, "missing"),
			Handler: testhelpers.EmptyHandler(http.StatusNotFound),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceCertificate()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{})
	d.SetId("missing")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("expected NotFound to clear id without error, got: %v", diags)
	}
	if d.Id() != "" {
		t.Errorf("expected id cleared on 404, got %s", d.Id())
	}
}
