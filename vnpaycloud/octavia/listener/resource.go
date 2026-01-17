package listener

import (
	"context"
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

func ResourceListener() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceListenerCreate,
		ReadContext:   resourceListenerRead,
		UpdateContext: resourceListenerUpdate,
		DeleteContext: resourceListenerDelete,
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

			"protocol": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"TCP", "UDP", "SCTP", "HTTP", "HTTPS", "TERMINATED_HTTPS", "PROMETHEUS",
				}, false),
			},

			"protocol_port": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},

			"tenant_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"loadbalancer_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"default_pool_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"connection_limit": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},

			"default_tls_container_ref": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"sni_container_refs": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
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

			"client_authentication": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"NONE", "OPTIONAL", "MANDATORY",
				}, false),
				DiffSuppressFunc: func(k, o, n string, d *schema.ResourceData) bool {
					return o == "NONE" && n == ""
				},
				DiffSuppressOnRefresh: true,
			},

			"client_ca_tls_container_ref": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"client_crl_container_ref": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"hsts_include_subdomains": {
				Type:         schema.TypeBool,
				Optional:     true,
				RequiredWith: []string{"hsts_max_age"},
			},

			"hsts_max_age": {
				Type:     schema.TypeInt,
				Optional: true,
			},

			"hsts_preload": {
				Type:         schema.TypeBool,
				Optional:     true,
				RequiredWith: []string{"hsts_max_age"},
			},

			"tls_ciphers": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true, // unsetting this parameter results in a default value
			},

			"tls_versions": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true, // unsetting this parameter is not possible due to a bug in Octavia
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

			"timeout_tcp_inspect": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},

			"insert_headers": {
				Type:     schema.TypeMap,
				Optional: true,
			},

			"allowed_cidrs": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
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

func resourceListenerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	tfClient, err := client.NewClient(ctx, config.ConsoleClientConfig)
	if err != nil {
		return diag.Errorf("Error creating VNPAY Cloud Terrform Client: %s", err)
	}

	timeout := d.Timeout(schema.TimeoutCreate)

	// Wait for LoadBalancer to become active before continuing.
	err = shared.WaitForLBLoadBalancer(ctx, tfClient, d.Get("loadbalancer_id").(string), "ACTIVE", shared.GetLbPendingStatuses(), timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	//adminStateUp := d.Get("admin_state_up").(bool)
	createOpts := dto.CreateListenerOpts{
		// Protocol SCTP requires octavia minor version 2.23
		Protocol:     dto.Protocol(d.Get("protocol").(string)),
		ProtocolPort: d.Get("protocol_port").(int),
		//ProjectID:               d.Get("tenant_id").(string),
		LoadbalancerID: d.Get("loadbalancer_id").(string),
		Name:           d.Get("name").(string),
		//DefaultPoolID:           d.Get("default_pool_id").(string),
		Description: d.Get("description").(string),
		//DefaultTlsContainerRef:  d.Get("default_tls_container_ref").(string),
		//SniContainerRefs:        util.ExpandToStringSlice(d.Get("sni_container_refs").([]interface{})),
		//ALPNProtocols:           util.ExpandToStringSlice(d.Get("alpn_protocols").(*schema.Set).List()),
		//ClientAuthentication:    listeners.ClientAuthentication(d.Get("client_authentication").(string)),
		//ClientCATLSContainerRef: d.Get("client_ca_tls_container_ref").(string),
		//ClientCRLContainerRef:   d.Get("client_crl_container_ref").(string),
		//HSTSIncludeSubdomains:   d.Get("hsts_include_subdomains").(bool),
		//HSTSMaxAge:              d.Get("hsts_max_age").(int),
		//HSTSPreload:             d.Get("hsts_preload").(bool),
		//TLSCiphers:              d.Get("tls_ciphers").(string),
		//InsertHeaders:           util.ExpandToMapStringString(d.Get("insert_headers").(map[string]interface{})),
		AllowedCIDRs: util.ExpandToStringSlice(d.Get("allowed_cidrs").([]interface{})),
		//AdminStateUp:            &adminStateUp,
		//Tags: util.ExpandToStringSlice(d.Get("tags").(*schema.Set).List()),
	}

	//if v, ok := d.GetOk("tls_versions"); ok {
	//	createOpts.TLSVersions = shared.ExpandLBListenerTLSVersion(v.(*schema.Set).List())
	//}
	//
	//if v, ok := d.GetOk("connection_limit"); ok {
	//	connectionLimit := v.(int)
	//	createOpts.ConnLimit = &connectionLimit
	//}
	//
	//if v, ok := d.GetOk("timeout_client_data"); ok {
	//	timeoutClientData := v.(int)
	//	createOpts.TimeoutClientData = &timeoutClientData
	//}
	//
	//if v, ok := d.GetOk("timeout_member_connect"); ok {
	//	timeoutMemberConnect := v.(int)
	//	createOpts.TimeoutMemberConnect = &timeoutMemberConnect
	//}
	//
	//if v, ok := d.GetOk("timeout_member_data"); ok {
	//	timeoutMemberData := v.(int)
	//	createOpts.TimeoutMemberData = &timeoutMemberData
	//}
	//
	//if v, ok := d.GetOk("timeout_tcp_inspect"); ok {
	//	timeoutTCPInspect := v.(int)
	//	createOpts.TimeoutTCPInspect = &timeoutTCPInspect
	//}

	log.Printf("[DEBUG] vnpaycloud_lb_listener create options: %#v", createOpts)
	listenerResp := &dto.CreateListenerResponse{}

	err = retry.RetryContext(ctx, timeout, func() *retry.RetryError {
		_, err = tfClient.Post(ctx, client.ApiPath.LbaasListener, dto.CreateListenerRequest{Listener: createOpts}, listenerResp, nil)
		if err != nil {
			return util.CheckForRetryableError(err)
		}
		return nil
	})

	if err != nil {
		return diag.Errorf("Error creating vnpaycloud_lb_listener: %s", err)
	}
	listener := &listenerResp.Listener

	// Wait for the listener to become ACTIVE.
	err = shared.WaitForLBListener(ctx, tfClient, listener, "ACTIVE", shared.GetLbPendingStatuses(), timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(listener.ID)

	return resourceListenerRead(ctx, d, meta)
}

func resourceListenerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	tfClient, err := client.NewClient(ctx, config.ConsoleClientConfig)
	if err != nil {
		return diag.Errorf("Error creating VNPAY Cloud Terrform Client: %s", err)
	}

	listenerResp := &dto.GetListenerResponse{}
	_, err = tfClient.Get(ctx, client.ApiPath.LbaasListenerWithId(d.Id()), listenerResp, &client.RequestOpts{})
	if err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "vnpaycloud_lb_listener"))
	}

	listener := &listenerResp.Listener
	log.Printf("[DEBUG] Retrieved vnpaycloud_lb_listener %s: %#v", d.Id(), listener)

	d.Set("name", listener.Name)
	d.Set("protocol", listener.Protocol)
	d.Set("tenant_id", listener.ProjectID)
	d.Set("description", listener.Description)
	d.Set("protocol_port", listener.ProtocolPort)
	d.Set("admin_state_up", listener.AdminStateUp)
	d.Set("default_pool_id", listener.DefaultPoolID)
	d.Set("connection_limit", listener.ConnLimit)
	d.Set("timeout_client_data", listener.TimeoutClientData)
	d.Set("timeout_member_connect", listener.TimeoutMemberConnect)
	d.Set("timeout_member_data", listener.TimeoutMemberData)
	d.Set("timeout_tcp_inspect", listener.TimeoutTCPInspect)
	d.Set("sni_container_refs", listener.SniContainerRefs)
	d.Set("default_tls_container_ref", listener.DefaultTlsContainerRef)
	d.Set("allowed_cidrs", listener.AllowedCIDRs)
	d.Set("alpn_protocols", listener.ALPNProtocols)
	d.Set("client_authentication", listener.ClientAuthentication)
	d.Set("client_ca_tls_container_ref", listener.ClientCATLSContainerRef)
	d.Set("client_crl_container_ref", listener.ClientCRLContainerRef)
	d.Set("hsts_include_subdomains", listener.HSTSIncludeSubdomains)
	d.Set("hsts_max_age", listener.HSTSMaxAge)
	d.Set("hsts_preload", listener.HSTSPreload)
	d.Set("tls_ciphers", listener.TLSCiphers)
	d.Set("tls_versions", listener.TLSVersions)
	d.Set("region", util.GetRegion(d, config))
	d.Set("tags", listener.Tags)

	// Required by import.
	//if len(listener.Loadbalancers) > 0 {
	//	d.Set("loadbalancer_id", listener.Loadbalancers[0].ID)
	//}

	if err := d.Set("insert_headers", listener.InsertHeaders); err != nil {
		return diag.Errorf("Unable to set vnpaycloud_lb_listener insert_headers: %s", err)
	}

	return nil
}

func resourceListenerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceListenerRead(ctx, d, meta)
}

func resourceListenerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	tfClient, err := client.NewClient(ctx, config.ConsoleClientConfig)
	if err != nil {
		return diag.Errorf("Error creating VNPAY Cloud Terrform Client: %s", err)
	}

	// Get a clean copy of the listener.
	listenerResp := &dto.GetListenerResponse{}
	_, err = tfClient.Get(ctx, client.ApiPath.LbaasListenerWithId(d.Id()), listenerResp, &client.RequestOpts{})
	if err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Unable to retrieve vnpaycloud_lb_listener"))
	}
	listener := &listenerResp.Listener

	timeout := d.Timeout(schema.TimeoutDelete)

	log.Printf("[DEBUG] Deleting vnpaycloud_lb_listener %s", d.Id())
	err = retry.RetryContext(ctx, timeout, func() *retry.RetryError {
		_, err = tfClient.Delete(ctx, client.ApiPath.LbaasListenerWithId(d.Id()), nil)
		if err != nil {
			return util.CheckForRetryableError(err)
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error deleting vnpaycloud_lb_listener"))
	}

	// Wait for the listener to become DELETED.
	err = shared.WaitForLBListener(ctx, tfClient, listener, "DELETED", shared.GetLbPendingDeleteStatuses(), timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
