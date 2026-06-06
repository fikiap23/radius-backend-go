package dto

import "github.com/radius/radius-backend/internal/storage/domain"

type PresignUploadInput struct {
	Body struct {
		FileName    string               `json:"fileName" minLength:"1" maxLength:"255" doc:"Original file name (used for extension validation only)"`
		ContentType string               `json:"contentType,omitempty" maxLength:"120" doc:"Optional MIME type for validation"`
		Purpose     domain.UploadPurpose `json:"purpose" enum:"avatar,project_cover,attachment" doc:"Upload purpose — drives allowed file types"`
	}
}

type PresignUploadOutput struct {
	UploadURL string `json:"uploadUrl"`
	TempKey   string `json:"tempKey"`
	ExpiresIn int    `json:"expiresIn"`
	Method    string `json:"method"`
}
