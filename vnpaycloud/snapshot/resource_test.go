package snapshot

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// testSnapshot returns a fully populated dto.Snapshot for use in tests.
func testSnapshot() dto.Snapshot {
	return dto.Snapshot{
		ID:          "snap-001",
		Name:        "test-snapshot",
		Description: "a test snapshot",
		VolumeID:    "vol-001",
		SizeGB:      50,
		Status:      "available",
		CreatedAt:   "2025-01-15T10:00:00Z",
		ProjectID:   testhelpers.TestProjectID,
		ZoneID:      testhelpers.TestZoneID,
	}
}

func TestResourceSnapshotCreate(t *testing.T) {
	snap := testSnapshot()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "POST",
			Pattern: "/v2/iac/projects/test-project-id/snapshots",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.SnapshotResponse{Snapshot: snap}),
		},
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/snapshots/snap-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.SnapshotResponse{Snapshot: snap}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceSnapshot()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":        "test-snapshot",
		"description": "a test snapshot",
		"volume_id":   "vol-001",
	})

	diags := res.CreateContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "snap-001" {
		t.Errorf("expected ID snap-001, got %s", d.Id())
	}
	if v := d.Get("name").(string); v != "test-snapshot" {
		t.Errorf("expected name test-snapshot, got %s", v)
	}
	if v := d.Get("status").(string); v != "available" {
		t.Errorf("expected status available, got %s", v)
	}
}

func TestResourceSnapshotRead(t *testing.T) {
	snap := testSnapshot()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/snapshots/snap-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.SnapshotResponse{Snapshot: snap}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceSnapshot()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":        "test-snapshot",
		"description": "",
		"volume_id":   "vol-001",
	})
	d.SetId("snap-001")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if v := d.Get("name").(string); v != "test-snapshot" {
		t.Errorf("expected name test-snapshot, got %s", v)
	}
	if v := d.Get("description").(string); v != "a test snapshot" {
		t.Errorf("expected description 'a test snapshot', got %s", v)
	}
	if v := d.Get("volume_id").(string); v != "vol-001" {
		t.Errorf("expected volume_id vol-001, got %s", v)
	}
	if v := d.Get("size").(int); v != 50 {
		t.Errorf("expected size 50, got %d", v)
	}
	if v := d.Get("status").(string); v != "available" {
		t.Errorf("expected status available, got %s", v)
	}
	if v := d.Get("created_at").(string); v != "2025-01-15T10:00:00Z" {
		t.Errorf("expected created_at 2025-01-15T10:00:00Z, got %s", v)
	}
}

func TestResourceSnapshotRead_NotFound(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/snapshots/snap-gone",
			Handler: testhelpers.EmptyHandler(http.StatusNotFound),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceSnapshot()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":        "",
		"description": "",
		"volume_id":   "",
	})
	d.SetId("snap-gone")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error on 404: %v", diags)
	}

	if d.Id() != "" {
		t.Errorf("expected resource ID to be cleared after 404, got %s", d.Id())
	}
}

func TestResourceSnapshotDelete(t *testing.T) {
	snap := testSnapshot()
	deletedCalled := false

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Pattern: "/v2/iac/projects/test-project-id/snapshots/snap-001",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case "GET":
					if deletedCalled {
						w.WriteHeader(http.StatusNotFound)
						return
					}
					testhelpers.JSONHandler(t, http.StatusOK, dto.SnapshotResponse{Snapshot: snap})(w, r)
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

	res := ResourceSnapshot()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":        "test-snapshot",
		"description": "",
		"volume_id":   "vol-001",
	})
	d.SetId("snap-001")

	diags := res.DeleteContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if !deletedCalled {
		t.Error("expected DELETE to have been called")
	}
}
