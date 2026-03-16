package flavor

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestDataSourceFlavorRead_ByID(t *testing.T) {
	flavorResp := dto.FlavorResponse{
		Flavor: dto.Flavor{
			ID:       "flavor-123",
			Name:     "m1.small",
			VCPUs:    2,
			RAMMB:    2048,
			DiskGB:   20,
			IsPublic: true,
			Zone:     "zone-a",
		},
	}

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.FlavorWithID("flavor-123"),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, flavorResp),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := DataSourceFlavor()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"id": "flavor-123",
	})

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "flavor-123" {
		t.Errorf("expected ID 'flavor-123', got '%s'", d.Id())
	}
	if got := d.Get("name").(string); got != "m1.small" {
		t.Errorf("expected name 'm1.small', got '%s'", got)
	}
	if got := d.Get("vcpus").(int); got != 2 {
		t.Errorf("expected vcpus 2, got %d", got)
	}
	if got := d.Get("ram_mb").(int); got != 2048 {
		t.Errorf("expected ram_mb 2048, got %d", got)
	}
	if got := d.Get("disk_gb").(int); got != 20 {
		t.Errorf("expected disk_gb 20, got %d", got)
	}
	if got := d.Get("is_public").(bool); !got {
		t.Errorf("expected is_public true, got false")
	}
	if got := d.Get("zone").(string); got != "zone-a" {
		t.Errorf("expected zone 'zone-a', got '%s'", got)
	}
}

func TestDataSourceFlavorRead_ByName(t *testing.T) {
	listResp := dto.ListFlavorsResponse{
		Flavors: []dto.Flavor{
			{
				ID:       "flavor-aaa",
				Name:     "m1.large",
				VCPUs:    4,
				RAMMB:    8192,
				DiskGB:   40,
				IsPublic: true,
				Zone:     "zone-b",
			},
			{
				ID:       "flavor-bbb",
				Name:     "m1.xlarge",
				VCPUs:    8,
				RAMMB:    16384,
				DiskGB:   80,
				IsPublic: false,
				Zone:     "zone-b",
			},
		},
	}

	// The client calls GET /v2/iac/flavors?zone=test-zone-id.
	// http.ServeMux strips query params before matching, so register the base path.
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/flavors",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, listResp),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := DataSourceFlavor()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name": "m1.xlarge",
	})

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "flavor-bbb" {
		t.Errorf("expected ID 'flavor-bbb', got '%s'", d.Id())
	}
	if got := d.Get("name").(string); got != "m1.xlarge" {
		t.Errorf("expected name 'm1.xlarge', got '%s'", got)
	}
	if got := d.Get("vcpus").(int); got != 8 {
		t.Errorf("expected vcpus 8, got %d", got)
	}
}

func TestDataSourceFlavorsRead(t *testing.T) {
	listResp := dto.ListFlavorsResponse{
		Flavors: []dto.Flavor{
			{
				ID:       "flavor-1",
				Name:     "m1.small",
				VCPUs:    2,
				RAMMB:    2048,
				DiskGB:   20,
				IsPublic: true,
				Zone:     "zone-a",
			},
			{
				ID:       "flavor-2",
				Name:     "m1.medium",
				VCPUs:    4,
				RAMMB:    4096,
				DiskGB:   40,
				IsPublic: true,
				Zone:     "zone-a",
			},
		},
	}

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/flavors",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, listResp),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := DataSourceFlavors()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{})

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	expectedID := "flavors-" + testhelpers.TestZoneID
	if d.Id() != expectedID {
		t.Errorf("expected ID '%s', got '%s'", expectedID, d.Id())
	}

	flavors := d.Get("flavors").([]interface{})
	if len(flavors) != 2 {
		t.Fatalf("expected 2 flavors, got %d", len(flavors))
	}

	first := flavors[0].(map[string]interface{})
	if first["id"] != "flavor-1" {
		t.Errorf("expected first flavor id 'flavor-1', got '%s'", first["id"])
	}
	if first["name"] != "m1.small" {
		t.Errorf("expected first flavor name 'm1.small', got '%s'", first["name"])
	}

	second := flavors[1].(map[string]interface{})
	if second["id"] != "flavor-2" {
		t.Errorf("expected second flavor id 'flavor-2', got '%s'", second["id"])
	}
}
