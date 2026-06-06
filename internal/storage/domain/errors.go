package domain

import "errors"

var (
	ErrUnsupportedUploadPurpose = errors.New("unsupported upload purpose")
	ErrInvalidFileType            = errors.New("invalid file type for upload purpose")
	ErrInvalidContentType         = errors.New("invalid content type for upload purpose")
	ErrTempFileNotFound           = errors.New("uploaded temp file not found")
	ErrInvalidTempKey             = errors.New("invalid temp key")
)
