package models

import (
	"time"
)

// Severity represents the severity level of an incident
type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityMedium   Severity = "medium"
	SeverityLow      Severity = "low"
	SeverityUnknown  Severity = "unknown"
)

// IncidentStatus represents the status of an incident
type IncidentStatus string

const (
	StatusOpen       IncidentStatus = "open"
	StatusInProgress IncidentStatus = "in_progress"
	StatusResolved   IncidentStatus = "resolved"
	StatusClosed     IncidentStatus = "closed"
)

// Incident represents an incident entity
type Incident struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Source      string                 `json:"source,omitempty"`
	Status      IncidentStatus         `json:"status"`
	Severity    Severity               `json:"severity"`
	Logs        []string               `json:"logs,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	AssignedTo  string                 `json:"assigned_to,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	ResolvedAt  *time.Time             `json:"resolved_at,omitempty"`
	AIAnalysis  *AIAnalysis            `json:"ai_analysis,omitempty"`
	RCADocument *RCADocument           `json:"rca_document,omitempty"`
}

// AIAnalysis represents AI-generated analysis for an incident
type AIAnalysis struct {
	Summary            string    `json:"summary"`
	Findings           []string  `json:"findings"`
	RootCauses         []string  `json:"root_causes"`
	RecommendedActions []string  `json:"recommended_actions"`
	SeveritySuggestion Severity  `json:"severity_suggestion"`
	GeneratedAt        time.Time `json:"generated_at"`
	Model              string    `json:"model"`
	Provider           string    `json:"provider"`
}

// RCADocument represents a root cause analysis document
type RCADocument struct {
	Timeline            []string  `json:"timeline"`
	RootCause           string    `json:"root_cause"`
	Impact              string    `json:"impact"`
	ImmediateResolution string    `json:"immediate_resolution"`
	PreventiveMeasures  []string  `json:"preventive_measures"`
	LessonsLearned      []string  `json:"lessons_learned"`
	GeneratedAt         time.Time `json:"generated_at"`
	Model               string    `json:"model"`
	Provider            string    `json:"provider"`
}

// CreateIncidentRequest represents a request to create an incident
type CreateIncidentRequest struct {
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Source      string                 `json:"source,omitempty"`
	Severity    *Severity              `json:"severity,omitempty"`
	Logs        []string               `json:"logs,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	AssignedTo  string                 `json:"assigned_to,omitempty"`
}

// UpdateIncidentRequest represents a request to update an incident
type UpdateIncidentRequest struct {
	Title       *string                `json:"title,omitempty"`
	Description *string                `json:"description,omitempty"`
	Status      *IncidentStatus        `json:"status,omitempty"`
	Severity    *Severity              `json:"severity,omitempty"`
	Logs        []string               `json:"logs,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	AssignedTo  *string                `json:"assigned_to,omitempty"`
}

// LogSummarizeRequest represents a request to summarize logs
type LogSummarizeRequest struct {
	Logs []string `json:"logs"`
}

// LogSummarizeResponse represents the response from log summarization
type LogSummarizeResponse struct {
	Summary     string    `json:"summary"`
	KeyInsights []string  `json:"key_insights"`
	Alerts      []string  `json:"alerts"`
	GeneratedAt time.Time `json:"generated_at"`
}
