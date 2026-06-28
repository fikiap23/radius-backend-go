package rest

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	projectdomain "github.com/radius/radius-backend/internal/projects/domain"
	"github.com/radius/radius-backend/internal/shared/humaapi"
	"github.com/radius/radius-backend/internal/shared/middleware"
	"github.com/radius/radius-backend/internal/storage/application/dto"
	"github.com/radius/radius-backend/internal/storage/application/services"
	"github.com/radius/radius-backend/internal/storage/domain"
	taskdomain "github.com/radius/radius-backend/internal/tasks/domain"
	"go.uber.org/zap"
)

var storageErrors = []humaapi.ErrorMapping{
	{Err: domain.ErrUnsupportedUploadPurpose, Status: http.StatusBadRequest, Code: "unsupported_upload_purpose", Message: "Unsupported upload purpose."},
	{Err: domain.ErrInvalidFileType, Status: http.StatusBadRequest, Code: "invalid_file_type", Message: "Invalid file type for upload purpose."},
	{Err: domain.ErrInvalidContentType, Status: http.StatusBadRequest, Code: "invalid_content_type", Message: "Invalid content type for upload purpose."},
	{Err: domain.ErrPresignAuthRequired, Status: http.StatusUnauthorized, Code: "presign_auth_required", Message: "Authentication required for this upload purpose."},
	{Err: domain.ErrPresignTaskIDRequired, Status: http.StatusBadRequest, Code: "presign_task_id_required", Message: "context.taskId is required for task_attachment uploads.", Param: "context.taskId"},
	{Err: taskdomain.ErrTaskNotFound, Status: http.StatusNotFound, Code: "task_not_found", Message: "Task not found.", Param: "context.taskId"},
	{Err: projectdomain.ErrProjectForbidden, Status: http.StatusForbidden, Code: "project_forbidden", Message: "You do not have access to this project."},
}

func RegisterStorage(api huma.API, svc *services.StorageService, auth *middleware.AuthMiddleware, logger *zap.Logger) {
	optionalAuthMW := humaapi.OptionalAuth(auth, api)

	huma.Register(api, huma.Operation{
		OperationID:   "storage-presign-upload",
		Method:        http.MethodPost,
		Path:          "/storage/presign-upload",
		Summary:       "Generate presigned upload URL",
		Tags:          []string{"storage"},
		DefaultStatus: http.StatusCreated,
		Middlewares:   huma.Middlewares{optionalAuthMW},
	}, func(ctx context.Context, in *dto.PresignUploadInput) (*humaapi.CreatedOutput, error) {
		userID, _ := humaapi.UserIDFromContext(ctx)
		if in.Body.Purpose == domain.UploadPurposeTaskAttachment && userID == "" {
			return nil, humaapi.MapError(domain.ErrPresignAuthRequired, storageErrors, logger)
		}

		result, err := svc.HandlePresignUpload(ctx, userID, *in)
		if err != nil {
			return nil, humaapi.MapError(err, storageErrors, logger)
		}
		return humaapi.Created(result), nil
	})
}
