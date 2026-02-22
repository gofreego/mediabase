package storage

import (
	"context"
	"io"
	"time"
)

// Storage defines the interface for object storage operations
// This abstraction allows easy migration between different storage providers (MinIO, S3, GCS, etc.)
type Storage interface {
	// GeneratePresignedUploadURL creates a presigned URL or POST policy for uploading a file
	// Parameters:
	//   - ctx: context for the operation
	//   - bucketName: name of the bucket
	//   - objectKey: the key/path where the object will be stored
	//   - contentType: MIME type of the file
	//   - expiryDuration: how long the URL should remain valid
	//   - maxSize: maximum allowed file size in bytes (enforced by storage)
	// Returns:
	//   - URL string
	//   - Form data map (for POST uploads)
	//   - error if operation fails
	GeneratePresignedUploadURL(ctx context.Context, bucketName, objectKey, contentType string, expiryDuration time.Duration, maxSize int64) (string, map[string]string, error)

	// GeneratePresignedDownloadURL creates a presigned URL for downloading a file
	// Parameters:
	//   - ctx: context for the operation
	//   - bucketName: name of the bucket
	//   - objectKey: the key/path of the object to download
	//   - expiryDuration: how long the URL should remain valid
	// Returns:
	//   - presigned URL string
	//   - error if operation fails
	GeneratePresignedDownloadURL(ctx context.Context, bucketName, objectKey string, expiryDuration time.Duration) (string, error)

	// DeleteObject removes a file from storage
	// Parameters:
	//   - ctx: context for the operation
	//   - bucketName: name of the bucket
	//   - objectKey: the key/path of the object to delete
	// Returns:
	//   - error if operation fails
	DeleteObject(ctx context.Context, bucketName, objectKey string) error

	// PutObject uploads a file directly to storage
	// Parameters:
	//   - ctx: context for the operation
	//   - bucketName: name of the bucket
	//   - objectKey: the key/path where the object will be stored
	//   - reader: data stream to upload
	//   - objectSize: size of the object in bytes
	//   - contentType: MIME type of the file
	// Returns:
	//   - error if operation fails
	PutObject(ctx context.Context, bucketName, objectKey string, reader io.Reader, objectSize int64, contentType string) error

	// GetObject downloads a file from storage
	// Parameters:
	//   - ctx: context for the operation
	//   - bucketName: name of the bucket
	//   - objectKey: the key/path of the object to download
	// Returns:
	//   - reader for the object data
	//   - error if operation fails
	GetObject(ctx context.Context, bucketName, objectKey string) (io.ReadCloser, error)

	// ObjectExists checks if an object exists in storage
	// Parameters:
	//   - ctx: context for the operation
	//   - bucketName: name of the bucket
	//   - objectKey: the key/path of the object to check
	// Returns:
	//   - true if object exists, false otherwise
	//   - error if operation fails
	ObjectExists(ctx context.Context, bucketName, objectKey string) (bool, error)

	// CreateBucket creates a new bucket if it doesn't exist
	// Parameters:
	//   - ctx: context for the operation
	//   - bucketName: name of the bucket to create
	// Returns:
	//   - error if operation fails
	CreateBucket(ctx context.Context, bucketName string) error

	// SetBucketPolicy sets the access policy for a bucket
	// Parameters:
	//   - ctx: context for the operation
	//   - bucketName: name of the bucket
	//   - policy: JSON policy string
	// Returns:
	//   - error if operation fails
	SetBucketPolicy(ctx context.Context, bucketName string, policy string) error
}

// Config holds common configuration for storage providers
type Config struct {
	Endpoint        string `yaml:"Endpoint"`
	AccessKeyID     string `yaml:"AccessKeyID"`
	SecretAccessKey string `yaml:"SecretAccessKey"`
	Region          string `yaml:"Region"`
	UseSSL          bool   `yaml:"UseSSL"`
}
