package humaapi

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/humatest"
)

func TestAPIErrorResponse(t *testing.T) {
	cfg := huma.DefaultConfig("Test API", "1.0.0")
	cfg.CreateHooks = nil
	installEnvelopeErrors()

	_, api := humatest.New(t, cfg)

	huma.Register(api, huma.Operation{
		OperationID: "test-login-fail",
		Method:      http.MethodPost,
		Path:        "/auth/login",
	}, func(ctx context.Context, _ *struct{}) (*OKOutput, error) {
		return nil, huma.NewError(http.StatusUnauthorized, "INVALID_CREDENTIALS")
	})

	resp := api.Post("/auth/login", map[string]string{}, "Host: localhost:8080")
	if resp.Code != http.StatusUnauthorized {
		t.Fatalf("status: got %d want %d", resp.Code, http.StatusUnauthorized)
	}

	var body map[string]any
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	errObj, ok := body["error"].(map[string]any)
	if !ok {
		t.Fatalf("expected error object: %+v", body)
	}
	if errObj["type"] != "authentication_error" {
		t.Fatalf("type: got %v want authentication_error", errObj["type"])
	}
	if errObj["code"] != "invalid_credentials" {
		t.Fatalf("code: got %v want invalid_credentials", errObj["code"])
	}
	if errObj["message"] == "" {
		t.Fatalf("message must not be empty: %+v", errObj)
	}
	if _, ok := body["isSuccess"]; ok {
		t.Fatalf("response must not use legacy envelope: %+v", body)
	}
	if _, ok := body["Body"]; ok {
		t.Fatalf("response must not wrap error in Body: %+v", body)
	}
	if _, ok := body["$schema"]; ok {
		t.Fatalf("response must not include $schema: %+v", body)
	}
}
