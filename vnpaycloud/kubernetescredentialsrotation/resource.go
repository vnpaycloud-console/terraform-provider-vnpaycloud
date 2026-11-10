package kubernetescredentialsrotation

import (
	"context"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/kubernetescluster"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceKubernetesCredentialsRotation() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRotationCreate,
		ReadContext:   resourceRotationRead,
		DeleteContext: resourceRotationDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
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
			"rotated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceRotationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	clusterID := d.Get("cluster_id").(string)

	tflog.Debug(ctx, "vnpaycloud_kubernetes_credentials_rotation trigger", map[string]interface{}{"cluster_id": clusterID})

	if _, err := cfg.Client.Post(ctx, client.ApiPath.ClusterRotateCredentials(cfg.ProjectID, clusterID), struct{}{}, &dto.K8sClusterResponse{}, nil); err != nil {
		return diag.Errorf("Error rotating credentials of cluster %s: %s", clusterID, err)
	}

	if err := kubernetescluster.WaitClusterActive(ctx, cfg.Client, cfg.ProjectID, clusterID, d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.Errorf("Error waiting for cluster %s credential rotation: %s", clusterID, err)
	}

	rotatedAt := time.Now().UTC().Format(time.RFC3339)
	d.SetId(clusterID + ":" + rotatedAt)
	d.Set("rotated_at", rotatedAt)
	return nil
}

// A rotation is a fire-once action with no queryable backend object.
func resourceRotationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceRotationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	d.SetId("")
	return nil
}
