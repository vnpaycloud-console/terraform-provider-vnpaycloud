package peeringconnection

import (
	"context"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourcePeeringConnection() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePeeringConnectionCreate,
		ReadContext:   resourcePeeringConnectionRead,
		UpdateContext: resourcePeeringConnectionUpdate,
		DeleteContext: resourcePeeringConnectionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
			},
			"peer_vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"peer_org_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"side": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"requester", "accepter"}, false),
			},
			"request_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"peer_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"port_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourcePeeringConnectionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	switch d.Get("side").(string) {
	case "requester":
		return createPeeringConnectionRequest(ctx, d, meta)
	case "accepter":
		return approvePeeringConnectionRequest(ctx, d, meta)
	default:
		return diag.Errorf("Error invalid peering connection side: %s", d.Get("side").(string))
	}
}

func createPeeringConnectionRequest(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	c, err := client.NewClient(ctx, config.ConsoleClientConfig)

	if err != nil {
		return diag.Errorf("Error creating VNPAY Cloud Peering Connection client: %s", err)
	}

	createOpts := CreatePeeringConnectionRequest{
		PeeringConnectionRequest: CreatePeeringConnectionRequestOpts{
			PeerVPCId:   d.Get("peer_vpc_id").(string),
			PeerOrgId:   d.Get("peer_org_id").(string),
			VPCId:       d.Get("vpc_id").(string),
			Description: d.Get("description").(string),
		},
	}

	tflog.Debug(ctx, "vnpaycloud_peering_connection request options", map[string]interface{}{"create_opts": createOpts})
	createResp := &CreatePeeringConnectionRequestResponse{}
	_, err = c.Post(ctx, client.ApiPath.PeeringConnectionRequest, createOpts, createResp, nil)

	if err != nil {
		return diag.Errorf("Error creating vnpaycloud_peering_connection request: %s", err)
	}

	peeringConnectionRequest := createResp.PeeringConnectionRequest

	d.SetId(peeringConnectionRequest.PeerId)
	d.Set("request_id", peeringConnectionRequest.ID)

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"OS_INITIATING"},
		Target:     []string{"OS_ACTIVE", "OS_CREATED"},
		Refresh:    peeringConnectionRequestStateRefreshFunc(ctx, c, peeringConnectionRequest.ID),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)

	if err != nil {
		return diag.Errorf(
			"Error waiting for vnpaycloud_peering_connection %s to become ready: %s", peeringConnectionRequest.ID, err)
	}

	return resourcePeeringConnectionRead(ctx, d, meta)
}

func approvePeeringConnectionRequest(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	c, err := client.NewClient(ctx, config.ConsoleClientConfig)

	if err != nil {
		return diag.Errorf("Error creating VNPAY Cloud Peering Connection client: %s", err)
	}

	listOtps := ListPeeringConnectApprovalRequest{
		VPCId:     d.Get("vpc_id").(string),
		PeerVPCId: d.Get("peer_vpc_id").(string),
		PeerOrgId: d.Get("peer_org_id").(string),
		Status:    "PCS_PENDING",
	}

	tflog.Debug(ctx, "vnpaycloud_peering_connection list approvals options", map[string]interface{}{"list_otps": listOtps})
	listResp := &ListPeeringConnectApprovalResponse{}
	_, err = c.Get(ctx, client.ApiPath.ListPeeringConnectionApproval(listOtps), listResp, nil)

	if err != nil {
		return diag.Errorf("Unable to query vnpaycloud_peering_connection: %s", err)
	}

	if len(listResp.PeeringConnectApprovals) > 1 {
		return diag.Errorf("Your vnpaycloud_peering_connection query returns multiple approvals")

	}

	if len(listResp.PeeringConnectApprovals) == 0 {
		return diag.Errorf("Your vnpaycloud_peering_connection query did not return approval results.")
	}

	updateOtps := UpdatePeeringConnectApprovalRequest{
		PeeringConnectApproval: PeeringConnectApprovalOpts{
			Accept: true,
		},
	}

	tflog.Debug(ctx, "vnpaycloud_peering_connection approval options", map[string]interface{}{"update_otps": updateOtps})
	updateResp := &UpdatePeeringConnectApprovalResponse{}
	_, err = c.Put(ctx, client.ApiPath.PeeringConnectionApprovalWithId(listResp.PeeringConnectApprovals[0].ID), updateOtps, updateResp, nil)

	if err != nil {
		tflog.Error(ctx, "Error updating VNPAY Cloud Peering Connection Approval", map[string]interface{}{"err": err})
	}

	peeringConnectionApproval := updateResp.PeeringConnectApproval

	d.SetId(peeringConnectionApproval.PeerId)
	tflog.Debug(ctx, "vnpaycloud_peering_connection approval id", map[string]interface{}{"id": peeringConnectionApproval.ID})

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"OS_INITIATING"},
		Target:     []string{"OS_ACTIVE", "OS_CREATED"},
		Refresh:    peeringConnectionStateRefreshFunc(ctx, c, peeringConnectionApproval.PeerId),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      15 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)

	if err != nil {
		return diag.Errorf(
			"Error waiting for vnpaycloud_peering_connection %s to become ready: %s", peeringConnectionApproval.PeerId, err)
	}

	return resourcePeeringConnectionRead(ctx, d, meta)
}

func resourcePeeringConnectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	switch d.Get("side").(string) {
	case "requester":
		return resourcePeeringConnectionRequestRead(ctx, d, meta)
	case "accepter":
		return resourcePeeringConnectionApprovalRead(ctx, d, meta)
	default:
		return diag.Errorf("Error invalid peering connection side: %s", d.Get("side").(string))
	}
}

func resourcePeeringConnectionRequestRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	c, err := client.NewClient(ctx, config.ConsoleClientConfig)

	if err != nil {
		return diag.Errorf("Error creating VNPAY Cloud Peering Connection client: %s", err)
	}

	getPeeringConnectionRequestResp := &GetPeeringConnectionRequestResponse{}
	_, err = c.Get(ctx, client.ApiPath.PeeringConnectionRequestWithId(d.Get("request_id").(string)), getPeeringConnectionRequestResp, nil)

	if err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error retrieving peering connection request"))
	}

	peeringConnectionRequest := getPeeringConnectionRequestResp.PeeringConnectionRequest

	tflog.Debug(ctx, "Retrieved peering connection request "+d.Id(), map[string]interface{}{"request": peeringConnectionRequest})

	d.Set("vpc_id", peeringConnectionRequest.VpcId)
	d.Set("request_id", peeringConnectionRequest.ID)
	d.Set("peer_vpc_id", peeringConnectionRequest.PeerVpcId)
	d.Set("peer_org_id", peeringConnectionRequest.PeerOrgId)
	d.Set("description", peeringConnectionRequest.Description)
	d.Set("peer_status", peeringConnectionRequest.RequestStatus)

	if peeringConnectionRequest.RequestStatus != "PCS_ACCEPTED" {
		return nil
	}

	getPeeringConnectionResp := &GetPeeringConnectionResponse{}

	_, err = c.Get(ctx, client.ApiPath.PeeringConnectionWithId(d.Id()), getPeeringConnectionResp, nil)

	if err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error retrieving peering connection"))
	}

	peeringConnection := getPeeringConnectionResp.PeeringConnection

	d.Set("peer_status", peeringConnection.PeerStatus)
	d.Set("status", peeringConnection.Status)
	d.Set("port_id", peeringConnection.PortId)

	return nil
}

func resourcePeeringConnectionApprovalRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	c, err := client.NewClient(ctx, config.ConsoleClientConfig)

	if err != nil {
		return diag.Errorf("Error creating VNPAY Cloud Peering Connection client: %s", err)
	}

	getPeeringConnectionResp := &GetPeeringConnectionResponse{}
	_, err = c.Get(ctx, client.ApiPath.PeeringConnectionWithId(d.Id()), getPeeringConnectionResp, nil)

	if err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error retrieving peering connection"))
	}

	peeringConnection := getPeeringConnectionResp.PeeringConnection

	tflog.Debug(ctx, "Retrieved vnpaycloud_peering_connection "+d.Id(), map[string]interface{}{"peering_connection": peeringConnection})

	d.Set("vpc_id", peeringConnection.VpcId)
	d.Set("peer_vpc_id", peeringConnection.PeerVpcId)
	d.Set("peer_org_id", peeringConnection.PeerOrgId)
	d.Set("description", peeringConnection.Description)
	d.Set("peer_status", peeringConnection.PeerStatus)
	d.Set("status", peeringConnection.Status)
	d.Set("port_id", peeringConnection.PortId)

	return nil
}

func resourcePeeringConnectionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourcePeeringConnectionRead(ctx, d, meta)
}

func resourcePeeringConnectionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	c, err := client.NewClient(ctx, config.ConsoleClientConfig)

	if err != nil {
		return diag.Errorf("Error creating VNPAY Cloud Peering Connection client: %s", err)
	}

	resp := &GetPeeringConnectionResponse{}
	_, err = c.Get(ctx, client.ApiPath.PeeringConnectionWithId(d.Id()), resp, nil)

	if err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error retrieving vnpaycloud_peering_connection"))
	}

	if resp.PeeringConnection.Status != "OS_DELETING" {
		if _, err := c.Delete(ctx, client.ApiPath.PeeringConnectionWithId(d.Id()), nil); err != nil {
			return diag.FromErr(util.CheckNotFound(d, err, "Error deleting vnpaycloud_peering_connection"))
		}
	}

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"OS_DELETING", "OS_ACTIVE", "OS_CREATED"},
		Target:     []string{"OS_DELETED"},
		Refresh:    peeringConnectionStateRefreshFunc(ctx, c, resp.PeeringConnection.ID),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)

	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_peering_connection %s to Delete:  %s", d.Id(), err)
	}

	return nil
}
