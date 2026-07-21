package openai_compatible

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/noodl-labs/ConnorLLM/services/runtime/internal/runtime/domain"
	"github.com/noodl-labs/ConnorLLM/services/runtime/internal/runtime/domain/entities"
)

// Client calls OpenAI-compatible POST {baseURL}/chat/completions.
// BASE_URL examples: https://api.openai.com/v1, https://openrouter.ai/api/v1
type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

// NewClient builds a client. baseURL must include /v1 (no trailing slash required).
func NewClient(baseURL, apiKey string, httpClient *http.Client) (*Client, error) {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		return nil, ErrEmptyBaseURL
	}
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{
		BaseURL:    baseURL,
		APIKey:     apiKey,
		HTTPClient: httpClient,
	}, nil
}

// Target returns the configured OpenAI-compatible base URL (includes /v1).
func (c *Client) Target() string {
	return c.BaseURL
}

// NewClientFromEnv uses CONNOR_BASE_URL and CONNOR_API_KEY.
func NewClientFromEnv(httpClient *http.Client) (*Client, error) {
	base := strings.TrimSpace(os.Getenv("CONNOR_BASE_URL"))
	key := os.Getenv("CONNOR_API_KEY")
	if base == "" {
		return nil, ErrEnvNotWired
	}
	return NewClient(base, key, httpClient)
}

var _ domain.ProviderExecutor = (*Client)(nil)

// Execute performs one non-stream chat completion attempt.
func (c *Client) Execute(ctx context.Context, req entities.Request) (entities.Response, error) {
	if req.Stream {
		return entities.Response{}, ErrStreamingBeta1
	}

	start := time.Now()
	latency := func() int64 { return time.Since(start).Milliseconds() }

	payload, err := buildRequest(req)
	if err != nil {
		return entities.Response{}, err
	}

	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.BaseURL+"/chat/completions",
		bytes.NewReader(payload),
	)
	if err != nil {
		return entities.Response{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)
	}

	httpResp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		// Network / DNS / TLS — Go error for ExecuteCase.
		return entities.Response{}, err
	}
	defer httpResp.Body.Close()

	raw, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return entities.Response{}, err
	}

	ms := latency()
	status := httpResp.StatusCode

	// Non-2xx: return Response + status, nil error (retry decided by ExecuteCase).
	if status < 200 || status >= 300 {
		return entities.Response{
			HTTPStatus: status,
			LatencyMs:  ms,
			Body:       string(raw), // optional: raw error JSON for logs; or "" in MVP
		}, nil
	}

	parsed, err := ParseCompletion(raw)
	if err != nil {
		return entities.NewFailedResponse(
			entities.FailureProvider,
			status,
			1,
			ms,
			err,
		), nil
	}

	resp := entities.NewSuccessResponse(parsed.Content, status, ms, 0, 1)
	resp.ToolCalls = parsed.ToolCalls
	resp.PromptTokens = parsed.PromptTokens
	resp.CompletionTokens = parsed.CompletionTokens
	return resp, nil
}

func buildRequest(req entities.Request) ([]byte, error) {
	msgs := make([]chatMessage, len(req.Messages))
	for i, m := range req.Messages {
		msgs[i] = chatMessage{Role: m.Role, Content: m.Content}
	}
	body := chatCompletionRequest{
		Model:    req.Model,
		Messages: msgs,
		Stream:   false,
	}
	return json.Marshal(body)
}
