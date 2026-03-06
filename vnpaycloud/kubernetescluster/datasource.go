package kubernetescluster

import (
	"context"
	"fmt"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceKubernetesCluster() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceClusterRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"k8s_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"purpose": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"private_gw_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"subnet_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cni_plugin": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"pod_cidr": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"service_cidr": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_size": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"zone": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"api_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"private_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if id, ok := d.GetOk("id"); ok {
		resp := &dto.K8sClusterResponse{}
		_, err := cfg.Client.Get(ctx, client.ApiPath.ClusterWithID(cfg.ProjectID, id.(string)), resp, nil)
		if err != nil {
			return diag.Errorf("Error fetching vnpaycloud_kubernetes_cluster %s: %s", id, err)
		}
		return setClusterData(d, &resp.Cluster)
	}

	listResp := &dto.ListK8sClustersResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.Clusters(cfg.ProjectID), listResp, nil)
	if err != nil {
		return diag.Errorf("Error listing vnpaycloud_kubernetes_cluster: %s", err)
	}

	nameFilter, nameOk := d.GetOk("name")

	for _, cluster := range listResp.Clusters {
		if nameOk && cluster.Name != nameFilter.(string) {
			continue
		}
		return setClusterData(d, &cluster)
	}

	return diag.Errorf("No vnpaycloud_kubernetes_cluster found matching the criteria")
}

func setClusterData(d *schema.ResourceData, c *dto.K8sCluster) diag.Diagnostics {
	d.SetId(c.ID)
	d.Set("name", c.Name)
	d.Set("k8s_version", c.K8sVersion)
	d.Set("purpose", c.Purpose)
	d.Set("private_gw_id", c.PrivateGwID)
	d.Set("subnet_id", c.SubnetID)
	d.Set("cni_plugin", c.CniPlugin)
	d.Set("pod_cidr", c.PodCidr)
	d.Set("service_cidr", c.ServiceCidr)
	d.Set("cluster_size", c.ClusterSize)
	d.Set("zone", c.Zone)
	d.Set("api_endpoint", c.ApiEndpoint)
	d.Set("private_ip", c.PrivateIP)
	d.Set("status", c.Status)
	d.Set("created_at", c.CreatedAt)
	return nil
}

func DataSourceKubernetesClusters() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceClustersRead,
		Schema: map[string]*schema.Schema{
			"clusters": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id":           {Type: schema.TypeString, Computed: true},
						"name":         {Type: schema.TypeString, Computed: true},
						"k8s_version":  {Type: schema.TypeString, Computed: true},
						"cluster_size": {Type: schema.TypeString, Computed: true},
						"zone":         {Type: schema.TypeString, Computed: true},
						"api_endpoint": {Type: schema.TypeString, Computed: true},
						"status":       {Type: schema.TypeString, Computed: true},
						"created_at":   {Type: schema.TypeString, Computed: true},
					},
				},
			},
		},
	}
}

func dataSourceClustersRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	listResp := &dto.ListK8sClustersResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.Clusters(cfg.ProjectID), listResp, nil)
	if err != nil {
		return diag.Errorf("Error listing vnpaycloud_kubernetes_clusters: %s", err)
	}

	var clusters []map[string]interface{}
	for _, c := range listResp.Clusters {
		clusters = append(clusters, map[string]interface{}{
			"id":           c.ID,
			"name":         c.Name,
			"k8s_version":  c.K8sVersion,
			"cluster_size": c.ClusterSize,
			"zone":         c.Zone,
			"api_endpoint": c.ApiEndpoint,
			"status":       c.Status,
			"created_at":   c.CreatedAt,
		})
	}

	d.SetId(fmt.Sprintf("clusters-%s", cfg.ProjectID))
	d.Set("clusters", clusters)

	return nil
}
