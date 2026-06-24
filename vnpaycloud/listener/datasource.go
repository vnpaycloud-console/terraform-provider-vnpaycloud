package listener

import (
	"context"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceListener() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceListenerRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"load_balancer_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"protocol": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"protocol_port": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"default_pool_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"insert_headers": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"allowed_cidrs": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"connection_limit": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"timeout_client_data": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"timeout_member_connect": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"timeout_member_data": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"certificate_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"certificate_authority_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"sni_certificate_ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
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

func dataSourceListenerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	id := d.Get("id").(string)

	resp := &dto.ListenerResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.ListenerWithID(cfg.ProjectID, id), resp, nil)
	if err != nil {
		return diag.Errorf("Error retrieving vnpaycloud_lb_listener %s: %s", id, err)
	}

	d.SetId(resp.Listener.ID)
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

func DataSourceListeners() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceListenersRead,
		Schema: map[string]*schema.Schema{
			"listeners": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id":                       {Type: schema.TypeString, Computed: true},
						"name":                     {Type: schema.TypeString, Computed: true},
						"description":              {Type: schema.TypeString, Computed: true},
						"load_balancer_id":         {Type: schema.TypeString, Computed: true},
						"protocol":                 {Type: schema.TypeString, Computed: true},
						"protocol_port":            {Type: schema.TypeInt, Computed: true},
						"default_pool_id":          {Type: schema.TypeString, Computed: true},
						"insert_headers":           {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
						"allowed_cidrs":            {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
						"connection_limit":         {Type: schema.TypeInt, Computed: true},
						"timeout_client_data":      {Type: schema.TypeInt, Computed: true},
						"timeout_member_connect":   {Type: schema.TypeInt, Computed: true},
						"timeout_member_data":      {Type: schema.TypeInt, Computed: true},
						"certificate_id":           {Type: schema.TypeString, Computed: true},
						"certificate_authority_id": {Type: schema.TypeString, Computed: true},
						"sni_certificate_ids":      {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
						"status":                   {Type: schema.TypeString, Computed: true},
						"created_at":               {Type: schema.TypeString, Computed: true},
					},
				},
			},
		},
	}
}

func dataSourceListenersRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	resp := &dto.ListListenersResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.Listeners(cfg.ProjectID), resp, nil)
	if err != nil {
		return diag.Errorf("Error listing vnpaycloud_lb_listeners: %s", err)
	}

	var listeners []map[string]interface{}
	for _, l := range resp.Listeners {
		listeners = append(listeners, map[string]interface{}{
			"id":                       l.ID,
			"name":                     l.Name,
			"description":              l.Description,
			"load_balancer_id":         l.LoadBalancerID,
			"protocol":                 l.Protocol,
			"protocol_port":            l.ProtocolPort,
			"default_pool_id":          l.DefaultPoolID,
			"insert_headers":           l.InsertHeaders,
			"allowed_cidrs":            l.AllowedCidrs,
			"connection_limit":         l.ConnectionLimit,
			"timeout_client_data":      l.TimeoutClientData,
			"timeout_member_connect":   l.TimeoutMemberConnect,
			"timeout_member_data":      l.TimeoutMemberData,
			"certificate_id":           l.CertificateID,
			"certificate_authority_id": l.CertificateAuthorityID,
			"sni_certificate_ids":      l.SniCertificateIDs,
			"status":                   l.Status,
			"created_at":               l.CreatedAt,
		})
	}

	d.SetId("listeners")
	d.Set("listeners", listeners)

	return nil
}
