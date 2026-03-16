package kubernetescluster

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// testCluster returns a fully populated dto.K8sCluster for use in tests.
func testCluster() dto.K8sCluster {
	return dto.K8sCluster{
		ID:          "cluster-001",
		Name:        "test-cluster",
		Zone:        "zone-a",
		K8sVersion:  "1.28",
		Purpose:     "development",
		SubnetID:    "subnet-001",
		CniPlugin:   "calico",
		PodCidr:     "10.244.0.0/16",
		ServiceCidr: "10.96.0.0/12",
		PrivateGwID: "pgw-001",
		ClusterSize: "small",
		ApiEndpoint: "https://k8s.example.com:6443",
		PrivateIP:   "10.0.0.100",
		Status:      "active",
		CreatedAt:   "2025-01-15T10:00:00Z",
	}
}

func TestResourceClusterCreate(t *testing.T) {
	cluster := testCluster()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "POST",
			Pattern: "/v2/iac/projects/test-project-id/clusters",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.K8sClusterResponse{Cluster: cluster}),
		},
		{
			Pattern: "/v2/iac/projects/test-project-id/clusters/cluster-001",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" {
					http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
					return
				}
				testhelpers.JSONHandler(t, http.StatusOK, dto.K8sClusterResponse{Cluster: cluster})(w, r)
			},
		},
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/clusters/cluster-001/kubeconfig",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.KubeconfigResponse{Kubeconfig: "apiVersion: v1\nkind: Config"}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceKubernetesCluster()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":                       "test-cluster",
		"k8s_version":                "1.28",
		"purpose":                    "development",
		"private_gw_id":              "pgw-001",
		"subnet_id":                  "subnet-001",
		"cni_plugin":                 "calico",
		"pod_cidr":                   "10.244.0.0/16",
		"service_cidr":               "10.96.0.0/12",
		"cluster_size":               "small",
		"default_worker_name":        "default-wg",
		"default_worker_flavor":      "v1.small",
		"default_worker_count":       3,
		"default_worker_volume_type": "SSD",
		"default_worker_volume_size": 50,
		"default_worker_ssh_key_id":  "key-001",
	})

	diags := res.CreateContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "cluster-001" {
		t.Errorf("expected ID cluster-001, got %s", d.Id())
	}
	if v := d.Get("name").(string); v != "test-cluster" {
		t.Errorf("expected name test-cluster, got %s", v)
	}
	if v := d.Get("status").(string); v != "active" {
		t.Errorf("expected status active, got %s", v)
	}
	if v := d.Get("api_endpoint").(string); v != "https://k8s.example.com:6443" {
		t.Errorf("expected api_endpoint https://k8s.example.com:6443, got %s", v)
	}
	if v := d.Get("kubeconfig").(string); v != "apiVersion: v1\nkind: Config" {
		t.Errorf("expected kubeconfig to be set, got %q", v)
	}
}

func TestResourceClusterRead(t *testing.T) {
	cluster := testCluster()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/clusters/cluster-001",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.K8sClusterResponse{Cluster: cluster}),
		},
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/clusters/cluster-001/kubeconfig",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.KubeconfigResponse{Kubeconfig: "kubeconfig-data"}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceKubernetesCluster()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":                       "",
		"subnet_id":                  "",
		"default_worker_flavor":      "",
		"default_worker_count":       1,
		"k8s_version":                "",
		"purpose":                    "",
		"private_gw_id":              "",
		"cni_plugin":                 "",
		"pod_cidr":                   "",
		"service_cidr":               "",
		"cluster_size":               "",
		"default_worker_name":        "",
		"default_worker_volume_type": "",
		"default_worker_volume_size": 0,
		"default_worker_ssh_key_id":  "",
	})
	d.SetId("cluster-001")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if v := d.Get("name").(string); v != "test-cluster" {
		t.Errorf("expected name test-cluster, got %s", v)
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
	if v := d.Get("kubeconfig").(string); v != "kubeconfig-data" {
		t.Errorf("expected kubeconfig kubeconfig-data, got %s", v)
	}
}

func TestResourceClusterRead_NotFound(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/clusters/cluster-gone",
			Handler: testhelpers.EmptyHandler(http.StatusNotFound),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceKubernetesCluster()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":                       "",
		"subnet_id":                  "",
		"default_worker_flavor":      "",
		"default_worker_count":       1,
		"k8s_version":                "",
		"purpose":                    "",
		"private_gw_id":              "",
		"cni_plugin":                 "",
		"pod_cidr":                   "",
		"service_cidr":               "",
		"cluster_size":               "",
		"default_worker_name":        "",
		"default_worker_volume_type": "",
		"default_worker_volume_size": 0,
		"default_worker_ssh_key_id":  "",
	})
	d.SetId("cluster-gone")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error on 404: %v", diags)
	}

	if d.Id() != "" {
		t.Errorf("expected resource ID to be cleared after 404, got %s", d.Id())
	}
}

func TestResourceClusterDelete(t *testing.T) {
	cluster := testCluster()
	deletedCalled := false

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Pattern: "/v2/iac/projects/test-project-id/clusters/cluster-001",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case "GET":
					if deletedCalled {
						w.WriteHeader(http.StatusNotFound)
						return
					}
					testhelpers.JSONHandler(t, http.StatusOK, dto.K8sClusterResponse{Cluster: cluster})(w, r)
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

	res := ResourceKubernetesCluster()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"name":                       "test-cluster",
		"subnet_id":                  "subnet-001",
		"default_worker_flavor":      "v1.small",
		"default_worker_count":       1,
		"k8s_version":                "",
		"purpose":                    "",
		"private_gw_id":              "",
		"cni_plugin":                 "",
		"pod_cidr":                   "",
		"service_cidr":               "",
		"cluster_size":               "",
		"default_worker_name":        "",
		"default_worker_volume_type": "",
		"default_worker_volume_size": 0,
		"default_worker_ssh_key_id":  "",
	})
	d.SetId("cluster-001")

	diags := res.DeleteContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if !deletedCalled {
		t.Error("expected DELETE to have been called")
	}
}
