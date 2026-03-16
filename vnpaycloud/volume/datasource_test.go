package volume

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestDataSourceVolumeRead_ByID(t *testing.T) {
	vol := testVolume()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/volumes/vol-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.VolumeResponse{Volume: vol}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceVolume()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"id": "vol-001",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "vol-001" {
		t.Errorf("expected ID vol-001, got %s", d.Id())
	}
	if v := d.Get("name").(string); v != "test-volume" {
		t.Errorf("expected name test-volume, got %s", v)
	}
	if v := d.Get("description").(string); v != "a test volume" {
		t.Errorf("expected description 'a test volume', got %s", v)
	}
	if v := d.Get("size").(int); v != 50 {
		t.Errorf("expected size 50, got %d", v)
	}
	if v := d.Get("volume_type").(string); v != "SSD" {
		t.Errorf("expected volume_type SSD, got %s", v)
	}
	if v := d.Get("zone").(string); v != "zone-a" {
		t.Errorf("expected zone zone-a, got %s", v)
	}
	if v := d.Get("status").(string); v != "available" {
		t.Errorf("expected status available, got %s", v)
	}
	if v := d.Get("iops").(int); v != 3000 {
		t.Errorf("expected iops 3000, got %d", v)
	}
	if v := d.Get("is_encrypted").(bool); !v {
		t.Error("expected is_encrypted true, got false")
	}
	if v := d.Get("is_multiattach").(bool); v {
		t.Error("expected is_multiattach false, got true")
	}
	if v := d.Get("is_bootable").(bool); v {
		t.Error("expected is_bootable false, got true")
	}
	if v := d.Get("created_at").(string); v != "2025-01-15T10:00:00Z" {
		t.Errorf("expected created_at 2025-01-15T10:00:00Z, got %s", v)
	}
}

func TestDataSourceVolumeRead_ByName(t *testing.T) {
	vol := testVolume()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method: "GET",
			Pattern: "/v2/iac/projects/test-project-id/volumes",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListVolumesResponse{
				Volumes: []dto.Volume{vol},
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceVolume()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"name": "test-volume",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "vol-001" {
		t.Errorf("expected ID vol-001, got %s", d.Id())
	}
	if v := d.Get("name").(string); v != "test-volume" {
		t.Errorf("expected name test-volume, got %s", v)
	}
}

func TestDataSourceVolumeRead_NotFound(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method: "GET",
			Pattern: "/v2/iac/projects/test-project-id/volumes",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListVolumesResponse{
				Volumes: []dto.Volume{},
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceVolume()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"name": "nonexistent",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if !diags.HasError() {
		t.Fatal("expected error for nonexistent volume, got none")
	}
}

func TestDataSourceVolumesRead(t *testing.T) {
	vol1 := testVolume()
	vol2 := dto.Volume{
		ID:                 "vol-002",
		Name:               "test-volume-2",
		Description:        "second volume",
		SizeGB:             100,
		VolumeType:         "HDD",
		Zone:               "zone-b",
		Status:             "in-use",
		IOPS:               1000,
		IsEncrypted:        false,
		IsMultiattach:      true,
		IsBootable:         true,
		AttachedServerID:   "srv-001",
		AttachedServerName: "my-server",
		CreatedAt:          "2025-01-16T12:00:00Z",
		ProjectID:          testhelpers.TestProjectID,
		ZoneID:             testhelpers.TestZoneID,
	}

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method: "GET",
			Pattern: "/v2/iac/projects/test-project-id/volumes",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListVolumesResponse{
				Volumes: []dto.Volume{vol1, vol2},
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceVolumes()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	volumes := d.Get("volumes").([]interface{})
	if len(volumes) != 2 {
		t.Fatalf("expected 2 volumes, got %d", len(volumes))
	}

	first := volumes[0].(map[string]interface{})
	if first["id"] != "vol-001" {
		t.Errorf("expected first volume id vol-001, got %v", first["id"])
	}
	if first["name"] != "test-volume" {
		t.Errorf("expected first volume name test-volume, got %v", first["name"])
	}

	second := volumes[1].(map[string]interface{})
	if second["id"] != "vol-002" {
		t.Errorf("expected second volume id vol-002, got %v", second["id"])
	}
	if second["name"] != "test-volume-2" {
		t.Errorf("expected second volume name test-volume-2, got %v", second["name"])
	}
	if second["status"] != "in-use" {
		t.Errorf("expected second volume status in-use, got %v", second["status"])
	}
}
