package listener

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

func ResourceListener() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceListenerCreate,
		ReadContext:   resourceListenerRead,
		UpdateContext: resourceListenerUpdate,
		DeleteContext: resourceListenerDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				cfg := meta.(*config.Config)
				resp := &dto.ListenerResponse{}
				if _, err := cfg.Client.Get(ctx, client.ApiPath.ListenerWithID(cfg.ProjectID, d.Id()), resp, nil); err != nil {
					return nil, fmt.Errorf("vnpaycloud_lb_listener %q not found: %w", d.Id(), err)
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
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"load_balancer_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"protocol": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"protocol_port": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"default_pool_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				Description: "ID of the default pool. The platform does not support detaching a default_pool " +
					"once attached — removing this field from config will NOT clear the attachment (drift is " +
					"suppressed). To change: swap to another pool's ID. To remove entirely: destroy and recreate the listener.",
				DiffSuppressFunc: func(_, old, new string, _ *schema.ResourceData) bool {
					return new == "" && old != ""
				},
			},
			"insert_headers": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"allowed_cidrs": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"connection_limit": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"timeout_client_data": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"timeout_member_connect": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"timeout_member_data": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"certificate_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Server certificate ID. Required for protocol `HTTPS`.",
			},
			"certificate_authority_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Client CA certificate ID for mutual TLS. Only valid for `HTTPS`.",
			},
			"sni_certificate_ids": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "SNI certificate IDs. Only valid for `HTTPS`.",
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceListenerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	createOpts := dto.CreateListenerRequest{
		Name:                 d.Get("name").(string),
		Description:          d.Get("description").(string),
		LoadBalancerID:       d.Get("load_balancer_id").(string),
		Protocol:             d.Get("protocol").(string),
		ProtocolPort:         d.Get("protocol_port").(int),
		ConnectionLimit:      d.Get("connection_limit").(int),
		TimeoutClientData:    d.Get("timeout_client_data").(int),
		TimeoutMemberConnect: d.Get("timeout_member_connect").(int),
		TimeoutMemberData:    d.Get("timeout_member_data").(int),
	}

	if v, ok := d.GetOk("default_pool_id"); ok {
		createOpts.DefaultPoolID = v.(string)
	}

	if v, ok := d.GetOk("insert_headers"); ok {
		for _, h := range v.([]interface{}) {
			createOpts.InsertHeaders = append(createOpts.InsertHeaders, h.(string))
		}
	}

	if v, ok := d.GetOk("allowed_cidrs"); ok {
		for _, c := range v.([]interface{}) {
			createOpts.AllowedCidrs = append(createOpts.AllowedCidrs, c.(string))
		}
	}

	if v, ok := d.GetOk("certificate_id"); ok {
		createOpts.CertificateID = v.(string)
	}
	if v, ok := d.GetOk("certificate_authority_id"); ok {
		createOpts.CertificateAuthorityID = v.(string)
	}
	if v, ok := d.GetOk("sni_certificate_ids"); ok {
		for _, c := range v.([]interface{}) {
			createOpts.SniCertificateIDs = append(createOpts.SniCertificateIDs, c.(string))
		}
	}

	tflog.Debug(ctx, "vnpaycloud_lb_listener create options", map[string]interface{}{"create_opts": createOpts})

	createResp := &dto.ListenerResponse{}
	_, err := cfg.Client.Post(ctx, client.ApiPath.Listeners(cfg.ProjectID), createOpts, createResp, nil)
	if err != nil {
		return diag.Errorf("Error creating vnpaycloud_lb_listener: %s", err)
	}

	d.SetId(createResp.Listener.ID)

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"initiating", "creating", "pending_create"},
		Target:     []string{"active", "created"},
		Refresh:    listenerStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, createResp.Listener.ID),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_lb_listener %s to become ready: %s", createResp.Listener.ID, err)
	}

	return resourceListenerRead(ctx, d, meta)
}

func resourceListenerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	resp := &dto.ListenerResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.ListenerWithID(cfg.ProjectID, d.Id()), resp, nil)
	if err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error retrieving vnpaycloud_lb_listener"))
	}

	tflog.Debug(ctx, "Retrieved vnpaycloud_lb_listener "+d.Id(), map[string]interface{}{"listener": resp.Listener})

	d.Set("name", resp.Listener.Name)
	d.Set("description", resp.Listener.Description)
	d.Set("load_balancer_id", resp.Listener.LoadBalancerID)
	d.Set("protocol", resp.Listener.Protocol)
	d.Set("protocol_port", resp.Listener.ProtocolPort)
	d.Set("default_pool_id", resp.Listener.DefaultPoolID)
	d.Set("insert_headers", resp.Listener.InsertHeaders)
	d.Set("allowed_cidrs", resp.Listener.AllowedCidrs)
	d.Set("connection_limit", resp.Listener.ConnectionLimit)
	d.Set("timeout_client_data", resp.Listener.TimeoutClientData)
	d.Set("timeout_member_connect", resp.Listener.TimeoutMemberConnect)
	d.Set("timeout_member_data", resp.Listener.TimeoutMemberData)
	d.Set("certificate_id", resp.Listener.CertificateID)
	d.Set("certificate_authority_id", resp.Listener.CertificateAuthorityID)
	d.Set("sni_certificate_ids", resp.Listener.SniCertificateIDs)
	d.Set("status", resp.Listener.Status)
	d.Set("created_at", resp.Listener.CreatedAt)

	return nil
}

func resourceListenerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if d.HasChanges("name", "description", "default_pool_id", "insert_headers", "allowed_cidrs", "connection_limit", "timeout_client_data", "timeout_member_connect", "timeout_member_data", "certificate_id", "certificate_authority_id", "sni_certificate_ids") {
		waitBefore := &retry.StateChangeConf{
			Pending:    []string{"initiating", "creating", "pending_create", "pending_update"},
			Target:     []string{"active", "created"},
			Refresh:    listenerStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, d.Id()),
			Timeout:    d.Timeout(schema.TimeoutUpdate),
			Delay:      3 * time.Second,
			MinTimeout: 3 * time.Second,
		}
		if _, err := waitBefore.WaitForStateContext(ctx); err != nil {
			return diag.Errorf("Error waiting for vnpaycloud_lb_listener %s to become ready before update: %s", d.Id(), err)
		}

		updateOpts := dto.UpdateListenerRequest{
			Name:                   d.Get("name").(string),
			Description:            d.Get("description").(string),
			DefaultPoolID:          d.Get("default_pool_id").(string),
			CertificateID:          d.Get("certificate_id").(string),
			CertificateAuthorityID: d.Get("certificate_authority_id").(string),
			ConnectionLimit:        d.Get("connection_limit").(int),
			TimeoutClientData:      d.Get("timeout_client_data").(int),
			TimeoutMemberConnect:   d.Get("timeout_member_connect").(int),
			TimeoutMemberData:      d.Get("timeout_member_data").(int),
		}

		// allowed_cidrs and sni_certificate_ids are set unconditionally (no GetOk
		// guard) so that removing them from the config clears them on the backend.
		for _, c := range d.Get("sni_certificate_ids").([]interface{}) {
			updateOpts.SniCertificateIDs = append(updateOpts.SniCertificateIDs, c.(string))
		}

		if v, ok := d.GetOk("insert_headers"); ok {
			for _, h := range v.([]interface{}) {
				updateOpts.InsertHeaders = append(updateOpts.InsertHeaders, h.(string))
			}
		}

		for _, c := range d.Get("allowed_cidrs").([]interface{}) {
			updateOpts.AllowedCidrs = append(updateOpts.AllowedCidrs, c.(string))
		}

		tflog.Debug(ctx, "vnpaycloud_lb_listener update options", map[string]interface{}{"update_opts": updateOpts})

		err := util.RetryLBPendingPut(ctx, d.Timeout(schema.TimeoutUpdate), func() error {
			_, putErr := cfg.Client.Put(ctx, client.ApiPath.ListenerWithID(cfg.ProjectID, d.Id()), updateOpts, nil, nil)
			return putErr
		})
		if err != nil {
			return diag.Errorf("Error updating vnpaycloud_lb_listener %s: %s", d.Id(), err)
		}

		waitAfter := &retry.StateChangeConf{
			Pending:    []string{"pending_update", "creating"},
			Target:     []string{"active", "created"},
			Refresh:    listenerStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, d.Id()),
			Timeout:    d.Timeout(schema.TimeoutUpdate),
			Delay:      3 * time.Second,
			MinTimeout: 3 * time.Second,
		}
		if _, err := waitAfter.WaitForStateContext(ctx); err != nil {
			return diag.Errorf("Error waiting for vnpaycloud_lb_listener %s to converge after update: %s", d.Id(), err)
		}
	}

	return resourceListenerRead(ctx, d, meta)
}

func resourceListenerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	deleteErr := retry.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *retry.RetryError {
		_, err := cfg.Client.Delete(ctx, client.ApiPath.ListenerWithID(cfg.ProjectID, d.Id()), nil)
		if err != nil && strings.Contains(err.Error(), "not active") {
			return retry.RetryableError(err)
		}
		if err != nil {
			return retry.NonRetryableError(err)
		}
		return nil
	})
	if deleteErr != nil {
		return diag.FromErr(util.CheckDeleted(d, deleteErr, "Error deleting vnpaycloud_lb_listener"))
	}

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"deleting", "active", "created", "pending_delete"},
		Target:     []string{"deleted"},
		Refresh:    listenerStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_lb_listener %s to delete: %s", d.Id(), err)
	}

	return nil
}
