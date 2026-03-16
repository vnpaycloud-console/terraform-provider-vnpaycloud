package snapshot

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestDataSourceSnapshotRead_ByID(t *testing.T) {
	snap := testSnapshot()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/snapshots/snap-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.SnapshotResponse{Snapshot: snap}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceSnapshot()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"id": "snap-001",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "snap-001" {
		t.Errorf("expected ID snap-001, got %s", d.Id())
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

func TestDataSourceSnapshotRead_ByName(t *testing.T) {
	snap := testSnapshot()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/snapshots",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListSnapshotsResponse{
				Snapshots: []dto.Snapshot{snap},
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceSnapshot()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"name": "test-snapshot",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "snap-001" {
		t.Errorf("expected ID snap-001, got %s", d.Id())
	}
	if v := d.Get("name").(string); v != "test-snapshot" {
		t.Errorf("expected name test-snapshot, got %s", v)
	}
}

func TestDataSourceSnapshotRead_NotFound(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/snapshots",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListSnapshotsResponse{
				Snapshots: []dto.Snapshot{},
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceSnapshot()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"name": "nonexistent",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if !diags.HasError() {
		t.Fatal("expected error for nonexistent snapshot, got none")
	}
}

func TestDataSourceSnapshotsRead(t *testing.T) {
	snap1 := testSnapshot()
	snap2 := dto.Snapshot{
		ID:          "snap-002",
		Name:        "test-snapshot-2",
		Description: "second snapshot",
		VolumeID:    "vol-002",
		SizeGB:      100,
		Status:      "available",
		CreatedAt:   "2025-01-16T12:00:00Z",
		ProjectID:   testhelpers.TestProjectID,
		ZoneID:      testhelpers.TestZoneID,
	}

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/snapshots",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListSnapshotsResponse{
				Snapshots: []dto.Snapshot{snap1, snap2},
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceSnapshots()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	snapshots := d.Get("snapshots").([]interface{})
	if len(snapshots) != 2 {
		t.Fatalf("expected 2 snapshots, got %d", len(snapshots))
	}

	first := snapshots[0].(map[string]interface{})
	if first["id"] != "snap-001" {
		t.Errorf("expected first snapshot id snap-001, got %v", first["id"])
	}
	if first["name"] != "test-snapshot" {
		t.Errorf("expected first snapshot name test-snapshot, got %v", first["name"])
	}

	second := snapshots[1].(map[string]interface{})
	if second["id"] != "snap-002" {
		t.Errorf("expected second snapshot id snap-002, got %v", second["id"])
	}
	if second["name"] != "test-snapshot-2" {
		t.Errorf("expected second snapshot name test-snapshot-2, got %v", second["name"])
	}
}
