package models

import "time"

// Severity levels for incidents
type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityMedium   Severity = "medium"
	SeverityLow      Severity = "low"
)

// IncidentStatus represents the current state of an incident
type IncidentStatus string

const (
	StatusOpen       IncidentStatus = "open"
	StatusInProgress IncidentStatus = "in_progress"
	StatusResolved   IncidentStatus = "resolved"
	StatusClosed     IncidentStatus = "closed"
)

// Incident represents an incident in the system
type Incident struct {
	ID          string         `json:"id"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Severity    Severity       `json:"severity"`
	Status      IncidentStatus `json:"status"`
	Source      string         `json:"source"` // e.g., "prometheus", "cloudwatch", "manual"
	AlertData   string         `json:"alert_data,omitempty"`
	Logs        []string       `json:"logs,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	ResolvedAt  *time.Time     `json:"resolved_at,omitempty"`
}

// AIAnalysis represents AI-generated analysis of an incident
type AIAnalysis struct {
	IncidentID       string    `json:"incident_id"`
	Summary          string    `json:"summary"`
	SuggestedSeverity Severity `json:"suggested_severity"`
	KeyFindings      []string  `json:"key_findings"`
	PotentialCauses  []string  `json:"potential_causes"`
	RecommendedActions []string `json:"recommended_actions"`
	GeneratedAt      time.Time `json:"generated_at"`
	Provider         string    `json:"provider"` // "openai" or "anthropic"
}

// RCADocument represents a Root Cause Analysis document
type RCADocument struct {
	IncidentID       string    `json:"incident_id"`
	Summary          string    `json:"summary"`
	Timeline         []string  `json:"timeline"`
	RootCause        string    `json:"root_cause"`
	ImpactAnalysis   string    `json:"impact_analysis"`
	Resolution       string    `json:"resolution"`
	PreventiveMeasures []string `json:"preventive_measures"`
	LessonsLearned   []string  `json:"lessons_learned"`
	GeneratedAt      time.Time `json:"generated_at"`
	GeneratedBy      string    `json:"generated_by"` // "ai" or username
}

// CreateIncidentRequest represents the request to create a new incident
type CreateIncidentRequest struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Severity    Severity `json:"severity,omitempty"`
	Source      string   `json:"source"`
	AlertData   string   `json:"alert_data,omitempty"`
	Logs        []string `json:"logs,omitempty"`
}

// AnalysisRequest represents a request to analyze an incident
type AnalysisRequest struct {
	IncidentID string `json:"incident_id"`
	Provider   string `json:"provider,omitempty"` // "openai" or "anthropic", defaults to configured
}
