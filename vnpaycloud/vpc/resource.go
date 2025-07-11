package vpc

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
	"github.com/vnpaycloud-console/gophercloud/v2/openstack/networking/v2/vpcs"
)

func ResourceVpc() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVpcCreate,
		ReadContext:   resourceVpcRead,
		UpdateContext: resourceVpcUpdate,
		DeleteContext: resourceVpcDelete,
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
				ForceNew: false,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
			},
			"cidr_block": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
		},
	}
}

func resourceVpcCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	vpcClient, err := client.NewClient(ctx, config.ConsoleClientConfig)

	if err != nil {
		return diag.Errorf("Error creating VNPAY Cloud VPC client: %s", err)
	}

	createOpts := CreateVpcDto{
		VPC: struct {
			Name        string "json:\"name,omitempty\""
			Description string "json:\"description,omitempty\""
			CIDR        string "json:\"cidr,omitempty\""
		}{
			Name:        d.Get("name").(string),
			Description: d.Get("description").(string),
			CIDR:        d.Get("cidr_block").(string),
		},
	}

	tflog.Debug(ctx, "vnpaycloud_vpc create options", map[string]interface{}{"create_opts": createOpts})

	createVpcResp := &CreateVpcDtoResponse{}
	_, err = vpcClient.Post(ctx, client.ApiPath.VPC, createOpts, createVpcResp, nil)

	if err != nil {
		return diag.Errorf("Error creating vnpaycloud_vpc: %s", err)
	}

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"OS_INITIATING"},
		Target:     []string{"OS_ACTIVE", "OS_CREATED"},
		Refresh:    vpcStateRefreshFunc(ctx, vpcClient, createVpcResp.VPC.ID),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)

	if err != nil {
		return diag.Errorf(
			"Error waiting for vnpaycloud_vpc %s to become ready: %s", createVpcResp.VPC.ID, err)
	}

	d.SetId(createVpcResp.VPC.ID)

	return resourceVpcRead(ctx, d, meta)
}

func resourceVpcRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	vpcClient, err := client.NewClient(ctx, config.ConsoleClientConfig)

	if err != nil {
		return diag.Errorf("Error creating VNPAY Cloud VPC client: %s", err)
	}

	vpcResp := &GetVpcDtoResponse{}
	otps := &client.RequestOpts{}
	_, err = vpcClient.Get(ctx, client.ApiPath.VPCWithId(d.Id()), vpcResp, otps)

	if err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error retrieving vnpaycloud_vpc"))
	}

	tflog.Debug(ctx, "Retrieved vnpaycloud_vpc "+d.Id(), map[string]interface{}{"vpc": vpcResp.VPC})

	d.Set("name", vpcResp.VPC.Name)
	d.Set("description", vpcResp.VPC.Description)
	d.Set("cidr_block", vpcResp.VPC.CIDR)

	return nil
}

func resourceVpcUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	vpcClient, err := config.NetworkingV2Client(ctx, util.GetRegion(d, config))

	if err != nil {
		return diag.Errorf("Error creating VNPAY Cloud VPC client: %s", err)
	}

	name := d.Get("name").(string)
	description := d.Get("description").(string)
	updateOpts := vpcs.UpdateOpts{
		Name:        name,
		Description: description,
	}

	_, err = vpcs.Update(ctx, vpcClient, d.Id(), updateOpts).Extract()

	if err != nil {
		return diag.Errorf("Error updating vnpaycloud_vpc %s: %s", d.Id(), err)
	}

	return resourceVpcRead(ctx, d, meta)
}

func resourceVpcDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	vpcClient, err := client.NewClient(ctx, config.ConsoleClientConfig)

	if err != nil {
		return diag.Errorf("Error creating VNPAY Cloud VPC client: %s", err)
	}

	vpcResp := &GetVpcDtoResponse{}
	_, err = vpcClient.Get(ctx, client.ApiPath.VPCWithId(d.Id()), vpcResp, nil)

	if err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error retrieving vnpaycloud_vpc"))
	}

	vpc := vpcResp.VPC

	if vpc.Status != "OS_DELETING" {
		if _, err := vpcClient.Delete(ctx, client.ApiPath.VPCWithId(d.Id()), nil); err != nil {
			return diag.FromErr(util.CheckDeleted(d, err, "Error deleting vnpaycloud_vpc"))
		}
	}

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"OS_DELETING", "OS_ACTIVE", "OS_CREATED"},
		Target:     []string{"OS_DELETED"},
		Refresh:    vpcStateRefreshFunc(ctx, vpcClient, vpc.ID),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)

	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_vpc %s to Delete:  %s", d.Id(), err)
	}

	return nil
}
