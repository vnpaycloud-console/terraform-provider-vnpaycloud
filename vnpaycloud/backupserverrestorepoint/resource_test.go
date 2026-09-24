package backupserverrestorepoint

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func testRestorePoint(status string) dto.BackupServerRestorePoint {
	return dto.BackupServerRestorePoint{
		ID:                "rp-001",
		Name:              "duytq_20260828T0932Z07",
		BackupServerID:    "bks-001",
		BackupPolicyID:    "pol-001",
		Type:              "vm",
		Purpose:           "on_demand",
		ZoneID:            testhelpers.TestZoneID,
		CreateByBackupNow: true,
		BackupPoint:       "2026-08-28T02:32:56Z",
		Status:            status,
		CreatedAt:         "2026-08-28T02:32:58Z",
	}
}

func TestResourceBackupServerRestorePointSchema(t *testing.T) {
	res := ResourceBackupServerRestorePoint()

	if !res.Schema["backup_server_id"].Required || !res.Schema["backup_server_id"].ForceNew {
		t.Error("expected backup_server_id to be Required+ForceNew")
	}
	if res.UpdateContext != nil {
		t.Error("expected no UpdateContext (restore point is not updatable)")
	}
	for _, key := range []string{"name", "backup_policy_id", "type", "purpose", "zone_id", "status", "backup_point"} {
		if !res.Schema[key].Computed {
			t.Errorf("expected %s to be Computed", key)
		}
	}
}

func TestResourceBackupServerRestorePointCreate(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Pattern: client.ApiPath.BackupServerRestorePoints(testhelpers.TestProjectID),
			Handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodPost {
					testhelpers.JSONHandler(t, http.StatusOK,
						dto.BackupServerRestorePointResponse{RestorePoint: testRestorePoint("active")})(w, r)
					return
				}
				testhelpers.JSONHandler(t, http.StatusOK,
					dto.ListBackupServerRestorePointsResponse{
						RestorePoints: []dto.BackupServerRestorePoint{testRestorePoint("active")},
					})(w, r)
			},
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceBackupServerRestorePoint()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"backup_server_id": "bks-001",
	})
	d.Set("backup_server_id", "bks-001")

	diags := res.CreateContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
	if d.Id() != "rp-001" {
		t.Errorf("expected id rp-001, got %s", d.Id())
	}
	if v := d.Get("status").(string); v != "active" {
		t.Errorf("expected status active, got %s", v)
	}
	if !d.Get("create_by_backup_now").(bool) {
		t.Error("expected create_by_backup_now true")
	}
}

// Rate limits on this path clear within seconds (per-second limits on the calls behind
// the create, a few-per-minute quota on on-demand backups), so the create must back off
// and retry like any other request instead of failing on the first rejection.
func TestResourceBackupServerRestorePointCreateRetriesRateLimit(t *testing.T) {
	var posts int32
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Pattern: client.ApiPath.BackupServerRestorePoints(testhelpers.TestProjectID),
			Handler: func(w http.ResponseWriter, r *http.Request) {
				atomic.AddInt32(&posts, 1)
				w.WriteHeader(http.StatusServiceUnavailable)
				_, _ = w.Write([]byte(`{"code":14,"message":"backup_server_restore_point \"bks-001\": Too Many Requests"}`))
			},
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceBackupServerRestorePoint()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"backup_server_id": "bks-001",
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	start := time.Now()
	diags := res.CreateContext(ctx, d, cfg)
	elapsed := time.Since(start)

	if !diags.HasError() {
		t.Fatal("expected an error once the context ran out")
	}
	if elapsed < time.Second {
		t.Errorf("expected the create to back off and wait, it failed after %s", elapsed)
	}
	if n := atomic.LoadInt32(&posts); n != 1 {
		t.Errorf("expected 1 POST before the backoff was cut short, got %d", n)
	}
}

func TestResourceBackupServerRestorePointReadGone(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.BackupServerRestorePoints(testhelpers.TestProjectID),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListBackupServerRestorePointsResponse{}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceBackupServerRestorePoint()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{})
	d.SetId("rp-001")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("expected a missing restore point to clear the ID, got: %v", diags)
	}
	if d.Id() != "" {
		t.Errorf("expected ID cleared, got %s", d.Id())
	}
}

// Delete must not return until the restore point actually disappears: the
// backend keeps it in "deleting" long after the DELETE call is accepted.
func TestResourceBackupServerRestorePointDeleteWaits(t *testing.T) {
	var lists int32
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Pattern: client.ApiPath.BackupServerRestorePointWithID(testhelpers.TestProjectID, "rp-001"),
			Handler: testhelpers.EmptyHandler(http.StatusOK),
		},
		{
			Method:  "GET",
			Pattern: client.ApiPath.BackupServerRestorePoints(testhelpers.TestProjectID),
			Handler: func(w http.ResponseWriter, r *http.Request) {
				if atomic.AddInt32(&lists, 1) < 2 {
					testhelpers.JSONHandler(t, http.StatusOK,
						dto.ListBackupServerRestorePointsResponse{
							RestorePoints: []dto.BackupServerRestorePoint{testRestorePoint("deleting")},
						})(w, r)
					return
				}
				testhelpers.JSONHandler(t, http.StatusOK, dto.ListBackupServerRestorePointsResponse{})(w, r)
			},
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceBackupServerRestorePoint()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"backup_server_id": "bks-001",
	})
	d.SetId("rp-001")
	d.Set("backup_server_id", "bks-001")

	diags := res.DeleteContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
	if n := atomic.LoadInt32(&lists); n < 2 {
		t.Errorf("expected the delete to poll until the restore point was gone, polled %d times", n)
	}
}

// A restore point that never shows up is a rolled-back backup, and the wait must
// say so rather than time out on an opaque "couldn't find resource".
func TestResourceBackupServerRestorePointCreateRolledBack(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Pattern: client.ApiPath.BackupServerRestorePoints(testhelpers.TestProjectID),
			Handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodPost {
					testhelpers.JSONHandler(t, http.StatusOK,
						dto.BackupServerRestorePointResponse{RestorePoint: testRestorePoint("creating")})(w, r)
					return
				}
				testhelpers.JSONHandler(t, http.StatusOK, dto.ListBackupServerRestorePointsResponse{})(w, r)
			},
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceBackupServerRestorePoint()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"backup_server_id": "bks-001",
	})
	d.Set("backup_server_id", "bks-001")

	defer func(orig time.Duration) { notFoundGrace = orig }(notFoundGrace)
	notFoundGrace = 0

	diags := res.CreateContext(context.Background(), d, cfg)
	if !diags.HasError() {
		t.Fatal("expected an error for a rolled-back restore point")
	}
	if !strings.Contains(diags[0].Summary, "rolled back") {
		t.Errorf("expected a rollback message, got: %s", diags[0].Summary)
	}
}

// On timeout the wait hands back no object, so the error must still name the
// status the backend last reported instead of a bare deadline message.
func TestResourceBackupServerRestorePointDeleteTimeoutReportsStatus(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Pattern: client.ApiPath.BackupServerRestorePointWithID(testhelpers.TestProjectID, "rp-001"),
			Handler: testhelpers.EmptyHandler(http.StatusOK),
		},
		{
			Method:  "GET",
			Pattern: client.ApiPath.BackupServerRestorePoints(testhelpers.TestProjectID),
			Handler: testhelpers.JSONHandler(t, http.StatusOK,
				dto.ListBackupServerRestorePointsResponse{
					RestorePoints: []dto.BackupServerRestorePoint{testRestorePoint("deleting")},
				}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceBackupServerRestorePoint()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"backup_server_id": "bks-001",
	})
	d.SetId("rp-001")
	d.Set("backup_server_id", "bks-001")

	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	diags := res.DeleteContext(ctx, d, cfg)
	if !diags.HasError() {
		t.Fatal("expected the delete to fail once the wait ran out")
	}
	if !strings.Contains(diags[0].Summary, `status "deleting"`) {
		t.Errorf("expected the last observed status in the error, got: %s", diags[0].Summary)
	}
}

func latestBackupGuardHandler(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
	_, _ = w.Write([]byte(`{"code":3,"message":"backup_server_restore_point \"rp-001\": unable to delete tier-1 latest backup"}`))
}

// Destroying several restore points of one server in a single run makes the backend refuse
// the newest one while the older ones are still deleting; the delete must wait that out.
func TestResourceBackupServerRestorePointDeleteRetriesLatestGuard(t *testing.T) {
	var deletes, lists int32
	older := testRestorePoint("deleting")
	older.ID = "rp-000"

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Pattern: client.ApiPath.BackupServerRestorePointWithID(testhelpers.TestProjectID, "rp-001"),
			Handler: func(w http.ResponseWriter, r *http.Request) {
				if atomic.AddInt32(&deletes, 1) == 1 {
					latestBackupGuardHandler(w)
					return
				}
				w.WriteHeader(http.StatusOK)
			},
		},
		{
			Method:  "GET",
			Pattern: client.ApiPath.BackupServerRestorePoints(testhelpers.TestProjectID),
			Handler: func(w http.ResponseWriter, r *http.Request) {
				points := []dto.BackupServerRestorePoint{older}
				if atomic.AddInt32(&lists, 1) == 1 {
					points = append(points, testRestorePoint("active"))
				}
				testhelpers.JSONHandler(t, http.StatusOK,
					dto.ListBackupServerRestorePointsResponse{RestorePoints: points})(w, r)
			},
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	defer func(orig time.Duration) { latestBackupGuardGrace = orig }(latestBackupGuardGrace)
	latestBackupGuardGrace = 0

	res := ResourceBackupServerRestorePoint()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"backup_server_id": "bks-001",
	})
	d.SetId("rp-001")

	if diags := res.DeleteContext(context.Background(), d, cfg); diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
	if n := atomic.LoadInt32(&deletes); n != 2 {
		t.Errorf("expected the refused delete to be retried once, got %d DELETE calls", n)
	}
}

// With no older restore point being deleted the refusal will not clear on its own, so the
// delete must fail straight away and say what to do instead of waiting out the timeout.
func TestResourceBackupServerRestorePointDeleteLatestGuardWithoutPendingDeletion(t *testing.T) {
	var deletes int32
	older := testRestorePoint("active")
	older.ID = "rp-000"

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Pattern: client.ApiPath.BackupServerRestorePointWithID(testhelpers.TestProjectID, "rp-001"),
			Handler: func(w http.ResponseWriter, r *http.Request) {
				atomic.AddInt32(&deletes, 1)
				latestBackupGuardHandler(w)
			},
		},
		{
			Method:  "GET",
			Pattern: client.ApiPath.BackupServerRestorePoints(testhelpers.TestProjectID),
			Handler: testhelpers.JSONHandler(t, http.StatusOK,
				dto.ListBackupServerRestorePointsResponse{
					RestorePoints: []dto.BackupServerRestorePoint{older, testRestorePoint("active")},
				}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	defer func(orig time.Duration) { latestBackupGuardGrace = orig }(latestBackupGuardGrace)
	latestBackupGuardGrace = 0

	res := ResourceBackupServerRestorePoint()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"backup_server_id": "bks-001",
	})
	d.SetId("rp-001")

	start := time.Now()
	diags := res.DeleteContext(context.Background(), d, cfg)
	if !diags.HasError() {
		t.Fatal("expected the delete to fail")
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Errorf("expected an immediate failure, took %s", elapsed)
	}
	if n := atomic.LoadInt32(&deletes); n != 1 {
		t.Errorf("expected a single DELETE, got %d", n)
	}
	if !strings.Contains(diags[0].Summary, "delete those first") {
		t.Errorf("expected guidance in the error, got: %s", diags[0].Summary)
	}
}

// The backend rejects a second on-demand backup while one is still running on the same
// server backup, so creates for one server must run one after another.
func TestResourceBackupServerRestorePointCreateSerializesPerBackupServer(t *testing.T) {
	var (
		mu      sync.Mutex
		running string
		posts   int
		polls   = map[string]int{}
	)
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Pattern: client.ApiPath.BackupServerRestorePoints(testhelpers.TestProjectID),
			Handler: func(w http.ResponseWriter, r *http.Request) {
				mu.Lock()
				defer mu.Unlock()

				if r.Method == http.MethodPost {
					if running != "" {
						w.WriteHeader(http.StatusBadRequest)
						_, _ = w.Write([]byte(`{"code":3,"message":"ServerBK is being backed up. Please wait a few minutes and try again!"}`))
						return
					}
					posts++
					running = fmt.Sprintf("rp-%03d", posts)
					rp := testRestorePoint("creating")
					rp.ID = running
					testhelpers.JSONHandler(t, http.StatusOK, dto.BackupServerRestorePointResponse{RestorePoint: rp})(w, r)
					return
				}

				var points []dto.BackupServerRestorePoint
				for i := 1; i <= posts; i++ {
					rp := testRestorePoint("active")
					rp.ID = fmt.Sprintf("rp-%03d", i)
					if rp.ID == running {
						polls[running]++
						if polls[running] < 2 {
							rp.Status = "creating"
						} else {
							running = ""
						}
					}
					points = append(points, rp)
				}
				testhelpers.JSONHandler(t, http.StatusOK,
					dto.ListBackupServerRestorePointsResponse{RestorePoints: points})(w, r)
			},
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)
	res := ResourceBackupServerRestorePoint()

	var wg sync.WaitGroup
	errs := make([]error, 2)
	for i := range errs {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
				"backup_server_id": "bks-001",
			})
			if diags := res.CreateContext(context.Background(), d, cfg); diags.HasError() {
				errs[i] = fmt.Errorf("%s", diags[0].Summary)
			}
		}(i)
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Errorf("create %d failed: %v", i, err)
		}
	}
	if posts != 2 {
		t.Errorf("expected 2 accepted POSTs, got %d", posts)
	}
}
