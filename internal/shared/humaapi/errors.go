package humaapi

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/danielgtaylor/huma/v2"
	"go.uber.org/zap"
)

type ErrorMapping struct {
	Err     error
	Status  int
	Code    string
	Message string
	Type    string
	Param   string
}

// APIErrorDetail is the nested error object (Stripe-style).
type APIErrorDetail struct {
	Type    string       `json:"type"`
	Code    string       `json:"code"`
	Message string       `json:"message"`
	Param   string       `json:"param,omitempty"`
	Errors  []FieldError `json:"errors,omitempty"`
}

// FieldError describes a single invalid request field.
type FieldError struct {
	Param   string `json:"param"`
	Message string `json:"message"`
}

// apiErrorResponse is serialized as the HTTP body as-is (Huma does not unwrap a Body field on errors).
type apiErrorResponse struct {
	status  int
	ErrBody APIErrorDetail `json:"error"`
}

func (e *apiErrorResponse) Error() string {
	return e.ErrBody.Message
}

func (e *apiErrorResponse) GetStatus() int {
	return e.status
}

var (
	requiredPropertyRE = regexp.MustCompile(`^expected required property (.+) to be present$`)
	oneOfRE            = regexp.MustCompile(`^expected value to be one of "(.+)"$`)
	minLengthRE        = regexp.MustCompile(`^expected length >= (\d+)$`)
	maxLengthRE        = regexp.MustCompile(`^expected length <= (\d+)$`)
	minNumberRE        = regexp.MustCompile(`^expected number >= (.+)$`)
	maxNumberRE        = regexp.MustCompile(`^expected number <= (.+)$`)
)

var installOnce sync.Once

func installEnvelopeErrors() {
	installOnce.Do(func() {
		huma.NewError = func(status int, msg string, errs ...error) huma.StatusError {
			details := collectErrorDetails(errs)
			if len(details) > 0 {
				return buildValidationAPIError(status, details)
			}
			if status == http.StatusInternalServerError {
				return newAPIError(status, "internal_error", "An unexpected error occurred.", "", "api_error", nil)
			}
			return newAPIError(status, defaultCode(status, msg), defaultMessage(status, msg), "", errorTypeForStatus(status), nil)
		}
	})
}

func collectErrorDetails(errs []error) []*huma.ErrorDetail {
	details := make([]*huma.ErrorDetail, 0, len(errs))
	for _, err := range errs {
		if err == nil {
			continue
		}
		var detailer huma.ErrorDetailer
		if errors.As(err, &detailer) {
			if detail := detailer.ErrorDetail(); detail != nil {
				details = append(details, detail)
			}
			continue
		}
		details = append(details, &huma.ErrorDetail{Message: err.Error()})
	}
	return details
}

func buildValidationAPIError(status int, details []*huma.ErrorDetail) huma.StatusError {
	fieldErrors := make([]FieldError, 0, len(details))
	seen := make(map[string]struct{}, len(details))
	for _, detail := range details {
		param := locationToParam(detail.Location)
		key := param + "\x00" + detail.Message
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		fieldErrors = append(fieldErrors, FieldError{
			Param:   param,
			Message: formatFieldMessage(param, detail.Message),
		})
	}

	message := "Request validation failed."
	param := ""
	if len(fieldErrors) == 1 {
		message = fieldErrors[0].Message
		param = fieldErrors[0].Param
	} else if len(fieldErrors) > 1 {
		parts := make([]string, len(fieldErrors))
		for i, fe := range fieldErrors {
			parts[i] = fe.Message
		}
		message = strings.Join(parts, " ")
	}

	if status == 0 {
		status = http.StatusUnprocessableEntity
	}

	var errorsOut []FieldError
	if len(fieldErrors) > 1 {
		errorsOut = fieldErrors
	}

	return newAPIError(status, "validation_error", message, param, "invalid_request_error", errorsOut)
}

func locationToParam(location string) string {
	location = strings.TrimSpace(location)
	if location == "" {
		return ""
	}
	if idx := strings.Index(location, "."); idx >= 0 {
		return location[idx+1:]
	}
	return location
}

func formatFieldMessage(param, raw string) string {
	raw = strings.TrimSpace(raw)
	field := effectiveField(param, raw)
	if raw == "" {
		if field == "" {
			return "Request validation failed."
		}
		return fmt.Sprintf("Field '%s' is invalid.", field)
	}
	if field == "" {
		return humanizeValidationMessage(raw)
	}

	if matches := requiredPropertyRE.FindStringSubmatch(raw); len(matches) == 2 {
		return fmt.Sprintf("Field '%s' is required.", matches[1])
	}
	if matches := oneOfRE.FindStringSubmatch(raw); len(matches) == 2 {
		values := strings.ReplaceAll(matches[1], ",", ", ")
		return fmt.Sprintf("Field '%s' must be one of: %s.", field, values)
	}
	if matches := minLengthRE.FindStringSubmatch(raw); len(matches) == 2 {
		return fmt.Sprintf("Field '%s' must be at least %s characters.", field, matches[1])
	}
	if matches := maxLengthRE.FindStringSubmatch(raw); len(matches) == 2 {
		return fmt.Sprintf("Field '%s' must be at most %s characters.", field, matches[1])
	}
	if matches := minNumberRE.FindStringSubmatch(raw); len(matches) == 2 {
		return fmt.Sprintf("Field '%s' must be greater than or equal to %s.", field, matches[1])
	}
	if matches := maxNumberRE.FindStringSubmatch(raw); len(matches) == 2 {
		return fmt.Sprintf("Field '%s' must be less than or equal to %s.", field, matches[1])
	}
	switch raw {
	case "request body is required":
		return "Request body is required."
	case "expected string to be RFC 4122 uuid":
		return fmt.Sprintf("Field '%s' must be a valid UUID.", field)
	case "unexpected property":
		return fmt.Sprintf("Field '%s' is not allowed.", field)
	case "expected string":
		return fmt.Sprintf("Field '%s' must be a string.", field)
	case "expected boolean":
		return fmt.Sprintf("Field '%s' must be a boolean.", field)
	case "expected number":
		return fmt.Sprintf("Field '%s' must be a number.", field)
	case "expected integer":
		return fmt.Sprintf("Field '%s' must be an integer.", field)
	case "expected array":
		return fmt.Sprintf("Field '%s' must be an array.", field)
	case "expected object":
		return fmt.Sprintf("Field '%s' must be an object.", field)
	default:
		if strings.HasPrefix(raw, "Unknown content type:") {
			return "Request body must be JSON (Content-Type: application/json)."
		}
		return fmt.Sprintf("Field '%s': %s.", field, humanizeValidationMessage(raw))
	}
}

func humanizeValidationMessage(raw string) string {
	if raw == "" {
		return "Request validation failed"
	}
	if !strings.HasSuffix(raw, ".") {
		raw += "."
	}
	if raw[0] >= 'a' && raw[0] <= 'z' {
		return strings.ToUpper(raw[:1]) + raw[1:]
	}
	return raw
}

func effectiveField(param, raw string) string {
	if matches := requiredPropertyRE.FindStringSubmatch(raw); len(matches) == 2 {
		return matches[1]
	}
	if param != "" && param != "body" {
		return param
	}
	return param
}

func newAPIError(status int, code, message, param, errType string, fieldErrors []FieldError) huma.StatusError {
	if errType == "" {
		errType = errorTypeForStatus(status)
	}
	if message == "" {
		message = defaultErrorMessage(code)
	}
	return &apiErrorResponse{
		status: status,
		ErrBody: APIErrorDetail{
			Type:    errType,
			Code:    code,
			Message: message,
			Param:   param,
			Errors:  fieldErrors,
		},
	}
}

func defaultCode(status int, msg string) string {
	switch status {
	case http.StatusUnauthorized:
		return "unauthorized"
	case http.StatusForbidden:
		return "forbidden"
	case http.StatusNotFound:
		return "not_found"
	case http.StatusMethodNotAllowed:
		return "method_not_allowed"
	case http.StatusUnsupportedMediaType:
		return "unsupported_media_type"
	case http.StatusBadRequest:
		return "bad_request"
	default:
		return toSnakeCode(msg)
	}
}

func defaultMessage(status int, msg string) string {
	switch status {
	case http.StatusUnauthorized:
		return "Authentication required. Provide a valid Bearer token."
	case http.StatusForbidden:
		return "You do not have permission to perform this action."
	case http.StatusNotFound:
		return "The requested endpoint or resource was not found."
	case http.StatusMethodNotAllowed:
		return "This HTTP method is not allowed for the requested endpoint."
	case http.StatusUnsupportedMediaType:
		return "Unsupported Content-Type. Use application/json."
	case http.StatusBadRequest:
		if msg != "" && !strings.EqualFold(msg, "bad request") {
			return humanizeValidationMessage(msg)
		}
		return "The request could not be processed."
	default:
		if msg != "" {
			return humanizeValidationMessage(msg)
		}
		return defaultErrorMessage(defaultCode(status, msg))
	}
}

func errorTypeForStatus(status int) string {
	switch status {
	case http.StatusUnauthorized:
		return "authentication_error"
	case http.StatusForbidden:
		return "permission_error"
	case http.StatusInternalServerError, http.StatusServiceUnavailable:
		return "api_error"
	default:
		return "invalid_request_error"
	}
}

func toSnakeCode(msg string) string {
	if msg == "" {
		return "unknown_error"
	}
	msg = strings.TrimSpace(strings.ToLower(msg))
	msg = strings.ReplaceAll(msg, " ", "_")
	return msg
}

func defaultErrorMessage(code string) string {
	words := strings.Split(code, "_")
	for i, w := range words {
		if w == "" {
			continue
		}
		words[i] = strings.ToUpper(w[:1]) + w[1:]
	}
	return strings.Join(words, " ") + "."
}

func MapError(err error, mappings []ErrorMapping, logger *zap.Logger) error {
	if err == nil {
		return nil
	}
	for _, m := range mappings {
		if errors.Is(err, m.Err) {
			return newAPIError(m.Status, m.Code, m.Message, m.Param, m.Type, nil)
		}
	}
	if logger != nil {
		logger.Error("handler error", zap.Error(err))
	}
	return newAPIError(http.StatusInternalServerError, "internal_error", "An unexpected error occurred.", "", "api_error", nil)
}
