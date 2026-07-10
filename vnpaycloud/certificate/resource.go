package certificate

import (
	"context"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	typeSelfSigned = "self_signed"
	typeImport     = "import"
	typeImportCA   = "import_ca"
)

func certTypeToResourceType(certType string) string {
	switch certType {
	case "CT_SELF_SIGNED":
		return typeSelfSigned
	case "CT_CA":
		return typeImportCA
	default:
		return typeImport
	}
}

func ResourceCertificate() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCertificateCreate,
		ReadContext:   resourceCertificateRead,
		DeleteContext: resourceCertificateDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{typeSelfSigned, typeImport, typeImportCA}, false),
				Description:  "How the certificate is created: self_signed, import, or import_ca.",
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(nameRegexp, "must be 3-255 chars: letters, digits, space, '-', '_', '.'"),
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"country":           {Type: schema.TypeString, Optional: true, ForceNew: true},
			"province":          {Type: schema.TypeString, Optional: true, ForceNew: true},
			"city":              {Type: schema.TypeString, Optional: true, ForceNew: true},
			"organization":      {Type: schema.TypeString, Optional: true, ForceNew: true},
			"organization_unit": {Type: schema.TypeString, Optional: true, ForceNew: true},
			"domain_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Description: "Common Name (CN) / domain for the certificate.",
			},
			"email": {Type: schema.TypeString, Optional: true, ForceNew: true},
			"key_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"2048 RSA", "4096 RSA"}, false),
				Description:  "Private key type for self_signed (default `4096 RSA`).",
			},
			"expiration": {
				Type:        schema.TypeInt,
				Optional:    true,
				ForceNew:    true,
				Description: "Lifetime of a self-signed certificate in days (default `3650`).",
			},
			"digest_algorithm": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"SHA-256", "SHA-512"}, false),
				Description:  "Signing digest for self_signed (default `SHA-512`).",
			},

			"private_key": {
				Type:      schema.TypeString,
				Optional:  true,
				ForceNew:  true,
				Sensitive: true,
			},
			"certificate_body": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Sensitive:   true,
				Description: "PEM certificate body. Required for import and import_ca.",
			},
			"intermediate_ca": {
				Type:      schema.TypeString,
				Optional:  true,
				ForceNew:  true,
				Sensitive: true,
			},

			"cert_type":         {Type: schema.TypeString, Computed: true},
			"expires_at":        {Type: schema.TypeString, Computed: true},
			"status":            {Type: schema.TypeString, Computed: true},
			"zone_id":           {Type: schema.TypeString, Computed: true},
			"load_balancer_ids": {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
		},
	}
}

func resourceCertificateCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	certResp := &dto.CertificateResponse{}

	switch d.Get("type").(string) {
	case typeSelfSigned:
		var missing []string
		for _, f := range []string{"country", "province", "organization", "organization_unit", "domain_name", "email"} {
			if d.Get(f).(string) == "" {
				missing = append(missing, f)
			}
		}
		if len(missing) > 0 {
			return diag.Errorf("%v are required when type = self_signed", missing)
		}
		keyType := d.Get("key_type").(string)
		if keyType == "" {
			keyType = "4096 RSA"
		}
		expiration := d.Get("expiration").(int)
		if expiration == 0 {
			expiration = 3650
		}
		digest := d.Get("digest_algorithm").(string)
		if digest == "" {
			digest = "SHA-512"
		}
		opts := dto.CreateSelfSignedCertificateRequest{
			Name:             d.Get("name").(string),
			Description:      d.Get("description").(string),
			Country:          d.Get("country").(string),
			Province:         d.Get("province").(string),
			City:             d.Get("city").(string),
			Organization:     d.Get("organization").(string),
			OrganizationUnit: d.Get("organization_unit").(string),
			DomainName:       d.Get("domain_name").(string),
			Email:            d.Get("email").(string),
			KeyType:          keyType,
			Expiration:       expiration,
			DigestAlgorithm:  digest,
		}
		tflog.Debug(ctx, "vnpaycloud_certificate self-signed create", map[string]interface{}{"name": opts.Name})
		if _, err := cfg.Client.Post(ctx, client.ApiPath.CertificateSelfSigned(cfg.ProjectID), opts, certResp, nil); err != nil {
			return diag.Errorf("Error creating self-signed vnpaycloud_certificate: %s", err)
		}

	case typeImport:
		if d.Get("private_key").(string) == "" || d.Get("certificate_body").(string) == "" {
			return diag.Errorf("`private_key` and `certificate_body` are required when type = import")
		}
		opts := dto.UploadCertificateRequest{
			Name:           d.Get("name").(string),
			Description:    d.Get("description").(string),
			PrivateKey:     d.Get("private_key").(string),
			Cert:           d.Get("certificate_body").(string),
			IntermediateCA: d.Get("intermediate_ca").(string),
		}
		if _, err := cfg.Client.Post(ctx, client.ApiPath.CertificateUpload(cfg.ProjectID), opts, certResp, nil); err != nil {
			return diag.Errorf("Error uploading vnpaycloud_certificate: %s", err)
		}

	case typeImportCA:
		if d.Get("certificate_body").(string) == "" {
			return diag.Errorf("`certificate_body` is required when type = import_ca")
		}
		opts := dto.UploadCACertificateRequest{
			Name:        d.Get("name").(string),
			Description: d.Get("description").(string),
			Cert:        d.Get("certificate_body").(string),
		}
		if _, err := cfg.Client.Post(ctx, client.ApiPath.CertificateUploadCA(cfg.ProjectID), opts, certResp, nil); err != nil {
			return diag.Errorf("Error uploading CA vnpaycloud_certificate: %s", err)
		}
	}

	d.SetId(certResp.Certificate.ID)
	return resourceCertificateRead(ctx, d, meta)
}

func resourceCertificateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	certResp := &dto.CertificateResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.CertificateWithID(cfg.ProjectID, d.Id()), certResp, nil)
	if err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error retrieving vnpaycloud_certificate"))
	}

	c := certResp.Certificate
	tflog.Debug(ctx, "Retrieved vnpaycloud_certificate "+d.Id(), map[string]interface{}{"certificate": c})

	d.Set("type", certTypeToResourceType(c.CertType))
	d.Set("name", c.Name)
	// The backend list/get does not echo description or domain_name, so only
	// overwrite them when the API actually returns a value — otherwise we would
	// clobber the configured (ForceNew) values and trigger a spurious replace.
	if c.Description != "" {
		d.Set("description", c.Description)
	}
	if c.DomainName != "" {
		d.Set("domain_name", c.DomainName)
	}
	d.Set("cert_type", c.CertType)
	d.Set("expires_at", c.Expiration)
	d.Set("status", c.Status)
	d.Set("zone_id", c.ZoneID)
	d.Set("load_balancer_ids", c.LoadBalancerIDs)

	return nil
}

func resourceCertificateDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if _, err := cfg.Client.Delete(ctx, client.ApiPath.CertificateWithID(cfg.ProjectID, d.Id()), nil); err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error deleting vnpaycloud_certificate"))
	}

	return nil
}
