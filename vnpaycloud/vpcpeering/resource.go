package vpcpeering

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

func ResourceVPCPeering() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVPCPeeringCreate,
		ReadContext:   resourceVPCPeeringRead,
		UpdateContext: resourceVPCPeeringUpdate,
		DeleteContext: resourceVPCPeeringDelete,
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
				Optional: true,
				Computed: true,
			},
			"src_vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"dest_vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"peering_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"src_vpc_cidr": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dest_vpc_cidr": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"reverse_peering_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceVPCPeeringCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	srcVpcID := d.Get("src_vpc_id").(string)

	createOpts := dto.CreatePeeringConnectionRequest{
		SrcVpcID:    srcVpcID,
		DestVpcID:   d.Get("dest_vpc_id").(string),
		Description: d.Get("description").(string),
	}

	tflog.Debug(ctx, "vnpaycloud_vpc_peering create options", map[string]interface{}{"create_opts": createOpts})

	// Create returns ListPeeringConnectionsResponse (both directions).
	// We need to find the one matching our src_vpc_id.
	createResp := &dto.ListPeeringConnectionsResponse{}
	_, err := cfg.Client.Post(ctx, client.ApiPath.PeeringConnections(), createOpts, createResp, nil)
	if err != nil {
		return diag.Errorf("Error creating vnpaycloud_vpc_peering: %s", err)
	}

	// Create returns peerings (bidirectional). Pick our direction as primary.
	if len(createResp.PeeringConnections) == 0 {
		return diag.Errorf("Error creating vnpaycloud_vpc_peering: no peering connection returned")
	}

	var primary *dto.PeeringConnection
	var reversePeeringID string

	if len(createResp.PeeringConnections) >= 2 {
		// Both directions returned — pick by src_vpc_id match
		p0 := &createResp.PeeringConnections[0]
		p1 := &createResp.PeeringConnections[1]
		if p0.SrcVpcID == srcVpcID {
			primary, reversePeeringID = p0, p1.ID
		} else {
			primary, reversePeeringID = p1, p0.ID
		}
	} else {
		// Only 1 returned — use it, find reverse via list after state poll
		primary = &createResp.PeeringConnections[0]
	}

	d.SetId(primary.ID)

	tflog.Info(ctx, "Created peering", map[string]interface{}{
		"primary_id": primary.ID, "reverse_id": reversePeeringID,
	})

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"creating", "unknown"},
		Target:     []string{"active"},
		Refresh:    vpcPeeringStateRefreshFunc(ctx, cfg.Client, primary.ID),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_vpc_peering %s to become ready: %s", primary.ID, err)
	}

	// Find reverse peering if not already known (bidirectional cleanup)
	if reversePeeringID == "" {
		reversePeeringID = findReversePeeringID(ctx, cfg.Client, primary.ID)
	}
	d.Set("reverse_peering_id", reversePeeringID)

	return resourceVPCPeeringRead(ctx, d, meta)
}

func resourceVPCPeeringRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	peeringResp := &dto.PeeringConnectionResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.PeeringConnectionWithID(d.Id()), peeringResp, nil)
	if err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error retrieving vnpaycloud_vpc_peering"))
	}

	tflog.Debug(ctx, "Retrieved vnpaycloud_vpc_peering "+d.Id(), map[string]interface{}{"peering": peeringResp.PeeringConnection})

	d.Set("name", peeringResp.PeeringConnection.Name)
	d.Set("src_vpc_id", peeringResp.PeeringConnection.SrcVpcID)
	d.Set("dest_vpc_id", peeringResp.PeeringConnection.DestVpcID)
	d.Set("status", peeringResp.PeeringConnection.Status)
	d.Set("peering_status", peeringResp.PeeringConnection.PeeringStatus)
	d.Set("src_vpc_cidr", peeringResp.PeeringConnection.SrcVpcCIDR)
	d.Set("dest_vpc_cidr", peeringResp.PeeringConnection.DestVpcCIDR)
	d.Set("created_at", peeringResp.PeeringConnection.CreatedAt)

	return nil
}

func resourceVPCPeeringUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if d.HasChange("name") {
		updateOpts := dto.UpdatePeeringConnectionRequest{
			Name: d.Get("name").(string),
		}

		tflog.Debug(ctx, "vnpaycloud_vpc_peering update options", map[string]interface{}{"update_opts": updateOpts})

		_, err := cfg.Client.Put(ctx, client.ApiPath.PeeringConnectionWithID(d.Id()), updateOpts, nil, nil)
		if err != nil {
			return diag.Errorf("Error updating vnpaycloud_vpc_peering %s: %s", d.Id(), err)
		}
	}

	return resourceVPCPeeringRead(ctx, d, meta)
}

func resourceVPCPeeringDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	// Delete primary peering
	if diags := deletePeeringByID(ctx, cfg, d, d.Id()); diags != nil {
		return diags
	}

	// Delete reverse peering (bidirectional: backend creates both directions)
	if reverseID, ok := d.GetOk("reverse_peering_id"); ok && reverseID.(string) != "" {
		tflog.Info(ctx, "Deleting reverse peering connection", map[string]interface{}{"reverse_peering_id": reverseID})
		if diags := deletePeeringByID(ctx, cfg, d, reverseID.(string)); diags != nil {
			tflog.Warn(ctx, "Failed to delete reverse peering, may already be deleted", map[string]interface{}{"reverse_peering_id": reverseID, "error": diags[0].Summary})
		}
	}

	return nil
}

func deletePeeringByID(ctx context.Context, cfg *config.Config, d *schema.ResourceData, peeringID string) diag.Diagnostics {
	peeringResp := &dto.PeeringConnectionResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.PeeringConnectionWithID(peeringID), peeringResp, nil)
	if err != nil {
		// Already gone
		if util.ResponseCodeIs(err, 404) {
			return nil
		}
		return diag.FromErr(util.CheckDeleted(d, err, "Error retrieving vnpaycloud_vpc_peering"))
	}

	if peeringResp.PeeringConnection.Status == "deleted" {
		return nil
	}

	if peeringResp.PeeringConnection.Status != "deleting" {
		if _, err := cfg.Client.Delete(ctx, client.ApiPath.PeeringConnectionWithID(peeringID), nil); err != nil {
			if util.ResponseCodeIs(err, 404) {
				return nil
			}
			return diag.Errorf("Error deleting vnpaycloud_vpc_peering %s: %s", peeringID, err)
		}
	}

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"deleting", "active"},
		Target:     []string{"deleted"},
		Refresh:    vpcPeeringStateRefreshFunc(ctx, cfg.Client, peeringID),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_vpc_peering %s to delete: %s", peeringID, err)
	}

	return nil
}
