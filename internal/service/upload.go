package service

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/gofreego/goutils/logger"
	"github.com/gofreego/mediabase/api/mediabase_v1"
	"github.com/google/uuid"
)

const (
	// Default expiry durations
	defaultUploadExpiry   = 60 * time.Second   // 60 seconds for upload
	defaultDownloadExpiry = 3600 * time.Second // 1 hour for download
)

// PresignUpload generates a presigned URL for uploading a file
func (s *Service) PresignUpload(ctx context.Context, req *mediabase_v1.PresignUploadRequest) (*mediabase_v1.PresignUploadResponse, error) {
	logger.Debug(ctx, "PresignUpload request received, bucket: %s, content_type: %s, max_file_size: %d", req.BucketName, req.ContentType, req.MaxFileSize)

	// Validate content type
	if !s.isValidContentType(req.ContentType) {
		return nil, fmt.Errorf("invalid content type: %s", req.ContentType)
	}

	// Validate requested max file size against server hard limit
	if req.MaxFileSize > s.maxFileSize {
		return nil, fmt.Errorf("requested max file size %d exceeds server maximum allowed size %d", req.MaxFileSize, s.maxFileSize)
	}

	// Generate unique object key
	objectKey := generateObjectKey(req.Path, req.FileName, req.ContentType)

	// Generate presigned URL/POST policy using the requested max size
	// This ensures the storage provider strictly enforces this exact limit
	presignedURL, formData, err := s.storage.GeneratePresignedUploadURL(ctx, req.BucketName, objectKey, req.ContentType, defaultUploadExpiry, req.MaxFileSize)
	if err != nil {
		logger.Error(ctx, "Failed to generate presigned upload URL: %v", err)
		return nil, fmt.Errorf("failed to generate presigned upload URL: %w", err)
	}

	logger.Debug(ctx, "Presigned upload URL generated successfully for object: %s in bucket: %s", objectKey, req.BucketName)

	return &mediabase_v1.PresignUploadResponse{
		PresignedUrl: presignedURL,
		ObjectKey:    objectKey,
		ExpiresIn:    int32(defaultUploadExpiry.Seconds()),
		FormData:     formData,
	}, nil
}

// PresignDownload generates a presigned URL for downloading a file
func (s *Service) PresignDownload(ctx context.Context, req *mediabase_v1.PresignDownloadRequest) (*mediabase_v1.PresignDownloadResponse, error) {
	logger.Debug(ctx, "PresignDownload request received, bucket: %s, object_key: %s", req.BucketName, req.ObjectKey)

	// Check if object exists
	exists, err := s.storage.ObjectExists(ctx, req.BucketName, req.ObjectKey)
	if err != nil {
		logger.Error(ctx, "Failed to check object existence: %v", err)
		return nil, fmt.Errorf("failed to check object existence: %w", err)
	}

	if !exists {
		return nil, fmt.Errorf("object not found: %s in bucket: %s", req.ObjectKey, req.BucketName)
	}

	// Generate presigned URL
	presignedURL, err := s.storage.GeneratePresignedDownloadURL(ctx, req.BucketName, req.ObjectKey, defaultDownloadExpiry)
	if err != nil {
		logger.Error(ctx, "Failed to generate presigned download URL: %v", err)
		return nil, fmt.Errorf("failed to generate presigned download URL: %w", err)
	}

	logger.Debug(ctx, "Presigned download URL generated successfully for object: %s", req.ObjectKey)

	return &mediabase_v1.PresignDownloadResponse{
		PresignedUrl: presignedURL,
		ExpiresIn:    int32(defaultDownloadExpiry.Seconds()),
	}, nil
}

// DeleteObject deletes a file from storage
func (s *Service) DeleteObject(ctx context.Context, req *mediabase_v1.DeleteObjectRequest) (*mediabase_v1.DeleteObjectResponse, error) {
	logger.Debug(ctx, "DeleteObject request received, bucket: %s, object_key: %s", req.BucketName, req.ObjectKey)

	// Delete the object
	err := s.storage.DeleteObject(ctx, req.BucketName, req.ObjectKey)
	if err != nil {
		logger.Error(ctx, "Failed to delete object: %v", err)
		return nil, fmt.Errorf("failed to delete object: %w", err)
	}

	logger.Debug(ctx, "Object deleted successfully: %s", req.ObjectKey)

	return &mediabase_v1.DeleteObjectResponse{
		Success: true,
	}, nil
}

// CreateBucket creates a bucket and optionally sets it to public read
func (s *Service) CreateBucket(ctx context.Context, req *mediabase_v1.CreateBucketRequest) (*mediabase_v1.CreateBucketResponse, error) {
	logger.Debug(ctx, "CreateBucket request received, bucket_name: %s, is_public: %v", req.BucketName, req.IsPublic)

	// Create bucket if it doesn't exist
	err := s.storage.CreateBucket(ctx, req.BucketName)
	if err != nil {
		logger.Error(ctx, "Failed to create bucket: %v", err)
		return nil, fmt.Errorf("failed to create bucket: %w", err)
	}

	if req.IsPublic {
		// Set public read policy
		// This policy allows s3:GetObject for all principals on all objects in the bucket
		policy := fmt.Sprintf(`{
			"Version": "2012-10-17",
			"Statement": [
				{
					"Effect": "Allow",
					"Principal": {"AWS": ["*"]},
					"Action": ["s3:GetObject"],
					"Resource": ["arn:aws:s3:::%s/*"]
				}
			]
		}`, req.BucketName)

		err = s.storage.SetBucketPolicy(ctx, req.BucketName, policy)
		if err != nil {
			logger.Error(ctx, "Failed to set bucket policy: %v", err)
			return nil, fmt.Errorf("failed to set bucket policy: %w", err)
		}
		logger.Debug(ctx, "Bucket created and policy set to public read: %s", req.BucketName)
	} else {
		logger.Debug(ctx, "Bucket created with private policy: %s", req.BucketName)
	}

	return &mediabase_v1.CreateBucketResponse{
		Success: true,
	}, nil
}

// Helper functions

// isValidContentType checks if the content type is allowed
func (s *Service) isValidContentType(contentType string) bool {
	return s.allowedContentTypes[contentType]
}

// generateObjectKey creates a unique object key with proper extension under the given path
func generateObjectKey(path, fileName, contentType string) string {
	var name string
	if fileName != "" {
		name = fileName
	} else {
		// Generate UUID for unique filename
		id := uuid.New().String()

		// Determine file extension based on content type
		var ext string
		switch contentType {
		case "image/jpeg":
			ext = ".jpg"
		case "image/png":
			ext = ".png"
		case "image/webp":
			ext = ".webp"
		default:
			ext = ".bin"
		}
		name = id + ext
	}

	if path != "" {
		return filepath.Join(path, name)
	}
	return name
}
