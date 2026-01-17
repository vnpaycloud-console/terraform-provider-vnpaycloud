package member

import (
	"context"
	"fmt"
	"log"
	"strings"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/shared"
	"terraform-provider-vnpaycloud/vnpaycloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceMember() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMemberCreate,
		ReadContext:   resourceMemberRead,
		UpdateContext: resourceMemberUpdate,
		DeleteContext: resourceMemberDelete,
		Importer: &schema.ResourceImporter{
			State: resourceMemberImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
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
				Optional: true,
			},

			"tenant_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"address": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"protocol_port": {
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntBetween(1, 65535),
			},

			"weight": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(0, 256),
			},

			"subnet_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"admin_state_up": {
				Type:     schema.TypeBool,
				Default:  true,
				Optional: true,
			},

			"pool_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"backup": {
				Type:     schema.TypeBool,
				Optional: true,
			},

			"monitor_address": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  nil,
				ForceNew: false,
			},

			"monitor_port": {
				Type:         schema.TypeInt,
				Default:      nil,
				Optional:     true,
				ForceNew:     false,
				ValidateFunc: validation.IntBetween(1, 65535),
			},

			"tags": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
		},
	}
}

func resourceMemberCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	tfClient, err := client.NewClient(ctx, config.ConsoleClientConfig)
	if err != nil {
		return diag.Errorf("Error creating VNPAY Cloud Terrform Client: %s", err)
	}

	//adminStateUp := d.Get("admin_state_up").(bool)

	createOpts := dto.CreateMemberOpts{
		Name: d.Get("name").(string),
		//ProjectID:    d.Get("tenant_id").(string),
		Address:      d.Get("address").(string),
		ProtocolPort: d.Get("protocol_port").(int),
		//AdminStateUp: &adminStateUp,
	}

	// Must omit if not set
	if v, ok := d.GetOk("subnet_id"); ok {
		createOpts.SubnetID = v.(string)
	}

	// Set the weight only if it's defined in the configuration.
	// This prevents all members from being created with a default weight of 0.
	if v, ok := util.GetOkExists(d, "weight"); ok {
		weight := v.(int)
		createOpts.Weight = weight
	}

	//if v, ok := d.GetOk("monitor_address"); ok {
	//	createOpts.MonitorAddress = v.(string)
	//}
	//
	//if v, ok := d.GetOk("monitor_port"); ok {
	//	monitorPort := v.(int)
	//	createOpts.MonitorPort = &monitorPort
	//}

	// Only set backup if it is defined by user as it requires
	// version 2.1 or later
	//if v, ok := d.GetOk("backup"); ok {
	//	backup := v.(bool)
	//	createOpts.Backup = &backup
	//}

	//if v, ok := d.GetOk("tags"); ok {
	//	tags := v.(*schema.Set).List()
	//	createOpts.Tags = util.ExpandToStringSlice(tags)
	//}

	log.Printf("[DEBUG] Create Options: %#v", createOpts)

	// Get a clean copy of the parent pool.
	poolID := d.Get("pool_id").(string)

	poolResp := &dto.GetPoolResponse{}
	_, err = tfClient.Get(ctx, client.ApiPath.LbaasPoolWithId(poolID), poolResp, &client.RequestOpts{})
	if err != nil {
		return diag.Errorf("Unable to retrieve parent pool %s: %s", poolID, err)
	}
	parentPool := &poolResp.Pool

	// Wait for parent pool to become active before continuing
	timeout := d.Timeout(schema.TimeoutCreate)
	err = shared.WaitForLBPool(ctx, tfClient, parentPool, "ACTIVE", shared.GetLbPendingStatuses(), timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Attempting to create member")
	//var member *pools.Member
	memberResp := &dto.CreateMemberResponse{}
	err = retry.RetryContext(ctx, timeout, func() *retry.RetryError {
		_, err = tfClient.Post(ctx, client.ApiPath.LbaasPoolMember(poolID), &dto.CreateMemberRequest{Member: createOpts}, memberResp, nil)

		if err != nil {
			return util.CheckForRetryableError(err)
		}
		return nil
	})

	if err != nil {
		return diag.Errorf("Error creating member: %s", err)
	}
	member := &memberResp.Member

	// Wait for member to become active before continuing
	err = shared.WaitForLBMember(ctx, tfClient, parentPool, member, "ACTIVE", shared.GetLbPendingStatuses(), timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(member.ID)

	return resourceMemberRead(ctx, d, meta)
}

func resourceMemberRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	tfClient, err := client.NewClient(ctx, config.ConsoleClientConfig)
	if err != nil {
		return diag.Errorf("Error creating VNPAY Cloud Terrform Client: %s", err)
	}

	poolID := d.Get("pool_id").(string)

	memberResp := &dto.GetMemberResponse{}
	_, err = tfClient.Get(ctx, client.ApiPath.LbaasPoolMemberWithId(poolID, d.Id()), memberResp, &client.RequestOpts{})
	if err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "member"))
	}
	member := &memberResp.Member

	log.Printf("[DEBUG] Retrieved member %s: %#v", d.Id(), member)

	d.Set("name", member.Name)
	d.Set("weight", member.Weight)
	d.Set("admin_state_up", member.AdminStateUp)
	d.Set("tenant_id", member.ProjectID)
	d.Set("subnet_id", member.SubnetID)
	d.Set("address", member.Address)
	d.Set("protocol_port", member.ProtocolPort)
	d.Set("region", util.GetRegion(d, config))
	d.Set("monitor_address", member.MonitorAddress)
	d.Set("monitor_port", member.MonitorPort)
	d.Set("backup", member.Backup)
	d.Set("tags", member.Tags)

	return nil
}

func resourceMemberUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceMemberRead(ctx, d, meta)
}

func resourceMemberDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	tfClient, err := client.NewClient(ctx, config.ConsoleClientConfig)
	if err != nil {
		return diag.Errorf("Error creating VNPAY Cloud Terrform Client: %s", err)
	}

	// Get a clean copy of the parent pool.
	poolID := d.Get("pool_id").(string)
	poolResp := &dto.GetPoolResponse{}
	_, err = tfClient.Get(ctx, client.ApiPath.LbaasPoolWithId(poolID), poolResp, &client.RequestOpts{})
	if err != nil {
		return diag.Errorf("Unable to retrieve parent pool (%s) for the member: %s", poolID, err)
	}
	parentPool := &poolResp.Pool

	// Get a clean copy of the member.
	memberResp := &dto.GetMemberResponse{}
	_, err = tfClient.Get(ctx, client.ApiPath.LbaasPoolMemberWithId(poolID, d.Id()), memberResp, &client.RequestOpts{})
	if err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Unable to retrieve member"))
	}
	member := &memberResp.Member

	// Wait for parent pool to become active before continuing.
	timeout := d.Timeout(schema.TimeoutDelete)
	err = shared.WaitForLBPool(ctx, tfClient, parentPool, "ACTIVE", shared.GetLbPendingStatuses(), timeout)
	if err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error waiting for the members pool status"))
	}

	log.Printf("[DEBUG] Attempting to delete member %s", d.Id())
	err = retry.RetryContext(ctx, timeout, func() *retry.RetryError {
		_, err := tfClient.Delete(ctx, client.ApiPath.LbaasPoolMemberWithId(poolID, d.Id()), nil)
		if err != nil {
			return util.CheckForRetryableError(err)
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error deleting member"))
	}

	// Wait for the member to become DELETED.
	err = shared.WaitForLBMember(ctx, tfClient, parentPool, member, "DELETED", shared.GetLbPendingDeleteStatuses(), timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceMemberImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.SplitN(d.Id(), "/", 2)
	if len(parts) != 2 {
		err := fmt.Errorf("Invalid format specified for Member. Format must be <pool id>/<member id>")
		return nil, err
	}

	poolID := parts[0]
	memberID := parts[1]

	d.SetId(memberID)
	d.Set("pool_id", poolID)

	return []*schema.ResourceData{d}, nil
}
