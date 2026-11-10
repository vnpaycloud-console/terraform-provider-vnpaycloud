package kubernetesapiwhitelist

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

func ResourceKubernetesApiWhitelist() *schema.Resource {
	s := map[string]*schema.Schema{
		"cluster_id": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"entry": {
			Type:     schema.TypeSet,
			Required: true,
			MinItems: 1,
			MaxItems: 50,
			Elem:     entrySchema(),
		},
	}
	for k, v := range computedSchema() {
		s[k] = v
	}

	return &schema.Resource{
		CreateContext: resourceApiWhitelistUpsert,
		ReadContext:   resourceApiWhitelistRead,
		UpdateContext: resourceApiWhitelistUpsert,
		DeleteContext: resourceApiWhitelistDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				d.Set("cluster_id", d.Id())
				return []*schema.ResourceData{d}, nil
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
			Update: schema.DefaultTimeout(15 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: s,
	}
}

func resourceApiWhitelistUpsert(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	clusterID := d.Get("cluster_id").(string)

	timeout := d.Timeout(schema.TimeoutUpdate)
	if d.IsNewResource() {
		timeout = d.Timeout(schema.TimeoutCreate)
	}

	if err := applyWhitelist(ctx, cfg, clusterID, expandEntries(d.Get("entry")), timeout); err != nil {
		return diag.Errorf("Error applying vnpaycloud_kubernetes_api_whitelist for cluster %s: %s", clusterID, err)
	}

	d.SetId(clusterID)
	return resourceApiWhitelistRead(ctx, d, meta)
}

func resourceApiWhitelistRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	resp := &dto.ClusterApiWhitelistResponse{}
	if _, err := cfg.Client.Get(ctx, client.ApiPath.ClusterApiWhitelist(cfg.ProjectID, d.Id()), resp, nil); err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error retrieving vnpaycloud_kubernetes_api_whitelist"))
	}

	tflog.Debug(ctx, "Retrieved vnpaycloud_kubernetes_api_whitelist "+d.Id(), map[string]interface{}{"whitelist": resp})

	d.Set("cluster_id", d.Id())
	setWhitelistData(d, resp)
	return nil
}

func resourceApiWhitelistDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if err := applyWhitelist(ctx, cfg, d.Id(), []dto.ApiWhitelistEntry{}, d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error clearing vnpaycloud_kubernetes_api_whitelist"))
	}
	return nil
}

// applyWhitelist PUTs the desired list and waits for the cluster to leave
// "updating": the backend validates synchronously, then applies the CIDRs to
// the API load balancer in the background.
func applyWhitelist(ctx context.Context, cfg *config.Config, clusterID string, entries []dto.ApiWhitelistEntry, timeout time.Duration) error {
	body := dto.UpdateClusterApiWhitelistRequest{Entries: entries}
	tflog.Debug(ctx, "vnpaycloud_kubernetes_api_whitelist apply", map[string]interface{}{"cluster_id": clusterID, "entries": len(entries)})

	if _, err := cfg.Client.Put(ctx, client.ApiPath.ClusterApiWhitelist(cfg.ProjectID, clusterID), body, &dto.ClusterApiWhitelistResponse{}, nil); err != nil {
		return err
	}
	return kubernetescluster.WaitClusterActive(ctx, cfg.Client, cfg.ProjectID, clusterID, timeout)
}
