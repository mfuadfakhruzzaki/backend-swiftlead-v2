package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/swiftlead/backend-swiftlet/internal/config"
	"github.com/swiftlead/backend-swiftlet/pkg/logger"
)

// MinIOClient wraps the MinIO client
type MinIOClient struct {
	client    *minio.Client
	bucket    string
	publicURL string
}

// NewMinIOClient creates a new MinIO client
func NewMinIOClient(cfg *config.Config) (*MinIOClient, error) {
	client, err := minio.New(cfg.MinIOEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinIOAccessKey, cfg.MinIOSecretKey, ""),
		Secure: cfg.MinIOUseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	mc := &MinIOClient{
		client:    client,
		bucket:    cfg.MinIOBucket,
		publicURL: cfg.MinIOPublicURL,
	}

	// Ensure bucket exists
	if err := mc.ensureBucket(context.Background()); err != nil {
		return nil, err
	}

	logger.Info("Connected to MinIO at %s", cfg.MinIOEndpoint)
	return mc, nil
}

// ensureBucket creates the bucket if it doesn't exist
func (m *MinIOClient) ensureBucket(ctx context.Context) error {
	exists, err := m.client.BucketExists(ctx, m.bucket)
	if err != nil {
		return fmt.Errorf("failed to check bucket: %w", err)
	}

	if !exists {
		if err := m.client.MakeBucket(ctx, m.bucket, minio.MakeBucketOptions{}); err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
		logger.Info("Created MinIO bucket: %s", m.bucket)
	}

	return nil
}

// Upload uploads a file to MinIO
func (m *MinIOClient) Upload(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error) {
	_, err := m.client.PutObject(ctx, m.bucket, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload object: %w", err)
	}

	// Return public URL
	return m.GetPublicURL(objectName), nil
}

// UploadFromFile uploads a file from local path
func (m *MinIOClient) UploadFromFile(ctx context.Context, objectName, filePath, contentType string) (string, error) {
	_, err := m.client.FPutObject(ctx, m.bucket, objectName, filePath, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	return m.GetPublicURL(objectName), nil
}

// GetPresignedURL generates a presigned URL for downloading
func (m *MinIOClient) GetPresignedURL(ctx context.Context, objectName string, expiry time.Duration) (string, error) {
	presignedURL, err := m.client.PresignedGetObject(ctx, m.bucket, objectName, expiry, url.Values{})
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}
	return presignedURL.String(), nil
}

// GetPublicURL returns the public URL for an object
func (m *MinIOClient) GetPublicURL(objectName string) string {
	if m.publicURL != "" {
		return fmt.Sprintf("%s/%s/%s", m.publicURL, m.bucket, objectName)
	}
	return fmt.Sprintf("http://%s/%s/%s", m.client.EndpointURL().Host, m.bucket, objectName)
}

// Delete removes an object from MinIO
func (m *MinIOClient) Delete(ctx context.Context, objectName string) error {
	err := m.client.RemoveObject(ctx, m.bucket, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}
	return nil
}

// ObjectExists checks if an object exists
func (m *MinIOClient) ObjectExists(ctx context.Context, objectName string) (bool, error) {
	_, err := m.client.StatObject(ctx, m.bucket, objectName, minio.StatObjectOptions{})
	if err != nil {
		errResp := minio.ToErrorResponse(err)
		if errResp.Code == "NoSuchKey" {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GenerateObjectName generates a unique object name
func GenerateObjectName(prefix, filename string) string {
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("%s/%d_%s", prefix, timestamp, filename)
}
