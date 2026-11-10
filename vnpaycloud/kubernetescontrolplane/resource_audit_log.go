package kubernetescontrolplane

import (
	"context"
	"net/http"
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

func ResourceKubernetesAuditLog() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAuditLogCreate,
		ReadContext:   resourceAuditLogRead,
		DeleteContext: resourceAuditLogDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				d.Set("cluster_id", d.Id())
				return []*schema.ResourceData{d}, nil
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(90 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func resourceAuditLogCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	clusterID := d.Get("cluster_id").(string)
	timeout := d.Timeout(schema.TimeoutCreate)

	baseline, err := getBaseline(ctx, cfg, clusterID)
	if err != nil {
		return diag.Errorf("Error retrieving control plane baseline of cluster %s: %s", clusterID, err)
	}
	if !baseline.Supported {
		return diag.Errorf("Audit log is not supported for cluster %s: %s", clusterID, baseline.UnsupportedReason)
	}

	if !baseline.AuditLogEnabled {
		if baseline.ControlPlaneRolling {
			return diag.Errorf("Control plane of cluster %s is being updated, retry once it is active", clusterID)
		}
		if !baseline.AuditLogReady {
			return diag.Errorf("Cluster %s control plane does not meet the baseline yet (outdated: %v); apply a vnpaycloud_kubernetes_control_plane_rolling_update for it first and add it to depends_on", clusterID, baseline.OutdatedItems)
		}

		tflog.Debug(ctx, "vnpaycloud_kubernetes_audit_log enable", map[string]interface{}{"cluster_id": clusterID})
		_, err := cfg.Client.Post(ctx, client.ApiPath.ClusterAuditLogEnable(cfg.ProjectID, clusterID), struct{}{}, &dto.K8sClusterResponse{}, nil)
		if err != nil && !util.ResponseCodeIs(err, http.StatusConflict) {
			return diag.Errorf("Error enabling audit log on cluster %s: %s", clusterID, err)
		}
		if err := kubernetescluster.WaitClusterActive(ctx, cfg.Client, cfg.ProjectID, clusterID, timeout); err != nil {
			return diag.Errorf("Error waiting for cluster %s to finish enabling audit log: %s", clusterID, err)
		}
		if baseline, err = getBaseline(ctx, cfg, clusterID); err != nil {
			return diag.Errorf("Error retrieving control plane baseline of cluster %s: %s", clusterID, err)
		}
		if !baseline.AuditLogEnabled {
			return diag.Errorf("Audit log of cluster %s is still disabled after the enable action finished", clusterID)
		}
	}

	d.SetId(clusterID)
	d.Set("enabled", true)
	return nil
}

func resourceAuditLogRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	baseline, err := getBaseline(ctx, cfg, d.Id())
	if err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error retrieving vnpaycloud_kubernetes_audit_log"))
	}
	if !baseline.AuditLogEnabled {
		tflog.Warn(ctx, "audit log of cluster "+d.Id()+" is no longer enabled, removing from state")
		d.SetId("")
		return nil
	}

	d.Set("cluster_id", d.Id())
	d.Set("enabled", true)
	return nil
}

// The platform offers no disable action; destroying only forgets the resource.
func resourceAuditLogDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	d.SetId("")
	return nil
}
