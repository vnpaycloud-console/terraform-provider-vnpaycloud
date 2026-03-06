package routetable

import (
	"context"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceRouteTable() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRouteTableCreate,
		ReadContext:   resourceRouteTableRead,
		DeleteContext: resourceRouteTableDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"dest_cidr": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"target_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"target_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"target_name": {
				Type:     schema.TypeString,
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

func resourceRouteTableCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	createOpts := dto.CreateRouteTableRequest{
		VpcID:      d.Get("vpc_id").(string),
		DestCIDR:   d.Get("dest_cidr").(string),
		TargetID:   d.Get("target_id").(string),
		TargetType: d.Get("target_type").(string),
	}

	tflog.Debug(ctx, "vnpaycloud_route_table create options", map[string]interface{}{"create_opts": createOpts})

	createResp := &dto.RouteTableResponse{}
	_, err := cfg.Client.Post(ctx, client.ApiPath.RouteTables(cfg.ProjectID), createOpts, createResp, nil)
	if err != nil {
		return diag.Errorf("Error creating vnpaycloud_route_table: %s", err)
	}

	d.SetId(createResp.RouteTable.ID)

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"initiating", "creating"},
		Target:     []string{"active", "created"},
		Refresh:    routeTableStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, createResp.RouteTable.ID),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_route_table %s to become ready: %s", createResp.RouteTable.ID, err)
	}

	return resourceRouteTableRead(ctx, d, meta)
}

func resourceRouteTableRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	rt, err := findRouteTableByID(ctx, cfg.Client, cfg.ProjectID, d.Id())
	if err != nil {
		return diag.Errorf("Error reading vnpaycloud_route_table %s: %s", d.Id(), err)
	}

	if rt == nil {
		tflog.Info(ctx, "Route table not found, removing from state", map[string]interface{}{"id": d.Id()})
		d.SetId("")
		return nil
	}

	tflog.Debug(ctx, "Retrieved vnpaycloud_route_table "+d.Id(), map[string]interface{}{"route_table": rt})

	d.Set("vpc_id", rt.VpcID)
	d.Set("dest_cidr", rt.DestCIDR)
	d.Set("target_id", rt.TargetID)
	d.Set("target_type", rt.TargetType)
	d.Set("name", rt.Name)
	d.Set("target_name", rt.TargetName)
	d.Set("status", rt.Status)
	d.Set("created_at", rt.CreatedAt)

	return nil
}

func resourceRouteTableDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	tflog.Debug(ctx, "Deleting vnpaycloud_route_table", map[string]interface{}{"id": d.Id()})

	if _, err := cfg.Client.Delete(ctx, client.ApiPath.RouteTableWithID(cfg.ProjectID, d.Id()), nil); err != nil {
		return diag.Errorf("Error deleting vnpaycloud_route_table %s: %s", d.Id(), err)
	}

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"deleting", "active", "created"},
		Target:     []string{"deleted"},
		Refresh:    routeTableStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_route_table %s to delete: %s", d.Id(), err)
	}

	return nil
}
