package bucket

import (
	"context"
	"net/http"
	"testing"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// testBucket returns a fully populated dto.S3Bucket for use in tests.
func testBucket() dto.S3Bucket {
	return dto.S3Bucket{
		BucketName:  "test-bucket",
		Region:      "vn-hanoi-1",
		CreatedAt:   "2025-01-15T10:00:00Z",
		PolicyName:  "standard",
		SizeBytes:   1024,
		ObjectCount: 10,
	}
}

func TestResourceBucketCreate(t *testing.T) {
	b := testBucket()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Pattern: "/v2/iac/projects/test-project-id/buckets",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case "POST":
					w.WriteHeader(http.StatusOK)
				case "GET":
					testhelpers.JSONHandler(t, http.StatusOK, dto.ListBucketsResponse{
						Buckets: []dto.S3Bucket{b},
					})(w, r)
				default:
					http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				}
			},
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceBucket()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"bucket_name":        "test-bucket",
		"region":             "vn-hanoi-1",
		"storage_policy_id":  "policy-001",
		"enable_object_lock": false,
	})

	diags := res.CreateContext(context.Background(), d, cfg)
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
}

func TestResourceBucketRead(t *testing.T) {
	b := testBucket()

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method: "GET",
			Pattern: "/v2/iac/projects/test-project-id/buckets",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListBucketsResponse{
				Buckets: []dto.S3Bucket{b},
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceBucket()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"bucket_name":        "",
		"region":             "",
		"storage_policy_id":  "",
		"enable_object_lock": false,
	})
	d.SetId("test-bucket")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
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
}

func TestResourceBucketRead_NotFound(t *testing.T) {
	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method: "GET",
			Pattern: "/v2/iac/projects/test-project-id/buckets",
			Handler: testhelpers.JSONHandler(t, http.StatusOK, dto.ListBucketsResponse{
				Buckets: []dto.S3Bucket{},
			}),
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceBucket()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"bucket_name":        "",
		"region":             "",
		"storage_policy_id":  "",
		"enable_object_lock": false,
	})
	d.SetId("nonexistent-bucket")

	diags := res.ReadContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error on not found: %v", diags)
	}

	if d.Id() != "" {
		t.Errorf("expected resource ID to be cleared when bucket not found, got %s", d.Id())
	}
}

func TestResourceBucketDelete(t *testing.T) {
	deletedCalled := false

	srv := testhelpers.NewMockServer(t, []testhelpers.Route{
		{
			Method: "DELETE",
			Pattern: "/v2/iac/projects/test-project-id/buckets/test-bucket",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				deletedCalled = true
				w.WriteHeader(http.StatusNoContent)
			},
		},
	})
	cfg := testhelpers.NewMockConfig(t, srv.URL)

	res := ResourceBucket()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{
		"bucket_name":        "test-bucket",
		"region":             "vn-hanoi-1",
		"storage_policy_id":  "",
		"enable_object_lock": false,
	})
	d.SetId("test-bucket")

	diags := res.DeleteContext(context.Background(), d, cfg)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if !deletedCalled {
		t.Error("expected DELETE to have been called")
	}
}
