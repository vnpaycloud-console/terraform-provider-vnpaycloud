package maasaccesskey

import (
	"context"
	"net/url"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceMaasAccessKey() *schema.Resource {
	ds := &schema.Resource{
		ReadContext: dataSourceMaasAccessKeyRead,
		Schema:      accessKeyComputedSchema(),
	}

	ds.Schema["id"] = &schema.Schema{
		Type:         schema.TypeString,
		Optional:     true,
		Computed:     true,
		AtLeastOneOf: []string{"id", "name"},
	}
	ds.Schema["name"] = &schema.Schema{
		Type:         schema.TypeString,
		Optional:     true,
		Computed:     true,
		AtLeastOneOf: []string{"id", "name"},
	}

	return ds
}

func dataSourceMaasAccessKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if id, ok := d.GetOk("id"); ok && id.(string) != "" {
		resp := &dto.MaasAccessKeyResponse{}
		if _, err := cfg.Client.Get(ctx, client.ApiPath.MaasAccessKeyWithID(cfg.ProjectID, id.(string)), resp, nil); err != nil {
			return diag.Errorf("Error fetching vnpaycloud_maas_access_key %s: %s", id, err)
		}
		if v, ok := d.GetOk("name"); ok && resp.AccessKey.Name != v.(string) {
			return diag.Errorf("vnpaycloud_maas_access_key %q does not match name %q", id.(string), v.(string))
		}
		d.SetId(resp.AccessKey.ID)
		setAccessKeyData(d, &resp.AccessKey)
		return nil
	}

	name := d.Get("name").(string)
	listResp := &dto.ListMaasAccessKeysResponse{}
	path := client.ApiPath.MaasAccessKeys(cfg.ProjectID) + "?name=" + url.QueryEscape(name)
	if _, err := cfg.Client.Get(ctx, path, listResp, nil); err != nil {
		return diag.Errorf("Error listing vnpaycloud_maas_access_key: %s", err)
	}

	var matched []dto.MaasAccessKey
	for _, k := range listResp.AccessKeys {
		if k.Name == name {
			matched = append(matched, k)
		}
	}

	if len(matched) == 0 {
		return diag.Errorf("No vnpaycloud_maas_access_key found with name %q", name)
	}
	if len(matched) > 1 {
		return diag.Errorf("Found %d vnpaycloud_maas_access_key matching name %q; use id to select one", len(matched), name)
	}

	d.SetId(matched[0].ID)
	setAccessKeyData(d, &matched[0])
	return nil
}

func DataSourceMaasAccessKeys() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceMaasAccessKeysRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"access_keys": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: accessKeyComputedSchema(),
				},
			},
		},
	}
}

func dataSourceMaasAccessKeysRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	path := client.ApiPath.MaasAccessKeys(cfg.ProjectID)
	if v, ok := d.GetOk("name"); ok {
		path += "?name=" + url.QueryEscape(v.(string))
	}

	listResp := &dto.ListMaasAccessKeysResponse{}
	if _, err := cfg.Client.Get(ctx, path, listResp, nil); err != nil {
		return diag.Errorf("Unable to query vnpaycloud_maas_access_keys: %s", err)
	}

	keys := make([]map[string]interface{}, 0, len(listResp.AccessKeys))
	for _, k := range listResp.AccessKeys {
		keys = append(keys, map[string]interface{}{
			"id":                   k.ID,
			"name":                 k.Name,
			"description":          k.Description,
			"log_permission":       k.LogPermission,
			"metric_permission":    k.MetricPermission,
			"log_pipeline_id":      k.LogPipelineID,
			"metric_pipeline_id":   k.MetricPipelineID,
			"log_label_matcher":    flattenLabelMatchers(k.LogLabelMatchers),
			"metric_label_matcher": flattenLabelMatchers(k.MetricLabelMatchers),
			"endpoints":            flattenEndpoints(k.Endpoints),
			"api_key":              k.APIKey,
			"status":               k.Status,
			"created_at":           k.CreatedAt,
		})
	}

	d.SetId(cfg.ProjectID)
	d.Set("access_keys", keys)

	return nil
}
