package peeringconnection

import (
	"context"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vnpaycloud-console/gophercloud/v2/openstack/networking/v2/peeringconnectionapprovals"
	"github.com/vnpaycloud-console/gophercloud/v2/openstack/networking/v2/peeringconnectionrequests"
	"github.com/vnpaycloud-console/gophercloud/v2/openstack/networking/v2/peeringconnections"
	"github.com/vnpaycloud-console/gophercloud/v2/openstack/networking/v2/vpcs"
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
	client, err := config.NetworkingV2Client(ctx, util.GetRegion(d, config))

	if err != nil {
		return diag.Errorf("Error creating VNPAY Cloud Peering Connection client: %s", err)
	}

	createOpts := peeringconnectionrequests.CreateOpts{
		PeerVPCId:   d.Get("peer_vpc_id").(string),
		PeerOrgId:   d.Get("peer_org_id").(string),
		VPCId:       d.Get("vpc_id").(string),
		Description: d.Get("description").(string),
	}

	tflog.Debug(ctx, "vnpaycloud_peering_connection request options", map[string]interface{}{"create_opts": createOpts})

	peeringConnectionRequest, err := peeringconnectionrequests.Create(ctx, client, createOpts).Extract()

	if err != nil {
		return diag.Errorf("Error creating vnpaycloud_peering_connection request: %s", err)
	}

	d.SetId(peeringConnectionRequest.PeerId)
	d.Set("request_id", peeringConnectionRequest.ID)

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"OS_INITIATING"},
		Target:     []string{"OS_ACTIVE", "OS_CREATED"},
		Refresh:    peeringConnectionRequestStateRefreshFunc(ctx, client, peeringConnectionRequest.ID),
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
	client, err := config.NetworkingV2Client(ctx, util.GetRegion(d, config))

	if err != nil {
		return diag.Errorf("Error creating VNPAY Cloud Peering Connection client: %s", err)
	}

	listOtps := peeringconnectionapprovals.ListOpts{
		VPCId:     d.Get("vpc_id").(string),
		PeerVPCId: d.Get("peer_vpc_id").(string),
		PeerOrgId: d.Get("peer_org_id").(string),
		Status:    "PCS_PENDING",
	}

	tflog.Debug(ctx, "vnpaycloud_peering_connection list approvals options", map[string]interface{}{"list_otps": listOtps})

	allPages, err := peeringconnectionapprovals.List(client, listOtps).AllPages(ctx)

	if err != nil {
		return diag.Errorf("Unable to query vnpaycloud_peering_connection: %s", err)
	}

	var allApprovals []peeringconnectionapprovals.PeeringConnectApproval
	err = peeringconnectionapprovals.ExtractPeeringConnectApprovalsInto(allPages, &allApprovals)

	if err != nil {
		return diag.Errorf("Unable to retrieve vnpaycloud_peering_connection approvals: %s", err)
	}

	if len(allApprovals) > 1 {
		return diag.Errorf("Your vnpaycloud_peering_connection query returns multiple approvals")

	}

	if len(allApprovals) == 0 {
		return diag.Errorf("Your vnpaycloud_peering_connection query did not return approval results.")
	}

	updateOtps := peeringconnectionapprovals.UpdateOpts{
		Accept: true,
	}

	tflog.Debug(ctx, "vnpaycloud_peering_connection approval options", map[string]interface{}{"update_otps": updateOtps})

	peeringConnectionApproval, err := peeringconnectionapprovals.Update(ctx, client, allApprovals[0].ID, updateOtps).Extract()

	d.SetId(peeringConnectionApproval.PeerId)
	tflog.Debug(ctx, "vnpaycloud_peering_connection approval id", map[string]interface{}{"id": peeringConnectionApproval.ID})
	stateConf := &retry.StateChangeConf{
		Pending:    []string{"OS_INITIATING"},
		Target:     []string{"OS_ACTIVE", "OS_CREATED"},
		Refresh:    peeringConnectionStateRefreshFunc(ctx, client, peeringConnectionApproval.PeerId),
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
	client, err := config.NetworkingV2Client(ctx, util.GetRegion(d, config))

	if err != nil {
		return diag.Errorf("Error creating VNPAY Cloud Peering Connection client: %s", err)
	}

	peeringConnectionRequest, err := peeringconnectionrequests.Get(ctx, client, d.Get("request_id").(string)).Extract()

	if err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error retrieving peering connection request"))
	}

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

	peeringConnection, err := peeringconnections.Get(ctx, client, d.Id()).Extract()

	if err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error retrieving peering connection"))
	}

	d.Set("peer_status", peeringConnection.PeerStatus)
	d.Set("status", peeringConnection.Status)

	return nil
}

func resourcePeeringConnectionApprovalRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	client, err := config.NetworkingV2Client(ctx, util.GetRegion(d, config))

	if err != nil {
		return diag.Errorf("Error creating VNPAY Cloud Peering Connection client: %s", err)
	}

	peeringConnection, err := peeringconnections.Get(ctx, client, d.Id()).Extract()

	if err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error retrieving peering connection"))
	}
	tflog.Debug(ctx, "Retrieved vnpaycloud_peering_connection "+d.Id(), map[string]interface{}{"peering_connection": peeringConnection})

	d.Set("vpc_id", peeringConnection.VpcId)
	d.Set("peer_vpc_id", peeringConnection.PeerVpcId)
	d.Set("peer_org_id", peeringConnection.PeerOrgId)
	d.Set("description", peeringConnection.Description)
	d.Set("peer_status", peeringConnection.PeerStatus)
	d.Set("status", peeringConnection.Status)

	return nil
}

func resourcePeeringConnectionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	client, err := config.NetworkingV2Client(ctx, util.GetRegion(d, config))

	if err != nil {
		return diag.Errorf("Error creating VNPAY Cloud Peering Connection client: %s", err)
	}

	name := d.Get("name").(string)
	description := d.Get("description").(string)
	updateOpts := vpcs.UpdateOpts{
		Name:        name,
		Description: description,
	}

	_, err = vpcs.Update(ctx, client, d.Id(), updateOpts).Extract()

	if err != nil {
		return diag.Errorf("Error updating vnpaycloud_peering_connection %s: %s", d.Id(), err)
	}

	return resourcePeeringConnectionRead(ctx, d, meta)
}

func resourcePeeringConnectionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	client, err := config.NetworkingV2Client(ctx, util.GetRegion(d, config))

	if err != nil {
		return diag.Errorf("Error creating VNPAY Cloud Peering Connection client: %s", err)
	}

	peeringConnection, err := peeringconnections.Get(ctx, client, d.Id()).Extract()

	if err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error retrieving vnpaycloud_peering_connection"))
	}

	if peeringConnection.Status != "OS_DELETING" {
		if err := peeringconnections.Delete(ctx, client, d.Id()).ExtractErr(); err != nil {
			return diag.FromErr(util.CheckDeleted(d, err, "Error deleting vnpaycloud_peering_connection"))
		}
	}

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"OS_DELETING", "OS_ACTIVE", "OS_CREATED"},
		Target:     []string{"OS_DELETED"},
		Refresh:    peeringConnectionStateRefreshFunc(ctx, client, peeringConnection.ID),
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
