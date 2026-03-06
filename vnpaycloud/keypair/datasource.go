package keypair

import (
	"context"
	"fmt"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceKeyPair() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceKeyPairRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"public_key": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"fingerprint": {
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

func dataSourceKeyPairRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	name := d.Get("name").(string)

	kpResp := &dto.KeyPairResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.KeyPairWithName(cfg.ProjectID, name), kpResp, nil)
	if err != nil {
		return diag.Errorf("Error fetching vnpaycloud_keypair %s: %s", name, err)
	}

	return setKeyPairData(d, &kpResp.KeyPair)
}

func setKeyPairData(d *schema.ResourceData, kp *dto.KeyPair) diag.Diagnostics {
	d.SetId(kp.Name)
	d.Set("name", kp.Name)
	d.Set("public_key", kp.PublicKey)
	d.Set("fingerprint", kp.Fingerprint)
	d.Set("created_at", kp.CreatedAt)
	return nil
}

func DataSourceKeyPairs() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceKeyPairsRead,
		Schema: map[string]*schema.Schema{
			"key_pairs": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name":        {Type: schema.TypeString, Computed: true},
						"public_key":  {Type: schema.TypeString, Computed: true},
						"fingerprint": {Type: schema.TypeString, Computed: true},
						"created_at":  {Type: schema.TypeString, Computed: true},
					},
				},
			},
		},
	}
}

func dataSourceKeyPairsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	listResp := &dto.ListKeyPairsResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.KeyPairs(cfg.ProjectID), listResp, nil)
	if err != nil {
		return diag.Errorf("Error listing vnpaycloud_keypairs: %s", err)
	}

	var keyPairs []map[string]interface{}
	for _, kp := range listResp.KeyPairs {
		keyPairs = append(keyPairs, map[string]interface{}{
			"name":        kp.Name,
			"public_key":  kp.PublicKey,
			"fingerprint": kp.Fingerprint,
			"created_at":  kp.CreatedAt,
		})
	}

	d.SetId(fmt.Sprintf("key-pairs-%s", cfg.ProjectID))
	d.Set("key_pairs", keyPairs)

	return nil
}
