package volume

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// testVolume returns a fully populated dto.Volume for use in tests.
func testVolume() dto.Volume {
	return dto.Volume{
		ID:                 "vol-001",
		Name:               "test-volume",
		Description:        "a test volume",
		SizeGB:             50,
		VolumeType:         "SSD",
		Zone:               "zone-a",
		Status:             "available",
		IOPS:               3000,
		IsEncrypted:        true,
		IsMultiattach:      false,
		IsBootable:         false,
		AttachedServerID:   "",
		AttachedServerName: "",
		CreatedAt:          "2025-01-15T10:00:00Z",
		ProjectID:          testhelpers.TestProjectID,
		ZoneID:             testhelpers.TestZoneID,
	}
}

func TestResourceVolumeCreate(t *testing.T) {
	vol := testVolume()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "POST",
			Pattern: "/v2/iac/projects/test-project-id/volumes",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.VolumeResponse{Volume: vol}),
		},
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/volumes/vol-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.VolumeResponse{Volume: vol}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceVolume()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":        "test-volume",
		"description": "a test volume",
		"size":        50,
		"volume_type": "SSD",
		"encrypt":     true,
		"multiattach": false,
		"snapshot_id": "",
	})

	diags := res.CreateContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "vol-001" {
		t.Errorf("expected ID vol-001, got %s", d.Id())
	}
	if v := d.Get("name").(string); v != "test-volume" {
		t.Errorf("expected name test-volume, got %s", v)
	}
	if v := d.Get("status").(string); v != "available" {
		t.Errorf("expected status available, got %s", v)
	}
}

func TestResourceVolumeRead(t *testing.T) {
	vol := testVolume()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/volumes/vol-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.VolumeResponse{Volume: vol}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceVolume()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":        "test-volume",
		"description": "",
		"size":        50,
		"volume_type": "SSD",
		"encrypt":     false,
		"multiattach": false,
		"snapshot_id": "",
	})
	d.SetId("vol-001")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
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

func TestResourceVolumeRead_NotFound(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/volumes/vol-gone",
			Handler: testhelpers.EmptyHandler(http.StatusNotFound),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceVolume()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":        "",
		"description": "",
		"size":        10,
		"volume_type": "SSD",
		"encrypt":     false,
		"multiattach": false,
		"snapshot_id": "",
	})
	d.SetId("vol-gone")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error on 404: %v", diags)
	}

	if d.Id() != "" {
		t.Errorf("expected resource ID to be cleared after 404, got %s", d.Id())
	}
}

func TestResourceVolumeDelete(t *testing.T) {
	vol := testVolume()
	deletedCalled := false

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Pattern: "/v2/iac/projects/test-project-id/volumes/vol-001",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case "GET":
					if deletedCalled {
						w.WriteHeader(http.StatusNotFound)
						return
					}
					testhelpers.JSONHandler(t, http.StatusOK, dto.VolumeResponse{Volume: vol})(w, r)
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

	res := ResourceVolume()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":        "test-volume",
		"description": "",
		"size":        50,
		"volume_type": "SSD",
		"encrypt":     false,
		"multiattach": false,
		"snapshot_id": "",
	})
	d.SetId("vol-001")

	diags := res.DeleteContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if !deletedCalled {
		t.Error("expected DELETE to have been called")
	}
}
