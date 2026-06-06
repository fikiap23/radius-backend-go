package domain

import (
	"context"
	"time"
)

type UploadPurpose string

const (
	UploadPurposeAvatar       UploadPurpose = "avatar"
	UploadPurposeProjectCover UploadPurpose = "project_cover"
	UploadPurposeAttachment   UploadPurpose = "attachment"
)

type PresignedUploadResult struct {
	UploadURL        string
	TempKey          string
	ExpiresInSeconds int
	Method           string
}

type TempCleanupResult struct {
	DeletedCount int
	FailedCount  int
	ScannedCount int
}

type ObjectStorage interface {
	EnsurePublicBucket(ctx context.Context) error
	CreatePresignedUploadURL(ctx context.Context, fileName string) (*PresignedUploadResult, error)
	PublicURLByKey(key string) string
	MoveObject(ctx context.Context, sourceKey, destinationKey string) error
	ValidateTempKey(key string) error
	PromoteUserAvatar(ctx context.Context, tempKey string) (publicURL string, err error)
	PromoteProjectCover(ctx context.Context, tempKey string) (publicURL string, err error)
	PromoteAttachment(ctx context.Context, tempKey string) (publicURL string, err error)
	CleanupTempUploadsOlderThan(ctx context.Context, olderThan time.Time) (*TempCleanupResult, error)
}
