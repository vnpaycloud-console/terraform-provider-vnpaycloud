package kubernetescredentialsrotation

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestResourceRotationSchema(t *testing.T) {
	res := ResourceKubernetesCredentialsRotation()
	if !res.Schema["cluster_id"].Required || !res.Schema["cluster_id"].ForceNew {
		t.Error("expected cluster_id to be Required+ForceNew")
	}
	if !res.Schema["triggers"].ForceNew || res.Schema["triggers"].Type != schema.TypeMap {
		t.Error("expected triggers to be a ForceNew map")
	}
	if !res.Schema["rotated_at"].Computed {
		t.Error("expected rotated_at to be Computed")
	}
	if res.UpdateContext != nil {
		t.Error("rotation must not have an Update, every change forces a new run")
	}
}

func TestResourceRotationCreateFailsFast(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "POST",
			Pattern: client.ApiPath.ClusterRotateCredentials(testhelpers.TestProjectID, "clu-001"),
			Handler: testhelpers.JSONHandler(t, http.StatusPreconditionFailed, map[string]string{"message": "Cluster clu-001 is not active"}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceKubernetesCredentialsRotation()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{"cluster_id": "clu-001"})

	if diags := res.CreateContext(context.Background(), d, cfg); !diags.HasError() {
		t.Fatal("expected an error when the backend rejects the rotation")
	}
	if d.Id() != "" {
		t.Error("expected no ID when the rotation was rejected")
	}
}

func TestResourceRotationCreate(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "POST",
			Pattern: client.ApiPath.ClusterRotateCredentials(testhelpers.TestProjectID, "clu-001"),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.K8sClusterResponse{Cluster: dto.K8sCluster{ID: "clu-001", Status: "updating"}}),
		},
		{
			Method:  "GET",
			Pattern: client.ApiPath.ClusterWithID(testhelpers.TestProjectID, "clu-001"),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.K8sClusterResponse{Cluster: dto.K8sCluster{ID: "clu-001", Status: "active"}}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceKubernetesCredentialsRotation()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{"cluster_id": "clu-001"})

	if diags := res.CreateContext(context.Background(), d, cfg); diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
	if d.Id() == "" || d.Get("rotated_at").(string) == "" {
		t.Errorf("expected ID and rotated_at to be set, got id=%q rotated_at=%q", d.Id(), d.Get("rotated_at"))
	}
}
