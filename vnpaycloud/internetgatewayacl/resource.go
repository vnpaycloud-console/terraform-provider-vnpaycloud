package internetgatewayacl

import (
	"context"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceInternetGatewayACL() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceInternetGatewayACLCreate,
		ReadContext:   resourceInternetGatewayACLRead,
		UpdateContext: resourceInternetGatewayACLUpdate,
		DeleteContext: resourceInternetGatewayACLDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"internet_gateway_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"match": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateBlacklistCIDR,
				},
			},
			"zone_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceInternetGatewayACLCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	igwID := d.Get("internet_gateway_id").(string)

	// A PUT replaces the whole list, so refuse to silently overwrite a blacklist set elsewhere.
	currentResp := &dto.InternetGatewayBlacklistResponse{}
	if _, err := cfg.Client.Get(ctx, client.ApiPath.InternetGatewayBlacklist(cfg.ProjectID, igwID), currentResp, nil); err != nil {
		return diag.Errorf("Error reading current ACL of internet gateway %s: %s", igwID, err)
	}
	if len(currentResp.Blacklist.Match) > 0 {
		return diag.Errorf("Internet gateway %s already has an ACL with %d CIDR block(s). "+
			"Import it with `terraform import vnpaycloud_internet_gateway_acl.<name> %s` or clear it before applying.",
			igwID, len(currentResp.Blacklist.Match), igwID)
	}

	updateOpts := dto.UpdateInternetGatewayBlacklistRequest{
		Match: expandMatch(d.Get("match")),
	}

	tflog.Debug(ctx, "vnpaycloud_internet_gateway_acl create options", map[string]interface{}{"update_opts": updateOpts})

	updateResp := &dto.InternetGatewayBlacklistResponse{}
	_, err := cfg.Client.Put(ctx, client.ApiPath.InternetGatewayBlacklist(cfg.ProjectID, igwID), updateOpts, updateResp, nil)
	if err != nil {
		return diag.Errorf("Error creating vnpaycloud_internet_gateway_acl for internet gateway %s: %s", igwID, err)
	}

	d.SetId(igwID)

	return resourceInternetGatewayACLRead(ctx, d, meta)
}

func resourceInternetGatewayACLRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	aclResp := &dto.InternetGatewayBlacklistResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.InternetGatewayBlacklist(cfg.ProjectID, d.Id()), aclResp, nil)
	if err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error retrieving vnpaycloud_internet_gateway_acl"))
	}

	tflog.Debug(ctx, "Retrieved vnpaycloud_internet_gateway_acl "+d.Id(), map[string]interface{}{"internet_gateway_acl": aclResp.Blacklist})

	// An internet gateway with no blacklist entry left is the same as the resource being gone.
	if len(aclResp.Blacklist.Match) == 0 {
		d.SetId("")
		return nil
	}

	d.Set("internet_gateway_id", d.Id())
	d.Set("match", aclResp.Blacklist.Match)
	d.Set("zone_id", aclResp.Blacklist.ZoneID)

	return nil
}

func resourceInternetGatewayACLUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if d.HasChange("match") {
		updateOpts := dto.UpdateInternetGatewayBlacklistRequest{
			Match: expandMatch(d.Get("match")),
		}

		tflog.Debug(ctx, "vnpaycloud_internet_gateway_acl update options", map[string]interface{}{"update_opts": updateOpts})

		_, err := cfg.Client.Put(ctx, client.ApiPath.InternetGatewayBlacklist(cfg.ProjectID, d.Id()), updateOpts, nil, nil)
		if err != nil {
			return diag.Errorf("Error updating vnpaycloud_internet_gateway_acl %s: %s", d.Id(), err)
		}
	}

	return resourceInternetGatewayACLRead(ctx, d, meta)
}

func resourceInternetGatewayACLDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	updateOpts := dto.UpdateInternetGatewayBlacklistRequest{Match: []string{}}

	_, err := cfg.Client.Put(ctx, client.ApiPath.InternetGatewayBlacklist(cfg.ProjectID, d.Id()), updateOpts, nil, nil)
	if err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error deleting vnpaycloud_internet_gateway_acl"))
	}

	return nil
}

func expandMatch(raw interface{}) []string {
	set, ok := raw.(*schema.Set)
	if !ok {
		return []string{}
	}

	match := make([]string, 0, set.Len())
	for _, v := range set.List() {
		match = append(match, v.(string))
	}
	return match
}
