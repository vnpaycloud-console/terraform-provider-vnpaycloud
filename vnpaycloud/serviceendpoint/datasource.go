package serviceendpoint

import (
	"context"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func serviceEndpointElemSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"id":                  {Type: schema.TypeString, Computed: true},
		"name":                {Type: schema.TypeString, Computed: true},
		"description":         {Type: schema.TypeString, Computed: true},
		"provider_id":         {Type: schema.TypeString, Computed: true},
		"service_id":          {Type: schema.TypeString, Computed: true},
		"service_gateway_id":  {Type: schema.TypeString, Computed: true},
		"port":                {Type: schema.TypeInt, Computed: true},
		"allowed_cidrs":       {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
		"listener_id":         {Type: schema.TypeString, Computed: true},
		"pool_id":             {Type: schema.TypeString, Computed: true},
		"health_monitor_id":   {Type: schema.TypeString, Computed: true},
		"pool_member_ids":     {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
		"operating_status":    {Type: schema.TypeString, Computed: true},
		"provisioning_status": {Type: schema.TypeString, Computed: true},
		"status":              {Type: schema.TypeString, Computed: true},
		"created_at":          {Type: schema.TypeString, Computed: true},
	}
}

func flattenServiceEndpoint(se dto.ServiceEndpoint) map[string]interface{} {
	return map[string]interface{}{
		"id":                  se.ID,
		"name":                se.Name,
		"description":         se.Description,
		"provider_id":         se.ProviderID,
		"service_id":          se.ServiceID,
		"service_gateway_id":  se.ServiceGatewayID,
		"port":                se.Port,
		"allowed_cidrs":       se.AllowedCIDRs,
		"listener_id":         se.ListenerID,
		"pool_id":             se.PoolID,
		"health_monitor_id":   se.HealthMonitorID,
		"pool_member_ids":     se.PoolMemberIDs,
		"operating_status":    se.OperatingStatus,
		"provisioning_status": se.ProvisioningStatus,
		"status":              se.Status,
		"created_at":          se.CreatedAt,
	}
}

func DataSourceServiceEndpoint() *schema.Resource {
	elem := serviceEndpointElemSchema()
	elem["id"] = &schema.Schema{Type: schema.TypeString, Required: true}
	return &schema.Resource{
		ReadContext: dataSourceServiceEndpointRead,
		Schema:      elem,
	}
}

func dataSourceServiceEndpointRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	id := d.Get("id").(string)

	resp := &dto.ServiceEndpointResponse{}
	if _, err := cfg.Client.Get(ctx, client.ApiPath.ServiceEndpointWithID(cfg.ProjectID, id), resp, nil); err != nil {
		return diag.Errorf("Error retrieving vnpaycloud_service_endpoint %s: %s", id, err)
	}

	d.SetId(resp.ServiceEndpoint.ID)
	for k, v := range flattenServiceEndpoint(resp.ServiceEndpoint) {
		if k == "id" {
			continue
		}
		d.Set(k, v)
	}

	return nil
}

func DataSourceServiceEndpoints() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceServiceEndpointsRead,
		Schema: map[string]*schema.Schema{
			"service_gateway_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"service_endpoints": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Resource{Schema: serviceEndpointElemSchema()},
			},
		},
	}
}

func dataSourceServiceEndpointsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	path := client.ApiPath.ServiceEndpoints(cfg.ProjectID)
	if sgID := d.Get("service_gateway_id").(string); sgID != "" {
		path += "?serviceGatewayId=" + sgID
	}

	resp := &dto.ListServiceEndpointsResponse{}
	if _, err := cfg.Client.Get(ctx, path, resp, nil); err != nil {
		return diag.Errorf("Error listing vnpaycloud_service_endpoints: %s", err)
	}

	endpoints := make([]map[string]interface{}, 0, len(resp.ServiceEndpoints))
	for _, se := range resp.ServiceEndpoints {
		endpoints = append(endpoints, flattenServiceEndpoint(se))
	}

	d.SetId("service_endpoints")
	d.Set("service_endpoints", endpoints)

	return nil
}

func DataSourceServiceProviders() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceServiceProvidersRead,
		Schema: map[string]*schema.Schema{
			"providers": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id":     {Type: schema.TypeString, Computed: true},
						"name":   {Type: schema.TypeString, Computed: true},
						"status": {Type: schema.TypeString, Computed: true},
					},
				},
			},
		},
	}
}

func dataSourceServiceProvidersRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	resp := &dto.ListServiceProvidersResponse{}
	if _, err := cfg.Client.Get(ctx, client.ApiPath.ServiceProviders(cfg.ProjectID), resp, nil); err != nil {
		return diag.Errorf("Error listing vnpaycloud_service_providers: %s", err)
	}

	providers := make([]map[string]interface{}, 0, len(resp.Providers))
	for _, p := range resp.Providers {
		providers = append(providers, map[string]interface{}{
			"id":     p.ID,
			"name":   p.Name,
			"status": p.Status,
		})
	}

	d.SetId("service_providers")
	d.Set("providers", providers)

	return nil
}

func DataSourceServices() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceServicesRead,
		Schema: map[string]*schema.Schema{
			"provider_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"services": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id":             {Type: schema.TypeString, Computed: true},
						"name":           {Type: schema.TypeString, Computed: true},
						"description":    {Type: schema.TypeString, Computed: true},
						"provider_id":    {Type: schema.TypeString, Computed: true},
						"service_domain": {Type: schema.TypeString, Computed: true},
						"status":         {Type: schema.TypeString, Computed: true},
					},
				},
			},
		},
	}
}

func dataSourceServicesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	path := client.ApiPath.Services(cfg.ProjectID)
	sep := "?"
	if providerID := d.Get("provider_id").(string); providerID != "" {
		path += sep + "providerId=" + providerID
		sep = "&"
	}
	if name := d.Get("name").(string); name != "" {
		path += sep + "name=" + name
	}

	resp := &dto.ListServicesResponse{}
	if _, err := cfg.Client.Get(ctx, path, resp, nil); err != nil {
		return diag.Errorf("Error listing vnpaycloud_services: %s", err)
	}

	services := make([]map[string]interface{}, 0, len(resp.Services))
	for _, s := range resp.Services {
		services = append(services, map[string]interface{}{
			"id":             s.ID,
			"name":           s.Name,
			"description":    s.Description,
			"provider_id":    s.ProviderID,
			"service_domain": s.ServiceDomain,
			"status":         s.Status,
		})
	}

	d.SetId("services")
	d.Set("services", services)

	return nil
}
