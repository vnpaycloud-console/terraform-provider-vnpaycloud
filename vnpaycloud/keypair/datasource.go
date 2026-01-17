package keypair

import (
	"context"
	"log"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceComputeKeypair() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceComputeKeypairRead,

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			// computed-only
			"fingerprint": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"public_key": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"user_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func dataSourceComputeKeypairRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	computeClient, err := client.NewClient(ctx, config.ConsoleClientConfig)
	if err != nil {
		return diag.Errorf("Error creating VNPAYCLOUD compute client: %s", err)
	}

	opts := dto.GetKeyPairOpts{}

	// Check if searching for the keypair of another user
	userID := d.Get("user_id").(string)
	if userID != "" {
		opts.UserID = userID
	}

	name := d.Get("name").(string)
	kp := &dto.GetKeyPairResponse{}
	_, err = computeClient.Get(ctx, client.ApiPath.KeyPairWithParams(opts), kp, nil)
	if err != nil {
		return diag.Errorf("Error retrieving vnpaycloud_compute_keypair %s: %s", name, err)
	}

	d.SetId(name)

	log.Printf("[DEBUG] Retrieved vnpaycloud_compute_keypair %s: %#v", d.Id(), kp)

	d.Set("fingerprint", kp.KeyPair.Fingerprint)
	d.Set("public_key", kp.KeyPair.PublicKey)
	d.Set("region", util.GetRegion(d, config))
	d.Set("user_id", kp.KeyPair.UserID)

	return nil
}
