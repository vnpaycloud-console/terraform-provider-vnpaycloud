package members

import (
	"context"
	"fmt"
	"log"
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

func ResourceMembers() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMembersCreate,
		ReadContext:   resourceMembersRead,
		UpdateContext: resourceMembersUpdate,
		DeleteContext: resourceMembersDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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

			"pool_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"member": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"name": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"address": {
							Type:     schema.TypeString,
							Required: true,
						},

						"protocol_port": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(1, 65535),
						},

						"weight": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      1,
							ValidateFunc: validation.IntBetween(0, 256),
						},

						"monitor_port": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(1, 65535),
						},

						"monitor_address": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"subnet_id": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"backup": {
							Type:     schema.TypeBool,
							Optional: true,
						},

						"admin_state_up": {
							Type:     schema.TypeBool,
							Default:  true,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func resourceMembersCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	lbClient, err := client.NewClient(ctx, config.ConsoleClientConfig)
	if err != nil {
		return diag.Errorf("Error creating VNPAYCLOUD networking client: %s", err)
	}

	createOpts := shared.ExpandLBMembers(d.Get("member").(*schema.Set), lbClient)
	log.Printf("[DEBUG] Create Options: %#v", createOpts)

	// Get a clean copy of the parent pool.
	poolID := d.Get("pool_id").(string)
	parentPoolResp := dto.GetPoolResponse{}
	_, err = lbClient.Get(ctx, client.ApiPath.LbaasPoolWithId(poolID), &parentPoolResp, nil)
	if err != nil {
		return diag.Errorf("Unable to retrieve parent pool %s: %s", poolID, err)
	}

	parentPool := parentPoolResp.Pool

	timeout := d.Timeout(schema.TimeoutCreate)
	err = shared.WaitForLBPool(ctx, lbClient, &parentPool, "ACTIVE", shared.GetLbPendingStatuses(), timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Attempting to create members")
	createRequest := dto.BatchUpdateMemberRequest{
		Members: createOpts,
	}
	err = retry.RetryContext(ctx, timeout, func() *retry.RetryError {
		_, err = lbClient.Put(ctx, client.ApiPath.LbaasPoolMember(poolID), createRequest, nil, nil)
		if err != nil {
			return util.CheckForRetryableError(err)
		}
		return nil
	})

	if err != nil {
		return diag.Errorf("Error creating members: %s", err)
	}

	// Wait for parent pool to become active before continuing
	err = shared.WaitForLBPool(ctx, lbClient, &parentPool, "ACTIVE", shared.GetLbPendingStatuses(), timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(poolID)

	return resourceMembersRead(ctx, d, meta)
}

func resourceMembersRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	lbClient, err := client.NewClient(ctx, config.ConsoleClientConfig)
	if err != nil {
		return diag.Errorf("Error creating VNPAYCLOUD networking client: %s", err)
	}

	membersResp := dto.ListMembersResponse{}
	_, err = lbClient.Get(ctx, client.ApiPath.LbaasPoolMember(d.Id()), &membersResp, nil)
	if err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error getting vnpaycloud_lb_members"))
	}

	members := membersResp.Members

	log.Printf("[DEBUG] Retrieved members for the %s pool: %#v", d.Id(), members)

	d.Set("pool_id", d.Id())
	d.Set("member", shared.FlattenLBMembers(members))
	d.Set("region", util.GetRegion(d, config))

	return nil
}

func resourceMembersUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceMembersRead(ctx, d, meta)
}

func resourceMembersDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	lbClient, err := client.NewClient(ctx, config.ConsoleClientConfig)
	if err != nil {
		return diag.Errorf("Error creating VNPAYCLOUD networking client: %s", err)
	}

	// Get a clean copy of the parent pool.
	parentPoolResp := dto.GetPoolResponse{}
	_, err = lbClient.Get(ctx, client.ApiPath.LbaasPoolWithId(d.Id()), &parentPoolResp, nil)
	if err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, fmt.Sprintf("Unable to retrieve parent pool (%s) for the member", d.Id())))
	}

	parentPool := parentPoolResp.Pool

	// Wait for parent pool to become active before continuing.
	timeout := d.Timeout(schema.TimeoutDelete)
	err = shared.WaitForLBPool(ctx, lbClient, &parentPool, "ACTIVE", shared.GetLbPendingStatuses(), timeout)
	if err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error waiting for the members' pool status"))
	}

	log.Printf("[DEBUG] Attempting to delete %s pool members", d.Id())
	err = retry.RetryContext(ctx, timeout, func() *retry.RetryError {
		deleteRequest := dto.BatchUpdateMemberRequest{
			Members: []dto.BatchUpdateMemberOpts{},
		}
		_, err = lbClient.Put(ctx, client.ApiPath.LbaasPoolMember(d.Id()), deleteRequest, nil, nil)
		if err != nil {
			return util.CheckForRetryableError(err)
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error deleting members"))
	}

	// Wait for parent pool to become active before continuing.
	err = shared.WaitForLBPool(ctx, lbClient, &parentPool, "ACTIVE", shared.GetLbPendingStatuses(), timeout)
	if err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error waiting for the members' pool status"))
	}

	return nil
}
