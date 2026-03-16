package kubernetescluster

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestDataSourceClusterRead_ByID(t *testing.T) {
	cluster := testCluster()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/clusters/cluster-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.K8sClusterResponse{Cluster: cluster}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceKubernetesCluster()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"id": "cluster-001",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "cluster-001" {
		t.Errorf("expected ID cluster-001, got %s", d.Id())
	}
	if v := d.Get("name").(string); v != "test-cluster" {
		t.Errorf("expected name test-cluster, got %s", v)
	}
	if v := d.Get("k8s_version").(string); v != "1.28" {
		t.Errorf("expected k8s_version 1.28, got %s", v)
	}
	if v := d.Get("purpose").(string); v != "development" {
		t.Errorf("expected purpose development, got %s", v)
	}
	if v := d.Get("subnet_id").(string); v != "subnet-001" {
		t.Errorf("expected subnet_id subnet-001, got %s", v)
	}
	if v := d.Get("cni_plugin").(string); v != "calico" {
		t.Errorf("expected cni_plugin calico, got %s", v)
	}
	if v := d.Get("pod_cidr").(string); v != "10.244.0.0/16" {
		t.Errorf("expected pod_cidr 10.244.0.0/16, got %s", v)
	}
	if v := d.Get("service_cidr").(string); v != "10.96.0.0/12" {
		t.Errorf("expected service_cidr 10.96.0.0/12, got %s", v)
	}
	if v := d.Get("cluster_size").(string); v != "small" {
		t.Errorf("expected cluster_size small, got %s", v)
	}
	if v := d.Get("zone").(string); v != "zone-a" {
		t.Errorf("expected zone zone-a, got %s", v)
	}
	if v := d.Get("api_endpoint").(string); v != "https://k8s.example.com:6443" {
		t.Errorf("expected api_endpoint https://k8s.example.com:6443, got %s", v)
	}
	if v := d.Get("private_ip").(string); v != "10.0.0.100" {
		t.Errorf("expected private_ip 10.0.0.100, got %s", v)
	}
	if v := d.Get("status").(string); v != "active" {
		t.Errorf("expected status active, got %s", v)
	}
	if v := d.Get("created_at").(string); v != "2025-01-15T10:00:00Z" {
		t.Errorf("expected created_at 2025-01-15T10:00:00Z, got %s", v)
	}
}

func TestDataSourceClusterRead_ByName(t *testing.T) {
	cluster := testCluster()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method: "GET",
			Pattern: "/v2/iac/projects/test-project-id/clusters",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListK8sClustersResponse{
				Clusters: []dto.K8sCluster{cluster},
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceKubernetesCluster()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"name": "test-cluster",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "cluster-001" {
		t.Errorf("expected ID cluster-001, got %s", d.Id())
	}
	if v := d.Get("name").(string); v != "test-cluster" {
		t.Errorf("expected name test-cluster, got %s", v)
	}
}

func TestDataSourceClusterRead_NotFound(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method: "GET",
			Pattern: "/v2/iac/projects/test-project-id/clusters",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListK8sClustersResponse{
				Clusters: []dto.K8sCluster{},
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceKubernetesCluster()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"name": "nonexistent",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if !diags.HasError() {
		t.Fatal("expected error for nonexistent cluster, got none")
	}
}

func TestDataSourceClustersRead(t *testing.T) {
	cluster1 := testCluster()
	cluster2 := dto.K8sCluster{
		ID:          "cluster-002",
		Name:        "prod-cluster",
		K8sVersion:  "1.29",
		ClusterSize: "large",
		Zone:        "zone-b",
		ApiEndpoint: "https://k8s-prod.example.com:6443",
		Status:      "active",
		CreatedAt:   "2025-02-01T08:00:00Z",
	}

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method: "GET",
			Pattern: "/v2/iac/projects/test-project-id/clusters",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListK8sClustersResponse{
				Clusters: []dto.K8sCluster{cluster1, cluster2},
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceKubernetesClusters()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	clusters := d.Get("clusters").([]interface{})
	if len(clusters) != 2 {
		t.Fatalf("expected 2 clusters, got %d", len(clusters))
	}

	first := clusters[0].(map[string]interface{})
	if first["id"] != "cluster-001" {
		t.Errorf("expected first cluster id cluster-001, got %v", first["id"])
	}
	if first["name"] != "test-cluster" {
		t.Errorf("expected first cluster name test-cluster, got %v", first["name"])
	}
	if first["k8s_version"] != "1.28" {
		t.Errorf("expected first cluster k8s_version 1.28, got %v", first["k8s_version"])
	}
	if first["status"] != "active" {
		t.Errorf("expected first cluster status active, got %v", first["status"])
	}

	second := clusters[1].(map[string]interface{})
	if second["id"] != "cluster-002" {
		t.Errorf("expected second cluster id cluster-002, got %v", second["id"])
	}
	if second["name"] != "prod-cluster" {
		t.Errorf("expected second cluster name prod-cluster, got %v", second["name"])
	}
	if second["k8s_version"] != "1.29" {
		t.Errorf("expected second cluster k8s_version 1.29, got %v", second["k8s_version"])
	}
	if second["cluster_size"] != "large" {
		t.Errorf("expected second cluster cluster_size large, got %v", second["cluster_size"])
	}
}
