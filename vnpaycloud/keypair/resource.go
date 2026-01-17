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

func ResourceComputeKeypair() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceComputeKeypairCreate,
		ReadContext:   resourceComputeKeypairRead,
		DeleteContext: resourceComputeKeypairDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"public_key": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"value_specs": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
			},

			// computed-only
			"private_key": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},

			"fingerprint": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"user_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
		},
	}
}

func resourceComputeKeypairCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	computeClient, err := client.NewClient(ctx, config.ConsoleClientConfig)
	if err != nil {
		return diag.Errorf("Error creating VNPAYCLOUD compute client: %s", err)
	}

	// userID := d.Get("user_id").(string)
	// if userID != "" {
	// 	computeClient.Microversion = computeKeyPairUserIDMicroversion
	// }

	name := d.Get("name").(string)
	createOpts := dto.CreateKeyPairOpts{
		Name:       name,
		PublicKey:  d.Get("public_key").(string),
		UserID:     d.Get("user_id").(string),
		ValueSpecs: util.MapValueSpecs(d),
	}

	log.Printf("[DEBUG] vnpaycloud_compute_keypair create options: %#v", createOpts)

	createReq := dto.CreateKeyPairRequest{
		KeyPair: createOpts,
	}

	createResp := &dto.CreateKeyPairResponse{}
	_, err = computeClient.Post(ctx, client.ApiPath.KeyPair, createReq, createResp, &client.RequestOpts{OkCodes: []int{200, 201, 202}})
	if err != nil {
		return diag.Errorf("Unable to create vnpaycloud_compute_keypair %s: %s", name, err)
	}

	kp := createResp.KeyPair

	d.SetId(kp.Name)
	d.Set("user_id", d.Get("user_id").(string))

	// Private Key is only available in the response to a create.
	d.Set("private_key", kp.PrivateKey)

	return resourceComputeKeypairRead(ctx, d, meta)
}

func resourceComputeKeypairRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	computeClient, err := client.NewClient(ctx, config.ConsoleClientConfig)
	if err != nil {
		return diag.Errorf("Error creating VNPAYCLOUD compute client: %s", err)
	}

	userID := d.Get("user_id").(string)

	kpopts := dto.GetKeyPairOpts{
		UserID: userID,
	}
	kp := &dto.GetKeyPairResponse{}

	_, err = computeClient.Get(ctx, client.ApiPath.KeyPairWithIdAndParams(d.Id(), kpopts), kp, nil)
	if err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error retrieving vnpaycloud_compute_keypair"))
	}

	log.Printf("[DEBUG] Retrieved vnpaycloud_compute_keypair %s: %#v", d.Id(), kp.KeyPair)

	d.Set("name", kp.KeyPair.Name)
	d.Set("public_key", kp.KeyPair.PublicKey)
	d.Set("fingerprint", kp.KeyPair.Fingerprint)
	d.Set("region", util.GetRegion(d, config))
	if userID != "" {
		d.Set("user_id", kp.KeyPair.UserID)
	}

	return nil
}

func resourceComputeKeypairDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	computeClient, err := client.NewClient(ctx, config.ConsoleClientConfig)
	if err != nil {
		return diag.Errorf("Error creating VNPAYCLOUD compute client: %s", err)
	}

	userID := d.Get("user_id").(string)
	log.Printf("[DEBUG] User ID %s", userID)

	kpopts := dto.DeleteKeyPairOpts{
		UserID: userID,
	}

	_, err = computeClient.Delete(ctx, client.ApiPath.KeyPairWithIdAndParams(d.Id(), kpopts), nil)
	if err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error deleting vnpaycloud_compute_keypair"))
	}

	return nil
}
