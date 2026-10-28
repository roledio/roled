package api

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/roledio/roled/auth/internal/configs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockOpenRouterService struct {
	mock.Mock
}

func (m *mockOpenRouterService) ExtractInvoiceFromImage(ctx context.Context, fileHeader *multipart.FileHeader) (any, error) {
	args := m.Called(ctx, fileHeader)
	return args.Get(0), args.Error(1)
}

func (m *mockOpenRouterService) ExtractInvoice(ctx context.Context, imageBytes []byte, mimeType string) (any, error) {
	args := m.Called(ctx, imageBytes, mimeType)
	return args.Get(0), args.Error(1)
}

func (m *mockOpenRouterService) ExtractInvoiceFromBase64OrURL(ctx context.Context, imageURLOrBase64 string) (any, error) {
	args := m.Called(ctx, imageURLOrBase64)
	return args.Get(0), args.Error(1)
}

func TestExtractImage_MultipartSuccess(t *testing.T) {
	mockService := new(mockOpenRouterService)
	expectedData := map[string]any{
		"invoice_number": "INV-100",
		"total":          250000.0,
	}

	mockService.On("ExtractInvoiceFromImage", mock.Anything, mock.Anything).Return(expectedData, nil)

	app := fiber.New()
	h := &handler{
		defaultConfig:     &configs.DefaultConfig{},
		openRouterService: mockService,
	}
	app.Post("/api/openrouter/extract-image", h.extractImage)

	// Create multipart form request
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("image", "invoice.png")
	require.NoError(t, err)
	_, err = part.Write([]byte("fake-image-data"))
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	req := httptest.NewRequest("POST", "/api/openrouter/extract-image", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	res, err := app.Test(req)
	require.NoError(t, err)
	defer res.Body.Close()

	assert.Equal(t, 200, res.StatusCode)

	var resp struct {
		Success bool           `json:"success"`
		Data    map[string]any `json:"data"`
	}
	err = json.NewDecoder(res.Body).Decode(&resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, "INV-100", resp.Data["invoice_number"])
}

func TestExtractImage_RawQueryParam(t *testing.T) {
	mockService := new(mockOpenRouterService)
	expectedData := map[string]any{
		"invoice_number": "INV-RAW-1",
	}

	mockService.On("ExtractInvoiceFromImage", mock.Anything, mock.Anything).Return(expectedData, nil)

	app := fiber.New()
	h := &handler{
		defaultConfig:     &configs.DefaultConfig{},
		openRouterService: mockService,
	}
	app.Post("/api/openrouter/extract-image", h.extractImage)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "invoice.jpg")
	require.NoError(t, err)
	_, err = part.Write([]byte("fake-jpg"))
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	req := httptest.NewRequest("POST", "/api/openrouter/extract-image?raw=true", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	res, err := app.Test(req)
	require.NoError(t, err)
	defer res.Body.Close()

	assert.Equal(t, 200, res.StatusCode)

	var resp map[string]any
	err = json.NewDecoder(res.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "INV-RAW-1", resp["invoice_number"])
}

func TestExtractImage_MissingFile(t *testing.T) {
	mockService := new(mockOpenRouterService)

	app := fiber.New()
	h := &handler{
		defaultConfig:     &configs.DefaultConfig{},
		openRouterService: mockService,
	}
	app.Post("/api/openrouter/extract-image", h.extractImage)

	req := httptest.NewRequest("POST", "/api/openrouter/extract-image", nil)
	res, err := app.Test(req)
	require.NoError(t, err)
	defer res.Body.Close()

	assert.Equal(t, 400, res.StatusCode)
}
