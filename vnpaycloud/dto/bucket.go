package dto

// S3Bucket matches the backend S3Bucket proto message.
type S3Bucket struct {
	BucketName  string `json:"bucketName"`
	Region      string `json:"region"`
	CreatedAt   string `json:"createdAt"`
	PolicyName  string `json:"policyName"`
	SizeBytes   uint64 `json:"sizeBytes,string"`
	ObjectCount uint64 `json:"objectCount,string"`
}

// CreateBucketRequest matches the backend CreateBucketRequest proto message.
// project_id is passed via URL path, not in the body.
type CreateBucketRequest struct {
	BucketName       string `json:"bucketName"`
	Region           string `json:"region"`
	StoragePolicyID  string `json:"storagePolicyId,omitempty"`
	EnableObjectLock bool   `json:"enableObjectLock,omitempty"`
}

// ListBucketsResponse matches the backend ListBucketsResponse proto message.
type ListBucketsResponse struct {
	Buckets []S3Bucket `json:"buckets"`
}

// GetBucketUsageResponse matches the backend GetBucketUsageResponse proto message.
type GetBucketUsageResponse struct {
	Bucket S3Bucket `json:"bucket"`
}
