package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Prakash-sa/terraform-aws/app/internal/ai"
	"github.com/Prakash-sa/terraform-aws/app/internal/incident"
	"github.com/Prakash-sa/terraform-aws/app/pkg/models"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

var (
	logger              *zap.Logger
	startTime           = time.Now()
	httpRequestsTotal   *prometheus.CounterVec
	httpRequestDuration *prometheus.HistogramVec
	activeConnections   prometheus.Gauge
)

func init() {
	var err error
	logger, err = zap.NewProduction()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}

	// Prometheus metrics
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	activeConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_connections",
			Help: "Number of active connections",
		},
	)
}

// Server represents the HTTP server
type Server struct {
	router          *mux.Router
	server          *http.Server
	incidentService *incident.Service
}

// HealthResponse represents health check response
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
	Uptime    string    `json:"uptime"`
}

// APIResponse represents a generic API response
type APIResponse struct {
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// NewServer creates a new HTTP server
func NewServer() *Server {
	s := &Server{
		router: mux.NewRouter(),
	}

	// Initialize AI client
	aiConfig := ai.Config{
		Provider:       ai.Provider(getEnv("AI_PROVIDER", "openai")),
		OpenAIKey:      getEnv("OPENAI_API_KEY", ""),
		AnthropicKey:   getEnv("ANTHROPIC_API_KEY", ""),
		OpenAIModel:    getEnv("OPENAI_MODEL", "gpt-4"),
		AnthropicModel: getEnv("ANTHROPIC_MODEL", "claude-3-5-sonnet-20241022"),
	}

	var aiClient ai.Client
	var err error
	if aiConfig.OpenAIKey != "" || aiConfig.AnthropicKey != "" {
		aiClient, err = ai.NewClient(aiConfig)
		if err != nil {
			logger.Warn("Failed to initialize AI client", zap.Error(err))
		} else {
			logger.Info("AI client initialized", zap.String("provider", string(aiConfig.Provider)))
		}
	} else {
		logger.Info("AI client not configured - API keys not provided")
	}

	// Initialize incident service
	s.incidentService = incident.NewService(aiClient)

	// Middleware
	s.router.Use(loggingMiddleware)
	s.router.Use(metricsMiddleware)

	// Routes
	s.router.HandleFunc("/", homeHandler).Methods("GET")
	s.router.HandleFunc("/health", healthHandler).Methods("GET")
	s.router.HandleFunc("/ready", readinessHandler).Methods("GET")
	s.router.HandleFunc("/api/v1/data", dataHandler).Methods("GET")
	s.router.HandleFunc("/api/v1/echo", echoHandler).Methods("POST")
	s.router.Handle("/metrics", promhttp.Handler()).Methods("GET")

	// Incident management endpoints
	s.router.HandleFunc("/api/v1/incidents", s.createIncidentHandler).Methods("POST")
	s.router.HandleFunc("/api/v1/incidents", s.listIncidentsHandler).Methods("GET")
	s.router.HandleFunc("/api/v1/incidents/{id}", s.getIncidentHandler).Methods("GET")
	s.router.HandleFunc("/api/v1/incidents/{id}/analyze", s.analyzeIncidentHandler).Methods("POST")
	s.router.HandleFunc("/api/v1/incidents/{id}/analysis", s.getAnalysisHandler).Methods("GET")
	s.router.HandleFunc("/api/v1/incidents/{id}/rca", s.getRCAHandler).Methods("GET")
	s.router.HandleFunc("/api/v1/incidents/{id}/rca/generate", s.generateRCAHandler).Methods("POST")
	s.router.HandleFunc("/api/v1/logs/summarize", s.summarizeLogsHandler).Methods("POST")

	port := getEnv("PORT", "8080")
	s.server = &http.Server{
		Addr:         ":" + port,
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return s
}

func (s *Server) Start() error {
	logger.Info("Starting server",
		zap.String("port", s.server.Addr),
		zap.Time("start_time", startTime),
	)
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	logger.Info("Shutting down server gracefully...")
	return s.server.Shutdown(ctx)
}

// Middleware
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		activeConnections.Inc()
		defer activeConnections.Dec()

		logger.Info("Request received",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("remote_addr", r.RemoteAddr),
		)

		next.ServeHTTP(w, r)

		logger.Info("Request completed",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Duration("duration", time.Since(start)),
		)
	})
}

func metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		route := mux.CurrentRoute(r)
		path, _ := route.GetPathTemplate()

		// Wrap response writer to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start).Seconds()
		httpRequestDuration.WithLabelValues(r.Method, path).Observe(duration)
		httpRequestsTotal.WithLabelValues(r.Method, path, fmt.Sprintf("%d", wrapped.statusCode)).Inc()
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Handlers
func homeHandler(w http.ResponseWriter, r *http.Request) {
	response := APIResponse{
		Message:   "Welcome to the Production-Ready Go API with AI-Powered Incident Automation",
		Timestamp: time.Now(),
		Data: map[string]string{
			"version":     getEnv("APP_VERSION", "1.0.0"),
			"environment": getEnv("ENVIRONMENT", "production"),
		},
	}
	respondJSON(w, http.StatusOK, response)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Version:   getEnv("APP_VERSION", "1.0.0"),
		Uptime:    time.Since(startTime).String(),
	}
	respondJSON(w, http.StatusOK, response)
}

func readinessHandler(w http.ResponseWriter, r *http.Request) {
	// Add your readiness checks here (database, cache, etc.)
	// For now, we'll just return ready
	response := map[string]interface{}{
		"ready":     true,
		"timestamp": time.Now(),
		"checks": map[string]string{
			"database": "ok",
			"cache":    "ok",
		},
	}
	respondJSON(w, http.StatusOK, response)
}

func dataHandler(w http.ResponseWriter, r *http.Request) {
	// Simulate data retrieval
	data := map[string]interface{}{
		"items": []map[string]interface{}{
			{"id": 1, "name": "Item 1", "price": 99.99},
			{"id": 2, "name": "Item 2", "price": 149.99},
			{"id": 3, "name": "Item 3", "price": 199.99},
		},
		"total": 3,
	}

	response := APIResponse{
		Message:   "Data retrieved successfully",
		Data:      data,
		Timestamp: time.Now(),
	}
	respondJSON(w, http.StatusOK, response)
}

func echoHandler(w http.ResponseWriter, r *http.Request) {
	var payload map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	response := APIResponse{
		Message:   "Echo response",
		Data:      payload,
		Timestamp: time.Now(),
	}
	respondJSON(w, http.StatusOK, response)
}

// Incident management handlers
func (s *Server) createIncidentHandler(w http.ResponseWriter, r *http.Request) {
	var req models.CreateIncidentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	incident, err := s.incidentService.CreateIncident(r.Context(), req)
	if err != nil {
		logger.Error("Failed to create incident", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "Failed to create incident")
		return
	}

	respondJSON(w, http.StatusCreated, incident)
}

func (s *Server) listIncidentsHandler(w http.ResponseWriter, r *http.Request) {
	incidents := s.incidentService.ListIncidents()
	respondJSON(w, http.StatusOK, incidents)
}

func (s *Server) getIncidentHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	incidentID := vars["id"]

	incident, err := s.incidentService.GetIncident(incidentID)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, incident)
}

func (s *Server) analyzeIncidentHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	incidentID := vars["id"]

	analysis, err := s.incidentService.AnalyzeIncident(r.Context(), incidentID)
	if err != nil {
		logger.Error("Failed to analyze incident", zap.Error(err))
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, analysis)
}

func (s *Server) getAnalysisHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	incidentID := vars["id"]

	analysis, err := s.incidentService.GetAnalysis(incidentID)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, analysis)
}

func (s *Server) generateRCAHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	incidentID := vars["id"]

	rca, err := s.incidentService.GenerateRCA(r.Context(), incidentID)
	if err != nil {
		logger.Error("Failed to generate RCA", zap.Error(err))
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, rca)
}

func (s *Server) getRCAHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	incidentID := vars["id"]

	rca, err := s.incidentService.GetRCA(incidentID)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, rca)
}

func (s *Server) summarizeLogsHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Logs []string `json:"logs"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	summary, err := s.incidentService.SummarizeLogs(r.Context(), req.Logs)
	if err != nil {
		logger.Error("Failed to summarize logs", zap.Error(err))
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"summary":   summary,
		"log_count": len(req.Logs),
		"timestamp": time.Now(),
	})
}

// Helper functions
func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		logger.Error("Failed to encode JSON response", zap.Error(err))
	}
}

func respondError(w http.ResponseWriter, status int, message string) {
	response := APIResponse{
		Message:   message,
		Timestamp: time.Now(),
	}
	respondJSON(w, status, response)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	defer logger.Sync()

	server := NewServer()

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			logger.Fatal("Server shutdown failed", zap.Error(err))
		}
	}()

	logger.Info("Server starting...")
	if err := server.Start(); err != nil && err != http.ErrServerClosed {
		logger.Fatal("Server failed to start", zap.Error(err))
	}
	logger.Info("Server stopped")
}
