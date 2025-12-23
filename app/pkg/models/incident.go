package models
package models

import (
	"time"
)

// Severity represents the incident severity level
type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityMedium   Severity = "medium"
	SeverityLow      Severity = "low"































































































}	GeneratedAt time.Time `json:"generated_at"`	Alerts     []string `json:"alerts,omitempty"`	KeyInsights []string `json:"key_insights"`	Summary    string   `json:"summary"`type LogSummarizeResponse struct {// LogSummarizeResponse represents the response from log summarization}	Context map[string]string `json:"context,omitempty"`	Logs   []string          `json:"logs" binding:"required"`type LogSummarizeRequest struct {// LogSummarizeRequest represents a request to summarize logs}	AssignedTo  *string                `json:"assigned_to,omitempty"`	Metadata    map[string]interface{} `json:"metadata,omitempty"`	Tags        []string               `json:"tags,omitempty"`	Logs        []string               `json:"logs,omitempty"`	Status      *IncidentStatus        `json:"status,omitempty"`	Severity    *Severity              `json:"severity,omitempty"`	Description *string                `json:"description,omitempty"`	Title       *string                `json:"title,omitempty"`type UpdateIncidentRequest struct {// UpdateIncidentRequest represents a request to update an incident}	AssignedTo  string                 `json:"assigned_to,omitempty"`	Metadata    map[string]interface{} `json:"metadata,omitempty"`	Tags        []string               `json:"tags,omitempty"`	Logs        []string               `json:"logs,omitempty"`	Severity    *Severity              `json:"severity,omitempty"`	Source      string                 `json:"source"`	Description string                 `json:"description" binding:"required"`	Title       string                 `json:"title" binding:"required"`type CreateIncidentRequest struct {// CreateIncidentRequest represents a request to create an incident}	Provider          string    `json:"provider"`	Model             string    `json:"model"`	GeneratedAt       time.Time `json:"generated_at"`	References        []string `json:"references,omitempty"`	LessonsLearned    []string `json:"lessons_learned"`	PreventiveMeasures []string `json:"preventive_measures"`	ImmediateResolution string `json:"immediate_resolution"`	Impact            string   `json:"impact"`	RootCause         string   `json:"root_cause"`	Timeline          string   `json:"timeline"`type RCADocument struct {// RCADocument represents a Root Cause Analysis document}	Provider         string    `json:"provider"`	Model            string    `json:"model"`	GeneratedAt      time.Time `json:"generated_at"`	SeveritySuggestion Severity `json:"severity_suggestion"`	RecommendedActions []string `json:"recommended_actions"`	RootCauses       []string `json:"root_causes"`	Findings         []string `json:"findings"`	Summary          string   `json:"summary"`type AIAnalysis struct {// AIAnalysis represents AI-generated analysis of an incident}	RCADocument     *RCADocument           `json:"rca_document,omitempty"`	AIAnalysis      *AIAnalysis            `json:"ai_analysis,omitempty"`	AssignedTo      string                 `json:"assigned_to,omitempty"`	ResolvedAt      *time.Time             `json:"resolved_at,omitempty"`	UpdatedAt       time.Time              `json:"updated_at"`	CreatedAt       time.Time              `json:"created_at"`	Metadata        map[string]interface{} `json:"metadata,omitempty"`	Tags            []string               `json:"tags,omitempty"`	Logs            []string               `json:"logs,omitempty"`	Status          IncidentStatus         `json:"status"`	Severity        Severity               `json:"severity"`	Source          string                 `json:"source"` // prometheus, logs, manual, etc.	Description     string                 `json:"description"`	Title           string                 `json:"title"`	ID              string                 `json:"id"`type Incident struct {// Incident represents a security or operational incident)	StatusClosed     IncidentStatus = "closed"	StatusResolved   IncidentStatus = "resolved"	StatusInProgress IncidentStatus = "in_progress"	StatusOpen       IncidentStatus = "open"const (type IncidentStatus string// IncidentStatus represents the current status of an incident)	SeverityUnknown  Severity = "unknown"