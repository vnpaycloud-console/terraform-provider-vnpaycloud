package l7policy

import (
	"context"
	"fmt"
	"regexp"
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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceL7Policy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceL7PolicyCreate,
		ReadContext:   resourceL7PolicyRead,
		UpdateContext: resourceL7PolicyUpdate,
		DeleteContext: resourceL7PolicyDelete,
		CustomizeDiff: func(_ context.Context, d *schema.ResourceDiff, _ interface{}) error {
			action := d.Get("action").(string)
			redirectPool := d.Get("redirect_pool_id").(string)
			redirectURL := d.Get("redirect_url").(string)
			redirectPoolKnown := d.NewValueKnown("redirect_pool_id")
			redirectURLKnown := d.NewValueKnown("redirect_url")
			switch action {
			case "REDIRECT_TO_URL":
				if redirectURLKnown && redirectURL == "" {
					return fmt.Errorf("redirect_url is required when action=%s", action)
				}
				if redirectPoolKnown && redirectPool != "" {
					return fmt.Errorf("redirect_pool_id must be empty when action=%s", action)
				}
			case "REDIRECT_TO_POOL":
				if redirectPoolKnown && redirectPool == "" {
					return fmt.Errorf("redirect_pool_id is required when action=REDIRECT_TO_POOL")
				}
				if redirectURLKnown && redirectURL != "" {
					return fmt.Errorf("redirect_url must be empty when action=REDIRECT_TO_POOL")
				}
			case "REJECT":
				if (redirectPoolKnown && redirectPool != "") || (redirectURLKnown && redirectURL != "") {
					return fmt.Errorf("redirect_pool_id and redirect_url must be empty when action=REJECT")
				}
			}
			return nil
		},
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				cfg := meta.(*config.Config)
				resp := &dto.L7PolicyResponse{}
				if _, err := cfg.Client.Get(ctx, client.ApiPath.L7PolicyWithID(cfg.ProjectID, d.Id()), resp, nil); err != nil {
					return nil, fmt.Errorf("vnpaycloud_lb_l7policy %q not found: %w", d.Id(), err)
				}
				return []*schema.ResourceData{d}, nil
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(3, 250),
					validation.StringMatch(regexp.MustCompile(`^([^ ].*[^ ])?$`), "name must not start or end with whitespace"),
				),
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 255),
			},
			"listener_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"action": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice(
					[]string{"REJECT", "REDIRECT_TO_URL", "REDIRECT_TO_POOL"}, false,
				),
			},
			"position": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntAtLeast(1),
			},
			"redirect_pool_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"redirect_url": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringMatch(
					regexp.MustCompile(`^(https?://|/).*`),
					"redirect_url must start with http://, https:// or /",
				),
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceL7PolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	createOpts := dto.CreateL7PolicyRequest{
		Name:           d.Get("name").(string),
		ListenerID:     d.Get("listener_id").(string),
		Action:         d.Get("action").(string),
		Position:       d.Get("position").(int),
		Description:    d.Get("description").(string),
		RedirectPoolID: d.Get("redirect_pool_id").(string),
		RedirectURL:    d.Get("redirect_url").(string),
	}

	tflog.Debug(ctx, "vnpaycloud_lb_l7policy create options", map[string]interface{}{"create_opts": createOpts})

	createResp := &dto.L7PolicyResponse{}
	_, err := cfg.Client.Post(ctx, client.ApiPath.L7Policies(cfg.ProjectID), createOpts, createResp, nil)
	if err != nil {
		return diag.Errorf("Error creating vnpaycloud_lb_l7policy: %s", err)
	}

	d.SetId(createResp.L7Policy.ID)

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"initiating", "creating", "pending_create"},
		Target:     []string{"active", "created"},
		Refresh:    l7PolicyStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, createResp.L7Policy.ID),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_lb_l7policy %s to become ready: %s", createResp.L7Policy.ID, err)
	}

	return resourceL7PolicyRead(ctx, d, meta)
}

func resourceL7PolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	resp := &dto.L7PolicyResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.L7PolicyWithID(cfg.ProjectID, d.Id()), resp, nil)
	if err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error retrieving vnpaycloud_lb_l7policy"))
	}

	tflog.Debug(ctx, "Retrieved vnpaycloud_lb_l7policy "+d.Id(), map[string]interface{}{"l7policy": resp.L7Policy})

	d.Set("name", resp.L7Policy.Name)
	d.Set("description", resp.L7Policy.Description)
	d.Set("listener_id", resp.L7Policy.ListenerID)
	d.Set("action", resp.L7Policy.Action)
	d.Set("position", resp.L7Policy.Position)
	d.Set("redirect_pool_id", resp.L7Policy.RedirectPoolID)
	d.Set("redirect_url", resp.L7Policy.RedirectURL)
	d.Set("status", resp.L7Policy.Status)

	return nil
}

func resourceL7PolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if d.HasChanges("name", "description", "action", "position", "redirect_pool_id", "redirect_url") {
		updateOpts := dto.UpdateL7PolicyRequest{
			Name:           d.Get("name").(string),
			Action:         d.Get("action").(string),
			Position:       d.Get("position").(int),
			Description:    d.Get("description").(string),
			RedirectPoolID: d.Get("redirect_pool_id").(string),
			RedirectURL:    d.Get("redirect_url").(string),
		}

		tflog.Debug(ctx, "vnpaycloud_lb_l7policy update options", map[string]interface{}{"update_opts": updateOpts})

		err := util.RetryLBPendingPut(ctx, d.Timeout(schema.TimeoutUpdate), func() error {
			_, putErr := cfg.Client.Put(ctx, client.ApiPath.L7PolicyWithID(cfg.ProjectID, d.Id()), updateOpts, nil, nil)
			return putErr
		})
		if err != nil {
			return diag.Errorf("Error updating vnpaycloud_lb_l7policy %s: %s", d.Id(), err)
		}

		stateConf := &retry.StateChangeConf{
			Pending:    []string{"pending_update", "creating"},
			Target:     []string{"active", "created"},
			Refresh:    l7PolicyStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, d.Id()),
			Timeout:    d.Timeout(schema.TimeoutUpdate),
			Delay:      3 * time.Second,
			MinTimeout: 3 * time.Second,
		}
		if _, err := stateConf.WaitForStateContext(ctx); err != nil {
			return diag.Errorf("Error waiting for vnpaycloud_lb_l7policy %s to converge after update: %s", d.Id(), err)
		}
	}

	return resourceL7PolicyRead(ctx, d, meta)
}

func resourceL7PolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	deleteErr := retry.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *retry.RetryError {
		_, err := cfg.Client.Delete(ctx, client.ApiPath.L7PolicyWithID(cfg.ProjectID, d.Id()), nil)
		if err != nil && strings.Contains(err.Error(), "not active") {
			return retry.RetryableError(err)
		}
		if err != nil {
			return retry.NonRetryableError(err)
		}
		return nil
	})
	if deleteErr != nil {
		return diag.FromErr(util.CheckDeleted(d, deleteErr, "Error deleting vnpaycloud_lb_l7policy"))
	}

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"deleting", "active", "created", "pending_delete"},
		Target:     []string{"deleted"},
		Refresh:    l7PolicyStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_lb_l7policy %s to delete: %s", d.Id(), err)
	}

	return nil
}
