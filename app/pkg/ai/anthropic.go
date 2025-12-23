package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// AnthropicClient implements the Client interface for Anthropic Claude
type AnthropicClient struct {
	apiKey      string
	model       string
	timeout     time.Duration
	temperature float32
	maxTokens   int
	httpClient  *http.Client
}

// Anthropic API request/response types
type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicRequest struct {
	Model       string             `json:"model"`
	Messages    []anthropicMessage `json:"messages"`
	Temperature float32            `json:"temperature"`
	MaxTokens   int                `json:"max_tokens"`
	System      string             `json:"system,omitempty"`
}

type anthropicContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type anthropicResponse struct {
	Content []anthropicContent `json:"content"`
}

const (
	anthropicAPIURL    = "https://api.anthropic.com/v1/messages"
	anthropicVersion   = "2023-06-01"
	defaultClaudeModel = "claude-3-5-sonnet-20241022"
)

// NewAnthropicClient creates a new Anthropic Claude client
func NewAnthropicClient(cfg ClientConfig) (*AnthropicClient, error) {
	if cfg.APIKey == "" {
		return nil, ErrNoAPIKey
	}

	model := cfg.Model
	if model == "" {
		model = defaultClaudeModel
	}

	timeout := time.Duration(cfg.Timeout) * time.Second
	if timeout == 0 {
		timeout = 60 * time.Second
	}

	temperature := cfg.Temperature
	if temperature == 0 {
		temperature = 0.7
	}

	maxTokens := cfg.MaxTokens
	if maxTokens == 0 {
		maxTokens = 2000
	}

	return &AnthropicClient{
		apiKey:      cfg.APIKey,
		model:       model,
		timeout:     timeout,
		temperature: temperature,
		maxTokens:   maxTokens,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

func (c *AnthropicClient) Health(ctx context.Context) error {
	req := anthropicRequest{
		Model: c.model,
		Messages: []anthropicMessage{
			{
				Role:    "user",
				Content: "ping",
			},
		},
		Temperature: 0,
		MaxTokens:   5,
	}

	_, err := c.call(ctx, req, "You are a helpful assistant.")
	return err
}

func (c *AnthropicClient) AnalyzeIncident(ctx context.Context, req AnalysisRequest) (*AnalysisResponse, error) {
	logsText := strings.Join(req.Logs, "\n")

	prompt := fmt.Sprintf(`Analyze this incident and provide structured analysis in JSON format:

Title: %s
Description: %s

Related Logs:
%s

Respond with a JSON object containing:
{
  "summary": "Brief summary of the incident",
  "findings": ["finding1", "finding2"],
  "root_causes": ["cause1", "cause2"],
  "recommended_actions": ["action1", "action2"],
  "suggested_severity": "critical|high|medium|low"
}

Only respond with the JSON object, no additional text.`, req.IncidentTitle, req.IncidentDesc, logsText)

	system := "You are an expert incident response analyst. Analyze incidents and provide structured JSON responses."

	anthropicReq := anthropicRequest{
		Model: c.model,
		Messages: []anthropicMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: c.temperature,
		MaxTokens:   c.maxTokens,
	}

	resp, err := c.call(ctx, anthropicReq, system)
	if err != nil {
		return nil, err
	}

	return parseAnalysisResponse(resp)
}

func (c *AnthropicClient) GenerateRCA(ctx context.Context, req RCARequest) (*RCAResponse, error) {
	analysisJSON, _ := json.Marshal(req.Analysis)
	timelineText := strings.Join(req.Timeline, "\n")

	prompt := fmt.Sprintf(`Generate a comprehensive Root Cause Analysis document for this incident:

Title: %s
Description: %s

Previous Analysis:
%s

Timeline:
%s

Respond with a JSON object containing:
{
  "timeline": "Detailed timeline of events",
  "root_cause": "Identified root cause",
  "impact": "Impact assessment",
  "immediate_resolution": "Steps taken to resolve",
  "preventive_measures": ["measure1", "measure2"],
  "lessons_learned": ["lesson1", "lesson2"]
}

Only respond with the JSON object, no additional text.`, req.IncidentTitle, req.IncidentDesc, string(analysisJSON), timelineText)

	system := "You are an expert in writing Root Cause Analysis (RCA) documents. Generate comprehensive, structured RCA documents in JSON format."

	anthropicReq := anthropicRequest{
		Model: c.model,
		Messages: []anthropicMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: c.temperature,
		MaxTokens:   c.maxTokens,
	}

	resp, err := c.call(ctx, anthropicReq, system)
	if err != nil {
		return nil, err
	}

	return parseRCAResponse(resp)
}

func (c *AnthropicClient) SummarizeLogs(ctx context.Context, req SummarizeRequest) (*SummarizeResponse, error) {
	logsText := strings.Join(req.Logs, "\n")

	prompt := fmt.Sprintf(`Summarize these logs and extract key insights:

Logs:
%s

Respond with a JSON object containing:
{
  "summary": "Brief summary of logs",
  "key_insights": ["insight1", "insight2"],
  "alerts": ["alert1", "alert2"]
}

Only respond with the JSON object, no additional text.`, logsText)

	system := "You are an expert at analyzing logs and extracting key insights. Respond with structured JSON."

	anthropicReq := anthropicRequest{
		Model: c.model,
		Messages: []anthropicMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: c.temperature,
		MaxTokens:   1500,
	}

	resp, err := c.call(ctx, anthropicReq, system)
	if err != nil {
		return nil, err
	}

	return parseSummarizeResponse(resp)
}

func (c *AnthropicClient) Provider() Provider {
	return ProviderAnthropic
}

func (c *AnthropicClient) Model() string {
	return c.model
}

func (c *AnthropicClient) call(ctx context.Context, req anthropicRequest, system string) (string, error) {
	req.System = system
	body, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", anthropicAPIURL, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}

	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", anthropicVersion)
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to call Anthropic API: %w", err)
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return "", err
	}

	if httpResp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Anthropic API error: %d - %s", httpResp.StatusCode, string(respBody))
	}

	var anthropicResp anthropicResponse
	if err := json.Unmarshal(respBody, &anthropicResp); err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidResponse, err)
	}

	if len(anthropicResp.Content) == 0 {
		return "", ErrInvalidResponse
	}

	return anthropicResp.Content[0].Text, nil
}
