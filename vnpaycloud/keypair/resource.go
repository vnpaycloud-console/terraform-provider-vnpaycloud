package keypair

import (
	"context"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceKeyPair() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeyPairCreate,
		ReadContext:   resourceKeyPairRead,
		DeleteContext: resourceKeyPairDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
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
			"private_key": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
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

func resourceKeyPairCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	createOpts := dto.CreateKeyPairRequest{
		Name:      d.Get("name").(string),
		PublicKey:  d.Get("public_key").(string),
	}

	tflog.Debug(ctx, "vnpaycloud_keypair create options", map[string]interface{}{"name": createOpts.Name})

	createResp := &dto.KeyPairResponse{}
	_, err := cfg.Client.Post(ctx, client.ApiPath.CreateKeyPair(), createOpts, createResp, nil)
	if err != nil {
		return diag.Errorf("Error creating vnpaycloud_keypair: %s", err)
	}

	d.SetId(createResp.KeyPair.ID)
	d.Set("name", createResp.KeyPair.Name)

	// Store private key if auto-generated (only available at creation time)
	if createResp.PrivateKey != "" {
		d.Set("private_key", createResp.PrivateKey)
	}

	d.Set("public_key", createResp.KeyPair.PublicKey)
	d.Set("fingerprint", createResp.KeyPair.Fingerprint)
	d.Set("created_at", createResp.KeyPair.CreatedAt)

	return nil
}

func resourceKeyPairRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	name := d.Get("name").(string)

	kpResp := &dto.KeyPairResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.KeyPairWithName(cfg.ProjectID, name), kpResp, nil)
	if err != nil {
		if client.ResponseCodeIs(err, 404) {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Error retrieving vnpaycloud_keypair %s: %s", d.Id(), err)
	}

	tflog.Debug(ctx, "Retrieved vnpaycloud_keypair "+d.Id(), map[string]interface{}{"keypair": kpResp.KeyPair})

	d.Set("name", kpResp.KeyPair.Name)
	d.Set("public_key", kpResp.KeyPair.PublicKey)
	d.Set("fingerprint", kpResp.KeyPair.Fingerprint)
	d.Set("created_at", kpResp.KeyPair.CreatedAt)
	// Note: private_key is only available at creation time, not on subsequent reads

	return nil
}

func resourceKeyPairDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	name := d.Get("name").(string)

	if _, err := cfg.Client.Delete(ctx, client.ApiPath.KeyPairWithName(cfg.ProjectID, name), nil); err != nil {
		if client.ResponseCodeIs(err, 404) {
			return nil
		}
		return diag.Errorf("Error deleting vnpaycloud_keypair %s: %s", d.Id(), err)
	}

	return nil
}
