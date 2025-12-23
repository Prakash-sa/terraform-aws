package ai

import (
	"context"
	"errors"
	"fmt"
)

// Provider represents an AI provider type
type Provider string

const (
	ProviderOpenAI    Provider = "openai"
	ProviderAnthropic Provider = "anthropic"
)

// ClientConfig holds configuration for AI clients
type ClientConfig struct {
	Provider    Provider
	APIKey      string
	Model       string
	Timeout     int // seconds
	Temperature float32
	MaxTokens   int
}

// AnalysisRequest represents a request for incident analysis
type AnalysisRequest struct {
	IncidentTitle     string
	IncidentDesc      string
	Logs              []string
	AdditionalContext map[string]string
}

// AnalysisResponse represents the response from incident analysis
type AnalysisResponse struct {
	Summary            string
	Findings           []string
	RootCauses         []string
	RecommendedActions []string
	SuggestedSeverity  string
	RawResponse        string
}

// RCARequest represents a request for RCA generation
type RCARequest struct {
	IncidentTitle     string
	IncidentDesc      string
	Analysis          AnalysisResponse
	Timeline          []string
	AdditionalContext map[string]string
}

// RCAResponse represents the response from RCA generation
type RCAResponse struct {
	Timeline            string
	RootCause           string
	Impact              string
	ImmediateResolution string
	PreventiveMeasures  []string
	LessonsLearned      []string
	RawResponse         string
}

// SummarizeRequest represents a request for log summarization
type SummarizeRequest struct {
	Logs          []string
	Context       map[string]string
	IncludeAlerts bool
}

// SummarizeResponse represents the response from log summarization
type SummarizeResponse struct {
	Summary     string
	KeyInsights []string
	Alerts      []string
	RawResponse string
}

// Client defines the interface for AI providers
type Client interface {
	// AnalyzeIncident generates analysis for an incident
	AnalyzeIncident(ctx context.Context, req AnalysisRequest) (*AnalysisResponse, error)

	// GenerateRCA generates a root cause analysis document
	GenerateRCA(ctx context.Context, req RCARequest) (*RCAResponse, error)

	// SummarizeLogs extracts insights from log collections
	SummarizeLogs(ctx context.Context, req SummarizeRequest) (*SummarizeResponse, error)

	// Health checks if the client is properly configured and accessible
	Health(ctx context.Context) error

	// Provider returns the provider type
	Provider() Provider

	// Model returns the model being used
	Model() string
}

// ErrNoAPIKey is returned when API key is not configured
var ErrNoAPIKey = errors.New("API key not configured")

// ErrProviderNotSupported is returned for unsupported providers
var ErrProviderNotSupported = errors.New("provider not supported")

// ErrTimeout is returned when API call times out
var ErrTimeout = errors.New("API call timeout")

// ErrInvalidResponse is returned when response parsing fails
var ErrInvalidResponse = errors.New("invalid API response")

// NewClient creates a new AI client based on the provider configuration
func NewClient(cfg ClientConfig) (Client, error) {
	if cfg.APIKey == "" {
		return nil, ErrNoAPIKey
	}

	switch cfg.Provider {
	case ProviderOpenAI:
		return NewOpenAIClient(cfg)
	case ProviderAnthropic:
		return NewAnthropicClient(cfg)
	default:
		return nil, fmt.Errorf("%w: %s", ErrProviderNotSupported, cfg.Provider)
	}
}

// NoOpClient is a client that does nothing, used when AI is not configured
type NoOpClient struct {
	provider Provider
	model    string
}

func NewNoOpClient(provider Provider, model string) *NoOpClient {
	return &NoOpClient{
		provider: provider,
		model:    model,
	}
}

func (c *NoOpClient) AnalyzeIncident(ctx context.Context, req AnalysisRequest) (*AnalysisResponse, error) {
	return &AnalysisResponse{
		Summary:            "AI analysis not available (provider not configured)",
		Findings:           []string{},
		RootCauses:         []string{},
		RecommendedActions: []string{},
		SuggestedSeverity:  "unknown",
	}, nil
}

func (c *NoOpClient) GenerateRCA(ctx context.Context, req RCARequest) (*RCAResponse, error) {
	return &RCAResponse{
		Timeline:            "AI RCA generation not available (provider not configured)",
		RootCause:           "",
		Impact:              "",
		ImmediateResolution: "",
		PreventiveMeasures:  []string{},
		LessonsLearned:      []string{},
	}, nil
}

func (c *NoOpClient) SummarizeLogs(ctx context.Context, req SummarizeRequest) (*SummarizeResponse, error) {
	return &SummarizeResponse{
		Summary:     "Log summarization not available (provider not configured)",
		KeyInsights: []string{},
		Alerts:      []string{},
	}, nil
}

func (c *NoOpClient) Health(ctx context.Context) error {
	return fmt.Errorf("AI provider not configured")
}

func (c *NoOpClient) Provider() Provider {
	return c.provider
}

func (c *NoOpClient) Model() string {
	return c.model
}
