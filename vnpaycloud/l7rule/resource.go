package l7rule

import (
	"context"
	"fmt"
	"strings"
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

func ResourceL7Rule() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceL7RuleCreate,
		ReadContext:   resourceL7RuleRead,
		UpdateContext: resourceL7RuleUpdate,
		DeleteContext: resourceL7RuleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				// Import ID format: <l7policy_id>/<rule_id>
				parts := strings.Split(d.Id(), "/")
				if len(parts) != 2 {
					return nil, fmt.Errorf("import id must be <l7policy_id>/<rule_id>, got: %s", d.Id())
				}
				cfg := meta.(*config.Config)
				resp := &dto.L7RuleResponse{}
				if _, err := cfg.Client.Get(ctx, client.ApiPath.L7RuleWithID(cfg.ProjectID, parts[0], parts[1]), resp, nil); err != nil {
					return nil, fmt.Errorf("vnpaycloud_lb_l7rule %q not found: %w", d.Id(), err)
				}
				d.SetId(parts[1])
				d.Set("l7policy_id", parts[0])
				return []*schema.ResourceData{d}, nil
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"l7policy_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"rule_type": {
				Type:     schema.TypeString,
				Required: true,
			},
			"compare_type": {
				Type:     schema.TypeString,
				Required: true,
			},
			"value": {
				Type:     schema.TypeString,
				Required: true,
			},
			"key": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Required when rule_type is COOKIE; must be empty otherwise.",
			},
			"invert": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceL7RuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	l7policyID := d.Get("l7policy_id").(string)

	createOpts := dto.CreateL7RuleRequest{
		RuleType:    d.Get("rule_type").(string),
		CompareType: d.Get("compare_type").(string),
		Value:       d.Get("value").(string),
		Key:         d.Get("key").(string),
		Invert:      d.Get("invert").(bool),
	}

	tflog.Debug(ctx, "vnpaycloud_lb_l7rule create options", map[string]interface{}{"create_opts": createOpts})

	createResp := &dto.L7RuleResponse{}
	_, err := cfg.Client.Post(ctx, client.ApiPath.L7Rules(cfg.ProjectID, l7policyID), createOpts, createResp, nil)
	if err != nil {
		return diag.Errorf("Error creating vnpaycloud_lb_l7rule: %s", err)
	}

	d.SetId(createResp.L7Rule.ID)

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"initiating", "creating", "pending_create"},
		Target:     []string{"active", "created"},
		Refresh:    l7RuleStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, l7policyID, createResp.L7Rule.ID),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_lb_l7rule %s to become ready: %s", createResp.L7Rule.ID, err)
	}

	return resourceL7RuleRead(ctx, d, meta)
}

func resourceL7RuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	l7policyID := d.Get("l7policy_id").(string)

	resp := &dto.L7RuleResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.L7RuleWithID(cfg.ProjectID, l7policyID, d.Id()), resp, nil)
	if err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error retrieving vnpaycloud_lb_l7rule"))
	}

	tflog.Debug(ctx, "Retrieved vnpaycloud_lb_l7rule "+d.Id(), map[string]interface{}{"l7rule": resp.L7Rule})

	d.Set("l7policy_id", resp.L7Rule.L7PolicyID)
	d.Set("rule_type", resp.L7Rule.RuleType)
	d.Set("compare_type", resp.L7Rule.CompareType)
	d.Set("value", resp.L7Rule.Value)
	d.Set("key", resp.L7Rule.Key)
	d.Set("invert", resp.L7Rule.Invert)
	d.Set("status", resp.L7Rule.Status)

	return nil
}

func resourceL7RuleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	l7policyID := d.Get("l7policy_id").(string)

	if d.HasChanges("rule_type", "compare_type", "value", "key", "invert") {
		updateOpts := dto.UpdateL7RuleRequest{
			RuleType:    d.Get("rule_type").(string),
			CompareType: d.Get("compare_type").(string),
			Value:       d.Get("value").(string),
			Key:         d.Get("key").(string),
			Invert:      d.Get("invert").(bool),
		}

		tflog.Debug(ctx, "vnpaycloud_lb_l7rule update options", map[string]interface{}{"update_opts": updateOpts})

		err := util.RetryLBPendingPut(ctx, d.Timeout(schema.TimeoutUpdate), func() error {
			_, putErr := cfg.Client.Put(ctx, client.ApiPath.L7RuleWithID(cfg.ProjectID, l7policyID, d.Id()), updateOpts, nil, nil)
			return putErr
		})
		if err != nil {
			return diag.Errorf("Error updating vnpaycloud_lb_l7rule %s: %s", d.Id(), err)
		}

		stateConf := &retry.StateChangeConf{
			Pending:    []string{"pending_update", "creating"},
			Target:     []string{"active", "created"},
			Refresh:    l7RuleStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, l7policyID, d.Id()),
			Timeout:    d.Timeout(schema.TimeoutUpdate),
			Delay:      3 * time.Second,
			MinTimeout: 3 * time.Second,
		}
		if _, err := stateConf.WaitForStateContext(ctx); err != nil {
			return diag.Errorf("Error waiting for vnpaycloud_lb_l7rule %s to converge: %s", d.Id(), err)
		}
	}

	return resourceL7RuleRead(ctx, d, meta)
}

func resourceL7RuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	l7policyID := d.Get("l7policy_id").(string)

	deleteErr := retry.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *retry.RetryError {
		_, err := cfg.Client.Delete(ctx, client.ApiPath.L7RuleWithID(cfg.ProjectID, l7policyID, d.Id()), nil)
		if err != nil && strings.Contains(err.Error(), "not active") {
			return retry.RetryableError(err)
		}
		if err != nil {
			return retry.NonRetryableError(err)
		}
		return nil
	})
	if deleteErr != nil {
		return diag.FromErr(util.CheckDeleted(d, deleteErr, "Error deleting vnpaycloud_lb_l7rule"))
	}

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"deleting", "active", "created", "pending_delete"},
		Target:     []string{"deleted"},
		Refresh:    l7RuleStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, l7policyID, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_lb_l7rule %s to delete: %s", d.Id(), err)
	}

	return nil
}
