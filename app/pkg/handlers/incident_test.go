package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Prakash-sa/terraform-aws/app/pkg/ai"
	"github.com/Prakash-sa/terraform-aws/app/pkg/models"
	"github.com/Prakash-sa/terraform-aws/app/pkg/service"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// MockAIClient for handler testing
type MockAIClient struct{}

func (m *MockAIClient) AnalyzeIncident(ctx context.Context, req ai.AnalysisRequest) (*ai.AnalysisResponse, error) {
	return &ai.AnalysisResponse{
		Summary:            "Mock analysis",
		Findings:           []string{},
		RootCauses:         []string{},
		RecommendedActions: []string{},
		SuggestedSeverity:  "medium",
	}, nil
}

func (m *MockAIClient) GenerateRCA(ctx context.Context, req ai.RCARequest) (*ai.RCAResponse, error) {
	return &ai.RCAResponse{
		Timeline:            "Mock timeline",
		RootCause:           "Mock cause",
		Impact:              "Mock impact",
		ImmediateResolution: "Mock resolution",
		PreventiveMeasures:  []string{},
		LessonsLearned:      []string{},
	}, nil
}

func (m *MockAIClient) SummarizeLogs(ctx context.Context, req ai.SummarizeRequest) (*ai.SummarizeResponse, error) {
	return &ai.SummarizeResponse{
		Summary:     "Mock summary",
		KeyInsights: []string{},
		Alerts:      []string{},
	}, nil
}

func (m *MockAIClient) Health(ctx context.Context) error {
	return nil
}

func (m *MockAIClient) Provider() ai.Provider {
	return "mock"
}

func (m *MockAIClient) Model() string {
	return "mock-model"
}

func setupTestHandler() *IncidentHandler {
	store := service.NewIncidentStore()
	mockAI := &MockAIClient{}
	logger := zap.NewNop()
	svc := service.NewIncidentService(store, mockAI, logger)
	return NewIncidentHandler(svc, logger)
}

func TestCreateIncidentHandler(t *testing.T) {
	handler := setupTestHandler()

	body := models.CreateIncidentRequest{
		Title:       "Test incident",
		Description: "Test description",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/incidents", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateIncident(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var incident models.Incident
	json.NewDecoder(w.Body).Decode(&incident)
	if incident.Title != "Test incident" {
		t.Errorf("expected title 'Test incident', got %q", incident.Title)
	}
}

func TestGetIncidentHandler(t *testing.T) {
	handler := setupTestHandler()

	// Create an incident first
	store := service.NewIncidentStore()
	svc := service.NewIncidentService(store, &MockAIClient{}, zap.NewNop())
	created, _ := svc.CreateIncident(&models.CreateIncidentRequest{
		Title:       "Test",
		Description: "Test",
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/incidents/"+created.ID, nil)
	w := httptest.NewRecorder()

	// Manually set the URL vars (normally done by router)
	vars := map[string]string{"id": created.ID}
	req = mux.SetURLVars(req, vars)

	handler.GetIncident(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestListIncidentsHandler(t *testing.T) {
	handler := setupTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/incidents", nil)
	w := httptest.NewRecorder()

	handler.ListIncidents(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var incidents []*models.Incident
	json.NewDecoder(w.Body).Decode(&incidents)
	if incidents == nil {
		t.Error("expected non-nil incidents array")
	}
}

func TestUpdateIncidentHandler(t *testing.T) {
	handler := setupTestHandler()

	// Create incident first
	store := service.NewIncidentStore()
	svc := service.NewIncidentService(store, &MockAIClient{}, zap.NewNop())
	created, _ := svc.CreateIncident(&models.CreateIncidentRequest{
		Title:       "Test",
		Description: "Test",
	})

	newTitle := "Updated"
	body := models.UpdateIncidentRequest{
		Title: &newTitle,
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/incidents/"+created.ID, bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	vars := map[string]string{"id": created.ID}
	req = mux.SetURLVars(req, vars)

	w := httptest.NewRecorder()
	handler.UpdateIncident(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestDeleteIncidentHandler(t *testing.T) {
	handler := setupTestHandler()

	// Create incident first
	store := service.NewIncidentStore()
	svc := service.NewIncidentService(store, &MockAIClient{}, zap.NewNop())
	created, _ := svc.CreateIncident(&models.CreateIncidentRequest{
		Title:       "Test",
		Description: "Test",
	})

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/incidents/"+created.ID, nil)
	vars := map[string]string{"id": created.ID}
	req = mux.SetURLVars(req, vars)

	w := httptest.NewRecorder()
	handler.DeleteIncident(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status %d, got %d", http.StatusNoContent, w.Code)
	}
}

func TestSummarizeLogsHandler(t *testing.T) {
	handler := setupTestHandler()

	body := models.LogSummarizeRequest{
		Logs: []string{"log 1", "log 2"},
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/logs/summarize", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.SummarizeLogs(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}
