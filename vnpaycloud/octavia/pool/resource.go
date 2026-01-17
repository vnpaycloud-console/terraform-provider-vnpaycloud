package pool

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

func ResourcePool() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePoolCreate,
		ReadContext:   resourcePoolRead,
		UpdateContext: resourcePoolUpdate,
		DeleteContext: resourcePoolDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourcePoolImport,
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

			"tenant_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"protocol": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"TCP", "UDP", "HTTP", "HTTPS", "PROXY", "SCTP", "PROXYV2",
				}, false),
			},

			// One of loadbalancer_id or listener_id must be provided
			"loadbalancer_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"loadbalancer_id", "listener_id"},
			},

			// One of loadbalancer_id or listener_id must be provided
			"listener_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"loadbalancer_id", "listener_id"},
			},

			"lb_method": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"ROUND_ROBIN", "LEAST_CONNECTIONS", "SOURCE_IP", "SOURCE_IP_PORT",
				}, false),
			},

			"persistence": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								"SOURCE_IP", "HTTP_COOKIE", "APP_COOKIE",
							}, false),
						},

						"cookie_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},

			"alpn_protocols": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true, // unsetting this parameter results in a default value
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{
						"http/1.0", "http/1.1", "h2",
					}, false),
				},
			},

			"ca_tls_container_ref": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"crl_container_ref": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"tls_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},

			"tls_ciphers": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true, // unsetting this parameter results in a default value
			},

			"tls_container_ref": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"tls_versions": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true, // unsetting this parameter results in a default value
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{
						"TLSv1", "TLSv1.1", "TLSv1.2", "TLSv1.3",
					}, false),
				},
			},

			"admin_state_up": {
				Type:     schema.TypeBool,
				Default:  true,
				Optional: true,
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

func resourcePoolCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	tfClient, err := client.NewClient(ctx, config.ConsoleClientConfig)
	if err != nil {
		return diag.Errorf("Error creating VNPAY Cloud Terrform Client: %s", err)
	}

	//adminStateUp := d.Get("admin_state_up").(bool)
	lbID := d.Get("loadbalancer_id").(string)
	listenerID := d.Get("listener_id").(string)

	createOpts := dto.CreatePoolOpts{
		//ProjectID:         d.Get("tenant_id").(string),
		Name:           d.Get("name").(string),
		Description:    d.Get("description").(string),
		Protocol:       d.Get("protocol").(string),
		LoadbalancerID: lbID,
		ListenerID:     listenerID,
		LBMethod:       d.Get("lb_method").(string),
		//ALPNProtocols:     util.ExpandToStringSlice(d.Get("alpn_protocols").(*schema.Set).List()),
		//CATLSContainerRef: d.Get("ca_tls_container_ref").(string),
		//CRLContainerRef:   d.Get("crl_container_ref").(string),
		//TLSEnabled:        d.Get("tls_enabled").(bool),
		//TLSCiphers:        d.Get("tls_ciphers").(string),
		//TLSContainerRef:   d.Get("tls_container_ref").(string),
		//AdminStateUp:      &adminStateUp,
	}

	//if v, ok := d.GetOk("tls_versions"); ok {
	//	createOpts.TLSVersions = shared.ExpandLBPoolTLSVersion(v.(*schema.Set).List())
	//}
	//
	//if v, ok := d.GetOk("persistence"); ok {
	//	createOpts.Persistence, err = shared.ExpandLBPoolPersistance(v.([]interface{}))
	//	if err != nil {
	//		return diag.FromErr(err)
	//	}
	//}
	//
	//if v, ok := d.GetOk("tags"); ok {
	//	tags := v.(*schema.Set).List()
	//	createOpts.Tags = util.ExpandToStringSlice(tags)
	//}

	log.Printf("[DEBUG] Create Options: %#v", createOpts)

	timeout := d.Timeout(schema.TimeoutCreate)

	// Wait for Listener or LoadBalancer to become active before continuing
	if listenerID != "" {
		listenerResp := &dto.GetListenerResponse{}
		_, err := tfClient.Get(ctx, client.ApiPath.LbaasListenerWithId(listenerID), listenerResp, &client.RequestOpts{})
		if err != nil {
			return diag.Errorf("Unable to get vnpaycloud_lb_listener %s: %s", listenerID, err)
		}

		waitErr := shared.WaitForLBListener(ctx, tfClient, &listenerResp.Listener, "ACTIVE", shared.GetLbPendingStatuses(), timeout)
		if waitErr != nil {
			return diag.Errorf(
				"Error waiting for vnpaycloud_lb_listener %s to become active: %s", listenerID, err)
		}
	} else {
		waitErr := shared.WaitForLBLoadBalancer(ctx, tfClient, lbID, "ACTIVE", shared.GetLbPendingStatuses(), timeout)
		if waitErr != nil {
			return diag.Errorf(
				"Error waiting for vnpaycloud_lb_loadbalancer %s to become active: %s", lbID, err)
		}
	}

	log.Printf("[DEBUG] Attempting to create pool")
	poolResp := &dto.CreatePoolResponse{}
	err = retry.RetryContext(ctx, timeout, func() *retry.RetryError {
		_, err = tfClient.Post(ctx, client.ApiPath.LbaasPool, &dto.CreatePoolRequest{Pool: createOpts}, poolResp, nil)
		if err != nil {
			return util.CheckForRetryableError(err)
		}
		return nil
	})

	if err != nil {
		return diag.Errorf("Error creating pool: %s", err)
	}
	pool := &poolResp.Pool

	// Pool was successfully created
	// Wait for pool to become active before continuing
	err = shared.WaitForLBPool(ctx, tfClient, pool, "ACTIVE", shared.GetLbPendingStatuses(), timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(pool.ID)

	return resourcePoolRead(ctx, d, meta)
}

func resourcePoolRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	tfClient, err := client.NewClient(ctx, config.ConsoleClientConfig)
	if err != nil {
		return diag.Errorf("Error creating VNPAY Cloud Terrform Client: %s", err)
	}

	poolResp := &dto.GetPoolResponse{}
	_, err = tfClient.Get(ctx, client.ApiPath.LbaasPoolWithId(d.Id()), poolResp, &client.RequestOpts{})
	if err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "pool"))
	}
	pool := poolResp.Pool

	log.Printf("[DEBUG] Retrieved pool %s: %#v", d.Id(), pool)

	d.Set("lb_method", pool.LBMethod)
	d.Set("protocol", pool.Protocol)
	d.Set("description", pool.Description)
	d.Set("tenant_id", pool.ProjectID)
	d.Set("admin_state_up", pool.AdminStateUp)
	d.Set("name", pool.Name)
	d.Set("persistence", shared.FlattenLBPoolPersistence(pool.Persistence))
	d.Set("alpn_protocols", pool.ALPNProtocols)
	d.Set("ca_tls_container_ref", pool.CATLSContainerRef)
	d.Set("crl_container_ref", pool.CRLContainerRef)
	d.Set("tls_enabled", pool.TLSEnabled)
	d.Set("tls_ciphers", pool.TLSCiphers)
	d.Set("tls_container_ref", pool.TLSContainerRef)
	d.Set("tls_versions", pool.TLSVersions)
	d.Set("region", util.GetRegion(d, config))
	d.Set("tags", pool.Tags)

	return nil
}

func resourcePoolUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourcePoolRead(ctx, d, meta)
}

func resourcePoolDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	tfClient, err := client.NewClient(ctx, config.ConsoleClientConfig)
	if err != nil {
		return diag.Errorf("Error creating VNPAY Cloud Terrform Client: %s", err)
	}

	timeout := d.Timeout(schema.TimeoutDelete)

	// Get a clean copy of the pool.
	poolResp := &dto.GetPoolResponse{}
	_, err = tfClient.Get(ctx, client.ApiPath.LbaasPoolWithId(d.Id()), poolResp, &client.RequestOpts{})
	if err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Unable to retrieve pool"))
	}
	pool := &poolResp.Pool

	log.Printf("[DEBUG] Attempting to delete pool %s", d.Id())
	err = retry.RetryContext(ctx, timeout, func() *retry.RetryError {
		_, err = tfClient.Delete(ctx, client.ApiPath.LbaasPoolWithId(d.Id()), nil)
		if err != nil {
			return util.CheckForRetryableError(err)
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error deleting pool"))
	}

	// Wait for Pool to delete
	err = shared.WaitForLBPool(ctx, tfClient, pool, "DELETED", shared.GetLbPendingDeleteStatuses(), timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourcePoolImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	config := meta.(*config.Config)
	tfClient, err := client.NewClient(ctx, config.ConsoleClientConfig)
	if err != nil {
		return nil, fmt.Errorf("Error creating VNPAY Cloud Terrform Client: %s", err)
	}

	poolResp := &dto.GetPoolResponse{}
	_, err = tfClient.Get(ctx, client.ApiPath.LbaasPoolWithId(d.Id()), poolResp, &client.RequestOpts{})
	if err != nil {
		return nil, util.CheckDeleted(d, err, "pool")
	}
	pool := &poolResp.Pool

	log.Printf("[DEBUG] Retrieved pool %s during the import: %#v", d.Id(), pool)

	if len(pool.Listeners) > 0 && pool.Listeners[0].ID != "" {
		d.Set("listener_id", pool.Listeners[0].ID)
	} else if len(pool.Loadbalancers) > 0 && pool.Loadbalancers[0].ID != "" {
		d.Set("loadbalancer_id", pool.Loadbalancers[0].ID)
	} else {
		return nil, fmt.Errorf("Unable to detect pool's Listener ID or Load Balancer ID")
	}

	return []*schema.ResourceData{d}, nil
}
