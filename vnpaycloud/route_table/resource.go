package routetable

import (
	"context"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/shared"
	"terraform-provider-vnpaycloud/vnpaycloud/util"
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
				Computed: true,
			},
			"cidr_block": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"peering_connection_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceRouteTableCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	c, err := client.NewClient(ctx, config.ConsoleClientConfig)

	if err != nil {
		return diag.Errorf("Error creating VNPAY Cloud Route Table client: %s", err)
	}

	portId, err := shared.PeeringConnectionId2PortId(ctx, d, meta, d.Get("peering_connection_id").(string))

	if err != nil {
		return diag.Errorf("Error retrieving Peer VPC ID: %s", err)
	}

	createOpts := CreateRouteTableRequest{
		RouteTable: CreateRouteTableOpts{
			Cidr:     d.Get("cidr_block").(string),
			TargetId: portId,
			VpcId:    d.Get("vpc_id").(string),
		},
	}

	tflog.Debug(ctx, "vnpaycloud_route_table request options", map[string]interface{}{"create_opts": createOpts})

	createResp := &CreateRouteTableResponse{}
	_, err = c.Post(ctx, client.ApiPath.RouteTable, createOpts, createResp, nil)

	if err != nil {
		return diag.Errorf("Error creating vnpaycloud_route_table request: %s", err)
	}

	routeTable := createResp.RouteTable

	d.SetId(routeTable.ID)

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"OS_INITIATING"},
		Target:     []string{"OS_ACTIVE", "OS_CREATED"},
		Refresh:    routeTableStateRefreshFunc(ctx, c, routeTable.ID),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)

	if err != nil {
		return diag.Errorf(
			"Error waiting for vnpaycloud_route_table %s to become ready: %s", routeTable.ID, err)
	}

	return resourceRouteTableRead(ctx, d, meta)
}

func resourceRouteTableRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	c, err := client.NewClient(ctx, config.ConsoleClientConfig)

	if err != nil {
		return diag.Errorf("Error creating VNPAY Cloud Route Table client: %s", err)
	}

	getRouteTableResp := &GetRouteTableResponse{}
	_, err = c.Get(ctx, client.ApiPath.RouteTableWithId(d.Id()), getRouteTableResp, nil)

	if err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error retrieving route table"))
	}

	routeTable := getRouteTableResp.RouteTable
	tflog.Debug(ctx, "Retrieved route table "+d.Id(), map[string]interface{}{"request": routeTable})

	d.Set("name", routeTable.Name)
	d.Set("cidr_block", routeTable.DestCidr)
	d.Set("vpc_id", routeTable.VpcId)

	return nil
}

func resourceRouteTableDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	c, err := client.NewClient(ctx, config.ConsoleClientConfig)

	if err != nil {
		return diag.Errorf("Error creating VNPAY Cloud Route Table client: %s", err)
	}

	resp := &GetRouteTableResponse{}
	_, err = c.Get(ctx, client.ApiPath.RouteTableWithId(d.Id()), resp, nil)

	if err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error retrieving vnpaycloud_route_table"))
	}

	if resp.RouteTable.Status != "OS_DELETING" {
		if _, err := c.Delete(ctx, client.ApiPath.RouteTableWithId(d.Id()), nil); err != nil {
			return diag.FromErr(util.CheckNotFound(d, err, "Error deleting vnpaycloud_route_table"))
		}
	}

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"OS_DELETING", "OS_ACTIVE", "OS_CREATED"},
		Target:     []string{"OS_DELETED"},
		Refresh:    routeTableStateRefreshFunc(ctx, c, resp.RouteTable.ID),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)

	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_route_table %s to Delete:  %s", d.Id(), err)
	}

	return nil
}
