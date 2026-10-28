package openrouter

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/roledio/roled/auth/internal/configs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractInvoiceFromBase64OrURL_Success(t *testing.T) {
	mockInvoiceData := map[string]any{
		"invoice_number": "INV-2026-001",
		"total_amount":   150000.0,
		"vendor": map[string]any{
			"name": "PT Example Tech",
		},
	}
	invoiceJSON, err := json.Marshal(mockInvoiceData)
	require.NoError(t, err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var req chatCompletionRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		assert.NoError(t, err)
		assert.Equal(t, "google/gemini-2.0-flash-001", req.Model)

		// Return content wrapped with markdown fences to test markdown cleaning
		resp := chatCompletionResponse{
			ID: "gen-123",
			Choices: []struct {
				Message struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				} `json:"message"`
				FinishReason string `json:"finish_reason"`
			}{
				{
					Message: struct {
						Role    string `json:"role"`
						Content string `json:"content"`
					}{
						Role:    "assistant",
						Content: "```json\n" + string(invoiceJSON) + "\n```",
					},
					FinishReason: "stop",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := &configs.DefaultConfig{
		OpenRouter: struct {
			BaseURL string `mapstructure:"base_url"`
			APIKey  string `mapstructure:"api_key"`
			Model   string `mapstructure:"model"`
		}{
			BaseURL: server.URL,
			APIKey:  "test-api-key",
			Model:   "google/gemini-2.0-flash-001",
		},
	}

	svc := NewServiceWithHTTPClient(cfg, server.Client())
	result, err := svc.ExtractInvoice(context.Background(), []byte("fake-image-bytes"), "image/png")
	require.NoError(t, err)

	resultMap, ok := result.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "INV-2026-001", resultMap["invoice_number"])
	assert.Equal(t, 150000.0, resultMap["total_amount"])
}

func TestExtractInvoice_MissingAPIKey(t *testing.T) {
	cfg := &configs.DefaultConfig{
		OpenRouter: struct {
			BaseURL string `mapstructure:"base_url"`
			APIKey  string `mapstructure:"api_key"`
			Model   string `mapstructure:"model"`
		}{
			BaseURL: "https://openrouter.ai/api/v1",
			APIKey:  "",
		},
	}

	svc := NewService(cfg)
	_, err := svc.ExtractInvoice(context.Background(), []byte("fake-image"), "image/jpeg")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "api_key is not configured")
}

func TestExtractInvoice_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error": {"message": "Invalid API Key"}}`))
	}))
	defer server.Close()

	cfg := &configs.DefaultConfig{
		OpenRouter: struct {
			BaseURL string `mapstructure:"base_url"`
			APIKey  string `mapstructure:"api_key"`
			Model   string `mapstructure:"model"`
		}{
			BaseURL: server.URL,
			APIKey:  "invalid-key",
		},
	}

	svc := NewServiceWithHTTPClient(cfg, server.Client())
	_, err := svc.ExtractInvoice(context.Background(), []byte("fake-image"), "image/jpeg")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "openrouter API error (status 401)")
}
