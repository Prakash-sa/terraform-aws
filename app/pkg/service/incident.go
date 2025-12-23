package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Prakash-sa/terraform-aws/app/pkg/ai"
	"github.com/Prakash-sa/terraform-aws/app/pkg/models"
	"go.uber.org/zap"
)

// IncidentStore provides thread-safe incident storage and retrieval
type IncidentStore struct {
	incidents map[string]*models.Incident
	mu        sync.RWMutex
	counter   int64
}

// IncidentService provides business logic for incident management
type IncidentService struct {
	store    *IncidentStore
	aiClient ai.Client
	logger   *zap.Logger
}

// NewIncidentStore creates a new incident store
func NewIncidentStore() *IncidentStore {
	return &IncidentStore{
		incidents: make(map[string]*models.Incident),
		counter:   0,
	}
}

// NewIncidentService creates a new incident service
func NewIncidentService(store *IncidentStore, aiClient ai.Client, logger *zap.Logger) *IncidentService {
	return &IncidentService{
		store:    store,
		aiClient: aiClient,
		logger:   logger,
	}
}

// CreateIncident creates a new incident with optional AI severity classification
func (s *IncidentService) CreateIncident(req *models.CreateIncidentRequest) (*models.Incident, error) {
	incident := &models.Incident{
		ID:          s.generateID(),
		Title:       req.Title,
		Description: req.Description,
		Source:      req.Source,
		Status:      models.StatusOpen,
		Logs:        req.Logs,
		Tags:        req.Tags,
		Metadata:    req.Metadata,
		AssignedTo:  req.AssignedTo,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Use provided severity or classify with AI
	if req.Severity != nil {
		incident.Severity = *req.Severity
	} else {
		// Try to classify severity using AI
		severity := s.classifySeverity(incident)
		incident.Severity = severity
	}

	if incident.Metadata == nil {
		incident.Metadata = make(map[string]interface{})
	}

	// Store the incident
	s.store.mu.Lock()
	s.store.incidents[incident.ID] = incident
	s.store.mu.Unlock()

	s.logger.Info("incident created", zap.String("id", incident.ID), zap.String("title", incident.Title))
	return incident, nil
}

// GetIncident retrieves an incident by ID
func (s *IncidentService) GetIncident(id string) (*models.Incident, error) {
	s.store.mu.RLock()
	incident, ok := s.store.incidents[id]
	s.store.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("incident not found: %s", id)
	}

	return incident, nil
}

// ListIncidents returns all incidents with optional filtering
func (s *IncidentService) ListIncidents(filterStatus *models.IncidentStatus, filterSeverity *models.Severity) ([]*models.Incident, error) {
	var results []*models.Incident

	s.store.mu.RLock()
	for _, incident := range s.store.incidents {
		// Check status filter
		if filterStatus != nil && incident.Status != *filterStatus {
			continue
		}

		// Check severity filter
		if filterSeverity != nil && incident.Severity != *filterSeverity {
			continue
		}

		results = append(results, incident)
	}
	s.store.mu.RUnlock()

	return results, nil
}

// UpdateIncident updates an existing incident
func (s *IncidentService) UpdateIncident(id string, req *models.UpdateIncidentRequest) (*models.Incident, error) {
	s.store.mu.Lock()
	incident, ok := s.store.incidents[id]
	s.store.mu.Unlock()

	if !ok {
		return nil, fmt.Errorf("incident not found: %s", id)
	}

	// Update fields if provided
	if req.Title != nil {
		incident.Title = *req.Title
	}

	if req.Description != nil {
		incident.Description = *req.Description
	}

	if req.Severity != nil {
		incident.Severity = *req.Severity
	}

	if req.Status != nil {
		oldStatus := incident.Status
		incident.Status = *req.Status

		// Set resolved time when status changes to resolved
		if *req.Status == models.StatusResolved && oldStatus != models.StatusResolved {
			now := time.Now()
			incident.ResolvedAt = &now
		}
	}

	if len(req.Logs) > 0 {
		incident.Logs = req.Logs
	}

	if req.Tags != nil {
		incident.Tags = req.Tags
	}

	if req.Metadata != nil {
		incident.Metadata = req.Metadata
	}

	if req.AssignedTo != nil {
		incident.AssignedTo = *req.AssignedTo
	}

	incident.UpdatedAt = time.Now()

	s.logger.Info("incident updated", zap.String("id", incident.ID))
	return incident, nil
}

// DeleteIncident deletes an incident
func (s *IncidentService) DeleteIncident(id string) error {
	s.store.mu.Lock()
	if _, ok := s.store.incidents[id]; !ok {
		s.store.mu.Unlock()
		return fmt.Errorf("incident not found: %s", id)
	}

	delete(s.store.incidents, id)
	s.store.mu.Unlock()

	s.logger.Info("incident deleted", zap.String("id", id))
	return nil
}

// AnalyzeIncident generates AI analysis for an incident
func (s *IncidentService) AnalyzeIncident(id string) (*models.Incident, error) {
	// Get the incident first
	incident, err := s.GetIncident(id)
	if err != nil {
		return nil, err
	}

	// Call AI client to analyze
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	analysisReq := ai.AnalysisRequest{
		IncidentTitle: incident.Title,
		IncidentDesc:  incident.Description,
		Logs:          incident.Logs,
	}

	analysis, err := s.aiClient.AnalyzeIncident(ctx, analysisReq)
	if err != nil {
		s.logger.Error("failed to analyze incident", zap.String("id", id), zap.Error(err))
		return incident, err
	}

	// Convert AI response to model
	s.store.mu.Lock()
	incident.AIAnalysis = &models.AIAnalysis{
		Summary:            analysis.Summary,
		Findings:           analysis.Findings,
		RootCauses:         analysis.RootCauses,
		RecommendedActions: analysis.RecommendedActions,
		SeveritySuggestion: models.Severity(analysis.SuggestedSeverity),
		GeneratedAt:        time.Now(),
		Model:              s.aiClient.Model(),
		Provider:           string(s.aiClient.Provider()),
	}
	incident.UpdatedAt = time.Now()
	s.store.mu.Unlock()

	s.logger.Info("incident analyzed", zap.String("id", id), zap.String("provider", string(s.aiClient.Provider())))
	return incident, nil
}

// GenerateRCA generates a root cause analysis document
func (s *IncidentService) GenerateRCA(id string) (*models.Incident, error) {
	// Get the incident first
	incident, err := s.GetIncident(id)
	if err != nil {
		return nil, err
	}

	// Use existing analysis or create empty one
	var analysis ai.AnalysisResponse
	if incident.AIAnalysis != nil {
		analysis = ai.AnalysisResponse{
			Summary:            incident.AIAnalysis.Summary,
			Findings:           incident.AIAnalysis.Findings,
			RootCauses:         incident.AIAnalysis.RootCauses,
			RecommendedActions: incident.AIAnalysis.RecommendedActions,
			SuggestedSeverity:  string(incident.AIAnalysis.SeveritySuggestion),
		}
	}

	rcaReq := ai.RCARequest{
		IncidentTitle: incident.Title,
		IncidentDesc:  incident.Description,
		Analysis:      analysis,
		Timeline:      buildTimeline(incident),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	rca, err := s.aiClient.GenerateRCA(ctx, rcaReq)
	if err != nil {
		s.logger.Error("failed to generate RCA", zap.String("id", id), zap.Error(err))
		return incident, err
	}

	// Convert AI response to model
	s.store.mu.Lock()
	incident.RCADocument = &models.RCADocument{
		Timeline:            buildTimeline(incident),
		RootCause:           rca.RootCause,
		Impact:              rca.Impact,
		ImmediateResolution: rca.ImmediateResolution,
		PreventiveMeasures:  rca.PreventiveMeasures,
		LessonsLearned:      rca.LessonsLearned,
		GeneratedAt:         time.Now(),
		Model:               s.aiClient.Model(),
		Provider:            string(s.aiClient.Provider()),
	}
	incident.UpdatedAt = time.Now()
	s.store.mu.Unlock()

	s.logger.Info("RCA generated", zap.String("id", id), zap.String("provider", string(s.aiClient.Provider())))
	return incident, nil
}

// SummarizeLogs extracts insights from log collections
func (s *IncidentService) SummarizeLogs(logs []string) (*models.LogSummarizeResponse, error) {
	summarizeReq := ai.SummarizeRequest{
		Logs: logs,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	summary, err := s.aiClient.SummarizeLogs(ctx, summarizeReq)
	if err != nil {
		s.logger.Error("failed to summarize logs", zap.Error(err))
		return nil, err
	}

	return &models.LogSummarizeResponse{
		Summary:     summary.Summary,
		KeyInsights: summary.KeyInsights,
		Alerts:      summary.Alerts,
		GeneratedAt: time.Now(),
	}, nil
}

// Private helper methods

// generateID creates a unique incident ID
func (s *IncidentService) generateID() string {
	s.store.mu.Lock()
	s.store.counter++
	defer s.store.mu.Unlock()
	return fmt.Sprintf("INC-%d-%d", time.Now().Unix(), s.store.counter)
}

// classifySeverity classifies incident severity based on keywords
func (s *IncidentService) classifySeverity(incident *models.Incident) models.Severity {
	// Basic heuristics for severity classification
	desc := incident.Title + " " + incident.Description

	if hasKeyword(desc, "critical", "production down", "data loss", "security breach") {
		return models.SeverityCritical
	}

	if hasKeyword(desc, "error", "failure", "down", "unavailable") {
		return models.SeverityHigh
	}

	if hasKeyword(desc, "warning", "degraded", "slow", "high memory") {
		return models.SeverityMedium
	}

	return models.SeverityLow
}

// Utility functions

// hasKeyword checks if text contains any of the keywords (case-insensitive)
func hasKeyword(text string, keywords ...string) bool {
	for _, kw := range keywords {
		if len(text) > 0 && len(kw) > 0 {
			// Case-insensitive search
			t := text
			for i := 0; i < len(t)-len(kw)+1; i++ {
				if t[i:i+len(kw)] == kw {
					return true
				}
			}
		}
	}
	return false
}

// buildTimeline builds a timeline of incident events
func buildTimeline(incident *models.Incident) []string {
	timeline := []string{
		fmt.Sprintf("Created: %s", incident.CreatedAt.Format(time.RFC3339)),
	}

	if incident.ResolvedAt != nil {
		timeline = append(timeline, fmt.Sprintf("Resolved: %s", incident.ResolvedAt.Format(time.RFC3339)))
	}

	return timeline
}
