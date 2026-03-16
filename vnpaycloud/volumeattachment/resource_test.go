package volumeattachment

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// testVolumeAttachment returns a fully populated dto.VolumeAttachment for use in tests.
func testVolumeAttachment() dto.VolumeAttachment {
	return dto.VolumeAttachment{
		ID:         "att-001",
		VolumeID:   "vol-001",
		ServerID:   "srv-001",
		Device:     "/dev/vdb",
		Status:     "attached",
		AttachedAt: "2025-01-15T10:00:00Z",
		ProjectID:  testhelpers.TestProjectID,
		ZoneID:     testhelpers.TestZoneID,
	}
}

func TestResourceVolumeAttachmentCreate(t *testing.T) {
	att := testVolumeAttachment()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "POST",
			Pattern: "/v2/iac/projects/test-project-id/volumes/vol-001/attach",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.VolumeAttachmentResponse{Attachment: att}),
		},
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/volume-attachments/att-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.VolumeAttachmentResponse{Attachment: att}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceVolumeAttachment()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"volume_id": "vol-001",
		"server_id": "srv-001",
	})

	diags := res.CreateContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "att-001" {
		t.Errorf("expected ID att-001, got %s", d.Id())
	}
	if v := d.Get("volume_id").(string); v != "vol-001" {
		t.Errorf("expected volume_id vol-001, got %s", v)
	}
	if v := d.Get("server_id").(string); v != "srv-001" {
		t.Errorf("expected server_id srv-001, got %s", v)
	}
	if v := d.Get("device").(string); v != "/dev/vdb" {
		t.Errorf("expected device /dev/vdb, got %s", v)
	}
	if v := d.Get("status").(string); v != "attached" {
		t.Errorf("expected status attached, got %s", v)
	}
	if v := d.Get("attached_at").(string); v != "2025-01-15T10:00:00Z" {
		t.Errorf("expected attached_at 2025-01-15T10:00:00Z, got %s", v)
	}
}

func TestResourceVolumeAttachmentRead(t *testing.T) {
	att := testVolumeAttachment()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/volume-attachments/att-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.VolumeAttachmentResponse{Attachment: att}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceVolumeAttachment()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"volume_id": "vol-001",
		"server_id": "srv-001",
	})
	d.SetId("att-001")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if v := d.Get("volume_id").(string); v != "vol-001" {
		t.Errorf("expected volume_id vol-001, got %s", v)
	}
	if v := d.Get("server_id").(string); v != "srv-001" {
		t.Errorf("expected server_id srv-001, got %s", v)
	}
	if v := d.Get("device").(string); v != "/dev/vdb" {
		t.Errorf("expected device /dev/vdb, got %s", v)
	}
	if v := d.Get("status").(string); v != "attached" {
		t.Errorf("expected status attached, got %s", v)
	}
	if v := d.Get("attached_at").(string); v != "2025-01-15T10:00:00Z" {
		t.Errorf("expected attached_at 2025-01-15T10:00:00Z, got %s", v)
	}
}

func TestResourceVolumeAttachmentRead_NotFound(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/volume-attachments/att-gone",
			Handler: testhelpers.EmptyHandler(http.StatusNotFound),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceVolumeAttachment()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"volume_id": "vol-001",
		"server_id": "srv-001",
	})
	d.SetId("att-gone")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error on 404: %v", diags)
	}

	if d.Id() != "" {
		t.Errorf("expected resource ID to be cleared after 404, got %s", d.Id())
	}
}

func TestResourceVolumeAttachmentDelete(t *testing.T) {
	detachCalled := false

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method: "POST",
			Pattern: "/v2/iac/projects/test-project-id/volumes/vol-001/detach",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				detachCalled = true
				w.WriteHeader(http.StatusAccepted)
			},
		},
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/volume-attachments/att-001",
			Handler: testhelpers.EmptyHandler(http.StatusNotFound),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceVolumeAttachment()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"volume_id": "vol-001",
		"server_id": "srv-001",
	})
	d.SetId("att-001")

	diags := res.DeleteContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if !detachCalled {
		t.Error("expected detach POST to have been called")
	}
}
