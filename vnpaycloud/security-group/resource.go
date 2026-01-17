package securityGroup

import (
	"context"
	"log"
	"time"

	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceNetworkingSecGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNetworkingSecGroupCreate,
		ReadContext:   resourceNetworkingSecGroupRead,
		UpdateContext: resourceNetworkingSecGroupUpdate,
		DeleteContext: resourceNetworkingSecGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceNetworkingSecGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)

	tfClient, err := client.NewClient(ctx, config.ConsoleClientConfig)

	if err != nil {
		return diag.Errorf("Error creating VNPAY Cloud networking client: %s", err)
	}

	createOpts := dto.CreateSecurityGroupOpts{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
	}

	log.Printf("[DEBUG] vnpaycloud_networking_secgroup create options: %#v", createOpts)

	createResp := &dto.CreateSecurityGroupResponse{}
	_, err = tfClient.Post(ctx, client.ApiPath.SecurityGroup, dto.CreateSecurityGroupRequest{SecurityGroup: createOpts}, createResp, nil)
	if err != nil {
		return diag.Errorf("Error creating vnpaycloud_networking_secgroup: %s", err)
	}

	d.SetId(createResp.SecurityGroup.ID)

	return resourceNetworkingSecGroupRead(ctx, d, meta)
}

func resourceNetworkingSecGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	tfClient, err := client.NewClient(ctx, config.ConsoleClientConfig)
	if err != nil {
		return diag.Errorf("Error creating VNPAY Cloud networking client: %s", err)
	}

	securityGroupResp := &dto.GetSecurityGroupResponse{}
	_, err = tfClient.Get(ctx, client.ApiPath.SecurityGroupWithId(d.Id()), securityGroupResp, nil)
	if err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error retrieving vnpaycloud_networking_secgroup"))
	}

	log.Printf("[DEBUG] Retrieved vnpaycloud_networking_secgroup %s: %#v", d.Id(), securityGroupResp.SecurityGroup)

	d.Set("description", securityGroupResp.SecurityGroup.Description)
	d.Set("tenant_id", securityGroupResp.SecurityGroup.TenantID)
	d.Set("name", securityGroupResp.SecurityGroup.Name)
	d.Set("region", util.GetRegion(d, config))

	return nil
}

func resourceNetworkingSecGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceNetworkingSecGroupRead(ctx, d, meta)
}

func resourceNetworkingSecGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	tfClient, err := client.NewClient(ctx, config.ConsoleClientConfig)
	if err != nil {
		return diag.Errorf("Error creating VNPAY Cloud networking client: %s", err)
	}

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"ACTIVE"},
		Target:     []string{"DELETED"},
		Refresh:    networkingSecgroupStateRefreshFuncDelete(ctx, tfClient, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error deleting vnpaycloud_networking_secgroup: %s", err)
	}

	return diag.FromErr(err)
}
