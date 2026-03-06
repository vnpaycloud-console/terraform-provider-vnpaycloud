package workergroup

import (
	"context"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceWorkerGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceWorkerGroupCreate,
		ReadContext:   resourceWorkerGroupRead,
		UpdateContext: resourceWorkerGroupUpdate,
		DeleteContext: resourceWorkerGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"flavor": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"num_workers": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntAtLeast(1),
			},
			"auto_scaling": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"min_workers": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(0),
			},
			"max_workers": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(0),
			},
			"volume_type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"volume_size": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
			"ssh_key_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			// Computed attributes
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

func resourceWorkerGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	clusterID := d.Get("cluster_id").(string)

	createOpts := dto.CreateWorkerGroupRequest{
		Name:       d.Get("name").(string),
		Flavor:     d.Get("flavor").(string),
		NumWorkers: d.Get("num_workers").(int),
	}

	if v, ok := d.GetOk("auto_scaling"); ok {
		createOpts.AutoScaling = v.(bool)
	}
	if v, ok := d.GetOk("min_workers"); ok {
		createOpts.MinWorkers = v.(int)
	}
	if v, ok := d.GetOk("max_workers"); ok {
		createOpts.MaxWorkers = v.(int)
	}
	if v, ok := d.GetOk("volume_type"); ok {
		createOpts.VolumeType = v.(string)
	}
	if v, ok := d.GetOk("volume_size"); ok {
		createOpts.VolumeSize = v.(int)
	}
	if v, ok := d.GetOk("ssh_key_id"); ok {
		createOpts.SshKeyID = v.(string)
	}
	if v, ok := d.GetOk("labels"); ok {
		labels := make(map[string]string)
		for k, val := range v.(map[string]interface{}) {
			labels[k] = val.(string)
		}
		createOpts.Labels = labels
	}

	tflog.Debug(ctx, "vnpaycloud_kubernetes_worker_group create options", map[string]interface{}{"create_opts": createOpts})

	createResp := &dto.WorkerGroupResponse{}
	_, err := cfg.Client.Post(ctx, client.ApiPath.WorkerGroups(cfg.ProjectID, clusterID), createOpts, createResp, nil)
	if err != nil {
		return diag.Errorf("Error creating vnpaycloud_kubernetes_worker_group: %s", err)
	}

	d.SetId(createResp.WorkerGroup.ID)

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"creating", "initiating", "pending_create", "unknown"},
		Target:     []string{"active"},
		Refresh:    workerGroupStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, clusterID, createResp.WorkerGroup.ID),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      30 * time.Second,
		MinTimeout: 15 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_kubernetes_worker_group %s to become ready: %s", createResp.WorkerGroup.ID, err)
	}

	return resourceWorkerGroupRead(ctx, d, meta)
}

func resourceWorkerGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	clusterID := d.Get("cluster_id").(string)

	resp := &dto.WorkerGroupResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.WorkerGroupWithID(cfg.ProjectID, clusterID, d.Id()), resp, nil)
	if err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error retrieving vnpaycloud_kubernetes_worker_group"))
	}

	tflog.Debug(ctx, "Retrieved vnpaycloud_kubernetes_worker_group "+d.Id(), map[string]interface{}{"worker_group": resp.WorkerGroup})

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

func resourceWorkerGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	clusterID := d.Get("cluster_id").(string)

	if d.HasChanges("num_workers", "auto_scaling", "min_workers", "max_workers") {
		updateOpts := dto.UpdateWorkerGroupRequest{
			NumWorkers:  d.Get("num_workers").(int),
			AutoScaling: d.Get("auto_scaling").(bool),
			MinWorkers:  d.Get("min_workers").(int),
			MaxWorkers:  d.Get("max_workers").(int),
		}

		tflog.Debug(ctx, "vnpaycloud_kubernetes_worker_group update options", map[string]interface{}{"update_opts": updateOpts})

		_, err := cfg.Client.Put(ctx, client.ApiPath.WorkerGroupWithID(cfg.ProjectID, clusterID, d.Id()), updateOpts, nil, nil)
		if err != nil {
			return diag.Errorf("Error updating vnpaycloud_kubernetes_worker_group %s: %s", d.Id(), err)
		}

		stateConf := &retry.StateChangeConf{
			Pending:    []string{"updating", "resizing", "unknown"},
			Target:     []string{"active"},
			Refresh:    workerGroupStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, clusterID, d.Id()),
			Timeout:    d.Timeout(schema.TimeoutUpdate),
			Delay:      15 * time.Second,
			MinTimeout: 10 * time.Second,
		}

		_, err = stateConf.WaitForStateContext(ctx)
		if err != nil {
			return diag.Errorf("Error waiting for vnpaycloud_kubernetes_worker_group %s to finish updating: %s", d.Id(), err)
		}
	}

	return resourceWorkerGroupRead(ctx, d, meta)
}

func resourceWorkerGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	clusterID := d.Get("cluster_id").(string)

	if _, err := cfg.Client.Delete(ctx, client.ApiPath.WorkerGroupWithID(cfg.ProjectID, clusterID, d.Id()), nil); err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error deleting vnpaycloud_kubernetes_worker_group"))
	}

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"deleting", "active", "unknown"},
		Target:     []string{"deleted"},
		Refresh:    workerGroupStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, clusterID, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      15 * time.Second,
		MinTimeout: 10 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_kubernetes_worker_group %s to delete: %s", d.Id(), err)
	}

	return nil
}
