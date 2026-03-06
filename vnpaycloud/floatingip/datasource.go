package floatingip

import (
	"context"
	"fmt"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceFloatingIP() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceFloatingIPRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"address": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"port_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"instance_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"instance_name": {
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

func dataSourceFloatingIPRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if id, ok := d.GetOk("id"); ok {
		fipResp := &dto.FloatingIPResponse{}
		_, err := cfg.Client.Get(ctx, client.ApiPath.FloatingIPWithID(cfg.ProjectID, id.(string)), fipResp, nil)
		if err != nil {
			return diag.Errorf("Error fetching vnpaycloud_floating_ip %s: %s", id, err)
		}
		return setFloatingIPData(d, &fipResp.FloatingIP)
	}

	// List and filter client-side
	listResp := &dto.ListFloatingIPsResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.FloatingIPs(cfg.ProjectID), listResp, nil)
	if err != nil {
		return diag.Errorf("Error listing vnpaycloud_floating_ip: %s", err)
	}

	addressFilter, addressOk := d.GetOk("address")

	for _, fip := range listResp.FloatingIPs {
		if addressOk && fip.Address != addressFilter.(string) {
			continue
		}
		return setFloatingIPData(d, &fip)
	}

	return diag.Errorf("No vnpaycloud_floating_ip found matching the criteria")
}

func setFloatingIPData(d *schema.ResourceData, fip *dto.FloatingIP) diag.Diagnostics {
	d.SetId(fip.ID)
	d.Set("address", fip.Address)
	d.Set("status", fip.Status)
	d.Set("port_id", fip.PortID)
	d.Set("instance_id", fip.InstanceID)
	d.Set("instance_name", fip.InstanceName)
	d.Set("created_at", fip.CreatedAt)
	return nil
}

func DataSourceFloatingIPs() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceFloatingIPsRead,
		Schema: map[string]*schema.Schema{
			"floating_ips": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id":            {Type: schema.TypeString, Computed: true},
						"address":       {Type: schema.TypeString, Computed: true},
						"status":        {Type: schema.TypeString, Computed: true},
						"port_id":       {Type: schema.TypeString, Computed: true},
						"instance_id":   {Type: schema.TypeString, Computed: true},
						"instance_name": {Type: schema.TypeString, Computed: true},
						"created_at":    {Type: schema.TypeString, Computed: true},
					},
				},
			},
		},
	}
}

func dataSourceFloatingIPsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	listResp := &dto.ListFloatingIPsResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.FloatingIPs(cfg.ProjectID), listResp, nil)
	if err != nil {
		return diag.Errorf("Error listing vnpaycloud_floating_ips: %s", err)
	}

	var floatingIPs []map[string]interface{}
	for _, fip := range listResp.FloatingIPs {
		floatingIPs = append(floatingIPs, map[string]interface{}{
			"id":            fip.ID,
			"address":       fip.Address,
			"status":        fip.Status,
			"port_id":       fip.PortID,
			"instance_id":   fip.InstanceID,
			"instance_name": fip.InstanceName,
			"created_at":    fip.CreatedAt,
		})
	}

	d.SetId(fmt.Sprintf("floating-ips-%s", cfg.ProjectID))
	d.Set("floating_ips", floatingIPs)

	return nil
}
