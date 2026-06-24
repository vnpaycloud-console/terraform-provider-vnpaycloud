package networkacl

import (
	"context"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceNetworkACL() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNetworkACLCreate,
		ReadContext:   resourceNetworkACLRead,
		UpdateContext: resourceNetworkACLUpdate,
		DeleteContext: resourceNetworkACLDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"subnet_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"total_rules": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceNetworkACLCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	createOpts := dto.CreateNetworkACLRequest{
		Name:        d.Get("name").(string),
		VpcID:       d.Get("vpc_id").(string),
		Description: d.Get("description").(string),
	}

	tflog.Debug(ctx, "vnpaycloud_network_acl create options", map[string]interface{}{"create_opts": createOpts})

	createResp := &dto.NetworkACLResponse{}
	_, err := cfg.Client.Post(ctx, client.ApiPath.NetworkACLs(cfg.ProjectID), createOpts, createResp, nil)
	if err != nil {
		return diag.Errorf("Error creating vnpaycloud_network_acl: %s", err)
	}

	d.SetId(createResp.NetworkACL.ID)

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"initiating", "creating", "unknown"},
		Target:     []string{"active", "created"},
		Refresh:    networkACLStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, createResp.NetworkACL.ID),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	if _, err = stateConf.WaitForStateContext(ctx); err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_network_acl %s to become ready: %s", createResp.NetworkACL.ID, err)
	}

	for _, subnetID := range stringSetValues(d.Get("subnet_ids").(*schema.Set)) {
		if err := mapNetworkACLSubnet(ctx, cfg, d.Id(), subnetID); err != nil {
			return diag.Errorf("Error mapping subnet %s to vnpaycloud_network_acl %s: %s", subnetID, d.Id(), err)
		}
	}

	return resourceNetworkACLRead(ctx, d, meta)
}

func resourceNetworkACLRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	resp := &dto.NetworkACLResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.NetworkACLWithID(cfg.ProjectID, d.Id()), resp, nil)
	if err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error retrieving vnpaycloud_network_acl"))
	}

	tflog.Debug(ctx, "Retrieved vnpaycloud_network_acl "+d.Id(), map[string]interface{}{"network_acl": resp.NetworkACL})
	setNetworkACLAttributes(d, resp.NetworkACL)

	return nil
}

func resourceNetworkACLUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if d.HasChange("subnet_ids") {
		oldRaw, newRaw := d.GetChange("subnet_ids")
		oldSet := oldRaw.(*schema.Set)
		newSet := newRaw.(*schema.Set)

		for _, raw := range newSet.Difference(oldSet).List() {
			subnetID := raw.(string)
			if err := mapNetworkACLSubnet(ctx, cfg, d.Id(), subnetID); err != nil {
				return diag.Errorf("Error mapping subnet %s to vnpaycloud_network_acl %s: %s", subnetID, d.Id(), err)
			}
		}

		for _, raw := range oldSet.Difference(newSet).List() {
			subnetID := raw.(string)
			if err := unmapNetworkACLSubnet(ctx, cfg, subnetID); err != nil {
				return diag.Errorf("Error unmapping subnet %s from vnpaycloud_network_acl %s: %s", subnetID, d.Id(), err)
			}
		}
	}

	return resourceNetworkACLRead(ctx, d, meta)
}

func resourceNetworkACLDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	resp := &dto.NetworkACLResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.NetworkACLWithID(cfg.ProjectID, d.Id()), resp, nil)
	if err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error retrieving vnpaycloud_network_acl"))
	}

	for _, subnetID := range resp.NetworkACL.SubnetIDs {
		if err := unmapNetworkACLSubnet(ctx, cfg, subnetID); err != nil {
			return diag.Errorf("Error unmapping subnet %s from vnpaycloud_network_acl %s before deletion: %s", subnetID, d.Id(), err)
		}
	}

	if resp.NetworkACL.Status != "deleting" {
		if _, err := cfg.Client.Delete(ctx, client.ApiPath.NetworkACLWithID(cfg.ProjectID, d.Id()), nil); err != nil {
			return diag.FromErr(util.CheckDeleted(d, err, "Error deleting vnpaycloud_network_acl"))
		}
	}

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"deleting", "active", "created"},
		Target:     []string{"deleted"},
		Refresh:    networkACLStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	if _, err = stateConf.WaitForStateContext(ctx); err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_network_acl %s to delete: %s", d.Id(), err)
	}

	return nil
}

func mapNetworkACLSubnet(ctx context.Context, cfg *config.Config, id, subnetID string) error {
	_, err := cfg.Client.Put(ctx, client.ApiPath.NetworkACLSubnet(cfg.ProjectID, id, subnetID), nil, nil, nil)
	return err
}

func unmapNetworkACLSubnet(ctx context.Context, cfg *config.Config, subnetID string) error {
	_, err := cfg.Client.Delete(ctx, client.ApiPath.NetworkACLSubnetRemove(cfg.ProjectID, subnetID), nil)
	return err
}
