package minio

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	sharedstorage "github.com/radius/radius-backend/internal/shared/storage"
	"github.com/radius/radius-backend/internal/storage/domain"
)

type ObjectStorage struct {
	client *sharedstorage.Client
}

func NewObjectStorage(client *sharedstorage.Client) *ObjectStorage {
	return &ObjectStorage{client: client}
}

func (s *ObjectStorage) EnsurePublicBucket(ctx context.Context) error {
	return s.client.EnsurePublicBucket(ctx)
}

func (s *ObjectStorage) CreatePresignedUploadURL(ctx context.Context, fileName string) (*domain.PresignedUploadResult, error) {
	result, err := s.client.CreatePresignedUploadURL(ctx, fileName)
	if err != nil {
		return nil, err
	}
	return &domain.PresignedUploadResult{
		UploadURL:        result.UploadURL,
		TempKey:          result.Key,
		ExpiresInSeconds: result.ExpiresInSeconds,
		Method:           "PUT",
	}, nil
}

func (s *ObjectStorage) PublicURLByKey(key string) string {
	return s.client.PublicURLByKey(key)
}

func (s *ObjectStorage) MoveObject(ctx context.Context, sourceKey, destinationKey string) error {
	if err := s.client.MoveObject(ctx, sourceKey, destinationKey); err != nil {
		if strings.Contains(err.Error(), "source object not found") {
			return domain.ErrTempFileNotFound
		}
		return fmt.Errorf("move object: %w", err)
	}
	return nil
}

func (s *ObjectStorage) ValidateTempKey(key string) error {
	if err := sharedstorage.ValidateTempKey(key); err != nil {
		return errors.Join(domain.ErrInvalidTempKey, err)
	}
	return nil
}

func (s *ObjectStorage) PromoteUserAvatar(ctx context.Context, tempKey string) (string, error) {
	return s.promote(ctx, tempKey, sharedstorage.BuildUserAvatarKey(tempKey))
}

func (s *ObjectStorage) PromoteProjectCover(ctx context.Context, tempKey string) (string, error) {
	return s.promote(ctx, tempKey, sharedstorage.BuildProjectCoverKey(tempKey))
}

func (s *ObjectStorage) PromoteAttachment(ctx context.Context, tempKey string) (string, error) {
	return s.promote(ctx, tempKey, sharedstorage.BuildAttachmentKey(tempKey))
}

func (s *ObjectStorage) promote(ctx context.Context, tempKey, destinationKey string) (string, error) {
	if err := s.ValidateTempKey(tempKey); err != nil {
		return "", err
	}
	url, err := s.client.PromoteTempObject(ctx, tempKey, destinationKey)
	if err != nil {
		if strings.Contains(err.Error(), "source object not found") {
			return "", domain.ErrTempFileNotFound
		}
		if strings.Contains(err.Error(), "invalid temp key") {
			return "", domain.ErrInvalidTempKey
		}
		return "", fmt.Errorf("promote temp object: %w", err)
	}
	return url, nil
}

func (s *ObjectStorage) CleanupTempUploadsOlderThan(ctx context.Context, olderThan time.Time) (*domain.TempCleanupResult, error) {
	result, err := s.client.CleanupTempUploadsOlderThan(ctx, olderThan, "temp/")
	if err != nil {
		return nil, fmt.Errorf("cleanup temp uploads: %w", err)
	}
	return &domain.TempCleanupResult{
		DeletedCount: result.DeletedCount,
		FailedCount:  result.FailedCount,
		ScannedCount: result.ScannedCount,
	}, nil
}

var _ domain.ObjectStorage = (*ObjectStorage)(nil)
