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

// OpenAIClient implements the Client interface using OpenAI API
type OpenAIClient struct {
	apiKey     string
	model      string
	httpClient *http.Client
}

// NewOpenAIClient creates a new OpenAI client
func NewOpenAIClient(apiKey, model string) (*OpenAIClient, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required")
	}
	if model == "" {
		model = "gpt-4" // Default model
	}
	return &OpenAIClient{
		apiKey: apiKey,
		model:  model,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}, nil
}

type openAIRequest struct {
	Model    string          `json:"model"`
	Messages []openAIMessage `json:"messages"`
	Temperature float64      `json:"temperature,omitempty"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIResponse struct {
	Choices []struct {
		Message openAIMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (c *OpenAIClient) callAPI(ctx context.Context, prompt string) (string, error) {
	reqBody := openAIRequest{
		Model: c.model,
		Messages: []openAIMessage{
			{
				Role:    "system",
				Content: "You are an expert DevOps engineer specializing in incident response and analysis. Provide clear, actionable insights.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 0.7,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call OpenAI API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("OpenAI API error (status %d): %s", resp.StatusCode, string(body))
	}

	var openAIResp openAIResponse
	if err := json.Unmarshal(body, &openAIResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if openAIResp.Error != nil {
		return "", fmt.Errorf("OpenAI API error: %s", openAIResp.Error.Message)
	}

	if len(openAIResp.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI API")
	}

	return openAIResp.Choices[0].Message.Content, nil
}

// AnalyzeIncident analyzes an incident using OpenAI
func (c *OpenAIClient) AnalyzeIncident(ctx context.Context, incident *models.Incident) (*models.AIAnalysis, error) {
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

	// Try to extract JSON from the response (it might be wrapped in markdown code blocks)
	jsonStr := response
	if strings.Contains(response, "```json") {
		start := strings.Index(response, "```json") + 7
		end := strings.LastIndex(response, "```")
		if start > 7 && end > start {
			jsonStr = strings.TrimSpace(response[start:end])
		}
	} else if strings.Contains(response, "```") {
		start := strings.Index(response, "```") + 3
		end := strings.LastIndex(response, "```")
		if start > 3 && end > start {
			jsonStr = strings.TrimSpace(response[start:end])
		}
	}

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
		Provider:           "openai",
	}, nil
}

// GenerateRCA generates a Root Cause Analysis document
func (c *OpenAIClient) GenerateRCA(ctx context.Context, incident *models.Incident, analysis *models.AIAnalysis) (*models.RCADocument, error) {
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

	// Try to extract JSON from the response
	jsonStr := response
	if strings.Contains(response, "```json") {
		start := strings.Index(response, "```json") + 7
		end := strings.LastIndex(response, "```")
		if start > 7 && end > start {
			jsonStr = strings.TrimSpace(response[start:end])
		}
	} else if strings.Contains(response, "```") {
		start := strings.Index(response, "```") + 3
		end := strings.LastIndex(response, "```")
		if start > 3 && end > start {
			jsonStr = strings.TrimSpace(response[start:end])
		}
	}

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
func (c *OpenAIClient) SummarizeLogs(ctx context.Context, logs []string) (string, error) {
	prompt := fmt.Sprintf(`Analyze and summarize the following logs, highlighting important events, errors, and patterns:

%s

Provide a concise summary of the key information from these logs.`,
		strings.Join(logs, "\n"),
	)

	return c.callAPI(ctx, prompt)
}

// ClassifySeverity suggests an appropriate severity level
func (c *OpenAIClient) ClassifySeverity(ctx context.Context, incident *models.Incident) (models.Severity, error) {
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
