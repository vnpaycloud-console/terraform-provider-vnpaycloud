package bucket

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestDataSourceBucketRead_ByName(t *testing.T) {
	b := testBucket()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/buckets/test-bucket/usage",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.GetBucketUsageResponse{
				Bucket: dto.S3Bucket{SizeBytes: 2048, ObjectCount: 25},
			}),
		},
		{
			Method: "GET",
			Pattern: "/v2/iac/projects/test-project-id/buckets",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListBucketsResponse{
				Buckets: []dto.S3Bucket{b},
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceBucket()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"bucket_name": "test-bucket",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if d.Id() != "test-bucket" {
		t.Errorf("expected ID test-bucket, got %s", d.Id())
	}
	if v := d.Get("bucket_name").(string); v != "test-bucket" {
		t.Errorf("expected bucket_name test-bucket, got %s", v)
	}
	if v := d.Get("region").(string); v != "vn-hanoi-1" {
		t.Errorf("expected region vn-hanoi-1, got %s", v)
	}
	if v := d.Get("created_at").(string); v != "2025-01-15T10:00:00Z" {
		t.Errorf("expected created_at 2025-01-15T10:00:00Z, got %s", v)
	}
	if v := d.Get("policy_name").(string); v != "standard" {
		t.Errorf("expected policy_name standard, got %s", v)
	}
	if v := d.Get("size_bytes").(int); v != 2048 {
		t.Errorf("expected size_bytes 2048, got %d", v)
	}
	if v := d.Get("object_count").(int); v != 25 {
		t.Errorf("expected object_count 25, got %d", v)
	}
}

func TestDataSourceBucketRead_NotFound(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method:  "GET",
			Pattern: "/v2/iac/projects/test-project-id/buckets/nonexistent/usage",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.GetBucketUsageResponse{
				Bucket: dto.S3Bucket{},
			}),
		},
		{
			Method: "GET",
			Pattern: "/v2/iac/projects/test-project-id/buckets",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListBucketsResponse{
				Buckets: []dto.S3Bucket{},
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceBucket()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{
		"bucket_name": "nonexistent",
	})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if !diags.HasError() {
		t.Fatal("expected error for nonexistent bucket, got none")
	}
}

func TestDataSourceBucketsRead(t *testing.T) {
	b1 := testBucket()
	b2 := dto.S3Bucket{
		BucketName: "logs-bucket",
		Region:     "vn-hcm-1",
		CreatedAt:  "2025-02-01T08:00:00Z",
		PolicyName: "archive",
	}

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method: "GET",
			Pattern: "/v2/iac/projects/test-project-id/buckets",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListBucketsResponse{
				Buckets: []dto.S3Bucket{b1, b2},
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	ds := DataSourceBuckets()
	d := schema.TestResourceDataRaw(t, ds.Schema, map[string]interface{}{})

	diags := ds.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	buckets := d.Get("buckets").([]interface{})
	if len(buckets) != 2 {
		t.Fatalf("expected 2 buckets, got %d", len(buckets))
	}

	first := buckets[0].(map[string]interface{})
	if first["bucket_name"] != "test-bucket" {
		t.Errorf("expected first bucket name test-bucket, got %v", first["bucket_name"])
	}
	if first["region"] != "vn-hanoi-1" {
		t.Errorf("expected first bucket region vn-hanoi-1, got %v", first["region"])
	}
	if first["created_at"] != "2025-01-15T10:00:00Z" {
		t.Errorf("expected first bucket created_at 2025-01-15T10:00:00Z, got %v", first["created_at"])
	}
	if first["policy_name"] != "standard" {
		t.Errorf("expected first bucket policy_name standard, got %v", first["policy_name"])
	}

	second := buckets[1].(map[string]interface{})
	if second["bucket_name"] != "logs-bucket" {
		t.Errorf("expected second bucket name logs-bucket, got %v", second["bucket_name"])
	}
	if second["region"] != "vn-hcm-1" {
		t.Errorf("expected second bucket region vn-hcm-1, got %v", second["region"])
	}
	if second["policy_name"] != "archive" {
		t.Errorf("expected second bucket policy_name archive, got %v", second["policy_name"])
	}
}
