package floatingip

import (
	"context"
	"log"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/vnpaycloud-console/gophercloud/v2/openstack/networking/v2/extensions/layer3/floatingips"
)

func ResourceNetworkingFloatingIPAssociateV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNetworkingFloatingIPAssociateV2Create,
		ReadContext:   resourceNetworkingFloatingIPAssociateV2Read,
		UpdateContext: resourceNetworkingFloatingIPAssociateV2Update,
		DeleteContext: resourceNetworkingFloatingIPAssociateV2Delete,
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

func resourceNetworkingFloatingIPAssociateV2Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	configMeta := meta.(*config.Config)
	networkingClient, err := configMeta.NetworkingV2Client(ctx, util.GetRegion(d, configMeta))
	if err != nil {
		return diag.Errorf("Error creating OpenStack network client: %s", err)
	}

	floatingIP := d.Get("floating_ip").(string)
	portID := d.Get("port_id").(string)

	fipID, err := networkingFloatingIPV2ID(ctx, networkingClient, floatingIP)
	if err != nil {
		return diag.Errorf("Unable to get ID of openstack_networking_floatingip_associate_v2 floating_ip %s: %s", floatingIP, err)
	}

	updateOpts := floatingips.UpdateOpts{
		PortID: &portID,
	}

	log.Printf("[DEBUG] openstack_networking_floatingip_associate_v2 create options: %#v", updateOpts)
	_, err = floatingips.Update(ctx, networkingClient, fipID, updateOpts).Extract()
	if err != nil {
		return diag.Errorf("Error associating openstack_networking_floatingip_associate_v2 floating_ip %s with port %s: %s", fipID, portID, err)
	}

	d.SetId(fipID)

	return resourceNetworkingFloatingIPAssociateV2Read(ctx, d, meta)
}

func resourceNetworkingFloatingIPAssociateV2Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	configMeta := meta.(*config.Config)
	networkingClient, err := configMeta.NetworkingV2Client(ctx, util.GetRegion(d, configMeta))
	if err != nil {
		return diag.Errorf("Error creating OpenStack network client: %s", err)
	}

	fip, err := floatingips.Get(ctx, networkingClient, d.Id()).Extract()
	if err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error getting openstack_networking_floatingip_associate_v2"))
	}

	log.Printf("[DEBUG] Retrieved openstack_networking_floatingip_associate_v2 %s: %#v", d.Id(), fip)

	d.Set("floating_ip", fip.FloatingIP)
	d.Set("port_id", fip.PortID)
	d.Set("region", util.GetRegion(d, configMeta))

	return nil
}

func resourceNetworkingFloatingIPAssociateV2Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	configMeta := meta.(*config.Config)
	networkingClient, err := configMeta.NetworkingV2Client(ctx, util.GetRegion(d, configMeta))
	if err != nil {
		return diag.Errorf("Error creating OpenStack network client: %s", err)
	}

	var updateOpts floatingips.UpdateOpts

	// port_id must always exist
	portID := d.Get("port_id").(string)
	updateOpts.PortID = &portID

	log.Printf("[DEBUG] openstack_networking_floatingip_associate_v2 %s update options: %#v", d.Id(), updateOpts)
	_, err = floatingips.Update(ctx, networkingClient, d.Id(), updateOpts).Extract()
	if err != nil {
		return diag.Errorf("Error updating openstack_networking_floatingip_associate_v2 %s: %s", d.Id(), err)
	}

	return resourceNetworkingFloatingIPAssociateV2Read(ctx, d, meta)
}

func resourceNetworkingFloatingIPAssociateV2Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	configMeta := meta.(*config.Config)
	networkingClient, err := configMeta.NetworkingV2Client(ctx, util.GetRegion(d, configMeta))
	if err != nil {
		return diag.Errorf("Error creating OpenStack network client: %s", err)
	}

	portID := d.Get("port_id").(string)
	updateOpts := floatingips.UpdateOpts{
		PortID: new(string),
	}

	log.Printf("[DEBUG] openstack_networking_floatingip_associate_v2 disassociating options: %#v", updateOpts)
	_, err = floatingips.Update(ctx, networkingClient, d.Id(), updateOpts).Extract()
	if err != nil {
		return diag.Errorf("Error disassociating openstack_networking_floatingip_associate_v2 floating_ip %s with port %s: %s", d.Id(), portID, err)
	}

	return nil
}
