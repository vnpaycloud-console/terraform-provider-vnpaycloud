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

func TestResourceKeyPairCreate_WithoutPublicKey(t *testing.T) {
	createResp := dto.KeyPairResponse{
		KeyPair: dto.KeyPair{
			ID:          "kp-123",
			Name:        "my-keypair",
			PublicKey:   "ssh-rsa AAAA...",
			Fingerprint: "aa:bb:cc:dd",
			CreatedAt:   "2025-01-15T10:00:00Z",
		},
		PrivateKey: "-----BEGIN RSA PRIVATE KEY-----\nMIIE...\n-----END RSA PRIVATE KEY-----",
	}

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "POST",
			Pattern: client.ApiPath.CreateKeyPair(),
			Handler: testhelpers.JSONHandler(t, http.StatusCreated, createResp),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceKeyPair()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name": "my-keypair",
	})

	diags := res.CreateContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "kp-123" {
		t.Errorf("expected ID 'kp-123', got '%s'", d.Id())
	}
	if got := d.Get("name").(string); got != "my-keypair" {
		t.Errorf("expected name 'my-keypair', got '%s'", got)
	}
	if got := d.Get("public_key").(string); got != "ssh-rsa AAAA..." {
		t.Errorf("expected public_key 'ssh-rsa AAAA...', got '%s'", got)
	}
	if got := d.Get("private_key").(string); got == "" {
		t.Error("expected private_key to be set when server generates keypair")
	}
	if got := d.Get("fingerprint").(string); got != "aa:bb:cc:dd" {
		t.Errorf("expected fingerprint 'aa:bb:cc:dd', got '%s'", got)
	}
	if got := d.Get("created_at").(string); got != "2025-01-15T10:00:00Z" {
		t.Errorf("expected created_at '2025-01-15T10:00:00Z', got '%s'", got)
	}
}

func TestResourceKeyPairCreate_WithPublicKey(t *testing.T) {
	createResp := dto.KeyPairResponse{
		KeyPair: dto.KeyPair{
			ID:          "kp-456",
			Name:        "my-keypair-2",
			PublicKey:   "ssh-rsa BBBBuser-provided",
			Fingerprint: "ee:ff:00:11",
			CreatedAt:   "2025-01-15T11:00:00Z",
		},
		// No PrivateKey when user provides their own public key
	}

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "POST",
			Pattern: client.ApiPath.CreateKeyPair(),
			Handler: testhelpers.JSONHandler(t, http.StatusCreated, createResp),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceKeyPair()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":       "my-keypair-2",
		"public_key": "ssh-rsa BBBBuser-provided",
	})

	diags := res.CreateContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "kp-456" {
		t.Errorf("expected ID 'kp-456', got '%s'", d.Id())
	}
	if got := d.Get("public_key").(string); got != "ssh-rsa BBBBuser-provided" {
		t.Errorf("expected public_key 'ssh-rsa BBBBuser-provided', got '%s'", got)
	}
	if got := d.Get("private_key").(string); got != "" {
		t.Errorf("expected private_key to be empty when user provides public key, got '%s'", got)
	}
}

func TestResourceKeyPairRead(t *testing.T) {
	readResp := dto.KeyPairResponse{
		KeyPair: dto.KeyPair{
			ID:          "kp-123",
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
			Handler: testhelpers.JSONHandler(t, http.StatusOK, readResp),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceKeyPair()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name": "my-keypair",
	})
	d.SetId("kp-123")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "kp-123" {
		t.Errorf("expected ID 'kp-123', got '%s'", d.Id())
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
}

func TestResourceKeyPairRead_NotFound(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.KeyPairWithName(testhelpers.TestProjectID, "gone-keypair"),
			Handler: testhelpers.EmptyHandler(http.StatusNotFound),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceKeyPair()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name": "gone-keypair",
	})
	d.SetId("kp-gone")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "" {
		t.Errorf("expected resource ID to be cleared on 404, got '%s'", d.Id())
	}
}

func TestResourceKeyPairDelete(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "DELETE",
			Pattern: client.ApiPath.KeyPairWithName(testhelpers.TestProjectID, "my-keypair"),
			Handler: testhelpers.EmptyHandler(http.StatusNoContent),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceKeyPair()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name": "my-keypair",
	})
	d.SetId("kp-123")

	diags := res.DeleteContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
}
