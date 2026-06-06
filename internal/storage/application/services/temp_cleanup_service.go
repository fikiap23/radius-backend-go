package services

import (
	"context"
	"fmt"
	"time"

	"github.com/radius/radius-backend/internal/storage/domain"
	"go.uber.org/zap"
)

type TempCleanupService struct {
	storage domain.ObjectStorage
	logger  *zap.Logger
	maxAge  time.Duration
}

func NewTempCleanupService(storage domain.ObjectStorage, maxAge time.Duration, logger *zap.Logger) *TempCleanupService {
	return &TempCleanupService{
		storage: storage,
		logger:  logger,
		maxAge:  maxAge,
	}
}

func (s *TempCleanupService) RunOnce(ctx context.Context) error {
	olderThan := time.Now().UTC().Add(-s.maxAge)
	result, err := s.storage.CleanupTempUploadsOlderThan(ctx, olderThan)
	if err != nil {
		return fmt.Errorf("cleanup temp uploads: %w", err)
	}

	s.logger.Info("temp upload cleanup finished",
		zap.Int("scanned", result.ScannedCount),
		zap.Int("deleted", result.DeletedCount),
		zap.Int("failed", result.FailedCount),
		zap.Time("olderThan", olderThan),
	)
	return nil
}
