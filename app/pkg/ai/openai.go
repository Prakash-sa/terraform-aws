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

// OpenAIClient implements the Client interface for OpenAI
type OpenAIClient struct {
	apiKey      string
	model       string
	timeout     time.Duration
	temperature float32
	maxTokens   int
	httpClient  *http.Client
}

// OpenAI API request/response types
type openaiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openaiRequest struct {
	Model       string          `json:"model"`
	Messages    []openaiMessage `json:"messages"`
	Temperature float32         `json:"temperature"`
	MaxTokens   int             `json:"max_tokens"`
}

type openaiChoice struct {
	Message struct {
		Content string `json:"content"`
	} `json:"message"`
}

type openaiResponse struct {
	Choices []openaiChoice `json:"choices"`
}

const (
	openaiAPIURL = "https://api.openai.com/v1/chat/completions"
	defaultModel = "gpt-4"
)

// NewOpenAIClient creates a new OpenAI client
func NewOpenAIClient(cfg ClientConfig) (*OpenAIClient, error) {
	if cfg.APIKey == "" {
		return nil, ErrNoAPIKey
	}

	model := cfg.Model
	if model == "" {
		model = defaultModel
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

	return &OpenAIClient{
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

func (c *OpenAIClient) Health(ctx context.Context) error {
	// Simple health check by sending a minimal request
	req := openaiRequest{
		Model: c.model,
		Messages: []openaiMessage{
			{
				Role:    "user",
				Content: "ping",
			},
		},
		Temperature: 0,
		MaxTokens:   5,
	}

	_, err := c.call(ctx, req)
	return err
}

func (c *OpenAIClient) AnalyzeIncident(ctx context.Context, req AnalysisRequest) (*AnalysisResponse, error) {
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

	openaiReq := openaiRequest{
		Model: c.model,
		Messages: []openaiMessage{
			{
				Role:    "system",
				Content: "You are an expert incident response analyst. Analyze incidents and provide structured JSON responses.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: c.temperature,
		MaxTokens:   c.maxTokens,
	}

	resp, err := c.call(ctx, openaiReq)
	if err != nil {
		return nil, err
	}

	return parseAnalysisResponse(resp)
}

func (c *OpenAIClient) GenerateRCA(ctx context.Context, req RCARequest) (*RCAResponse, error) {
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

	openaiReq := openaiRequest{
		Model: c.model,
		Messages: []openaiMessage{
			{
				Role:    "system",
				Content: "You are an expert in writing Root Cause Analysis (RCA) documents. Generate comprehensive, structured RCA documents in JSON format.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: c.temperature,
		MaxTokens:   c.maxTokens,
	}

	resp, err := c.call(ctx, openaiReq)
	if err != nil {
		return nil, err
	}

	return parseRCAResponse(resp)
}

func (c *OpenAIClient) SummarizeLogs(ctx context.Context, req SummarizeRequest) (*SummarizeResponse, error) {
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

	openaiReq := openaiRequest{
		Model: c.model,
		Messages: []openaiMessage{
			{
				Role:    "system",
				Content: "You are an expert at analyzing logs and extracting key insights. Respond with structured JSON.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: c.temperature,
		MaxTokens:   1500,
	}

	resp, err := c.call(ctx, openaiReq)
	if err != nil {
		return nil, err
	}

	return parseSummarizeResponse(resp)
}

func (c *OpenAIClient) Provider() Provider {
	return ProviderOpenAI
}

func (c *OpenAIClient) Model() string {
	return c.model
}

func (c *OpenAIClient) call(ctx context.Context, req openaiRequest) (string, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", openaiAPIURL, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to call OpenAI API: %w", err)
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return "", err
	}

	if httpResp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("OpenAI API error: %d - %s", httpResp.StatusCode, string(respBody))
	}

	var openaiResp openaiResponse
	if err := json.Unmarshal(respBody, &openaiResp); err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidResponse, err)
	}

	if len(openaiResp.Choices) == 0 {
		return "", ErrInvalidResponse
	}

	return openaiResp.Choices[0].Message.Content, nil
}
