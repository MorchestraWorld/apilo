package daemon

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync/atomic"
	"time"
)

// ClaudeClient handles Claude API requests with token tracking
type ClaudeClient struct {
	httpClient *http.Client
	apiKey     string
	endpoint   string
	metrics    *ClaudeMetrics
}

// ClaudeMetrics tracks Claude API token usage with thread-safe operations
type ClaudeMetrics struct {
	inputTokens  int64
	outputTokens int64
	totalCost    int64 // Cost in cents to avoid floating point issues
	requests     int64
}

// ClaudeRequest represents a request to Claude API
type ClaudeRequest struct {
	Model       string          `json:"model"`
	MaxTokens   int             `json:"max_tokens"`
	Messages    []ClaudeMessage `json:"messages"`
	Temperature float64         `json:"temperature,omitempty"`
}

// ClaudeMessage represents a message in Claude conversation
type ClaudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ClaudeResponse represents Claude API response
type ClaudeResponse struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Role    string `json:"role"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Model        string      `json:"model"`
	StopReason   string      `json:"stop_reason"`
	Usage        ClaudeUsage `json:"usage"`
	ErrorMessage string      `json:"error,omitempty"`
}

// ClaudeUsage contains token usage information
type ClaudeUsage struct {
	InputTokens  int64 `json:"input_tokens"`
	OutputTokens int64 `json:"output_tokens"`
}

// NewClaudeClient creates a new Claude API client
func NewClaudeClient() (*ClaudeClient, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY environment variable not set")
	}

	client := &ClaudeClient{
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		apiKey:   apiKey,
		endpoint: "https://api.anthropic.com/v1/messages",
		metrics:  &ClaudeMetrics{},
	}

	return client, nil
}

// MakeRequest sends a request to Claude API and tracks tokens
func (c *ClaudeClient) MakeRequest(prompt string, maxTokens int) (*ClaudeResponse, error) {
	atomic.AddInt64(&c.metrics.requests, 1)

	req := ClaudeRequest{
		Model:     "claude-sonnet-4-20250514",
		MaxTokens: maxTokens,
		Messages: []ClaudeMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 1.0,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.endpoint, bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer httpResp.Body.Close()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", httpResp.StatusCode, string(body))
	}

	var response ClaudeResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Track tokens and cost
	c.trackTokens(&response)

	response.Usage.InputTokens = response.Usage.InputTokens
	response.Usage.OutputTokens = response.Usage.OutputTokens

	return &response, nil
}

// trackTokens records token usage and calculates cost
func (c *ClaudeClient) trackTokens(resp *ClaudeResponse) {
	inputTokens := resp.Usage.InputTokens
	outputTokens := resp.Usage.OutputTokens

	atomic.AddInt64(&c.metrics.inputTokens, inputTokens)
	atomic.AddInt64(&c.metrics.outputTokens, outputTokens)

	// Sonnet 4 pricing: $3 per MTok input, $15 per MTok output
	// Store cost in cents to avoid floating point
	inputCostCents := (inputTokens * 3 * 100) / 1_000_000    // $3 per MTok
	outputCostCents := (outputTokens * 15 * 100) / 1_000_000 // $15 per MTok
	totalCostCents := inputCostCents + outputCostCents

	atomic.AddInt64(&c.metrics.totalCost, totalCostCents)
}

// GetMetrics returns current token usage metrics
func (c *ClaudeClient) GetMetrics() ClaudeTokenMetrics {
	return ClaudeTokenMetrics{
		InputTokens:   atomic.LoadInt64(&c.metrics.inputTokens),
		OutputTokens:  atomic.LoadInt64(&c.metrics.outputTokens),
		TotalTokens:   atomic.LoadInt64(&c.metrics.inputTokens) + atomic.LoadInt64(&c.metrics.outputTokens),
		Cost:          float64(atomic.LoadInt64(&c.metrics.totalCost)) / 100.0, // Convert cents to dollars
		TotalRequests: atomic.LoadInt64(&c.metrics.requests),
		Model:         "claude-sonnet-4-20250514",
	}
}

// ResetMetrics resets all token counters
func (c *ClaudeClient) ResetMetrics() {
	atomic.StoreInt64(&c.metrics.inputTokens, 0)
	atomic.StoreInt64(&c.metrics.outputTokens, 0)
	atomic.StoreInt64(&c.metrics.totalCost, 0)
	atomic.StoreInt64(&c.metrics.requests, 0)
}
