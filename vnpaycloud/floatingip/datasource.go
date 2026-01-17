package floatingip

import (
	"context"
	"log"
	"strings"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/shared"
	"terraform-provider-vnpaycloud/vnpaycloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceNetworkingFloatingIP() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceNetworkingFloatingIPRead,

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"address": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"pool": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"port_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"tenant_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"fixed_ip": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"status": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"tags": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"dns_name": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"dns_domain": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"all_tags": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceNetworkingFloatingIPRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	configMeta := meta.(*config.Config)
	networkingClient, err := client.NewClient(ctx, configMeta.ConsoleClientConfig)
	if err != nil {
		return diag.Errorf("Error creating VNPAYCloud networking client: %s", err)
	}

	listOpts := dto.ListFloatingIPOpts{}

	if v, ok := d.GetOk("description"); ok {
		listOpts.Description = v.(string)
	}

	if v, ok := d.GetOk("address"); ok {
		listOpts.FloatingIP = v.(string)
	}

	if v, ok := d.GetOk("tenant_id"); ok {
		listOpts.TenantID = v.(string)
	}

	if v, ok := d.GetOk("pool"); ok {
		listOpts.FloatingNetworkID = v.(string)
	}

	if v, ok := d.GetOk("port_id"); ok {
		listOpts.PortID = v.(string)
	}

	if v, ok := d.GetOk("fixed_ip"); ok {
		listOpts.FixedIP = v.(string)
	}

	if v, ok := d.GetOk("status"); ok {
		listOpts.Status = v.(string)
	}

	tags := shared.NetworkingAttributesTags(d)
	if len(tags) > 0 {
		listOpts.Tags = strings.Join(tags, ",")
	}

	listResp := dto.ListFloatingIPResponse{}

	_, err = networkingClient.All(ctx, client.ApiPath.FloatingIPWithParams(listOpts), listResp, nil)
	if err != nil {
		return diag.Errorf("Unable to list vnpaycloud_networking_floatingip: %s", err)
	}

	allFloatingIPs := listResp.FloatingIPs

	if len(allFloatingIPs) < 1 {
		return diag.Errorf("No vnpaycloud_networking_floatingip found")
	}

	if len(allFloatingIPs) > 1 {
		return diag.Errorf("More than one vnpaycloud_networking_floatingip found")
	}

	fip := allFloatingIPs[0]

	log.Printf("[DEBUG] Retrieved vnpaycloud_networking_floatingip %s: %+v", fip.ID, fip)
	d.SetId(fip.ID)

	d.Set("description", fip.Description)
	d.Set("address", fip.FloatingIP)
	d.Set("pool", fip.FloatingNetworkID)
	d.Set("port_id", fip.PortID)
	d.Set("fixed_ip", fip.FixedIP)
	d.Set("tenant_id", fip.TenantID)
	d.Set("status", fip.Status)
	d.Set("all_tags", fip.Tags)
	d.Set("dns_name", fip.DNSName)
	d.Set("dns_domain", fip.DNSDomain)
	d.Set("region", util.GetRegion(d, configMeta))

	return nil
}
