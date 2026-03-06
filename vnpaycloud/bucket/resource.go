package bucket

import (
	"context"
	"fmt"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceBucket() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceBucketCreate,
		ReadContext:   resourceBucketRead,
		DeleteContext: resourceBucketDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"bucket_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the S3 bucket.",
			},
			"region": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Region where the bucket is created.",
			},
			"storage_policy_id": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Storage policy ID for the bucket.",
			},
			"enable_object_lock": {
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    true,
				Description: "Enable object lock on the bucket.",
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"policy_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceBucketCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	createOpts := dto.CreateBucketRequest{
		BucketName:       d.Get("bucket_name").(string),
		Region:           d.Get("region").(string),
		StoragePolicyID:  d.Get("storage_policy_id").(string),
		EnableObjectLock: d.Get("enable_object_lock").(bool),
	}

	tflog.Debug(ctx, "vnpaycloud_bucket create", map[string]interface{}{"opts": fmt.Sprintf("%+v", createOpts)})

	_, err := cfg.Client.Post(ctx, client.ApiPath.Buckets(cfg.ProjectID), createOpts, nil, nil)
	if err != nil {
		return diag.Errorf("Error creating vnpaycloud_bucket: %s", err)
	}

	d.SetId(createOpts.BucketName)

	return resourceBucketRead(ctx, d, meta)
}

func resourceBucketRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	bucketName := d.Id()

	tflog.Debug(ctx, "vnpaycloud_bucket read", map[string]interface{}{"bucket_name": bucketName})

	bucket, err := findBucketByName(ctx, cfg, bucketName)
	if err != nil {
		return diag.FromErr(err)
	}
	if bucket == nil {
		d.SetId("")
		return nil
	}

	d.Set("bucket_name", bucket.BucketName)
	d.Set("region", bucket.Region)
	d.Set("created_at", bucket.CreatedAt)
	d.Set("policy_name", bucket.PolicyName)

	return nil
}

func resourceBucketDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	bucketName := d.Id()
	region := d.Get("region").(string)

	tflog.Debug(ctx, "vnpaycloud_bucket delete", map[string]interface{}{"bucket_name": bucketName, "region": region})

	_, err := cfg.Client.Delete(ctx, client.ApiPath.BucketDelete(cfg.ProjectID, bucketName, region), nil)
	if err != nil {
		return diag.Errorf("Error deleting vnpaycloud_bucket %s: %s", bucketName, err)
	}

	return nil
}

// findBucketByName lists all buckets and finds one by name.
func findBucketByName(ctx context.Context, cfg *config.Config, bucketName string) (*dto.S3Bucket, error) {
	listResp := &dto.ListBucketsResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.Buckets(cfg.ProjectID), listResp, nil)
	if err != nil {
		return nil, fmt.Errorf("error listing buckets: %s", err)
	}

	for _, b := range listResp.Buckets {
		if b.BucketName == bucketName {
			return &b, nil
		}
	}

	return nil, nil
}
