package minio

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/gofreego/mediabase/internal/storage"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinIOStorage implements the Storage interface using MinIO
type MinIOStorage struct {
	client *minio.Client
}

// NewMinIOStorage creates a new MinIO storage instance
func NewMinIOStorage(config storage.Config) (*MinIOStorage, error) {
	// Initialize MinIO client
	minioClient, err := minio.New(config.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKeyID, config.SecretAccessKey, ""),
		Secure: config.UseSSL,
		Region: config.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	minioClient.TraceOn(os.Stdout)
	return &MinIOStorage{
		client: minioClient,
	}, nil
}

// GeneratePresignedUploadURL creates a presigned POST policy for uploading a file with size constraints
func (m *MinIOStorage) GeneratePresignedUploadURL(ctx context.Context, bucketName, objectKey, contentType string, expiryDuration time.Duration, maxSize int64) (string, map[string]string, error) {
	// Create post policy
	policy := minio.NewPostPolicy()
	policy.SetBucket(bucketName)
	policy.SetKey(objectKey)
	policy.SetExpires(time.Now().Add(expiryDuration))
	policy.SetContentType(contentType)

	// Enforce size limit at the storage level
	policy.SetContentLengthRange(0, maxSize)

	// Generate presigned POST URL and form fields
	u, formData, err := m.client.PresignedPostPolicy(ctx, policy)
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate presigned post policy: %w", err)
	}

	// Flatten form data for the response
	fields := make(map[string]string)
	for k, v := range formData {
		fields[k] = v
	}

	return u.String(), fields, nil
}

// GeneratePresignedDownloadURL creates a presigned URL for downloading a file
func (m *MinIOStorage) GeneratePresignedDownloadURL(ctx context.Context, bucketName, objectKey string, expiryDuration time.Duration) (string, error) {
	// Generate presigned GET URL
	presignedURL, err := m.client.PresignedGetObject(ctx, bucketName, objectKey, expiryDuration, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned download URL: %w", err)
	}

	return presignedURL.String(), nil
}

// DeleteObject removes a file from storage
func (m *MinIOStorage) DeleteObject(ctx context.Context, bucketName, objectKey string) error {
	err := m.client.RemoveObject(ctx, bucketName, objectKey, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}

	return nil
}

// PutObject uploads a file directly to storage
func (m *MinIOStorage) PutObject(ctx context.Context, bucketName, objectKey string, reader io.Reader, objectSize int64, contentType string) error {
	_, err := m.client.PutObject(ctx, bucketName, objectKey, reader, objectSize, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("failed to put object: %w", err)
	}

	return nil
}

// GetObject downloads a file from storage
func (m *MinIOStorage) GetObject(ctx context.Context, bucketName, objectKey string) (io.ReadCloser, error) {
	object, err := m.client.GetObject(ctx, bucketName, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}

	return object, nil
}

// ObjectExists checks if an object exists in storage
func (m *MinIOStorage) ObjectExists(ctx context.Context, bucketName, objectKey string) (bool, error) {
	_, err := m.client.StatObject(ctx, bucketName, objectKey, minio.StatObjectOptions{})
	if err != nil {
		// Check if error is "not found"
		errResponse := minio.ToErrorResponse(err)
		if errResponse.Code == "NoSuchKey" {
			return false, nil
		}
		return false, fmt.Errorf("failed to check object existence: %w", err)
	}

	return true, nil
}

// CreateBucket creates a new bucket if it doesn't exist
func (m *MinIOStorage) CreateBucket(ctx context.Context, bucketName string) error {
	exists, err := m.client.BucketExists(ctx, bucketName)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		err = m.client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
	}
	return nil
}

// SetBucketPolicy sets the access policy for a bucket
func (m *MinIOStorage) SetBucketPolicy(ctx context.Context, bucketName string, policy string) error {
	err := m.client.SetBucketPolicy(ctx, bucketName, policy)
	if err != nil {
		return fmt.Errorf("failed to set bucket policy: %w", err)
	}
	return nil
}
