package applicationcredentials

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vnpaycloud-console/gophercloud/v2"
	"github.com/vnpaycloud-console/gophercloud/v2/openstack/identity/v3/applicationcredentials"
	"log"
	"net/http"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/util"
	"time"
)

func ResourceIdentityApplicationCredentialV3() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceIdentityApplicationCredentialV3Create,
		ReadContext:   resourceIdentityApplicationCredentialV3Read,
		DeleteContext: resourceIdentityApplicationCredentialV3Delete,
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
				ForceNew: true,
				Required: true,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"unrestricted": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},
			"secret": {
				Type:      schema.TypeString,
				Optional:  true,
				Computed:  true,
				Sensitive: true,
				ForceNew:  true,
			},

			"project_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"access_rules": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"path": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"service": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"method": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice([]string{
								"POST", "GET", "HEAD", "PATCH", "PUT", "DELETE",
							}, false),
						},
					},
				},
			},

			"roles": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"expires_at": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsRFC3339Time,
			},
		},
	}
}

func resourceIdentityApplicationCredentialV3Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	identityClient, err := cfg.IdentityV3Client(ctx, util.GetRegion(d, cfg))
	if err != nil {
		return diag.Errorf("Error creating OpenStack identity client: %s", err)
	}

	tokenInfo, err := util.GetTokenInfo(ctx, identityClient)
	if err != nil {
		return diag.FromErr(err)
	}

	var expiresAt *time.Time
	if v, err := time.Parse(time.RFC3339, d.Get("expires_at").(string)); err == nil {
		expiresAt = &v
	}

	createOpts := applicationcredentials.CreateOpts{
		Name:         d.Get("name").(string),
		Description:  d.Get("description").(string),
		Unrestricted: d.Get("unrestricted").(bool),
		Roles:        expandIdentityApplicationCredentialRolesV3(d.Get("roles").(*schema.Set).List()),
		AccessRules:  expandIdentityApplicationCredentialAccessRulesV3(d.Get("access_rules").(*schema.Set).List()),
		ExpiresAt:    expiresAt,
	}

	log.Printf("[DEBUG] openstack_identity_application_credential_v3 create options: %#v", createOpts)

	createOpts.Secret = d.Get("secret").(string)

	applicationCredential, err := applicationcredentials.Create(ctx, identityClient, tokenInfo.UserID, createOpts).Extract()
	if err != nil {
		if gophercloud.ResponseCodeIs(err, http.StatusNotFound) {
			err := err.(gophercloud.ErrUnexpectedResponseCode)
			return diag.Errorf("Error creating openstack_identity_application_credential_v3: %s", err.Body)
		}
		return diag.Errorf("Error creating openstack_identity_application_credential_v3: %s", err)
	}

	d.SetId(applicationCredential.ID)

	// Secret is returned only once
	d.Set("secret", applicationCredential.Secret)

	return resourceIdentityApplicationCredentialV3Read(ctx, d, meta)
}

func resourceIdentityApplicationCredentialV3Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	identityClient, err := cfg.IdentityV3Client(ctx, util.GetRegion(d, cfg))
	if err != nil {
		return diag.Errorf("Error creating OpenStack identity client: %s", err)
	}

	tokenInfo, err := util.GetTokenInfo(ctx, identityClient)
	if err != nil {
		return diag.FromErr(err)
	}

	applicationCredential, err := applicationcredentials.Get(ctx, identityClient, tokenInfo.UserID, d.Id()).Extract()
	if err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error retrieving openstack_identity_application_credential_v3"))
	}

	log.Printf("[DEBUG] Retrieved openstack_identity_application_credential_v3 %s: %#v", d.Id(), applicationCredential)

	d.Set("name", applicationCredential.Name)
	d.Set("description", applicationCredential.Description)
	d.Set("unrestricted", applicationCredential.Unrestricted)
	d.Set("roles", flattenIdentityApplicationCredentialRolesV3(applicationCredential.Roles))
	d.Set("access_rules", flattenIdentityApplicationCredentialAccessRulesV3(applicationCredential.AccessRules))
	d.Set("project_id", applicationCredential.ProjectID)
	d.Set("region", util.GetRegion(d, cfg))

	if applicationCredential.ExpiresAt == (time.Time{}) {
		d.Set("expires_at", "")
	} else {
		d.Set("expires_at", applicationCredential.ExpiresAt.UTC().Format(time.RFC3339))
	}

	return nil
}

func resourceIdentityApplicationCredentialV3Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	identityClient, err := cfg.IdentityV3Client(ctx, util.GetRegion(d, cfg))
	if err != nil {
		return diag.Errorf("Error creating OpenStack identity client: %s", err)
	}

	tokenInfo, err := util.GetTokenInfo(ctx, identityClient)
	if err != nil {
		return diag.FromErr(err)
	}

	err = applicationcredentials.Delete(ctx, identityClient, tokenInfo.UserID, d.Id()).ExtractErr()
	if err != nil {
		err = util.CheckDeleted(d, err, "Error deleting openstack_identity_application_credential_v3")
		if err != nil {
			return diag.FromErr(err)
		}
	}

	// cleanup access rules
	accessRules := expandIdentityApplicationCredentialAccessRulesV3(d.Get("access_rules").(*schema.Set).List())
	return diag.FromErr(applicationCredentialCleanupAccessRulesV3(ctx, identityClient, tokenInfo.UserID, d.Id(), accessRules))
}
