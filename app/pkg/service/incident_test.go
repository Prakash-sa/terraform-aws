package service

import (
	"context"
	"testing"

	"github.com/Prakash-sa/terraform-aws/app/pkg/ai"
	"github.com/Prakash-sa/terraform-aws/app/pkg/models"
	"go.uber.org/zap"
)

// MockAIClient is a mock implementation of ai.Client for testing
type MockAIClient struct {
	analyzeErr    error
	rcaErr        error
	summarizeErr  error
	lastAnalysis  ai.AnalysisRequest
	lastRCA       ai.RCARequest
	lastSummarize ai.SummarizeRequest
}

func (m *MockAIClient) AnalyzeIncident(ctx context.Context, req ai.AnalysisRequest) (*ai.AnalysisResponse, error) {
	m.lastAnalysis = req
	if m.analyzeErr != nil {
		return nil, m.analyzeErr
	}
	return &ai.AnalysisResponse{
		Summary:            "Test analysis summary",
		Findings:           []string{"Finding 1", "Finding 2"},
		RootCauses:         []string{"Root cause 1"},
		RecommendedActions: []string{"Action 1"},
		SuggestedSeverity:  "high",
	}, nil
}

func (m *MockAIClient) GenerateRCA(ctx context.Context, req ai.RCARequest) (*ai.RCAResponse, error) {
	m.lastRCA = req
	if m.rcaErr != nil {
		return nil, m.rcaErr
	}
	return &ai.RCAResponse{
		Timeline:            "Timeline details",
		RootCause:           "Root cause details",
		Impact:              "Impact details",
		ImmediateResolution: "Immediate resolution",
		PreventiveMeasures:  []string{"Measure 1"},
		LessonsLearned:      []string{"Lesson 1"},
	}, nil
}

func (m *MockAIClient) SummarizeLogs(ctx context.Context, req ai.SummarizeRequest) (*ai.SummarizeResponse, error) {
	m.lastSummarize = req
	if m.summarizeErr != nil {
		return nil, m.summarizeErr
	}
	return &ai.SummarizeResponse{
		Summary:     "Log summary",
		KeyInsights: []string{"Insight 1"},
		Alerts:      []string{"Alert 1"},
	}, nil
}

func (m *MockAIClient) Health(ctx context.Context) error {
	return nil
}

func (m *MockAIClient) Provider() ai.Provider {
	return ai.ProviderOpenAI
}

func (m *MockAIClient) Model() string {
	return "gpt-4"
}

func TestCreateIncident(t *testing.T) {
	store := NewIncidentStore()
	mockAI := &MockAIClient{}
	logger := zap.NewNop()
	service := NewIncidentService(store, mockAI, logger)

	req := &models.CreateIncidentRequest{
		Title:       "Test incident",
		Description: "Test description",
		Source:      "test",
	}

	incident, err := service.CreateIncident(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if incident.ID == "" {
		t.Error("incident ID should not be empty")
	}
	if incident.Title != "Test incident" {
		t.Errorf("expected title 'Test incident', got %q", incident.Title)
	}
	if incident.Status != models.StatusOpen {
		t.Errorf("expected status 'open', got %q", incident.Status)
	}
}

func TestGetIncident(t *testing.T) {
	store := NewIncidentStore()
	mockAI := &MockAIClient{}
	logger := zap.NewNop()
	service := NewIncidentService(store, mockAI, logger)

	// Create incident first
	req := &models.CreateIncidentRequest{
		Title:       "Test",
		Description: "Test",
	}
	created, _ := service.CreateIncident(req)

	// Get incident
	retrieved, err := service.GetIncident(created.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if retrieved.ID != created.ID {
		t.Errorf("expected ID %q, got %q", created.ID, retrieved.ID)
	}
}

func TestGetIncidentNotFound(t *testing.T) {
	store := NewIncidentStore()
	mockAI := &MockAIClient{}
	logger := zap.NewNop()
	service := NewIncidentService(store, mockAI, logger)

	_, err := service.GetIncident("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent incident")
	}
}

func TestListIncidents(t *testing.T) {
	store := NewIncidentStore()
	mockAI := &MockAIClient{}
	logger := zap.NewNop()
	service := NewIncidentService(store, mockAI, logger)

	// Create a few incidents
	for i := 0; i < 3; i++ {
		service.CreateIncident(&models.CreateIncidentRequest{
			Title:       "Test",
			Description: "Test",
		})
	}

	incidents, err := service.ListIncidents(nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(incidents) != 3 {
		t.Errorf("expected 3 incidents, got %d", len(incidents))
	}
}

func TestUpdateIncident(t *testing.T) {
	store := NewIncidentStore()
	mockAI := &MockAIClient{}
	logger := zap.NewNop()
	service := NewIncidentService(store, mockAI, logger)

	// Create incident first
	created, _ := service.CreateIncident(&models.CreateIncidentRequest{
		Title:       "Test",
		Description: "Test",
	})

	// Update incident
	newTitle := "Updated title"
	updated, err := service.UpdateIncident(created.ID, &models.UpdateIncidentRequest{
		Title: &newTitle,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if updated.Title != "Updated title" {
		t.Errorf("expected title 'Updated title', got %q", updated.Title)
	}
}

func TestDeleteIncident(t *testing.T) {
	store := NewIncidentStore()
	mockAI := &MockAIClient{}
	logger := zap.NewNop()
	service := NewIncidentService(store, mockAI, logger)

	// Create incident first
	created, _ := service.CreateIncident(&models.CreateIncidentRequest{
		Title:       "Test",
		Description: "Test",
	})

	// Delete incident
	err := service.DeleteIncident(created.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify it's deleted
	_, err = service.GetIncident(created.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestAnalyzeIncident(t *testing.T) {
	store := NewIncidentStore()
	mockAI := &MockAIClient{}
	logger := zap.NewNop()
	service := NewIncidentService(store, mockAI, logger)

	// Create incident first
	created, _ := service.CreateIncident(&models.CreateIncidentRequest{
		Title:       "Test",
		Description: "Test",
	})

	// Analyze incident
	analyzed, err := service.AnalyzeIncident(created.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if analyzed.AIAnalysis == nil {
		t.Error("expected AI analysis to be present")
	}
	if analyzed.AIAnalysis.Summary != "Test analysis summary" {
		t.Errorf("expected summary 'Test analysis summary', got %q", analyzed.AIAnalysis.Summary)
	}
}

func TestSummarizeLogs(t *testing.T) {
	store := NewIncidentStore()
	mockAI := &MockAIClient{}
	logger := zap.NewNop()
	service := NewIncidentService(store, mockAI, logger)

	logs := []string{"log 1", "log 2", "log 3"}
	summary, err := service.SummarizeLogs(logs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if summary.Summary != "Log summary" {
		t.Errorf("expected summary 'Log summary', got %q", summary.Summary)
	}
}
