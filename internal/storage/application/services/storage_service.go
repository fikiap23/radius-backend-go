package services

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/radius/radius-backend/internal/storage/application/dto"
	"github.com/radius/radius-backend/internal/storage/domain"
	"go.uber.org/zap"
)

type StorageService struct {
	storage domain.ObjectStorage
	logger  *zap.Logger
}

func NewStorageService(storage domain.ObjectStorage, logger *zap.Logger) *StorageService {
	return &StorageService{storage: storage, logger: logger}
}

func (s *StorageService) HandlePresignUpload(ctx context.Context, in dto.PresignUploadInput) (*dto.PresignUploadOutput, error) {
	if err := validateUploadPurpose(in.Body.FileName, in.Body.ContentType, in.Body.Purpose); err != nil {
		return nil, err
	}

	if err := s.storage.EnsurePublicBucket(ctx); err != nil {
		return nil, fmt.Errorf("ensure bucket: %w", err)
	}

	result, err := s.storage.CreatePresignedUploadURL(ctx, in.Body.FileName)
	if err != nil {
		return nil, fmt.Errorf("presign upload: %w", err)
	}

	return &dto.PresignUploadOutput{
		UploadURL: result.UploadURL,
		TempKey:   result.TempKey,
		ExpiresIn: result.ExpiresInSeconds,
		Method:    result.Method,
	}, nil
}

func validateUploadPurpose(fileName, contentType string, purpose domain.UploadPurpose) error {
	imageExtensions := map[string]struct{}{
		".jpg": {}, ".jpeg": {}, ".png": {}, ".webp": {},
	}
	imageContentTypes := map[string]struct{}{
		"image/jpeg": {}, "image/png": {}, "image/webp": {},
	}
	documentExtensions := map[string]struct{}{
		".pdf": {}, ".doc": {}, ".docx": {},
		".jpg": {}, ".jpeg": {}, ".png": {}, ".webp": {},
	}
	documentContentTypes := map[string]struct{}{
		"application/pdf": {},
		"application/msword": {},
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": {},
		"image/jpeg": {}, "image/png": {}, "image/webp": {},
	}

	var allowedExtensions, allowedContentTypes map[string]struct{}

	switch purpose {
	case domain.UploadPurposeAvatar, domain.UploadPurposeProjectCover:
		allowedExtensions = imageExtensions
		allowedContentTypes = imageContentTypes
	case domain.UploadPurposeAttachment:
		allowedExtensions = documentExtensions
		allowedContentTypes = documentContentTypes
	default:
		return domain.ErrUnsupportedUploadPurpose
	}

	ext := safeExtension(fileName)
	if _, ok := allowedExtensions[ext]; !ok {
		return domain.ErrInvalidFileType
	}

	if contentType != "" {
		if _, ok := allowedContentTypes[strings.ToLower(contentType)]; !ok {
			return domain.ErrInvalidContentType
		}
	}

	return nil
}

func safeExtension(filename string) string {
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
