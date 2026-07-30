package kubernetesrbac

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestResourceKubernetesRbacSchema(t *testing.T) {
	res := ResourceKubernetesRbac()

	for _, key := range []string{"cluster_id", "user_id", "role"} {
		if !res.Schema[key].Required || !res.Schema[key].ForceNew {
			t.Errorf("expected %s to be Required+ForceNew", key)
		}
	}
}

func TestResourceKubernetesRbacRead(t *testing.T) {
	rbac := dto.KubernetesRbac{
		ID:                   "rbac-001",
		ClusterID:            "cluster-001",
		UserID:               "user-001",
		Email:                "dev@example.com",
		Role:                 "viewer",
		BindingType:          "namespace",
		IsClusterRoleBinding: false,
		Namespace:            "default",
		Name:                 "dev-binding",
		Status:               "active",
		CreatedAt:            "2026-01-15T10:00:00Z",
	}

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: client.ApiPath.KubernetesRbacWithID(testhelpers.TestProjectID, "cluster-001", "rbac-001"),
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.KubernetesRbacResponse{Rbac: rbac}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceKubernetesRbac()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"cluster_id": "cluster-001",
	})
	d.SetId("rbac-001")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if v := d.Get("user_id").(string); v != "user-001" {
		t.Errorf("expected user_id user-001, got %s", v)
	}
	if v := d.Get("role").(string); v != "viewer" {
		t.Errorf("expected role viewer, got %s", v)
	}
	if v := d.Get("namespace").(string); v != "default" {
		t.Errorf("expected namespace default, got %s", v)
	}
	if v := d.Get("email").(string); v != "dev@example.com" {
		t.Errorf("expected email dev@example.com, got %s", v)
	}
}
