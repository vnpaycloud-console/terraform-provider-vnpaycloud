package vpngateway

import (
	"context"
	"fmt"
	"strings"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const vpnGatewayVPCAttachmentIDSeparator = "/"

func ResourceVPNGatewayVPCAttachment() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVPNGatewayVPCAttachmentCreate,
		ReadContext:   resourceVPNGatewayVPCAttachmentRead,
		DeleteContext: resourceVPNGatewayVPCAttachmentDelete,
		Description:   "Attaches a VPC to a VNPAY Cloud VPN gateway.",
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				vpnGatewayID, vpcID, err := parseVPNGatewayVPCAttachmentID(d.Id())
				if err != nil {
					return nil, err
				}

				d.Set("vpn_gateway_id", vpnGatewayID)
				d.Set("vpc_id", vpcID)

				return []*schema.ResourceData{d}, nil
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"vpn_gateway_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the VPN gateway.",
			},
			"vpc_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the VPC to attach to the VPN gateway.",
			},
		},
	}
}

func resourceVPNGatewayVPCAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	vpnGatewayID := d.Get("vpn_gateway_id").(string)
	vpcID := d.Get("vpc_id").(string)

	// Serialize all mutating operations on the same VPN gateway (attach/detach
	// and connection create/delete share this key) so concurrent applies don't
	// collide on the backend's "being modified by another operation" guard.
	cfg.MutexKV.Lock(vpnGatewayMutexKey(vpnGatewayID))
	defer cfg.MutexKV.Unlock(vpnGatewayMutexKey(vpnGatewayID))

	attached, diags := vpnGatewayHasAttachedVPC(ctx, cfg, d, vpnGatewayID, vpcID)
	if diags.HasError() {
		return diags
	}
	if attached {
		d.SetId(vpnGatewayVPCAttachmentID(vpnGatewayID, vpcID))
		return nil
	}

	err := retryVPNGatewayVPCAttachmentOperation(ctx, d.Timeout(schema.TimeoutCreate), func() error {
		return attachVPNGatewayToVPC(ctx, cfg, vpnGatewayID, vpcID)
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(vpnGatewayVPCAttachmentID(vpnGatewayID, vpcID))

	return resourceVPNGatewayVPCAttachmentRead(ctx, d, meta)
}

func resourceVPNGatewayVPCAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	vpnGatewayID := d.Get("vpn_gateway_id").(string)
	vpcID := d.Get("vpc_id").(string)

	attached, diags := vpnGatewayHasAttachedVPC(ctx, cfg, d, vpnGatewayID, vpcID)
	if diags.HasError() || attached {
		return diags
	}

	tflog.Info(ctx, "VPN gateway VPC attachment not found, removing from state", map[string]interface{}{
		"vpn_gateway_id": vpnGatewayID,
		"vpc_id":         vpcID,
	})
	d.SetId("")

	return nil
}

func resourceVPNGatewayVPCAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	vpnGatewayID := d.Get("vpn_gateway_id").(string)
	vpcID := d.Get("vpc_id").(string)

	cfg.MutexKV.Lock(vpnGatewayMutexKey(vpnGatewayID))
	defer cfg.MutexKV.Unlock(vpnGatewayMutexKey(vpnGatewayID))

	attached, diags := vpnGatewayHasAttachedVPC(ctx, cfg, d, vpnGatewayID, vpcID)
	if diags.HasError() {
		return diags
	}
	if !attached {
		d.SetId("")
		return nil
	}

	err := retryVPNGatewayVPCAttachmentOperation(ctx, d.Timeout(schema.TimeoutDelete), func() error {
		return detachVPNGatewayFromVPC(ctx, cfg, vpnGatewayID, vpcID)
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}

func vpnGatewayVPCAttachmentID(vpnGatewayID, vpcID string) string {
	return vpnGatewayID + vpnGatewayVPCAttachmentIDSeparator + vpcID
}

func vpnGatewayHasAttachedVPC(ctx context.Context, cfg *config.Config, d *schema.ResourceData, vpnGatewayID, vpcID string) (bool, diag.Diagnostics) {
	gwResp := &dto.VPNGatewayResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.VPNGatewayWithID(cfg.ProjectID, vpnGatewayID), gwResp, nil)
	if err != nil {
		return false, diag.FromErr(util.CheckNotFound(d, err, "Error retrieving vnpaycloud_vpn_gateway_vpc_attachment"))
	}

	for _, attachedVPCID := range gwResp.VPNGateway.AttachedVPCIDs {
		if attachedVPCID == vpcID {
			d.SetId(vpnGatewayVPCAttachmentID(vpnGatewayID, vpcID))
			return true, nil
		}
	}

	return false, nil
}

func parseVPNGatewayVPCAttachmentID(id string) (string, string, error) {
	parts := strings.Split(id, vpnGatewayVPCAttachmentIDSeparator)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("expected import ID format <vpn_gateway_id>/<vpc_id>, got %q", id)
	}

	return parts[0], parts[1], nil
}

func attachVPNGatewayToVPC(ctx context.Context, cfg *config.Config, vpnGatewayID, vpcID string) error {
	tflog.Debug(ctx, "Attaching vnpaycloud_vpn_gateway to VPC", map[string]interface{}{
		"vpn_gateway_id": vpnGatewayID,
		"vpc_id":         vpcID,
	})

	attachReq := dto.AttachVPCToVPNGatewayRequest{
		VPCID: vpcID,
	}
	_, err := cfg.Client.Post(ctx, client.ApiPath.VPNGatewayAttachVPC(cfg.ProjectID, vpnGatewayID), attachReq, nil, nil)
	if err != nil {
		return fmt.Errorf("error attaching vnpaycloud_vpn_gateway %s to VPC %s: %s", vpnGatewayID, vpcID, err)
	}

	return nil
}

func detachVPNGatewayFromVPC(ctx context.Context, cfg *config.Config, vpnGatewayID, vpcID string) error {
	tflog.Debug(ctx, "Detaching vnpaycloud_vpn_gateway from VPC", map[string]interface{}{
		"vpn_gateway_id": vpnGatewayID,
		"vpc_id":         vpcID,
	})

	detachReq := dto.DetachVPCFromVPNGatewayRequest{
		VPCID: vpcID,
	}
	_, err := cfg.Client.Post(ctx, client.ApiPath.VPNGatewayDetachVPC(cfg.ProjectID, vpnGatewayID), detachReq, nil, nil)
	if err != nil {
		return fmt.Errorf("error detaching vnpaycloud_vpn_gateway %s from VPC %s: %s", vpnGatewayID, vpcID, err)
	}

	return nil
}

func retryVPNGatewayVPCAttachmentOperation(ctx context.Context, timeout time.Duration, operation func() error) error {
	return retry.RetryContext(ctx, timeout, func() *retry.RetryError {
		err := operation()
		if err == nil {
			return nil
		}

		if strings.Contains(strings.ToLower(err.Error()), "currently being modified by another operation") {
			return retry.RetryableError(err)
		}

		return retry.NonRetryableError(err)
	})
}

// vpnGatewayMutexKey returns the MutexKV key used to serialize mutating
// operations on a single VPN gateway. The vpnconnection package locks on the
// same key format so connection create/delete and VPC attach/detach never run
// concurrently against the same gateway.
func vpnGatewayMutexKey(vpnGatewayID string) string {
	return "vnpaycloud_vpn_gateway/" + vpnGatewayID
}
