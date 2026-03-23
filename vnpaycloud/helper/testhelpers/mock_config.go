package testhelpers

import (
	"context"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/mutexkv"
)

const (
	TestProjectID = "test-project-id"
	TestZoneID    = "test-zone-id"
	TestToken     = "vtx_pat_test_token"
)

// NewMockConfig creates a config.Config with a real client pointing at the
// given test server URL. Use together with NewMockServer to build a fully
// functional test config.
func NewMockConfig(t *testing.T, serverURL string) *config.Config {
	t.Helper()

	c, err := client.NewClient(context.Background(), &client.ClientConfig{
		BaseURL: serverURL,
		Token:   TestToken,
	})
	if err != nil {
		t.Fatalf("failed to create test client: %v", err)
	}

	return &config.Config{
		MutexKV:   mutexkv.NewMutexKV(),
		Client:    c,
		ProjectID: TestProjectID,
		ZoneID:    TestZoneID,
	}
}
