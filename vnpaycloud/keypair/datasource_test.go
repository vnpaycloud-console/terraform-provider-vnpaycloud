package keypair

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestDataSourceKeyPairRead(t *testing.T) {
	kpResp := dto.KeyPairResponse{
		KeyPair: dto.KeyPair{
			ID:          "kp-ds-1",
			Name:        "my-keypair",
			PublicKey:   "ssh-rsa AAAA...",
			Fingerprint: "aa:bb:cc:dd",
			CreatedAt:   "2025-01-15T10:00:00Z",
		},
	}

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.KeyPairWithName(testhelpers.TestProjectID, "my-keypair"),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, kpResp),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := DataSourceKeyPair()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name": "my-keypair",
	})

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	// DataSource sets ID to the keypair name
	if d.Id() != "my-keypair" {
		t.Errorf("expected ID 'my-keypair', got '%s'", d.Id())
	}
	if got := d.Get("name").(string); got != "my-keypair" {
		t.Errorf("expected name 'my-keypair', got '%s'", got)
	}
	if got := d.Get("public_key").(string); got != "ssh-rsa AAAA..." {
		t.Errorf("expected public_key 'ssh-rsa AAAA...', got '%s'", got)
	}
	if got := d.Get("fingerprint").(string); got != "aa:bb:cc:dd" {
		t.Errorf("expected fingerprint 'aa:bb:cc:dd', got '%s'", got)
	}
	if got := d.Get("created_at").(string); got != "2025-01-15T10:00:00Z" {
		t.Errorf("expected created_at '2025-01-15T10:00:00Z', got '%s'", got)
	}
}

func TestDataSourceKeyPairsRead(t *testing.T) {
	listResp := dto.ListKeyPairsResponse{
		KeyPairs: []dto.KeyPair{
			{
				ID:          "kp-1",
				Name:        "keypair-one",
				PublicKey:   "ssh-rsa AAA1...",
				Fingerprint: "11:22:33:44",
				CreatedAt:   "2025-01-10T08:00:00Z",
			},
			{
				ID:          "kp-2",
				Name:        "keypair-two",
				PublicKey:   "ssh-rsa AAA2...",
				Fingerprint: "55:66:77:88",
				CreatedAt:   "2025-01-12T09:00:00Z",
			},
		},
	}

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.KeyPairs(testhelpers.TestProjectID),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, listResp),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := DataSourceKeyPairs()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{})

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	expectedID := "key-pairs-" + testhelpers.TestProjectID
	if d.Id() != expectedID {
		t.Errorf("expected ID '%s', got '%s'", expectedID, d.Id())
	}

	keyPairs := d.Get("key_pairs").([]interface{})
	if len(keyPairs) != 2 {
		t.Fatalf("expected 2 key_pairs, got %d", len(keyPairs))
	}

	first := keyPairs[0].(map[string]interface{})
	if first["name"] != "keypair-one" {
		t.Errorf("expected first key_pair name 'keypair-one', got '%s'", first["name"])
	}
	if first["fingerprint"] != "11:22:33:44" {
		t.Errorf("expected first key_pair fingerprint '11:22:33:44', got '%s'", first["fingerprint"])
	}

	second := keyPairs[1].(map[string]interface{})
	if second["name"] != "keypair-two" {
		t.Errorf("expected second key_pair name 'keypair-two', got '%s'", second["name"])
	}
}
