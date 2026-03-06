package servergroup

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

func ResourceServerGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceServerGroupCreate,
		ReadContext:   resourceServerGroupRead,
		DeleteContext: resourceServerGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"policy": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"member_ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceServerGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	createOpts := dto.CreateServerGroupRequest{
		Name:   d.Get("name").(string),
		Policy: d.Get("policy").(string),
	}

	tflog.Debug(ctx, "vnpaycloud_server_group create options", map[string]interface{}{"create_opts": createOpts})

	createResp := &dto.ServerGroupResponse{}
	_, err := cfg.Client.Post(ctx, client.ApiPath.ServerGroups(cfg.ProjectID), createOpts, createResp, nil)
	if err != nil {
		return diag.Errorf("Error creating vnpaycloud_server_group: %s", err)
	}

	d.SetId(createResp.ServerGroup.ID)

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"initiating", "creating"},
		Target:     []string{"active", "created"},
		Refresh:    serverGroupStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, createResp.ServerGroup.ID),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_server_group %s to become ready: %s", createResp.ServerGroup.ID, err)
	}

	return resourceServerGroupRead(ctx, d, meta)
}

func resourceServerGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	sgResp := &dto.ServerGroupResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.ServerGroupWithID(cfg.ProjectID, d.Id()), sgResp, nil)
	if err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error retrieving vnpaycloud_server_group"))
	}

	tflog.Debug(ctx, "Retrieved vnpaycloud_server_group "+d.Id(), map[string]interface{}{"server_group": sgResp.ServerGroup})

	d.Set("name", sgResp.ServerGroup.Name)
	d.Set("policy", sgResp.ServerGroup.Policy)
	d.Set("member_ids", sgResp.ServerGroup.MemberIDs)
	d.Set("created_at", sgResp.ServerGroup.CreatedAt)

	return nil
}

func resourceServerGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	sgResp := &dto.ServerGroupResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.ServerGroupWithID(cfg.ProjectID, d.Id()), sgResp, nil)
	if err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error retrieving vnpaycloud_server_group"))
	}

	if sgResp.ServerGroup.ID != "" {
		if _, err := cfg.Client.Delete(ctx, client.ApiPath.ServerGroupWithID(cfg.ProjectID, d.Id()), nil); err != nil {
			return diag.FromErr(util.CheckDeleted(d, err, "Error deleting vnpaycloud_server_group"))
		}
	}

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"deleting", "active", "created"},
		Target:     []string{"deleted"},
		Refresh:    serverGroupStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_server_group %s to delete: %s", d.Id(), err)
	}

	return nil
}
