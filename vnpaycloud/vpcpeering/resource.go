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
				DiffSuppressFunc: func(k, oldValue, newValue string, d *schema.ResourceData) bool {
					return d.Id() != ""
				},
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

	createResp := &dto.ListPeeringConnectionsResponse{}
	_, err := cfg.Client.Post(ctx, client.ApiPath.PeeringConnections(), createOpts, createResp, nil)
	if err != nil {
		return diag.Errorf("Error creating vnpaycloud_vpc_peering: %s", err)
	}

	if len(createResp.PeeringConnections) == 0 {
		return diag.Errorf("Error creating vnpaycloud_vpc_peering: no peering connection returned")
	}

	var primary *dto.PeeringConnection
	var reversePeeringID string

	if len(createResp.PeeringConnections) >= 2 {
		p0 := &createResp.PeeringConnections[0]
		p1 := &createResp.PeeringConnections[1]
		if p0.SrcVpcID == srcVpcID {
			primary, reversePeeringID = p0, p1.ID
		} else {
			primary, reversePeeringID = p1, p0.ID
		}
	} else {
		primary = &createResp.PeeringConnections[0]
	}

	d.SetId(primary.ID)

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

	if reversePeeringID == "" {
		reversePeeringID = findReversePeeringID(ctx, cfg.Client, primary.ID)
	}
	d.Set("reverse_peering_id", reversePeeringID)

	if name := d.Get("name").(string); name != "" {
		updateOpts := dto.UpdatePeeringConnectionRequest{Name: name}
		if _, err := cfg.Client.Put(ctx, client.ApiPath.PeeringConnectionWithID(primary.ID), updateOpts, nil, nil); err != nil {
			return diag.Errorf("Error setting name on vnpaycloud_vpc_peering %s after create: %s", primary.ID, err)
		}
	}

	return resourceVPCPeeringRead(ctx, d, meta)
}

func resourceVPCPeeringRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	peeringResp := &dto.PeeringConnectionResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.PeeringConnectionWithID(d.Id()), peeringResp, nil)
	if err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error retrieving vnpaycloud_vpc_peering"))
	}

	d.Set("name", peeringResp.PeeringConnection.Name)
	d.Set("src_vpc_id", peeringResp.PeeringConnection.SrcVpcID)
	d.Set("dest_vpc_id", peeringResp.PeeringConnection.DestVpcID)
	d.Set("status", peeringResp.PeeringConnection.Status)
	d.Set("peering_status", peeringResp.PeeringConnection.PeeringStatus)
	d.Set("src_vpc_cidr", peeringResp.PeeringConnection.SrcVpcCIDR)
	d.Set("dest_vpc_cidr", peeringResp.PeeringConnection.DestVpcCIDR)
	d.Set("created_at", peeringResp.PeeringConnection.CreatedAt)

	if d.Get("reverse_peering_id").(string) == "" {
		if rev := findReversePeeringID(ctx, cfg.Client, d.Id()); rev != "" {
			d.Set("reverse_peering_id", rev)
		}
	}

	return nil
}

func resourceVPCPeeringUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if d.HasChange("name") {
		updateOpts := dto.UpdatePeeringConnectionRequest{
			Name: d.Get("name").(string),
		}

		_, err := cfg.Client.Put(ctx, client.ApiPath.PeeringConnectionWithID(d.Id()), updateOpts, nil, nil)
		if err != nil {
			return diag.Errorf("Error updating vnpaycloud_vpc_peering %s: %s", d.Id(), err)
		}
	}

	return resourceVPCPeeringRead(ctx, d, meta)
}

func resourceVPCPeeringDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if diags := deletePeeringByID(ctx, cfg, d, d.Id()); diags != nil {
		return diags
	}

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
