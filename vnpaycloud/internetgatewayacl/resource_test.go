package internetgatewayacl

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const blacklistPath = "/v2/iac/projects/test-project-id/internet-gateways/igw-001/blacklist"

func testBlacklist(match ...string) dto.InternetGatewayBlacklistResponse {
	return dto.InternetGatewayBlacklistResponse{
		Blacklist: dto.InternetGatewayBlacklist{
			InternetGatewayID: "igw-001",
			Match:             match,
			ProjectID:         testhelpers.TestProjectID,
			ZoneID:            testhelpers.TestZoneID,
		},
	}
}

// blacklistRoute serves GET and PUT on the same path, mimicking a gateway whose
// blacklist starts at initial and is replaced wholesale by each PUT.
func blacklistRoute(t *testing.T, putBody *dto.UpdateInternetGatewayBlacklistRequest, initial ...string) testhelpers.Route {
	t.Helper()
	current := initial

	return testhelpers.Route{
		Pattern: blacklistPath,
		Handler: func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodPut:
				var body dto.UpdateInternetGatewayBlacklistRequest
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Errorf("failed to decode request body: %v", err)
				}
				if putBody != nil {
					*putBody = body
				}
				current = body.Match
			case http.MethodGet:
			default:
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(testBlacklist(current...)); err != nil {
				t.Errorf("failed to encode response body: %v", err)
			}
		},
	}
}

func TestResourceInternetGatewayACLCreate(t *testing.T) {
	var putBody dto.UpdateInternetGatewayBlacklistRequest

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		blacklistRoute(t, &putBody),
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceInternetGatewayACL()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"internet_gateway_id": "igw-001",
		"match":               []interface{}{"1.2.3.0/24", "8.8.8.8/32"},
	})

	diags := res.CreateContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "igw-001" {
		t.Errorf("expected ID igw-001, got %s", d.Id())
	}
	if len(putBody.Match) != 2 {
		t.Errorf("expected 2 match entries sent, got %v", putBody.Match)
	}
	if v := d.Get("match").(*schema.Set).Len(); v != 2 {
		t.Errorf("expected 2 match entries in state, got %d", v)
	}
	if v := d.Get("zone_id").(string); v != testhelpers.TestZoneID {
		t.Errorf("expected zone_id %s, got %s", testhelpers.TestZoneID, v)
	}
}

func TestResourceInternetGatewayACLRead(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		blacklistRoute(t, nil, "1.2.3.0/24"),
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceInternetGatewayACL()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{})
	d.SetId("igw-001")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if v := d.Get("internet_gateway_id").(string); v != "igw-001" {
		t.Errorf("expected internet_gateway_id igw-001, got %s", v)
	}
	if match := d.Get("match").(*schema.Set); !match.Contains("1.2.3.0/24") {
		t.Errorf("expected match to contain 1.2.3.0/24, got %v", match.List())
	}
}

// An internet gateway whose blacklist has been emptied elsewhere must drop out of state.
func TestResourceInternetGatewayACLReadEmptyRemovesFromState(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		blacklistRoute(t, nil),
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceInternetGatewayACL()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{})
	d.SetId("igw-001")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "" {
		t.Errorf("expected resource to be removed from state, got ID %s", d.Id())
	}
}

func TestResourceInternetGatewayACLUpdate(t *testing.T) {
	var putBody dto.UpdateInternetGatewayBlacklistRequest

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		blacklistRoute(t, &putBody, "1.1.1.0/24"),
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceInternetGatewayACL()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"internet_gateway_id": "igw-001",
		"match":               []interface{}{"9.9.9.0/24"},
	})
	d.SetId("igw-001")

	diags := res.UpdateContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if match := d.Get("match").(*schema.Set); !match.Contains("9.9.9.0/24") {
		t.Errorf("expected match to contain 9.9.9.0/24, got %v", match.List())
	}
}

// Delete clears the blacklist by sending an empty match list, there is no DELETE endpoint.
func TestResourceInternetGatewayACLDeleteSendsEmptyMatch(t *testing.T) {
	putBody := dto.UpdateInternetGatewayBlacklistRequest{Match: []string{"stale"}}

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		blacklistRoute(t, &putBody, "1.2.3.0/24"),
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceInternetGatewayACL()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"internet_gateway_id": "igw-001",
		"match":               []interface{}{"1.2.3.0/24"},
	})
	d.SetId("igw-001")

	diags := res.DeleteContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if len(putBody.Match) != 0 {
		t.Errorf("expected delete to send an empty match list, got %v", putBody.Match)
	}
}

// Create must not silently replace a blacklist that was set outside Terraform.
func TestResourceInternetGatewayACLCreateRefusesExistingBlacklist(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		blacklistRoute(t, nil, "5.5.5.0/24"),
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceInternetGatewayACL()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"internet_gateway_id": "igw-001",
		"match":               []interface{}{"1.2.3.0/24"},
	})

	diags := res.CreateContext(context.Background(), d, cfg)
	if !diags.HasError() {
		t.Fatal("expected create to fail when the gateway already has a blacklist")
	}
	if !strings.Contains(diags[0].Summary, "terraform import") {
		t.Errorf("expected the error to point at import, got %q", diags[0].Summary)
	}
}

func TestValidateBlacklistCIDR(t *testing.T) {
	cases := []struct {
		value   string
		wantErr bool
	}{
		{"1.2.3.0/24", false},
		{"8.8.8.8/32", false},
		{"8.8.8.8", true},
		{"not-a-cidr", true},
		{"2001:db8::/32", true},
		{"10.0.1.0/24", true},
		{"192.168.1.0/24", true},
		{"127.0.0.1/32", true},
		{"224.0.0.1/32", true},
	}

	for _, c := range cases {
		_, errs := validateBlacklistCIDR(c.value, "match")
		if c.wantErr && len(errs) == 0 {
			t.Errorf("%s: expected an error, got none", c.value)
		}
		if !c.wantErr && len(errs) > 0 {
			t.Errorf("%s: expected no error, got %v", c.value, errs)
		}
	}
}
