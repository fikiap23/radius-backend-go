package humaapi

import (
	"net/http"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatFieldMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		param   string
		raw     string
		want    string
	}{
		{
			param: "priority",
			raw:   "expected required property priority to be present",
			want:  "Field 'priority' is required.",
		},
		{
			param: "status",
			raw:   `expected value to be one of "backlog,todo,in_progress,review,done"`,
			want:  "Field 'status' must be one of: backlog, todo, in_progress, review, done.",
		},
		{
			param: "title",
			raw:   "expected length >= 1",
			want:  "Field 'title' must be at least 1 characters.",
		},
		{
			param: "projectId",
			raw:   "expected string to be RFC 4122 uuid",
			want:  "Field 'projectId' must be a valid UUID.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.param, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, formatFieldMessage(tt.param, tt.raw))
		})
	}
}

func TestFormatFieldMessageRequiredAtBodyLocation(t *testing.T) {
	t.Parallel()

	assert.Equal(
		t,
		"Field 'dueAt' is required.",
		formatFieldMessage("body", "expected required property dueAt to be present"),
	)
}

func TestFormatFieldMessageUnknownContentType(t *testing.T) {
	t.Parallel()

	assert.Equal(
		t,
		"Request body must be JSON (Content-Type: application/json).",
		formatFieldMessage("body", "Unknown content type: application/x-www-form-urlencoded."),
	)
}

func TestBuildValidationAPIError(t *testing.T) {
	t.Parallel()

	err := buildValidationAPIError(http.StatusUnprocessableEntity, []*huma.ErrorDetail{
		{
			Message:  "expected required property priority to be present",
			Location: "body.priority",
		},
	})

	apiErr, ok := err.(*apiErrorResponse)
	require.True(t, ok)
	assert.Equal(t, http.StatusUnprocessableEntity, apiErr.GetStatus())
	assert.Equal(t, "validation_error", apiErr.ErrBody.Code)
	assert.Equal(t, "Field 'priority' is required.", apiErr.ErrBody.Message)
	assert.Equal(t, "priority", apiErr.ErrBody.Param)
	assert.Nil(t, apiErr.ErrBody.Errors)
}

func TestBuildValidationAPIErrorMultipleFields(t *testing.T) {
	t.Parallel()

	err := buildValidationAPIError(http.StatusUnprocessableEntity, []*huma.ErrorDetail{
		{
			Message:  "expected required property priority to be present",
			Location: "body.priority",
		},
		{
			Message:  "expected length >= 1",
			Location: "body.title",
		},
	})

	apiErr, ok := err.(*apiErrorResponse)
	require.True(t, ok)
	assert.Contains(t, apiErr.ErrBody.Message, "Field 'priority' is required.")
	assert.Contains(t, apiErr.ErrBody.Message, "Field 'title' must be at least 1 characters.")
	require.Len(t, apiErr.ErrBody.Errors, 2)
}

func TestDefaultMessageUnauthorized(t *testing.T) {
	t.Parallel()

	installEnvelopeErrors()
	err := huma.NewError(http.StatusUnauthorized, "UNAUTHORIZED")

	apiErr, ok := err.(*apiErrorResponse)
	require.True(t, ok)
	assert.Equal(t, "Authentication required. Provide a valid Bearer token.", apiErr.ErrBody.Message)
}

func TestDefaultMessageNotFound(t *testing.T) {
	t.Parallel()

	installEnvelopeErrors()
	err := huma.NewError(http.StatusNotFound, "not found")

	apiErr, ok := err.(*apiErrorResponse)
	require.True(t, ok)
	assert.Equal(t, "The requested endpoint or resource was not found.", apiErr.ErrBody.Message)
	assert.Equal(t, "not_found", apiErr.ErrBody.Code)
}

func TestLocationToParam(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "columnId", locationToParam("body.columnId"))
	assert.Equal(t, "projectId", locationToParam("path.projectId"))
	assert.Equal(t, "", locationToParam(""))
}
