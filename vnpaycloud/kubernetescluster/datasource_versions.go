package kubernetescluster

import (
	"context"
	"fmt"

	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceKubernetesVersions() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceKubernetesVersionsRead,
		Schema: map[string]*schema.Schema{
			"versions": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"version":    {Type: schema.TypeString, Computed: true},
						"is_default": {Type: schema.TypeBool, Computed: true},
					},
				},
			},
			"default_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceKubernetesVersionsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	resp := &dto.ListK8sVersionsResponse{}
	if _, err := cfg.Client.Get(ctx, client.ApiPath.KubernetesVersions(cfg.ProjectID), resp, nil); err != nil {
		return diag.Errorf("Error listing vnpaycloud_kubernetes_versions: %s", err)
	}

	versions := make([]map[string]interface{}, 0, len(resp.Versions))
	defaultVersion := ""
	for _, v := range resp.Versions {
		versions = append(versions, map[string]interface{}{
			"version":    v.Version,
			"is_default": v.IsDefault,
		})
		if v.IsDefault {
			defaultVersion = v.Version
		}
	}

	d.SetId(fmt.Sprintf("kubernetes-versions-%s", cfg.ProjectID))
	d.Set("versions", versions)
	d.Set("default_version", defaultVersion)

	return nil
}
