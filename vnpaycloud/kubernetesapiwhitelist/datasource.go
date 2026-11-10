package kubernetesapiwhitelist

import (
	"context"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceKubernetesApiWhitelist() *schema.Resource {
	s := map[string]*schema.Schema{
		"cluster_id": {
			Type:     schema.TypeString,
			Required: true,
		},
		"entry": {
			Type:     schema.TypeList,
			Computed: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"cidr": {Type: schema.TypeString, Computed: true},
					"note": {Type: schema.TypeString, Computed: true},
				},
			},
		},
	}
	for k, v := range computedSchema() {
		s[k] = v
	}

	return &schema.Resource{
		ReadContext: dataSourceApiWhitelistRead,
		Schema:      s,
	}
}

func dataSourceApiWhitelistRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	clusterID := d.Get("cluster_id").(string)

	resp := &dto.ClusterApiWhitelistResponse{}
	if _, err := cfg.Client.Get(ctx, client.ApiPath.ClusterApiWhitelist(cfg.ProjectID, clusterID), resp, nil); err != nil {
		return diag.Errorf("Error retrieving vnpaycloud_kubernetes_api_whitelist for cluster %s: %s", clusterID, err)
	}

	d.SetId(clusterID)
	setWhitelistData(d, resp)
	return nil
}
