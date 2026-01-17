package floatingip

import (
	"context"
	"log"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceNetworkingFloatingIPAssociate() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNetworkingFloatingIPAssociateCreate,
		ReadContext:   resourceNetworkingFloatingIPAssociateRead,
		UpdateContext: resourceNetworkingFloatingIPAssociateUpdate,
		DeleteContext: resourceNetworkingFloatingIPAssociateDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"floating_ip": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"port_id": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceNetworkingFloatingIPAssociateCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	configMeta := meta.(*config.Config)
	networkingClient, err := client.NewClient(ctx, configMeta.ConsoleClientConfig)
	if err != nil {
		return diag.Errorf("Error creating VNPAYCloud network client: %s", err)
	}

	floatingIP := d.Get("floating_ip").(string)
	portID := d.Get("port_id").(string)

	fipID, err := networkingFloatingIPID(ctx, networkingClient, floatingIP)
	if err != nil {
		return diag.Errorf("Unable to get ID of vnpaycloud_networking_floatingip_associate floating_ip %s: %s", floatingIP, err)
	}

	updateOpts := dto.UpdateFloatingIPOpts{
		PortID: &portID,
	}
	updateReq := dto.UpdateFloatingIPRequest{
		FloatingIP: updateOpts,
	}

	log.Printf("[DEBUG] vnpaycloud_networking_floatingip_associate create options: %#v", updateOpts)
	_, err = networkingClient.Put(ctx, client.ApiPath.FloatingIPWithId(fipID), updateReq, nil, nil)
	if err != nil {
		return diag.Errorf("Error associating vnpaycloud_networking_floatingip_associate floating_ip %s with port %s: %s", fipID, portID, err)
	}

	d.SetId(fipID)

	return resourceNetworkingFloatingIPAssociateRead(ctx, d, meta)
}

func resourceNetworkingFloatingIPAssociateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	configMeta := meta.(*config.Config)
	networkingClient, err := client.NewClient(ctx, configMeta.ConsoleClientConfig)
	if err != nil {
		return diag.Errorf("Error creating VNPAYCloud network client: %s", err)
	}

	fipResp := dto.GetFloatingIPResponse{}
	_, err = networkingClient.Get(ctx, client.ApiPath.FloatingIPWithId(d.Id()), &fipResp, nil)
	if err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error getting vnpaycloud_networking_floatingip_associate"))
	}

	log.Printf("[DEBUG] Retrieved vnpaycloud_networking_floatingip_associate %s: %#v", d.Id(), fipResp)

	d.Set("floating_ip", fipResp.FloatingIP.FloatingIP)
	d.Set("port_id", fipResp.FloatingIP.PortID)
	d.Set("region", util.GetRegion(d, configMeta))

	return nil
}

func resourceNetworkingFloatingIPAssociateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	configMeta := meta.(*config.Config)
	networkingClient, err := client.NewClient(ctx, configMeta.ConsoleClientConfig)
	if err != nil {
		return diag.Errorf("Error creating VNPAYCloud network client: %s", err)
	}

	var updateOpts dto.UpdateFloatingIPOpts

	// port_id must always exist
	portID := d.Get("port_id").(string)
	updateOpts.PortID = &portID

	log.Printf("[DEBUG] vnpaycloud_networking_floatingip_associate %s update options: %#v", d.Id(), updateOpts)
	_, err = networkingClient.Put(ctx, client.ApiPath.FloatingIPWithId(d.Id()), updateOpts, nil, nil)
	if err != nil {
		return diag.Errorf("Error updating vnpaycloud_networking_floatingip_associate %s: %s", d.Id(), err)
	}

	return resourceNetworkingFloatingIPAssociateRead(ctx, d, meta)
}

func resourceNetworkingFloatingIPAssociateDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	configMeta := meta.(*config.Config)
	networkingClient, err := client.NewClient(ctx, configMeta.ConsoleClientConfig)
	if err != nil {
		return diag.Errorf("Error creating VNPAYCloud network client: %s", err)
	}

	portID := d.Get("port_id").(string)
	updateOpts := dto.UpdateFloatingIPOpts{
		PortID: new(string),
	}
	updateReq := dto.UpdateFloatingIPRequest{
		FloatingIP: updateOpts,
	}

	log.Printf("[DEBUG] vnpaycloud_networking_floatingip_associate disassociating options: %#v", updateOpts)
	_, err = networkingClient.Put(ctx, client.ApiPath.FloatingIPWithId(d.Id()), updateReq, nil, nil)
	if err != nil {
		return diag.Errorf("Error disassociating vnpaycloud_networking_floatingip_associate floating_ip %s with port %s: %s", d.Id(), portID, err)
	}

	return nil
}
