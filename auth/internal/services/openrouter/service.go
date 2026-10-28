package openrouter

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/roledio/roled/auth/internal/configs"
)

type Service interface {
	ExtractInvoiceFromImage(ctx context.Context, fileHeader *multipart.FileHeader) (any, error)
	ExtractInvoice(ctx context.Context, imageBytes []byte, mimeType string) (any, error)
	ExtractInvoiceFromBase64OrURL(ctx context.Context, imageURLOrBase64 string) (any, error)
}

type service struct {
	defaultConfig *configs.DefaultConfig
	httpClient    *http.Client
}

func NewService(defaultConfig *configs.DefaultConfig) Service {
	return NewServiceWithHTTPClient(defaultConfig, &http.Client{
		Timeout: 90 * time.Second,
	})
}

func NewServiceWithHTTPClient(defaultConfig *configs.DefaultConfig, httpClient *http.Client) Service {
	return &service{
		defaultConfig: defaultConfig,
		httpClient:    httpClient,
	}
}

type chatMessageContentPart struct {
	Type     string        `json:"type"`
	Text     string        `json:"text,omitempty"`
	ImageURL *chatImageURL `json:"image_url,omitempty"`
}

type chatImageURL struct {
	URL string `json:"url"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

type chatCompletionRequest struct {
	Model          string        `json:"model"`
	Messages       []chatMessage `json:"messages"`
	ResponseFormat *responseFmt  `json:"response_format,omitempty"`
}

type responseFmt struct {
	Type string `json:"type"`
}

type chatCompletionResponse struct {
	ID      string `json:"id"`
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Code    any    `json:"code"`
	} `json:"error"`
}

func (s *service) ExtractInvoiceFromImage(ctx context.Context, fileHeader *multipart.FileHeader) (any, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read uploaded file: %w", err)
	}

	mimeType := fileHeader.Header.Get("Content-Type")
	if mimeType == "" || mimeType == "application/octet-stream" {
		mimeType = http.DetectContentType(data)
	}

	return s.ExtractInvoice(ctx, data, mimeType)
}

func (s *service) ExtractInvoice(ctx context.Context, imageBytes []byte, mimeType string) (any, error) {
	if idx := strings.Index(mimeType, ";"); idx != -1 {
		mimeType = strings.TrimSpace(mimeType[:idx])
	}
	if mimeType == "" {
		mimeType = "image/jpeg"
	}

	dataURI := fmt.Sprintf("data:%s;base64,%s", mimeType, base64.StdEncoding.EncodeToString(imageBytes))
	return s.ExtractInvoiceFromBase64OrURL(ctx, dataURI)
}

func (s *service) ExtractInvoiceFromBase64OrURL(ctx context.Context, imageURLOrDataURI string) (any, error) {
	apiKey := strings.TrimSpace(s.defaultConfig.OpenRouter.APIKey)
	if apiKey == "" || apiKey == "your-openrouter-api-key" {
		return nil, errors.New("openrouter api_key is not configured in configs/default.yml or OPENROUTER_API_KEY environment variable")
	}

	baseURL := strings.TrimSpace(s.defaultConfig.OpenRouter.BaseURL)
	if baseURL == "" {
		baseURL = "https://openrouter.ai/api/v1"
	}
	baseURL = strings.TrimRight(baseURL, "/")

	model := strings.TrimSpace(s.defaultConfig.OpenRouter.Model)
	if model == "" {
		model = "google/gemini-2.0-flash-001"
	}

	apiURL := fmt.Sprintf("%s/chat/completions", baseURL)

	systemPrompt := `You are an expert OCR and invoice data extraction AI.
Extract all structured data from the provided invoice image.
Return ONLY a valid JSON object without any explanations, markdown code blocks, or backticks.
Include fields such as:
- invoice_number (string)
- invoice_date (string)
- due_date (string)
- currency (string)
- vendor: { name, address, phone, email, tax_id / npwp, website }
- customer: { name, address, phone, email, tax_id }
- items: [ { item_number, description, quantity, unit_price, discount, tax, amount } ]
- payment_details: { bank_name, account_number, account_name, method }
- financial_summary: { subtotal, discount_amount, tax_rate, tax_amount, shipping_fee, total_amount }
- notes (string)
If any field is not visible in the invoice, omit it or set it to null.`

	userPrompt := "Extract all invoice details from this invoice image and return the data as structured JSON."

	reqBody := chatCompletionRequest{
		Model: model,
		Messages: []chatMessage{
			{
				Role:    "system",
				Content: systemPrompt,
			},
			{
				Role: "user",
				Content: []chatMessageContentPart{
					{
						Type: "text",
						Text: userPrompt,
					},
					{
						Type: "image_url",
						ImageURL: &chatImageURL{
							URL: imageURLOrDataURI,
						},
					},
				},
			},
		},
		ResponseFormat: &responseFmt{
			Type: "json_object",
		},
	}

	content, err := s.sendOpenRouterRequest(ctx, apiURL, apiKey, &reqBody)
	if err != nil {
		return nil, err
	}

	// Clean any markdown formatting like ```json ... ```
	content = strings.TrimSpace(content)
	if strings.HasPrefix(content, "```") {
		if firstNewline := strings.Index(content, "\n"); firstNewline != -1 {
			content = content[firstNewline+1:]
		}
		if lastBackticks := strings.LastIndex(content, "```"); lastBackticks != -1 {
			content = content[:lastBackticks]
		}
		content = strings.TrimSpace(content)
	}

	var parsedData any
	if err := json.Unmarshal([]byte(content), &parsedData); err != nil {
		// If unmarshalling failed, return raw content wrapped in a JSON object
		return map[string]any{
			"raw_content": content,
		}, nil
	}

	return parsedData, nil
}

func (s *service) sendOpenRouterRequest(ctx context.Context, apiURL, apiKey string, reqBody *chatCompletionRequest) (string, error) {
	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to serialize request: %w", err)
	}
	fmt.Println(">>> body: ", string(reqBytes))
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewBuffer(reqBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create http request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	if s.defaultConfig.BaseURL != "" {
		httpReq.Header.Set("HTTP-Referer", s.defaultConfig.BaseURL)
	}
	httpReq.Header.Set("X-Title", "Roled Auth")

	httpResp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to send request to openrouter: %w", err)
	}
	defer httpResp.Body.Close()

	respBytes, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response from openrouter: %w", err)
	}
	fmt.Println("httpResp.StatusCode: ", httpResp.StatusCode)
	fmt.Println("respBytes: ", string(respBytes))
	// If model doesn't support response_format json_object, retry once without it
	if httpResp.StatusCode == http.StatusBadRequest && reqBody.ResponseFormat != nil && strings.Contains(strings.ToLower(string(respBytes)), "response_format") {
		reqBody.ResponseFormat = nil
		return s.sendOpenRouterRequest(ctx, apiURL, apiKey, reqBody)
	}

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return "", fmt.Errorf("openrouter API error (status %d): %s", httpResp.StatusCode, string(respBytes))
	}

	var completionResp chatCompletionResponse
	if err := json.Unmarshal(respBytes, &completionResp); err != nil {
		return "", fmt.Errorf("failed to parse openrouter response: %w", err)
	}

	if completionResp.Error != nil {
		return "", fmt.Errorf("openrouter error: %s", completionResp.Error.Message)
	}

	if len(completionResp.Choices) == 0 {
		return "", errors.New("openrouter returned no choices")
	}

	return completionResp.Choices[0].Message.Content, nil
}
