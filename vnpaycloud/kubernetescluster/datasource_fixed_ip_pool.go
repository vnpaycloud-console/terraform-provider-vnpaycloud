package kubernetescluster

import (
	"context"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceKubernetesFixedIPPool() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceFixedIPPoolRead,
		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"worker_group_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"slots": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name":         {Type: schema.TypeString, Computed: true},
						"ip_address":   {Type: schema.TypeString, Computed: true},
						"claimed":      {Type: schema.TypeBool, Computed: true},
						"machine_name": {Type: schema.TypeString, Computed: true},
					},
				},
			},
		},
	}
}

func dataSourceFixedIPPoolRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	clusterID := d.Get("cluster_id").(string)
	workerGroupID := d.Get("worker_group_id").(string)

	path := client.ApiPath.ClusterFixedIPPool(cfg.ProjectID, clusterID)
	id := clusterID
	if workerGroupID != "" {
		path = client.ApiPath.WorkerGroupFixedIPPool(cfg.ProjectID, clusterID, workerGroupID)
		id = clusterID + "/" + workerGroupID
	}

	resp := &dto.FixedIPPoolResponse{}
	if _, err := cfg.Client.Get(ctx, path, resp, nil); err != nil {
		return diag.Errorf("Error retrieving vnpaycloud_kubernetes_fixed_ip_pool %s: %s", id, err)
	}

	slots := make([]map[string]interface{}, 0, len(resp.Slots))
	for _, s := range resp.Slots {
		slots = append(slots, map[string]interface{}{
			"name":         s.Name,
			"ip_address":   s.IPAddress,
			"claimed":      s.Claimed,
			"machine_name": s.MachineName,
		})
	}

	d.SetId(id)
	d.Set("name", resp.Name)
	d.Set("slots", slots)
	return nil
}
