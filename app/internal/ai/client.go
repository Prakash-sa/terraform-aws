package ai

import (
	"context"
	"github.com/Prakash-sa/terraform-aws/app/pkg/models"
)

// Client is the interface for AI services
type Client interface {
	// AnalyzeIncident analyzes an incident and returns AI-generated insights
	AnalyzeIncident(ctx context.Context, incident *models.Incident) (*models.AIAnalysis, error)
	
	// GenerateRCA generates a Root Cause Analysis document for an incident
	GenerateRCA(ctx context.Context, incident *models.Incident, analysis *models.AIAnalysis) (*models.RCADocument, error)
	
	// SummarizeLogs summarizes a collection of logs
	SummarizeLogs(ctx context.Context, logs []string) (string, error)
	
	// ClassifySeverity suggests an appropriate severity level for an incident
	ClassifySeverity(ctx context.Context, incident *models.Incident) (models.Severity, error)
}

// Provider represents the AI service provider
type Provider string

const (
	ProviderOpenAI    Provider = "openai"
	ProviderAnthropic Provider = "anthropic"
)

// Config holds configuration for AI clients
type Config struct {
	Provider       Provider
	OpenAIKey      string
	AnthropicKey   string
	OpenAIModel    string
	AnthropicModel string
}

// NewClient creates a new AI client based on the configuration
func NewClient(config Config) (Client, error) {
	switch config.Provider {
	case ProviderOpenAI:
		return NewOpenAIClient(config.OpenAIKey, config.OpenAIModel)
	case ProviderAnthropic:
		return NewAnthropicClient(config.AnthropicKey, config.AnthropicModel)
	default:
		return NewOpenAIClient(config.OpenAIKey, config.OpenAIModel)
	}
}
