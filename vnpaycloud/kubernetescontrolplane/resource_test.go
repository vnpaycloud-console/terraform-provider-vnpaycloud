package kubernetescontrolplane

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func baselineServer(t *testing.T, clusterID string, b dto.ClusterControlPlaneBaselineResponse, extra ...testhelpers.Route) *testhelpers.Route {
	t.Helper()
	return &testhelpers.Route{
		Method:  "GET",
		Pattern: client.ApiPath.ClusterControlPlaneBaseline(testhelpers.TestProjectID, clusterID),
		Handler: testhelpers.JSONHandler(t, http.StatusOK, b),
	}
}

func TestRollingUpdateCreateNoopWhenUpToDate(t *testing.T) {
	posted := false
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		*baselineServer(t, "clu-001", dto.ClusterControlPlaneBaselineResponse{Supported: true, AuditLogReady: true}),
		{
			Method:  "POST",
			Pattern: client.ApiPath.ClusterControlPlaneRollingUpdate(testhelpers.TestProjectID, "clu-001"),
			Handler: func(w http.ResponseWriter, r *http.Request) { posted = true; w.WriteHeader(http.StatusOK) },
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceKubernetesControlPlaneRollingUpdate()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{"cluster_id": "clu-001"})

	if diags := res.CreateContext(context.Background(), d, cfg); diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
	if posted {
		t.Error("rolling update must not be triggered when the baseline is met")
	}
	if d.Id() != "clu-001" || d.Get("performed").(bool) || !d.Get("audit_log_ready").(bool) {
		t.Errorf("unexpected state: id=%s performed=%v", d.Id(), d.Get("performed"))
	}
}

func TestRollingUpdateCreateUnsupported(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		*baselineServer(t, "clu-001", dto.ClusterControlPlaneBaselineResponse{Supported: false, UnsupportedReason: "Not available in this zone yet"}),
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceKubernetesControlPlaneRollingUpdate()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{"cluster_id": "clu-001"})

	diags := res.CreateContext(context.Background(), d, cfg)
	if !diags.HasError() {
		t.Fatal("expected an error for an unsupported cluster")
	}
	if d.Id() != "" {
		t.Error("expected no ID to be set")
	}
}

func TestAuditLogCreateAlreadyEnabled(t *testing.T) {
	posted := false
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		*baselineServer(t, "clu-001", dto.ClusterControlPlaneBaselineResponse{Supported: true, AuditLogEnabled: true}),
		{
			Method:  "POST",
			Pattern: client.ApiPath.ClusterAuditLogEnable(testhelpers.TestProjectID, "clu-001"),
			Handler: func(w http.ResponseWriter, r *http.Request) { posted = true; w.WriteHeader(http.StatusOK) },
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceKubernetesAuditLog()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{"cluster_id": "clu-001"})

	if diags := res.CreateContext(context.Background(), d, cfg); diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
	if posted {
		t.Error("enable must not be called when audit log is already on")
	}
	if d.Id() != "clu-001" || !d.Get("enabled").(bool) {
		t.Errorf("unexpected state: id=%s enabled=%v", d.Id(), d.Get("enabled"))
	}
}

func TestAuditLogCreateRequiresBaseline(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		*baselineServer(t, "clu-001", dto.ClusterControlPlaneBaselineResponse{
			Supported: true, RollingUpdateNeeded: true, OutdatedItems: []string{"CONTROL_PLANE_DISK"},
		}),
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceKubernetesAuditLog()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{"cluster_id": "clu-001"})

	if diags := res.CreateContext(context.Background(), d, cfg); !diags.HasError() {
		t.Fatal("expected an error when the control plane is not ready for audit log")
	}
}

func TestAuditLogReadDisabledRemovesFromState(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		*baselineServer(t, "clu-001", dto.ClusterControlPlaneBaselineResponse{Supported: true, AuditLogEnabled: false}),
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceKubernetesAuditLog()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{})
	d.SetId("clu-001")

	if diags := res.ReadContext(context.Background(), d, cfg); diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
	if d.Id() != "" {
		t.Error("expected the resource to be removed from state when audit log is off")
	}
}

func TestDataSourceBaselineRead(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		*baselineServer(t, "clu-001", dto.ClusterControlPlaneBaselineResponse{
			Supported: true, RollingUpdateNeeded: true, OutdatedItems: []string{"AUDIT_POLICY"}, AuditLogEnabled: true,
		}),
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceKubernetesControlPlaneBaseline()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{"cluster_id": "clu-001"})

	if diags := ds.ReadContext(context.Background(), d, cfg); diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
	items := d.Get("outdated_items").([]interface{})
	if d.Id() != "clu-001" || !d.Get("rolling_update_needed").(bool) || len(items) != 1 || items[0] != "AUDIT_POLICY" {
		t.Errorf("unexpected state: %v", d.State().Attributes)
	}
}
