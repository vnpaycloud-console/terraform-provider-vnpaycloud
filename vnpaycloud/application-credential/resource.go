package applicationcredentials

import (
	"context"
	"log"
	"net/http"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
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
	config := meta.(*config.Config)
	identityClient, err := client.NewClient(ctx, config.ConsoleClientConfig)
	if err != nil {
		return diag.Errorf("Error creating VNPAYCloud identity client: %s", err)
	}

	tokenInfo, err := util.GetTokenInfo(ctx, identityClient)
	if err != nil {
		return diag.FromErr(err)
	}

	var expiresAt *time.Time
	if v, err := time.Parse(time.RFC3339, d.Get("expires_at").(string)); err == nil {
		expiresAt = &v
	}

	var accessRules []dto.AccessRule
	if v, ok := d.GetOk("access_rules"); ok && v.(*schema.Set).Len() > 0 {
		accessRules = expandIdentityApplicationCredentialAccessRulesV3(v.(*schema.Set).List())
	} else {
		accessRules = []dto.AccessRule{}
	}

	createOpts := dto.CreateApplicationCredentialOpts{
		Name:         d.Get("name").(string),
		Description:  d.Get("description").(string),
		Unrestricted: d.Get("unrestricted").(bool),
		Roles:        expandIdentityApplicationCredentialRolesV3(d.Get("roles").(*schema.Set).List()),
		AccessRules:  accessRules,
		ExpiresAt:    expiresAt,
	}

	log.Printf("[DEBUG] vnpaycloud_identity_application_credential create options: %#v", createOpts)

	createRequest := dto.CreateApplicationCredentialRequest{
		ApplicationCredential: createOpts,
	}

	createOpts.Secret = d.Get("secret").(string)
	createResponse := dto.CreateApplicationCredentialResponse{}

	_, err = identityClient.Post(ctx, client.ApiPath.ApplicationCredential(tokenInfo.UserID), createRequest, &createResponse, nil)
	if err != nil {
		if client.ResponseCodeIs(err, http.StatusNotFound) {
			err := err.(client.ErrUnexpectedResponseCode)
			return diag.Errorf("Error creating vnpaycloud_identity_application_credential: %s", err.Body)
		}
		return diag.Errorf("Error creating vnpaycloud_identity_application_credential: %s", err)
	}

	d.SetId(createResponse.ApplicationCredential.ID)

	// Secret is returned only once
	d.Set("secret", createResponse.ApplicationCredential.Secret)

	return resourceIdentityApplicationCredentialV3Read(ctx, d, meta)
}

func resourceIdentityApplicationCredentialV3Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	identityClient, err := client.NewClient(ctx, config.ConsoleClientConfig)
	if err != nil {
		return diag.Errorf("Error creating VNPAYCloud identity client: %s", err)
	}

	tokenInfo, err := util.GetTokenInfo(ctx, identityClient)
	if err != nil {
		return diag.FromErr(err)
	}

	getResponse := dto.GetApplicationCredentialResponse{}
	_, err = identityClient.Get(ctx, client.ApiPath.ApplicationCredentialWithId(tokenInfo.UserID, d.Id()), &getResponse, nil)
	if err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error retrieving vnpaycloud_identity_application_credential"))
	}

	applicationCredential := getResponse.ApplicationCredential

	log.Printf("[DEBUG] Retrieved vnpaycloud_identity_application_credential %s: %#v", d.Id(), applicationCredential)

	d.Set("name", applicationCredential.Name)
	d.Set("description", applicationCredential.Description)
	d.Set("unrestricted", applicationCredential.Unrestricted)
	d.Set("roles", flattenIdentityApplicationCredentialRolesV3(applicationCredential.Roles))
	d.Set("access_rules", flattenIdentityApplicationCredentialAccessRulesV3(applicationCredential.AccessRules))
	d.Set("project_id", applicationCredential.ProjectID)
	d.Set("region", util.GetRegion(d, config))

	if applicationCredential.ExpiresAt == (time.Time{}) {
		d.Set("expires_at", "")
	} else {
		d.Set("expires_at", applicationCredential.ExpiresAt.UTC().Format(time.RFC3339))
	}

	return nil
}

func resourceIdentityApplicationCredentialV3Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	identityClient, err := client.NewClient(ctx, config.ConsoleClientConfig)
	if err != nil {
		return diag.Errorf("Error creating VNPAYCloud identity client: %s", err)
	}

	tokenInfo, err := util.GetTokenInfo(ctx, identityClient)
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = identityClient.Delete(ctx, client.ApiPath.ApplicationCredentialWithId(tokenInfo.UserID, d.Id()), nil)
	if err != nil {
		err = util.CheckDeleted(d, err, "Error deleting vnpaycloud_identity_application_credential")
		if err != nil {
			return diag.FromErr(err)
		}
	}

	// cleanup access rules
	accessRules := expandIdentityApplicationCredentialAccessRulesV3(d.Get("access_rules").(*schema.Set).List())
	return diag.FromErr(applicationCredentialCleanupAccessRulesV3(ctx, identityClient, tokenInfo.UserID, d.Id(), accessRules))
}
