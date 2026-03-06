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

func DataSourceBucket() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceBucketRead,
		Schema: map[string]*schema.Schema{
			"bucket_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"region": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"policy_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"size_bytes": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"object_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func DataSourceBuckets() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceBucketsRead,
		Schema: map[string]*schema.Schema{
			"buckets": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bucket_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"region": {
							Type:     schema.TypeString,
							Computed: true,
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
				},
			},
		},
	}
}

func dataSourceBucketRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	bucketName := d.Get("bucket_name").(string)

	tflog.Debug(ctx, "vnpaycloud_bucket data source read", map[string]interface{}{"bucket_name": bucketName})

	// Get usage for size/object count
	usageResp := &dto.GetBucketUsageResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.BucketUsage(cfg.ProjectID, bucketName), usageResp, nil)
	if err != nil {
		return diag.Errorf("Error reading vnpaycloud_bucket %s: %s", bucketName, err)
	}

	// Get bucket metadata (region, policy_name) from list
	bucket, err := findBucketByName(ctx, cfg, bucketName)
	if err != nil {
		return diag.FromErr(err)
	}
	if bucket == nil {
		return diag.Errorf("Bucket %s not found", bucketName)
	}

	d.SetId(bucketName)
	d.Set("bucket_name", bucket.BucketName)
	d.Set("region", bucket.Region)
	d.Set("created_at", bucket.CreatedAt)
	d.Set("policy_name", bucket.PolicyName)
	d.Set("size_bytes", int(usageResp.Bucket.SizeBytes))
	d.Set("object_count", int(usageResp.Bucket.ObjectCount))

	return nil
}

func dataSourceBucketsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	tflog.Debug(ctx, "vnpaycloud_buckets data source read")

	listResp := &dto.ListBucketsResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.Buckets(cfg.ProjectID), listResp, nil)
	if err != nil {
		return diag.Errorf("Error listing vnpaycloud_buckets: %s", err)
	}

	var buckets []map[string]interface{}
	for _, b := range listResp.Buckets {
		buckets = append(buckets, map[string]interface{}{
			"bucket_name": b.BucketName,
			"region":      b.Region,
			"created_at":  b.CreatedAt,
			"policy_name": b.PolicyName,
		})
	}

	d.SetId(fmt.Sprintf("buckets-%s", cfg.ProjectID))
	d.Set("buckets", buckets)

	return nil
}
