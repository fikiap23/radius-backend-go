package rest

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/radius/radius-backend/internal/shared/humaapi"
	"github.com/radius/radius-backend/internal/storage/application/dto"
	"github.com/radius/radius-backend/internal/storage/application/services"
	"github.com/radius/radius-backend/internal/storage/domain"
	"go.uber.org/zap"
)

var storageErrors = []humaapi.ErrorMapping{
	{Err: domain.ErrUnsupportedUploadPurpose, Status: http.StatusBadRequest, Code: "unsupported_upload_purpose", Message: "Unsupported upload purpose."},
	{Err: domain.ErrInvalidFileType, Status: http.StatusBadRequest, Code: "invalid_file_type", Message: "Invalid file type for upload purpose."},
	{Err: domain.ErrInvalidContentType, Status: http.StatusBadRequest, Code: "invalid_content_type", Message: "Invalid content type for upload purpose."},
}

func RegisterStorage(api huma.API, svc *services.StorageService, logger *zap.Logger) {
	huma.Register(api, huma.Operation{
		OperationID:   "storage-presign-upload",
		Method:        http.MethodPost,
		Path:          "/storage/presign-upload",
		Summary:       "Generate presigned upload URL",
		Tags:          []string{"storage"},
		DefaultStatus: http.StatusCreated,
	}, func(ctx context.Context, in *dto.PresignUploadInput) (*humaapi.CreatedOutput, error) {
		result, err := svc.HandlePresignUpload(ctx, *in)
		if err != nil {
			return nil, humaapi.MapError(err, storageErrors, logger)
		}
		return humaapi.Created(result), nil
	})
}
