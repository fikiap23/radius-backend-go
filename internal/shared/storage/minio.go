package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/radius/radius-backend/internal/shared/config"
)

const defaultRegion = "us-east-1"

// Client wraps dual MinIO clients: internal (Docker network) and presign (browser-facing).
type Client struct {
	internal    *minio.Client
	presign     *minio.Client
	publicURL   string
	bucketName  string
	presignTTL  time.Duration
}

func NewClient(cfg config.MinIOConfig) (*Client, error) {
	endpoint := cfg.Endpoint
	if cfg.Port > 0 && !strings.Contains(endpoint, ":") {
		endpoint = fmt.Sprintf("%s:%d", endpoint, cfg.Port)
	}

	internal, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
		Region: defaultRegion,
	})
	if err != nil {
		return nil, fmt.Errorf("minio internal client: %w", err)
	}

	presignEndpoint := strings.TrimSpace(cfg.PresignEndpoint)
	if presignEndpoint == "" {
		presignEndpoint = cfg.PublicURL
	}

	presign, err := buildPresignClient(presignEndpoint, cfg.AccessKey, cfg.SecretKey)
	if err != nil {
		return nil, fmt.Errorf("minio presign client: %w", err)
	}

	ttl := cfg.PresignExpiry
	if ttl <= 0 {
		ttl = 15 * time.Minute
	}

	return &Client{
		internal:   internal,
		presign:    presign,
		publicURL:  strings.TrimRight(cfg.PublicURL, "/"),
		bucketName: cfg.BucketName,
		presignTTL: ttl,
	}, nil
}

func buildPresignClient(endpoint, accessKey, secretKey string) (*minio.Client, error) {
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("parse presign endpoint %q: %w", endpoint, err)
	}

	port := parsed.Port()
	if port == "" {
		if parsed.Scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}

	return minio.New(parsed.Hostname()+":"+port, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: parsed.Scheme == "https",
		Region: defaultRegion,
	})
}

func (c *Client) BucketName() string {
	return c.bucketName
}

func (c *Client) PresignExpiry() time.Duration {
	return c.presignTTL
}

func (c *Client) EnsurePublicBucket(ctx context.Context) error {
	if err := c.ensureBucket(ctx, c.bucketName); err != nil {
		return err
	}
	return c.setPublicReadPolicy(ctx, c.bucketName)
}

func (c *Client) ensureBucket(ctx context.Context, bucket string) error {
	exists, err := c.internal.BucketExists(ctx, bucket)
	if err != nil {
		return fmt.Errorf("check bucket %q: %w", bucket, err)
	}
	if exists {
		return nil
	}
	if err := c.internal.MakeBucket(ctx, bucket, minio.MakeBucketOptions{Region: defaultRegion}); err != nil {
		return fmt.Errorf("create bucket %q: %w", bucket, err)
	}
	return nil
}

func (c *Client) setPublicReadPolicy(ctx context.Context, bucket string) error {
	policy := map[string]any{
		"Version": "2012-10-17",
		"Statement": []map[string]any{
			{
				"Effect":    "Allow",
				"Principal": "*",
				"Action":    []string{"s3:GetObject"},
				"Resource":  []string{fmt.Sprintf("arn:aws:s3:::%s/*", bucket)},
			},
		},
	}
	raw, err := json.Marshal(policy)
	if err != nil {
		return fmt.Errorf("marshal bucket policy: %w", err)
	}
	if err := c.internal.SetBucketPolicy(ctx, bucket, string(raw)); err != nil {
		return fmt.Errorf("set bucket policy: %w", err)
	}
	return nil
}

// BuildTempObjectKey generates a server-side temp key: temp/YYYY-MM-DD/<uuid><ext>.
func BuildTempObjectKey(originalFilename string) string {
	return BuildObjectKey(originalFilename, "temp")
}

func BuildObjectKey(originalFilename, folder string) string {
	ext := SafeExtension(originalFilename)
	date := time.Now().UTC().Format("2006-01-02")
	return fmt.Sprintf("%s/%s/%s%s", folder, date, uuid.NewString(), ext)
}

func SafeExtension(filename string) string {
	ext := strings.ToLower(path.Ext(filename))
	if ext == "" {
		return ""
	}
	if len(ext) < 2 || len(ext) > 11 {
		return ""
	}
	for _, ch := range ext[1:] {
		if (ch < 'a' || ch > 'z') && (ch < '0' || ch > '9') {
			return ""
		}
	}
	return ext
}

type PresignedUpload struct {
	BucketName         string
	Key                string
	UploadURL          string
	ExpiresInSeconds   int
}

func (c *Client) CreatePresignedUploadURL(ctx context.Context, originalFilename string) (*PresignedUpload, error) {
	if err := c.ensureBucket(ctx, c.bucketName); err != nil {
		return nil, err
	}

	objectKey := BuildTempObjectKey(originalFilename)
	expiry := c.presignTTL

	uploadURL, err := c.presign.PresignedPutObject(ctx, c.bucketName, objectKey, expiry)
	if err != nil {
		return nil, fmt.Errorf("presign put object: %w", err)
	}

	return &PresignedUpload{
		BucketName:       c.bucketName,
		Key:              objectKey,
		UploadURL:        uploadURL.String(),
		ExpiresInSeconds: int(expiry.Seconds()),
	}, nil
}

func (c *Client) PublicURLByKey(key string) string {
	return fmt.Sprintf("%s/%s/%s", c.publicURL, c.bucketName, key)
}

func (c *Client) ObjectExists(ctx context.Context, key string) (bool, error) {
	_, err := c.internal.StatObject(ctx, c.bucketName, key, minio.StatObjectOptions{})
	if err != nil {
		resp := minio.ToErrorResponse(err)
		if resp.Code == "NoSuchKey" || resp.StatusCode == 404 {
			return false, nil
		}
		return false, fmt.Errorf("stat object %q: %w", key, err)
	}
	return true, nil
}

func (c *Client) MoveObject(ctx context.Context, sourceKey, destinationKey string) error {
	if err := c.ensureBucket(ctx, c.bucketName); err != nil {
		return err
	}

	exists, err := c.ObjectExists(ctx, sourceKey)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("source object not found: %s", sourceKey)
	}

	src := minio.CopySrcOptions{Bucket: c.bucketName, Object: sourceKey}
	dst := minio.CopyDestOptions{Bucket: c.bucketName, Object: destinationKey}
	if _, err := c.internal.CopyObject(ctx, dst, src); err != nil {
		return fmt.Errorf("copy object: %w", err)
	}
	if err := c.internal.RemoveObject(ctx, c.bucketName, sourceKey, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("remove source object: %w", err)
	}
	return nil
}

func ValidateTempKey(key string) error {
	if !strings.HasPrefix(key, "temp/") {
		return fmt.Errorf("invalid temp key: %s", key)
	}
	return nil
}

type ObjectInfo struct {
	Key          string
	LastModified time.Time
}

func (c *Client) PromoteTempObject(ctx context.Context, tempKey, destinationKey string) (string, error) {
	if err := ValidateTempKey(tempKey); err != nil {
		return "", err
	}
	if err := c.MoveObject(ctx, tempKey, destinationKey); err != nil {
		return "", err
	}
	return c.PublicURLByKey(destinationKey), nil
}

func (c *Client) RemoveObject(ctx context.Context, key string) error {
	if err := c.internal.RemoveObject(ctx, c.bucketName, key, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("remove object %q: %w", key, err)
	}
	return nil
}

func (c *Client) ListObjects(ctx context.Context, prefix string, recursive bool) ([]ObjectInfo, error) {
	ch := c.internal.ListObjects(ctx, c.bucketName, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: recursive,
	})

	var rows []ObjectInfo
	for obj := range ch {
		if obj.Err != nil {
			return nil, fmt.Errorf("list objects: %w", obj.Err)
		}
		rows = append(rows, ObjectInfo{
			Key:          obj.Key,
			LastModified: obj.LastModified,
		})
	}
	return rows, nil
}

type TempCleanupResult struct {
	DeletedCount int
	FailedCount  int
	ScannedCount int
}

func (c *Client) CleanupTempUploadsOlderThan(ctx context.Context, olderThan time.Time, prefix string) (*TempCleanupResult, error) {
	if prefix == "" {
		prefix = "temp/"
	}

	rows, err := c.ListObjects(ctx, prefix, true)
	if err != nil {
		return nil, err
	}

	result := &TempCleanupResult{ScannedCount: len(rows)}
	for _, row := range rows {
		if row.LastModified.IsZero() || !row.LastModified.Before(olderThan) {
			continue
		}
		if err := c.RemoveObject(ctx, row.Key); err != nil {
			result.FailedCount++
			continue
		}
		result.DeletedCount++
	}
	return result, nil
}
