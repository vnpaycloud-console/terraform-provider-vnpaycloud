package certificate

import (
	"context"

	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceCertificates() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceCertificatesRead,
		Schema: map[string]*schema.Schema{
			"certificates": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id":                {Type: schema.TypeString, Computed: true},
						"name":              {Type: schema.TypeString, Computed: true},
						"cert_type":         {Type: schema.TypeString, Computed: true},
						"domain_name":       {Type: schema.TypeString, Computed: true},
						"description":       {Type: schema.TypeString, Computed: true},
						"expiration":        {Type: schema.TypeString, Computed: true},
						"status":            {Type: schema.TypeString, Computed: true},
						"zone_id":           {Type: schema.TypeString, Computed: true},
						"load_balancer_ids": {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
					},
				},
			},
		},
	}
}

func dataSourceCertificatesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	resp := &dto.ListCertificatesResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.Certificates(cfg.ProjectID), resp, nil)
	if err != nil {
		return diag.Errorf("Error listing vnpaycloud_certificates: %s", err)
	}

	out := make([]map[string]interface{}, 0, len(resp.Certificates))
	for _, c := range resp.Certificates {
		out = append(out, map[string]interface{}{
			"id":                c.ID,
			"name":              c.Name,
			"cert_type":         c.CertType,
			"domain_name":       c.DomainName,
			"description":       c.Description,
			"expiration":        c.Expiration,
			"status":            c.Status,
			"zone_id":           c.ZoneID,
			"load_balancer_ids": c.LoadBalancerIDs,
		})
	}

	d.SetId("certificates")
	d.Set("certificates", out)

	return nil
}
