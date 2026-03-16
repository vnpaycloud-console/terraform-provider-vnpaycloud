package testhelpers

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
)

func TestNewMockServer(t *testing.T) {
	type vpc struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	srv := NewMockServer(t, []Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/vpcs/vpc-1",
			Handler: JSONHandler(t, http.StatusOK, vpc{ID: "vpc-1", Name: "test-vpc"}),
		},
		{
			Method:  "DELETE",
			Pattern: "/v2/iac/projects/test-project-id/vpcs/vpc-2",
			Handler: EmptyHandler(http.StatusNoContent),
		},
	})

	cfg := NewMockConfig(t, srv.URL)

	t.Run("GET with JSON response", func(t *testing.T) {
		var got vpc
		path := client.ApiPath.VPCWithID(cfg.ProjectID, "vpc-1")
		_, err := cfg.Client.Get(context.Background(), path, &got, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ID != "vpc-1" || got.Name != "test-vpc" {
			t.Errorf("expected {vpc-1 test-vpc}, got %+v", got)
		}
	})

	t.Run("DELETE with empty response", func(t *testing.T) {
		path := client.ApiPath.VPCWithID(cfg.ProjectID, "vpc-2")
		resp, err := cfg.Client.Delete(context.Background(), path, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("expected 204, got %d", resp.StatusCode)
		}
	})

	t.Run("method not allowed", func(t *testing.T) {
		path := client.ApiPath.VPCWithID(cfg.ProjectID, "vpc-1")
		_, err := cfg.Client.Delete(context.Background(), path, nil)
		if err == nil {
			t.Fatal("expected error for method not allowed")
		}
	})

	t.Run("config has correct fields", func(t *testing.T) {
		if cfg.ProjectID != TestProjectID {
			t.Errorf("expected ProjectID=%s, got %s", TestProjectID, cfg.ProjectID)
		}
		if cfg.ZoneID != TestZoneID {
			t.Errorf("expected ZoneID=%s, got %s", TestZoneID, cfg.ZoneID)
		}
		if cfg.MutexKV == nil {
			t.Error("expected non-nil MutexKV")
		}
		if cfg.Client == nil {
			t.Error("expected non-nil Client")
		}
	})
}
