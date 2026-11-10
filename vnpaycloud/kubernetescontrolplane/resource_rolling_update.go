package kubernetescontrolplane

import (
	"context"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/kubernetescluster"
	"terraform-provider-vnpaycloud/vnpaycloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceKubernetesControlPlaneRollingUpdate() *schema.Resource {
	s := map[string]*schema.Schema{
		"cluster_id": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"triggers": {
			Type:     schema.TypeMap,
			Optional: true,
			ForceNew: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
		"performed": {
			Type:     schema.TypeBool,
			Computed: true,
		},
	}
	for k, v := range baselineComputedSchema() {
		s[k] = v
	}

	return &schema.Resource{
		CreateContext: resourceRollingUpdateCreate,
		ReadContext:   resourceRollingUpdateRead,
		DeleteContext: resourceRollingUpdateDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(90 * time.Minute),
		},
		Schema: s,
	}
}

func resourceRollingUpdateCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	clusterID := d.Get("cluster_id").(string)
	timeout := d.Timeout(schema.TimeoutCreate)

	baseline, err := getBaseline(ctx, cfg, clusterID)
	if err != nil {
		return diag.Errorf("Error retrieving control plane baseline of cluster %s: %s", clusterID, err)
	}
	if !baseline.Supported {
		return diag.Errorf("Control plane rolling update is not supported for cluster %s: %s", clusterID, baseline.UnsupportedReason)
	}

	if baseline.ControlPlaneRolling {
		tflog.Info(ctx, "control plane of cluster "+clusterID+" is already rolling, waiting for it to settle")
		if err := kubernetescluster.WaitClusterActive(ctx, cfg.Client, cfg.ProjectID, clusterID, timeout); err != nil {
			return diag.Errorf("Error waiting for cluster %s control plane to settle: %s", clusterID, err)
		}
		if baseline, err = getBaseline(ctx, cfg, clusterID); err != nil {
			return diag.Errorf("Error retrieving control plane baseline of cluster %s: %s", clusterID, err)
		}
	}

	performed := false
	if baseline.RollingUpdateNeeded {
		tflog.Debug(ctx, "vnpaycloud_kubernetes_control_plane_rolling_update trigger", map[string]interface{}{
			"cluster_id":     clusterID,
			"outdated_items": baseline.OutdatedItems,
		})
		if _, err := cfg.Client.Post(ctx, client.ApiPath.ClusterControlPlaneRollingUpdate(cfg.ProjectID, clusterID), struct{}{}, &dto.K8sClusterResponse{}, nil); err != nil {
			return diag.Errorf("Error triggering control plane rolling update on cluster %s: %s", clusterID, err)
		}
		if err := kubernetescluster.WaitClusterActive(ctx, cfg.Client, cfg.ProjectID, clusterID, timeout); err != nil {
			return diag.Errorf("Error waiting for cluster %s control plane rolling update: %s", clusterID, err)
		}
		if baseline, err = getBaseline(ctx, cfg, clusterID); err != nil {
			return diag.Errorf("Error retrieving control plane baseline of cluster %s: %s", clusterID, err)
		}
		if baseline.RollingUpdateNeeded || baseline.ControlPlaneRolling {
			return diag.Errorf("Control plane of cluster %s did not reach the baseline after rolling update (outdated: %v)", clusterID, baseline.OutdatedItems)
		}
		performed = true
	}

	d.SetId(clusterID)
	d.Set("performed", performed)
	setBaselineData(d, baseline)
	return nil
}

func resourceRollingUpdateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	baseline, err := getBaseline(ctx, cfg, d.Id())
	if err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error retrieving vnpaycloud_kubernetes_control_plane_rolling_update"))
	}

	d.Set("cluster_id", d.Id())
	setBaselineData(d, baseline)
	return nil
}

// A rolling update cannot be undone; destroying only forgets the resource.
func resourceRollingUpdateDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	d.SetId("")
	return nil
}
