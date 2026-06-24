package routetable

import (
	"context"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceRouteTable() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceRouteTableRead,
		Schema: map[string]*schema.Schema{
			"id":          {Type: schema.TypeString, Required: true},
			"vpc_id":      {Type: schema.TypeString, Computed: true},
			"dest_cidr":   {Type: schema.TypeString, Computed: true},
			"target_id":   {Type: schema.TypeString, Computed: true},
			"target_type": {Type: schema.TypeString, Computed: true},
			"target_name": {Type: schema.TypeString, Computed: true},
			"name":        {Type: schema.TypeString, Computed: true},
			"status":      {Type: schema.TypeString, Computed: true},
			"created_at":  {Type: schema.TypeString, Computed: true},
		},
	}
}

func dataSourceRouteTableRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	id := d.Get("id").(string)

	resp := &dto.RouteTableResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.RouteTableWithID(cfg.ProjectID, id), resp, nil)
	if err != nil {
		return diag.Errorf("Error retrieving vnpaycloud_route_table %s: %s", id, err)
	}

	d.SetId(resp.RouteTable.ID)
	d.Set("vpc_id", resp.RouteTable.VpcID)
	d.Set("dest_cidr", resp.RouteTable.DestCIDR)
	d.Set("target_id", resp.RouteTable.TargetID)
	d.Set("target_type", resp.RouteTable.TargetType)
	d.Set("target_name", resp.RouteTable.TargetName)
	d.Set("name", resp.RouteTable.Name)
	d.Set("status", resp.RouteTable.Status)
	d.Set("created_at", resp.RouteTable.CreatedAt)

	return nil
}

func DataSourceRouteTables() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceRouteTablesRead,
		Schema: map[string]*schema.Schema{
			"vpc_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"route_tables": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id":          {Type: schema.TypeString, Computed: true},
						"vpc_id":      {Type: schema.TypeString, Computed: true},
						"dest_cidr":   {Type: schema.TypeString, Computed: true},
						"target_id":   {Type: schema.TypeString, Computed: true},
						"target_type": {Type: schema.TypeString, Computed: true},
						"target_name": {Type: schema.TypeString, Computed: true},
						"name":        {Type: schema.TypeString, Computed: true},
						"status":      {Type: schema.TypeString, Computed: true},
						"created_at":  {Type: schema.TypeString, Computed: true},
					},
				},
			},
		},
	}
}

func dataSourceRouteTablesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	path := client.ApiPath.RouteTables(cfg.ProjectID)
	vpcID := ""
	if v, ok := d.GetOk("vpc_id"); ok {
		vpcID = v.(string)
		path += "?vpc_id=" + vpcID
	}

	resp := &dto.ListRouteTablesResponse{}
	_, err := cfg.Client.Get(ctx, path, resp, nil)
	if err != nil {
		return diag.Errorf("Error listing vnpaycloud_route_tables: %s", err)
	}

	var rts []map[string]interface{}
	for _, rt := range resp.RouteTables {
		rts = append(rts, map[string]interface{}{
			"id":          rt.ID,
			"vpc_id":      rt.VpcID,
			"dest_cidr":   rt.DestCIDR,
			"target_id":   rt.TargetID,
			"target_type": rt.TargetType,
			"target_name": rt.TargetName,
			"name":        rt.Name,
			"status":      rt.Status,
			"created_at":  rt.CreatedAt,
		})
	}

	if vpcID != "" {
		d.SetId(vpcID)
	} else {
		d.SetId("route_tables")
	}
	d.Set("route_tables", rts)

	return nil
}
