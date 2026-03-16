package image

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestDataSourceImageRead_ByID(t *testing.T) {
	imageResp := dto.ImageResponse{
		Image: dto.Image{
			ID:        "img-123",
			Name:      "Ubuntu 22.04",
			OsType:    "linux",
			OsVersion: "22.04",
			MinDiskGB: 20,
			Status:    "active",
			Zone:      "zone-a",
		},
	}

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.ImageWithID("img-123"),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, imageResp),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := DataSourceImage()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"id": "img-123",
	})

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "img-123" {
		t.Errorf("expected ID 'img-123', got '%s'", d.Id())
	}
	if got := d.Get("name").(string); got != "Ubuntu 22.04" {
		t.Errorf("expected name 'Ubuntu 22.04', got '%s'", got)
	}
	if got := d.Get("os_type").(string); got != "linux" {
		t.Errorf("expected os_type 'linux', got '%s'", got)
	}
	if got := d.Get("os_version").(string); got != "22.04" {
		t.Errorf("expected os_version '22.04', got '%s'", got)
	}
	if got := d.Get("min_disk_gb").(int); got != 20 {
		t.Errorf("expected min_disk_gb 20, got %d", got)
	}
	if got := d.Get("status").(string); got != "active" {
		t.Errorf("expected status 'active', got '%s'", got)
	}
	if got := d.Get("zone").(string); got != "zone-a" {
		t.Errorf("expected zone 'zone-a', got '%s'", got)
	}
}

func TestDataSourceImageRead_ByName(t *testing.T) {
	listResp := dto.ListImagesResponse{
		Images: []dto.Image{
			{
				ID:        "img-aaa",
				Name:      "CentOS 9",
				OsType:    "linux",
				OsVersion: "9",
				MinDiskGB: 10,
				Status:    "active",
				Zone:      "zone-b",
			},
			{
				ID:        "img-bbb",
				Name:      "Windows Server 2022",
				OsType:    "windows",
				OsVersion: "2022",
				MinDiskGB: 40,
				Status:    "active",
				Zone:      "zone-b",
			},
		},
	}

	// The client calls GET /v2/iac/images?zone=test-zone-id.
	// http.ServeMux strips query params before matching, so register the base path.
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/images",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, listResp),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := DataSourceImage()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name": "Windows Server 2022",
	})

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "img-bbb" {
		t.Errorf("expected ID 'img-bbb', got '%s'", d.Id())
	}
	if got := d.Get("name").(string); got != "Windows Server 2022" {
		t.Errorf("expected name 'Windows Server 2022', got '%s'", got)
	}
	if got := d.Get("os_type").(string); got != "windows" {
		t.Errorf("expected os_type 'windows', got '%s'", got)
	}
}

func TestDataSourceImagesRead(t *testing.T) {
	listResp := dto.ListImagesResponse{
		Images: []dto.Image{
			{
				ID:        "img-1",
				Name:      "Ubuntu 22.04",
				OsType:    "linux",
				OsVersion: "22.04",
				MinDiskGB: 20,
				Status:    "active",
				Zone:      "zone-a",
			},
			{
				ID:        "img-2",
				Name:      "Debian 12",
				OsType:    "linux",
				OsVersion: "12",
				MinDiskGB: 10,
				Status:    "active",
				Zone:      "zone-a",
			},
		},
	}

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/images",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, listResp),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := DataSourceImages()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{})

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	expectedID := "images-" + testhelpers.TestZoneID
	if d.Id() != expectedID {
		t.Errorf("expected ID '%s', got '%s'", expectedID, d.Id())
	}

	images := d.Get("images").([]interface{})
	if len(images) != 2 {
		t.Fatalf("expected 2 images, got %d", len(images))
	}

	first := images[0].(map[string]interface{})
	if first["id"] != "img-1" {
		t.Errorf("expected first image id 'img-1', got '%s'", first["id"])
	}
	if first["name"] != "Ubuntu 22.04" {
		t.Errorf("expected first image name 'Ubuntu 22.04', got '%s'", first["name"])
	}

	second := images[1].(map[string]interface{})
	if second["id"] != "img-2" {
		t.Errorf("expected second image id 'img-2', got '%s'", second["id"])
	}
}
