package vpc

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

func ResourceVpc() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVpcCreate,
		ReadContext:   resourceVpcRead,
		UpdateContext: resourceVpcUpdate,
		DeleteContext: resourceVpcDelete,
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
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"cidr": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enable_snat": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"snat_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"subnet_ids": {
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

func resourceVpcCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	createOpts := dto.CreateVPCRequest{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		CIDR:        d.Get("cidr").(string),
	}

	tflog.Debug(ctx, "vnpaycloud_vpc create options", map[string]interface{}{"create_opts": createOpts})

	createResp := &dto.VPCResponse{}
	_, err := cfg.Client.Post(ctx, client.ApiPath.VPCs(cfg.ProjectID), createOpts, createResp, nil)
	if err != nil {
		return diag.Errorf("Error creating vnpaycloud_vpc: %s", err)
	}

	d.SetId(createResp.VPC.ID)

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"initiating", "creating"},
		Target:     []string{"active", "created"},
		Refresh:    vpcStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, createResp.VPC.ID),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_vpc %s to become ready: %s", createResp.VPC.ID, err)
	}

	// Enable SNAT if requested
	if d.Get("enable_snat").(bool) {
		snatReq := dto.SetVPCRouterSNATRequest{EnableSnat: true}
		_, err := cfg.Client.Put(ctx, client.ApiPath.VPCSetSNAT(cfg.ProjectID, d.Id()), snatReq, nil, nil)
		if err != nil {
			return diag.Errorf("Error enabling SNAT for vnpaycloud_vpc %s: %s", d.Id(), err)
		}
	}

	return resourceVpcRead(ctx, d, meta)
}

func resourceVpcRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	vpcResp := &dto.VPCResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.VPCWithID(cfg.ProjectID, d.Id()), vpcResp, nil)
	if err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error retrieving vnpaycloud_vpc"))
	}

	tflog.Debug(ctx, "Retrieved vnpaycloud_vpc "+d.Id(), map[string]interface{}{"vpc": vpcResp.VPC})

	d.Set("name", vpcResp.VPC.Name)
	d.Set("description", vpcResp.VPC.Description)
	d.Set("cidr", vpcResp.VPC.CIDR)
	d.Set("status", vpcResp.VPC.Status)
	d.Set("enable_snat", vpcResp.VPC.EnableSnat)
	d.Set("snat_address", vpcResp.VPC.SnatAddress)
	d.Set("subnet_ids", vpcResp.VPC.SubnetIDs)
	d.Set("created_at", vpcResp.VPC.CreatedAt)

	return nil
}

func resourceVpcUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if d.HasChanges("name", "description") {
		updateOpts := dto.UpdateVPCRequest{
			Name:        d.Get("name").(string),
			Description: d.Get("description").(string),
		}

		tflog.Debug(ctx, "vnpaycloud_vpc update options", map[string]interface{}{"update_opts": updateOpts})

		_, err := cfg.Client.Put(ctx, client.ApiPath.VPCWithID(cfg.ProjectID, d.Id()), updateOpts, nil, nil)
		if err != nil {
			return diag.Errorf("Error updating vnpaycloud_vpc %s: %s", d.Id(), err)
		}
	}

	if d.HasChange("enable_snat") {
		snatReq := dto.SetVPCRouterSNATRequest{EnableSnat: d.Get("enable_snat").(bool)}
		_, err := cfg.Client.Put(ctx, client.ApiPath.VPCSetSNAT(cfg.ProjectID, d.Id()), snatReq, nil, nil)
		if err != nil {
			return diag.Errorf("Error setting SNAT for vnpaycloud_vpc %s: %s", d.Id(), err)
		}
	}

	return resourceVpcRead(ctx, d, meta)
}

func resourceVpcDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	vpcResp := &dto.VPCResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.VPCWithID(cfg.ProjectID, d.Id()), vpcResp, nil)
	if err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error retrieving vnpaycloud_vpc"))
	}

	if vpcResp.VPC.Status != "deleting" {
		if _, err := cfg.Client.Delete(ctx, client.ApiPath.VPCWithID(cfg.ProjectID, d.Id()), nil); err != nil {
			return diag.FromErr(util.CheckDeleted(d, err, "Error deleting vnpaycloud_vpc"))
		}
	}

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"deleting", "active", "created"},
		Target:     []string{"deleted"},
		Refresh:    vpcStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_vpc %s to Delete: %s", d.Id(), err)
	}

	return nil
}
