package subnetsnat

import (
	"context"
	"fmt"
	"strings"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceSubnetSNAT() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSubnetSNATCreate,
		ReadContext:   resourceSubnetSNATRead,
		DeleteContext: resourceSubnetSNATDelete,
		Schema: map[string]*schema.Schema{
			"subnet_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"floating_ip_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceSubnetSNATCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	subnetID := d.Get("subnet_id").(string)
	floatingIPID := d.Get("floating_ip_id").(string)

	tflog.Debug(ctx, "Enabling SNAT for subnet", map[string]interface{}{
		"subnet_id":      subnetID,
		"floating_ip_id": floatingIPID,
	})

	snatReq := dto.EnableSubnetSNATRequest{FloatingIpID: floatingIPID}
	_, err := cfg.Client.Put(ctx, client.ApiPath.SubnetEnableSNAT(cfg.ProjectID, subnetID), snatReq, nil, nil)
	if err != nil {
		// If SNAT is already enabled, adopt the existing state
		if strings.Contains(err.Error(), "already has SNAT enabled") {
			tflog.Info(ctx, "SNAT already enabled for subnet, adopting existing state", map[string]interface{}{"subnet_id": subnetID})
		} else {
			return diag.Errorf("Error enabling SNAT for subnet %s: %s", subnetID, err)
		}
	}

	d.SetId(fmt.Sprintf("%s/snat", subnetID))
	return resourceSubnetSNATRead(ctx, d, meta)
}

func resourceSubnetSNATRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	subnetID := d.Get("subnet_id").(string)

	subnetResp := &dto.SubnetResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.SubnetWithID(cfg.ProjectID, subnetID), subnetResp, nil)
	if err != nil {
		d.SetId("")
		return diag.Errorf("Error reading subnet %s for SNAT status: %s", subnetID, err)
	}

	if !subnetResp.Subnet.EnableSnat {
		// SNAT was disabled externally
		d.SetId("")
		return nil
	}

	d.Set("floating_ip_id", subnetResp.Subnet.ExternalIpID)
	return nil
}

func resourceSubnetSNATDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	subnetID := d.Get("subnet_id").(string)

	tflog.Debug(ctx, "Disabling SNAT for subnet", map[string]interface{}{
		"subnet_id": subnetID,
	})

	_, err := cfg.Client.Put(ctx, client.ApiPath.SubnetDisableSNAT(cfg.ProjectID, subnetID), nil, nil, nil)
	if err != nil {
		return diag.Errorf("Error disabling SNAT for subnet %s: %s", subnetID, err)
	}

	return nil
}
