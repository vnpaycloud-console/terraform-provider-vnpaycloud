package kubernetescluster

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

func ResourceKubernetesCluster() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterCreate,
		ReadContext:   resourceClusterRead,
		DeleteContext: resourceClusterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			// Cluster Information
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"k8s_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"purpose": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"private_gw_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			// Network Information
			"subnet_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"cni_plugin": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"calico", "cilium"}, false),
			},
			"pod_cidr": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"service_cidr": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			// Master Information
			"cluster_size": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"small", "medium", "large", "extra_large"}, false),
			},

			// Initial Worker Group (required for cluster creation)
			"default_worker_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"default_worker_flavor": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"default_worker_count": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				Default:      1,
				ValidateFunc: validation.IntAtLeast(1),
			},
			"default_worker_volume_type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"default_worker_volume_size": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
			"default_worker_ssh_key_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			// Computed attributes
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
			"kubeconfig": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func resourceClusterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	createOpts := dto.CreateK8sClusterRequest{
		ClusterInformation: dto.K8sClusterInformation{
			Name: d.Get("name").(string),
		},
		NetworkInformation: dto.K8sNetworkInformation{
			SubnetID: d.Get("subnet_id").(string),
		},
		WorkerGroupInformation: dto.K8sWorkerGroupInformation{
			Flavor: d.Get("default_worker_flavor").(string),
		},
	}

	if v, ok := d.GetOk("k8s_version"); ok {
		createOpts.ClusterInformation.K8sVersion = v.(string)
	}
	if v, ok := d.GetOk("purpose"); ok {
		createOpts.ClusterInformation.Purpose = v.(string)
	}
	if v, ok := d.GetOk("private_gw_id"); ok {
		createOpts.ClusterInformation.PrivateGwID = v.(string)
	}
	if v, ok := d.GetOk("cni_plugin"); ok {
		createOpts.NetworkInformation.CniPlugin = v.(string)
	}
	if v, ok := d.GetOk("pod_cidr"); ok {
		createOpts.NetworkInformation.PodCidr = v.(string)
	}
	if v, ok := d.GetOk("service_cidr"); ok {
		createOpts.NetworkInformation.ServiceCidr = v.(string)
	}
	if v, ok := d.GetOk("cluster_size"); ok {
		createOpts.MasterInformation.ClusterSize = v.(string)
	}
	if v, ok := d.GetOk("default_worker_name"); ok {
		createOpts.WorkerGroupInformation.Name = v.(string)
	}
	if v, ok := d.GetOk("default_worker_count"); ok {
		createOpts.WorkerGroupInformation.NumWorkers = v.(int)
	}
	if v, ok := d.GetOk("default_worker_volume_type"); ok {
		createOpts.WorkerGroupInformation.VolumeType = v.(string)
	}
	if v, ok := d.GetOk("default_worker_volume_size"); ok {
		createOpts.WorkerGroupInformation.VolumeSize = v.(int)
	}
	if v, ok := d.GetOk("default_worker_ssh_key_id"); ok {
		createOpts.WorkerGroupInformation.SshKeyID = v.(string)
	}

	tflog.Debug(ctx, "vnpaycloud_kubernetes_cluster create options", map[string]interface{}{"create_opts": createOpts})

	createResp := &dto.K8sClusterResponse{}
	_, err := cfg.Client.Post(ctx, client.ApiPath.Clusters(cfg.ProjectID), createOpts, createResp, nil)
	if err != nil {
		return diag.Errorf("Error creating vnpaycloud_kubernetes_cluster: %s", err)
	}

	d.SetId(createResp.Cluster.ID)

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"creating", "initiating", "pending_create", "unknown"},
		Target:     []string{"active"},
		Refresh:    clusterStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, createResp.Cluster.ID),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      30 * time.Second,
		MinTimeout: 15 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_kubernetes_cluster %s to become ready: %s", createResp.Cluster.ID, err)
	}

	return resourceClusterRead(ctx, d, meta)
}

func resourceClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	resp := &dto.K8sClusterResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.ClusterWithID(cfg.ProjectID, d.Id()), resp, nil)
	if err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error retrieving vnpaycloud_kubernetes_cluster"))
	}

	tflog.Debug(ctx, "Retrieved vnpaycloud_kubernetes_cluster "+d.Id(), map[string]interface{}{"cluster": resp.Cluster})

	d.Set("name", resp.Cluster.Name)
	// k8s_version: biz layer resolves name→ID; don't read back to avoid drift
	d.Set("purpose", resp.Cluster.Purpose)
	d.Set("private_gw_id", resp.Cluster.PrivateGwID)
	d.Set("subnet_id", resp.Cluster.SubnetID)
	d.Set("cni_plugin", resp.Cluster.CniPlugin)
	d.Set("pod_cidr", resp.Cluster.PodCidr)
	d.Set("service_cidr", resp.Cluster.ServiceCidr)
	d.Set("cluster_size", resp.Cluster.ClusterSize)
	d.Set("zone", resp.Cluster.Zone)
	d.Set("api_endpoint", resp.Cluster.ApiEndpoint)
	d.Set("private_ip", resp.Cluster.PrivateIP)
	d.Set("status", resp.Cluster.Status)
	d.Set("created_at", resp.Cluster.CreatedAt)

	// Fetch kubeconfig if cluster is active.
	if resp.Cluster.Status == "active" {
		kcResp := &dto.KubeconfigResponse{}
		_, err := cfg.Client.Get(ctx, client.ApiPath.ClusterKubeconfig(cfg.ProjectID, d.Id()), kcResp, nil)
		if err != nil {
			tflog.Warn(ctx, "Failed to fetch kubeconfig for cluster "+d.Id(), map[string]interface{}{"error": err})
		} else {
			d.Set("kubeconfig", kcResp.Kubeconfig)
		}
	}

	return nil
}

func resourceClusterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if _, err := cfg.Client.Delete(ctx, client.ApiPath.ClusterWithID(cfg.ProjectID, d.Id()), nil); err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error deleting vnpaycloud_kubernetes_cluster"))
	}

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"deleting", "active", "unknown"},
		Target:     []string{"deleted"},
		Refresh:    clusterStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      15 * time.Second,
		MinTimeout: 10 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_kubernetes_cluster %s to delete: %s", d.Id(), err)
	}

	return nil
}
