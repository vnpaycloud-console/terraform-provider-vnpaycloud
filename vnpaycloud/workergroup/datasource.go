package workergroup

import (
	"context"
	"fmt"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceWorkerGroup() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceWorkerGroupRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"cluster_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"flavor": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"num_workers": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"auto_scaling": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"min_workers": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"max_workers": {
				Type:     schema.TypeInt,
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

func dataSourceWorkerGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	id := d.Get("id").(string)
	clusterID := d.Get("cluster_id").(string)

	resp := &dto.WorkerGroupResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.WorkerGroupWithID(cfg.ProjectID, clusterID, id), resp, nil)
	if err != nil {
		return diag.Errorf("Error retrieving vnpaycloud_kubernetes_worker_group %s: %s", id, err)
	}

	d.SetId(resp.WorkerGroup.ID)
	d.Set("cluster_id", resp.WorkerGroup.ClusterID)
	d.Set("name", resp.WorkerGroup.Name)
	d.Set("flavor", resp.WorkerGroup.Flavor)
	d.Set("num_workers", resp.WorkerGroup.NumWorkers)
	d.Set("auto_scaling", resp.WorkerGroup.AutoScaling)
	d.Set("min_workers", resp.WorkerGroup.MinWorkers)
	d.Set("max_workers", resp.WorkerGroup.MaxWorkers)
	d.Set("status", resp.WorkerGroup.Status)
	d.Set("created_at", resp.WorkerGroup.CreatedAt)

	return nil
}

func DataSourceWorkerGroups() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceWorkerGroupsRead,
		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"worker_groups": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id":           {Type: schema.TypeString, Computed: true},
						"name":         {Type: schema.TypeString, Computed: true},
						"flavor":       {Type: schema.TypeString, Computed: true},
						"num_workers":  {Type: schema.TypeInt, Computed: true},
						"auto_scaling": {Type: schema.TypeBool, Computed: true},
						"min_workers":  {Type: schema.TypeInt, Computed: true},
						"max_workers":  {Type: schema.TypeInt, Computed: true},
						"status":       {Type: schema.TypeString, Computed: true},
						"created_at":   {Type: schema.TypeString, Computed: true},
					},
				},
			},
		},
	}
}

func dataSourceWorkerGroupsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	clusterID := d.Get("cluster_id").(string)

	listResp := &dto.ListWorkerGroupsResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.WorkerGroups(cfg.ProjectID, clusterID), listResp, nil)
	if err != nil {
		return diag.Errorf("Error listing vnpaycloud_kubernetes_worker_groups: %s", err)
	}

	var workerGroups []map[string]interface{}
	for _, wg := range listResp.WorkerGroups {
		workerGroups = append(workerGroups, map[string]interface{}{
			"id":           wg.ID,
			"name":         wg.Name,
			"flavor":       wg.Flavor,
			"num_workers":  wg.NumWorkers,
			"auto_scaling": wg.AutoScaling,
			"min_workers":  wg.MinWorkers,
			"max_workers":  wg.MaxWorkers,
			"status":       wg.Status,
			"created_at":   wg.CreatedAt,
		})
	}

	d.SetId(fmt.Sprintf("worker-groups-%s-%s", cfg.ProjectID, clusterID))
	d.Set("worker_groups", workerGroups)

	return nil
}
