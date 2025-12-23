package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Prakash-sa/terraform-aws/app/pkg/models"
	"github.com/Prakash-sa/terraform-aws/app/pkg/service"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// IncidentHandler handles incident-related HTTP requests
type IncidentHandler struct {
	incidentService *service.IncidentService
	logger          *zap.Logger
}

// NewIncidentHandler creates a new incident handler
func NewIncidentHandler(incidentService *service.IncidentService, logger *zap.Logger) *IncidentHandler {
	return &IncidentHandler{
		incidentService: incidentService,
		logger:          logger,
	}
}

// RegisterRoutes registers all incident routes
func (h *IncidentHandler) RegisterRoutes(router *mux.Router) {
	// Create a subrouter for API v1
	v1 := router.PathPrefix("/api/v1").Subrouter()

	// Incident endpoints
	v1.HandleFunc("/incidents", h.CreateIncident).Methods(http.MethodPost)
	v1.HandleFunc("/incidents", h.ListIncidents).Methods(http.MethodGet)
	v1.HandleFunc("/incidents/{id}", h.GetIncident).Methods(http.MethodGet)
	v1.HandleFunc("/incidents/{id}", h.UpdateIncident).Methods(http.MethodPut)
	v1.HandleFunc("/incidents/{id}", h.DeleteIncident).Methods(http.MethodDelete)

	// Analysis endpoints
	v1.HandleFunc("/incidents/{id}/analyze", h.AnalyzeIncident).Methods(http.MethodPost)
	v1.HandleFunc("/incidents/{id}/rca/generate", h.GenerateRCA).Methods(http.MethodPost)

	// Log endpoints
	v1.HandleFunc("/logs/summarize", h.SummarizeLogs).Methods(http.MethodPost)
}

// CreateIncident handles POST /api/v1/incidents
func (h *IncidentHandler) CreateIncident(w http.ResponseWriter, r *http.Request) {
	var req models.CreateIncidentRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Title == "" || req.Description == "" {
		respondError(w, http.StatusBadRequest, "title and description are required")
		return
	}

	incident, err := h.incidentService.CreateIncident(&req)
	if err != nil {
		h.logger.Error("failed to create incident", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "failed to create incident")
		return
	}

	respondJSON(w, http.StatusCreated, incident)
}

// GetIncident handles GET /api/v1/incidents/{id}
func (h *IncidentHandler) GetIncident(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	incident, err := h.incidentService.GetIncident(id)
	if err != nil {
		respondError(w, http.StatusNotFound, fmt.Sprintf("incident not found: %s", id))
		return
	}

	respondJSON(w, http.StatusOK, incident)
}

// ListIncidents handles GET /api/v1/incidents
func (h *IncidentHandler) ListIncidents(w http.ResponseWriter, r *http.Request) {
	// Optional query parameters for filtering
	statusParam := r.URL.Query().Get("status")
	severityParam := r.URL.Query().Get("severity")

	var statusFilter *models.IncidentStatus
	var severityFilter *models.Severity

	if statusParam != "" {
		status := models.IncidentStatus(statusParam)
		statusFilter = &status
	}

	if severityParam != "" {
		severity := models.Severity(severityParam)
		severityFilter = &severity
	}

	incidents, err := h.incidentService.ListIncidents(statusFilter, severityFilter)
	if err != nil {
		h.logger.Error("failed to list incidents", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "failed to list incidents")
		return
	}

	if incidents == nil {
		incidents = []*models.Incident{}
	}

	respondJSON(w, http.StatusOK, incidents)
}

// UpdateIncident handles PUT /api/v1/incidents/{id}
func (h *IncidentHandler) UpdateIncident(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var req models.UpdateIncidentRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	incident, err := h.incidentService.UpdateIncident(id, &req)
	if err != nil {
		if err.Error() == fmt.Sprintf("incident not found: %s", id) {
			respondError(w, http.StatusNotFound, err.Error())
		} else {
			h.logger.Error("failed to update incident", zap.String("id", id), zap.Error(err))
			respondError(w, http.StatusInternalServerError, "failed to update incident")
		}
		return
	}

	respondJSON(w, http.StatusOK, incident)
}

// DeleteIncident handles DELETE /api/v1/incidents/{id}
func (h *IncidentHandler) DeleteIncident(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	err := h.incidentService.DeleteIncident(id)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AnalyzeIncident handles POST /api/v1/incidents/{id}/analyze
func (h *IncidentHandler) AnalyzeIncident(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	incident, err := h.incidentService.AnalyzeIncident(id)
	if err != nil {
		response := map[string]interface{}{
			"incident": incident,
			"error":    err.Error(),
		}
		// Still return the incident with error message
		h.logger.Warn("analysis encountered error but returning result", zap.String("id", id), zap.Error(err))
		respondJSON(w, http.StatusOK, response)
		return
	}

	respondJSON(w, http.StatusOK, incident)
}

// GenerateRCA handles POST /api/v1/incidents/{id}/rca/generate
func (h *IncidentHandler) GenerateRCA(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	incident, err := h.incidentService.GenerateRCA(id)
	if err != nil {
		response := map[string]interface{}{
			"incident": incident,
			"error":    err.Error(),
		}
		// Still return the incident with error message
		h.logger.Warn("RCA generation encountered error but returning result", zap.String("id", id), zap.Error(err))
		respondJSON(w, http.StatusOK, response)
		return
	}

	respondJSON(w, http.StatusOK, incident)
}

// SummarizeLogs handles POST /api/v1/logs/summarize
func (h *IncidentHandler) SummarizeLogs(w http.ResponseWriter, r *http.Request) {
	var req models.LogSummarizeRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if len(req.Logs) == 0 {
		respondError(w, http.StatusBadRequest, "logs array is required")
		return
	}

	summary, err := h.incidentService.SummarizeLogs(req.Logs)
	if err != nil {
		// Still return with error message
		h.logger.Warn("log summarization encountered error but returning result", zap.Error(err))
		summary = &models.LogSummarizeResponse{
			Summary:     fmt.Sprintf("Summarization failed: %v", err),
			KeyInsights: []string{},
			Alerts:      []string{},
			GeneratedAt: time.Now(),
		}
	}

	respondJSON(w, http.StatusOK, summary)
}

// Response helpers

// APIResponse represents a standard API response
type APIResponse struct {
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// respondJSON sends a JSON response
func respondJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// respondError sends an error JSON response
func respondError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(APIResponse{
		Error:     message,
		Timestamp: time.Now(),
	})
}
