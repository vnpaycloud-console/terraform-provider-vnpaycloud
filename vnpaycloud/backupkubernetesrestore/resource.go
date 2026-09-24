package backupkubernetesrestore

import (
	"context"
	"fmt"
	"regexp"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var idRegex = regexp.MustCompile(`^[a-zA-Z0-9-_.]*$`)

var restoreTypes = []string{"all_resources", "cluster_scoped_resources", "specific_namespaces"}

func ResourceBackupKubernetesRestore() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRestoreCreate,
		ReadContext:   resourceRestoreRead,
		DeleteContext: resourceRestoreDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
		},
		CustomizeDiff: func(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
			if d.Get("transform_lb_to_cluster_ip").(bool) && d.Get("keep_node_port_numbers").(bool) {
				return fmt.Errorf("keep_node_port_numbers cannot be true when transform_lb_to_cluster_ip is true")
			}
			return nil
		},
		Schema: map[string]*schema.Schema{
			"restore_point_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(3, 100),
					validation.StringMatch(idRegex, "must contain only letters, digits, hyphens, underscores and dots"),
				),
			},
			"dest_cluster_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(3, 100),
					validation.StringMatch(idRegex, "must contain only letters, digits, hyphens, underscores and dots"),
				),
			},
			"restore_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "all_resources",
				ValidateFunc: validation.StringInSlice(restoreTypes, false),
			},
			"transform_lb_to_cluster_ip": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Default:  false,
			},
			"keep_node_port_numbers": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Default:  false,
			},
			"restore_persistent_volumes": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Default:  true,
			},
			"include_cluster_scoped_resource": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Default:  true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"included_namespaces": {Type: schema.TypeList, Optional: true, ForceNew: true, Elem: &schema.Schema{Type: schema.TypeString}},
			"excluded_namespaces": {Type: schema.TypeList, Optional: true, ForceNew: true, Elem: &schema.Schema{Type: schema.TypeString}},
			"included_resources":  {Type: schema.TypeList, Optional: true, ForceNew: true, Elem: &schema.Schema{Type: schema.TypeString}},
			"excluded_resources":  {Type: schema.TypeList, Optional: true, ForceNew: true, Elem: &schema.Schema{Type: schema.TypeString}},
		},
	}
}

func resourceRestoreCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	opts := dto.CreateBackupKubernetesRestoreRequest{
		DestClusterID:                d.Get("dest_cluster_id").(string),
		RestorePointID:               d.Get("restore_point_id").(string),
		RestoreType:                  d.Get("restore_type").(string),
		TransformLbToClusterIP:       d.Get("transform_lb_to_cluster_ip").(bool),
		KeepNodePortNumbers:          d.Get("keep_node_port_numbers").(bool),
		RestorePersistentVolumes:     d.Get("restore_persistent_volumes").(bool),
		IncludeClusterScopedResource: d.Get("include_cluster_scoped_resource").(bool),
		IncludedNamespaces:           expandStringList(d.Get("included_namespaces")),
		ExcludedNamespaces:           expandStringList(d.Get("excluded_namespaces")),
		IncludedResources:            expandStringList(d.Get("included_resources")),
		ExcludedResources:            expandStringList(d.Get("excluded_resources")),
	}

	tflog.Debug(ctx, "vnpaycloud_backup_kubernetes_restore trigger", map[string]interface{}{
		"dest_cluster_id":  opts.DestClusterID,
		"restore_point_id": opts.RestorePointID,
	})

	resp := &dto.BackupKubernetesRestoreResponse{}
	if _, err := cfg.Client.Post(ctx, client.ApiPath.BackupKubernetesRestores(cfg.ProjectID), opts, resp, nil); err != nil {
		return diag.Errorf("Error triggering vnpaycloud_backup_kubernetes_restore: %s", err)
	}

	d.SetId(opts.DestClusterID + ":" + opts.RestorePointID)

	// Wait for the restore saga to reach a terminal state on the destination cluster.
	stateConf := &retry.StateChangeConf{
		Pending:    []string{"restoring", "none", "unknown"},
		Target:     []string{"success"},
		Refresh:    clusterRestoreStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, opts.DestClusterID),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      30 * time.Second,
		MinTimeout: 15 * time.Second,
	}
	last, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_backup_kubernetes_restore %s to complete: %s", d.Id(), err)
	}

	if cluster, ok := last.(*dto.K8sClusterResponse); ok && cluster != nil {
		d.Set("status", cluster.Cluster.RestoreStatus)
	}

	return nil
}

// resourceRestoreRead is a no-op: a restore is a fire-once action with no
// queryable backend object, so there is nothing to reconcile.
func resourceRestoreRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

// resourceRestoreDelete just drops the resource from state; a completed restore
// cannot be undone.
func resourceRestoreDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	d.SetId("")
	return nil
}

func expandStringList(v interface{}) []string {
	l, ok := v.([]interface{})
	if !ok {
		return nil
	}
	result := make([]string, 0, len(l))
	for _, item := range l {
		if s, ok := item.(string); ok {
			result = append(result, s)
		}
	}
	return result
}
