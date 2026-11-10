package kubernetescontrolplane

import (
	"context"
	"terraform-provider-vnpaycloud/vnpaycloud/config"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceKubernetesControlPlaneBaseline() *schema.Resource {
	s := map[string]*schema.Schema{
		"cluster_id": {
			Type:     schema.TypeString,
			Required: true,
		},
	}
	for k, v := range baselineComputedSchema() {
		s[k] = v
	}
	return &schema.Resource{
		ReadContext: dataSourceBaselineRead,
		Schema:      s,
	}
}

func dataSourceBaselineRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	clusterID := d.Get("cluster_id").(string)

	baseline, err := getBaseline(ctx, cfg, clusterID)
	if err != nil {
		return diag.Errorf("Error retrieving vnpaycloud_kubernetes_control_plane_baseline for cluster %s: %s", clusterID, err)
	}

	d.SetId(clusterID)
	setBaselineData(d, baseline)
	return nil
}
