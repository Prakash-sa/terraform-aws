package incident

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Prakash-sa/terraform-aws/app/internal/ai"
	"github.com/Prakash-sa/terraform-aws/app/pkg/models"
	"github.com/google/uuid"
)

// Service manages incidents and AI-powered analysis
type Service struct {
	incidents map[string]*models.Incident
	analyses  map[string]*models.AIAnalysis
	rcas      map[string]*models.RCADocument
	aiClient  ai.Client
	mu        sync.RWMutex
}

// NewService creates a new incident service
func NewService(aiClient ai.Client) *Service {
	return &Service{
		incidents: make(map[string]*models.Incident),
		analyses:  make(map[string]*models.AIAnalysis),
		rcas:      make(map[string]*models.RCADocument),
		aiClient:  aiClient,
	}
}

// CreateIncident creates a new incident
func (s *Service) CreateIncident(ctx context.Context, req models.CreateIncidentRequest) (*models.Incident, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	incident := &models.Incident{
		ID:          uuid.New().String(),
		Title:       req.Title,
		Description: req.Description,
		Severity:    req.Severity,
		Status:      models.StatusOpen,
		Source:      req.Source,
		AlertData:   req.AlertData,
		Logs:        req.Logs,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// If severity is not provided, try to auto-classify using AI
	if incident.Severity == "" && s.aiClient != nil {
		severity, err := s.aiClient.ClassifySeverity(ctx, incident)
		if err == nil {
			incident.Severity = severity
		} else {
			// Default to medium if AI classification fails
			incident.Severity = models.SeverityMedium
		}
	}

	s.incidents[incident.ID] = incident
	return incident, nil
}

// GetIncident retrieves an incident by ID
func (s *Service) GetIncident(incidentID string) (*models.Incident, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	incident, exists := s.incidents[incidentID]
	if !exists {
		return nil, fmt.Errorf("incident not found: %s", incidentID)
	}

	return incident, nil
}

// ListIncidents returns all incidents
func (s *Service) ListIncidents() []*models.Incident {
	s.mu.RLock()
	defer s.mu.RUnlock()

	incidents := make([]*models.Incident, 0, len(s.incidents))
	for _, incident := range s.incidents {
		incidents = append(incidents, incident)
	}

	return incidents
}

// UpdateIncidentStatus updates the status of an incident
func (s *Service) UpdateIncidentStatus(incidentID string, status models.IncidentStatus) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	incident, exists := s.incidents[incidentID]
	if !exists {
		return fmt.Errorf("incident not found: %s", incidentID)
	}

	incident.Status = status
	incident.UpdatedAt = time.Now()

	if status == models.StatusResolved || status == models.StatusClosed {
		now := time.Now()
		incident.ResolvedAt = &now
	}

	return nil
}

// AnalyzeIncident performs AI-powered analysis on an incident
func (s *Service) AnalyzeIncident(ctx context.Context, incidentID string) (*models.AIAnalysis, error) {
	s.mu.RLock()
	incident, exists := s.incidents[incidentID]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("incident not found: %s", incidentID)
	}

	if s.aiClient == nil {
		return nil, fmt.Errorf("AI client not configured")
	}

	// Perform AI analysis
	analysis, err := s.aiClient.AnalyzeIncident(ctx, incident)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze incident: %w", err)
	}

	// Store the analysis
	s.mu.Lock()
	s.analyses[incidentID] = analysis
	s.mu.Unlock()

	return analysis, nil
}

// GetAnalysis retrieves the AI analysis for an incident
func (s *Service) GetAnalysis(incidentID string) (*models.AIAnalysis, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	analysis, exists := s.analyses[incidentID]
	if !exists {
		return nil, fmt.Errorf("analysis not found for incident: %s", incidentID)
	}

	return analysis, nil
}

// GenerateRCA generates a Root Cause Analysis document for an incident
func (s *Service) GenerateRCA(ctx context.Context, incidentID string) (*models.RCADocument, error) {
	s.mu.RLock()
	incident, incidentExists := s.incidents[incidentID]
	analysis, analysisExists := s.analyses[incidentID]
	s.mu.RUnlock()

	if !incidentExists {
		return nil, fmt.Errorf("incident not found: %s", incidentID)
	}

	if s.aiClient == nil {
		return nil, fmt.Errorf("AI client not configured")
	}

	// If no analysis exists, generate one first
	if !analysisExists {
		var err error
		analysis, err = s.AnalyzeIncident(ctx, incidentID)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze incident for RCA: %w", err)
		}
	}

	// Generate RCA document
	rca, err := s.aiClient.GenerateRCA(ctx, incident, analysis)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RCA: %w", err)
	}

	// Store the RCA
	s.mu.Lock()
	s.rcas[incidentID] = rca
	s.mu.Unlock()

	return rca, nil
}

// GetRCA retrieves the RCA document for an incident
func (s *Service) GetRCA(incidentID string) (*models.RCADocument, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rca, exists := s.rcas[incidentID]
	if !exists {
		return nil, fmt.Errorf("RCA not found for incident: %s", incidentID)
	}

	return rca, nil
}

// SummarizeLogs provides an AI-powered summary of logs
func (s *Service) SummarizeLogs(ctx context.Context, logs []string) (string, error) {
	if s.aiClient == nil {
		return "", fmt.Errorf("AI client not configured")
	}

	return s.aiClient.SummarizeLogs(ctx, logs)
}
