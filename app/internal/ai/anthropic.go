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

	"github.com/Prakash-sa/terraform-aws/app/pkg/models"
)

// AnthropicClient implements the Client interface using Anthropic Claude API
type AnthropicClient struct {
	apiKey     string
	model      string
	httpClient *http.Client
}

// NewAnthropicClient creates a new Anthropic client
func NewAnthropicClient(apiKey, model string) (*AnthropicClient, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("Anthropic API key is required")
	}
	if model == "" {
		model = "claude-3-5-sonnet-20241022" // Default model
	}
	return &AnthropicClient{
		apiKey: apiKey,
		model:  model,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}, nil
}

type anthropicRequest struct {
	Model     string              `json:"model"`
	Messages  []anthropicMessage  `json:"messages"`
	MaxTokens int                 `json:"max_tokens"`
	System    string              `json:"system,omitempty"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (c *AnthropicClient) callAPI(ctx context.Context, prompt string) (string, error) {
	reqBody := anthropicRequest{
		Model: c.model,
		Messages: []anthropicMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		MaxTokens: 4096,
		System:    "You are an expert DevOps engineer specializing in incident response and analysis. Provide clear, actionable insights.",
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call Anthropic API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Anthropic API error (status %d): %s", resp.StatusCode, string(body))
	}

	var anthropicResp anthropicResponse
	if err := json.Unmarshal(body, &anthropicResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if anthropicResp.Error != nil {
		return "", fmt.Errorf("Anthropic API error: %s", anthropicResp.Error.Message)
	}

	if len(anthropicResp.Content) == 0 {
		return "", fmt.Errorf("no response from Anthropic API")
	}

	return anthropicResp.Content[0].Text, nil
}

// AnalyzeIncident analyzes an incident using Anthropic Claude
func (c *AnthropicClient) AnalyzeIncident(ctx context.Context, incident *models.Incident) (*models.AIAnalysis, error) {
	prompt := fmt.Sprintf(`Analyze the following incident and provide detailed insights:

Title: %s
Description: %s
Current Severity: %s
Source: %s
Alert Data: %s
Logs: %s

Please provide:
1. A concise summary of the incident
2. Suggested severity level (critical, high, medium, or low)
3. Key findings (as a list)
4. Potential causes (as a list)
5. Recommended actions (as a list)

Format your response as JSON with the following structure:
{
  "summary": "...",
  "suggested_severity": "...",
  "key_findings": ["...", "..."],
  "potential_causes": ["...", "..."],
  "recommended_actions": ["...", "..."]
}`,
		incident.Title,
		incident.Description,
		incident.Severity,
		incident.Source,
		incident.AlertData,
		strings.Join(incident.Logs, "\n"),
	)

	response, err := c.callAPI(ctx, prompt)
	if err != nil {
		return nil, err
	}

	// Parse the JSON response
	var analysisData struct {
		Summary            string   `json:"summary"`
		SuggestedSeverity  string   `json:"suggested_severity"`
		KeyFindings        []string `json:"key_findings"`
		PotentialCauses    []string `json:"potential_causes"`
		RecommendedActions []string `json:"recommended_actions"`
	}

	// Extract JSON from the response
	jsonStr := extractJSON(response)

	if err := json.Unmarshal([]byte(jsonStr), &analysisData); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	return &models.AIAnalysis{
		IncidentID:         incident.ID,
		Summary:            analysisData.Summary,
		SuggestedSeverity:  models.Severity(analysisData.SuggestedSeverity),
		KeyFindings:        analysisData.KeyFindings,
		PotentialCauses:    analysisData.PotentialCauses,
		RecommendedActions: analysisData.RecommendedActions,
		GeneratedAt:        time.Now(),
		Provider:           "anthropic",
	}, nil
}

// GenerateRCA generates a Root Cause Analysis document
func (c *AnthropicClient) GenerateRCA(ctx context.Context, incident *models.Incident, analysis *models.AIAnalysis) (*models.RCADocument, error) {
	prompt := fmt.Sprintf(`Generate a comprehensive Root Cause Analysis (RCA) document for the following incident:

Incident Title: %s
Description: %s
Severity: %s
Status: %s
Created: %s

AI Analysis Summary: %s
Key Findings: %s
Potential Causes: %s

Please provide a detailed RCA document with:
1. Executive summary
2. Timeline of events (as a list)
3. Root cause identification
4. Impact analysis
5. Resolution steps
6. Preventive measures (as a list)
7. Lessons learned (as a list)

Format your response as JSON with the following structure:
{
  "summary": "...",
  "timeline": ["...", "..."],
  "root_cause": "...",
  "impact_analysis": "...",
  "resolution": "...",
  "preventive_measures": ["...", "..."],
  "lessons_learned": ["...", "..."]
}`,
		incident.Title,
		incident.Description,
		incident.Severity,
		incident.Status,
		incident.CreatedAt.Format(time.RFC3339),
		analysis.Summary,
		strings.Join(analysis.KeyFindings, "; "),
		strings.Join(analysis.PotentialCauses, "; "),
	)

	response, err := c.callAPI(ctx, prompt)
	if err != nil {
		return nil, err
	}

	// Parse the JSON response
	var rcaData struct {
		Summary            string   `json:"summary"`
		Timeline           []string `json:"timeline"`
		RootCause          string   `json:"root_cause"`
		ImpactAnalysis     string   `json:"impact_analysis"`
		Resolution         string   `json:"resolution"`
		PreventiveMeasures []string `json:"preventive_measures"`
		LessonsLearned     []string `json:"lessons_learned"`
	}

	// Extract JSON from the response
	jsonStr := extractJSON(response)

	if err := json.Unmarshal([]byte(jsonStr), &rcaData); err != nil {
		return nil, fmt.Errorf("failed to parse RCA response: %w", err)
	}

	return &models.RCADocument{
		IncidentID:         incident.ID,
		Summary:            rcaData.Summary,
		Timeline:           rcaData.Timeline,
		RootCause:          rcaData.RootCause,
		ImpactAnalysis:     rcaData.ImpactAnalysis,
		Resolution:         rcaData.Resolution,
		PreventiveMeasures: rcaData.PreventiveMeasures,
		LessonsLearned:     rcaData.LessonsLearned,
		GeneratedAt:        time.Now(),
		GeneratedBy:        "ai",
	}, nil
}

// SummarizeLogs summarizes a collection of logs
func (c *AnthropicClient) SummarizeLogs(ctx context.Context, logs []string) (string, error) {
	prompt := fmt.Sprintf(`Analyze and summarize the following logs, highlighting important events, errors, and patterns:

%s

Provide a concise summary of the key information from these logs.`,
		strings.Join(logs, "\n"),
	)

	return c.callAPI(ctx, prompt)
}

// ClassifySeverity suggests an appropriate severity level
func (c *AnthropicClient) ClassifySeverity(ctx context.Context, incident *models.Incident) (models.Severity, error) {
	prompt := fmt.Sprintf(`Based on the following incident details, classify the severity as one of: critical, high, medium, or low.

Title: %s
Description: %s
Source: %s
Alert Data: %s

Respond with ONLY the severity level (critical, high, medium, or low), no additional text.`,
		incident.Title,
		incident.Description,
		incident.Source,
		incident.AlertData,
	)

	response, err := c.callAPI(ctx, prompt)
	if err != nil {
		return models.SeverityMedium, err
	}

	// Clean up the response
	severity := strings.ToLower(strings.TrimSpace(response))
	
	switch severity {
	case "critical":
		return models.SeverityCritical, nil
	case "high":
		return models.SeverityHigh, nil
	case "medium":
		return models.SeverityMedium, nil
	case "low":
		return models.SeverityLow, nil
	default:
		return models.SeverityMedium, nil
	}
}
