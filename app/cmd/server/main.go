package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/Prakash-sa/terraform-aws/app/pkg/handlers"
	"github.com/Prakash-sa/terraform-aws/app/pkg/service"
)

type ctxKey string

const (
	ctxKeyResponseWriter ctxKey = "responseWriter"
	ctxKeyRequestID      ctxKey = "requestID"
)

var (
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

	logger    *zap.Logger = zap.NewNop()
	startTime             = time.Now()
	once      sync.Once
)

type Server struct {
	router          *mux.Router
	server          *http.Server
	cfg             config
	incidentService *service.IncidentService
	incidentHandler *handlers.IncidentHandler
}

type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
	Uptime    string    `json:"uptime"`
}

type APIResponse struct {
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

type config struct {
	Port        string
	Environment string
	Version     string
	LogLevel    string
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func NewServer(cfg config) *Server {
	s := &Server{
		router: mux.NewRouter(),
		cfg:    cfg,
	}

	s.router.Use(recoverMiddleware)
	s.router.Use(requestContextMiddleware)
	s.router.Use(metricsMiddleware)
	s.router.Use(loggingMiddleware)

	s.router.HandleFunc("/", homeHandler(cfg)).Methods(http.MethodGet)
	s.router.HandleFunc("/health", healthHandler(cfg)).Methods(http.MethodGet)
	s.router.HandleFunc("/ready", readinessHandler).Methods(http.MethodGet)
	s.router.HandleFunc("/api/v1/data", dataHandler).Methods(http.MethodGet)
	s.router.HandleFunc("/api/v1/echo", echoHandler).Methods(http.MethodPost)
	s.router.Handle("/metrics", promhttp.Handler()).Methods(http.MethodGet)

	// Initialize incident management system
	aiCfg := config.LoadConfig()
	aiClient, err := aiCfg.AI.CreateAIClient()
	if err != nil {
		logger.Warn("failed to create AI client", zap.Error(err))
	}

	incidentStore := service.NewIncidentStore()
	incidentService := service.NewIncidentService(incidentStore, aiClient, logger)
	incidentHandler := handlers.NewIncidentHandler(incidentService, logger)
	incidentHandler.RegisterRoutes(s.router)

	s.incidentService = incidentService
	s.incidentHandler = incidentHandler

	s.server = &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	logger.Info("AI configuration loaded", zap.Any("ai_config", aiCfg.AI.Summary()))

	return s
}

func (s *Server) Start() error {
	logger.Info("starting server",
		zap.String("addr", s.server.Addr),
		zap.Time("start_time", startTime),
	)

	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	logger.Info("shutting down server gracefully...")
	return s.server.Shutdown(ctx)
}

func recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				logger.Error("panic recovered", zap.Any("error", rec))
				respondError(w, http.StatusInternalServerError, "internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func requestContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get("X-Request-ID")
		if reqID == "" {
			reqID = requestID()
		}

		rw := newStatusWriter(w)
		rw.Header().Set("X-Request-ID", reqID)

		ctx := context.WithValue(r.Context(), ctxKeyRequestID, reqID)
		ctx = context.WithValue(ctx, ctxKeyResponseWriter, rw)

		next.ServeHTTP(rw, r.WithContext(ctx))
	})
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		activeConnections.Inc()
		defer activeConnections.Dec()

		next.ServeHTTP(w, r)

		rw := responseWriterFromCtx(r.Context())
		status := http.StatusOK
		if rw != nil {
			status = rw.statusCode
		}

		logger.Info("http_request",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Int("status", status),
			zap.String("request_id", requestIDFromCtx(r.Context())),
			zap.Duration("duration", time.Since(start)),
		)
	})
}

func metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		route := mux.CurrentRoute(r)
		path, _ := route.GetPathTemplate()

		next.ServeHTTP(w, r)

		rw := responseWriterFromCtx(r.Context())
		status := http.StatusOK
		if rw != nil {
			status = rw.statusCode
		}

		duration := time.Since(start).Seconds()
		httpRequestDuration.WithLabelValues(r.Method, path).Observe(duration)
		httpRequestsTotal.WithLabelValues(r.Method, path, fmt.Sprintf("%d", status)).Inc()
	})
}

type statusWriter struct {
	http.ResponseWriter
	statusCode int
}

func newStatusWriter(w http.ResponseWriter) *statusWriter {
	return &statusWriter{ResponseWriter: w, statusCode: http.StatusOK}
}

func (rw *statusWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func responseWriterFromCtx(ctx context.Context) *statusWriter {
	rw, ok := ctx.Value(ctxKeyResponseWriter).(*statusWriter)
	if !ok {
		return nil
	}
	return rw
}

func requestIDFromCtx(ctx context.Context) string {
	reqID, ok := ctx.Value(ctxKeyRequestID).(string)
	if !ok {
		return ""
	}
	return reqID
}

func homeHandler(cfg config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response := APIResponse{
			Message:   "Welcome to the Production-Ready Go API",
			Timestamp: time.Now(),
			Data: map[string]string{
				"version":     cfg.Version,
				"environment": cfg.Environment,
			},
		}
		respondJSON(w, http.StatusOK, response)
	}
}

func healthHandler(cfg config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response := HealthResponse{
			Status:    "healthy",
			Timestamp: time.Now(),
			Version:   cfg.Version,
			Uptime:    time.Since(startTime).String(),
		}
		respondJSON(w, http.StatusOK, response)
	}
}

func readinessHandler(w http.ResponseWriter, r *http.Request) {
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
		respondError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	response := APIResponse{
		Message:   "Echo response",
		Data:      payload,
		Timestamp: time.Now(),
	}
	respondJSON(w, http.StatusOK, response)
}

func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		logger.Error("failed to encode JSON response", zap.Error(err))
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
	cfg := loadConfig()

	var err error
	logger, err = newLogger(cfg.LogLevel)
	if err != nil {
		panic(fmt.Sprintf("failed to initialize logger: %v", err))
	}
	defer logger.Sync()

	healthCheck := flag.Bool("health-check", false, "perform a local health check instead of starting the server")
	flag.Parse()

	if *healthCheck {
		if err := runHealthCheck(cfg); err != nil {
			logger.Fatal("health check failed", zap.Error(err))
		}
		logger.Info("health check passed")
		return
	}

	server := NewServer(cfg)

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			logger.Fatal("server shutdown failed", zap.Error(err))
		}
	}()

	logger.Info("server starting...")
	if err := server.Start(); err != nil && err != http.ErrServerClosed {
		logger.Fatal("server failed to start", zap.Error(err))
	}

	logger.Info("server stopped")
}

func newLogger(level string) (*zap.Logger, error) {
	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(parseLevel(level))
	cfg.Encoding = "json"
	cfg.EncoderConfig.TimeKey = "ts"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	return cfg.Build()
}

func parseLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

func loadConfig() config {
	return config{
		Port:        getEnv("PORT", "8080"),
		Environment: getEnv("ENVIRONMENT", "production"),
		Version:     getEnv("APP_VERSION", "1.0.0"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
	}
}

func requestID() string {
	return fmt.Sprintf("%d", rand.Int63())
}

func runHealthCheck(cfg config) error {
	client := &http.Client{
		Timeout: 3 * time.Second,
	}
	resp, err := client.Get("http://127.0.0.1:" + cfg.Port + "/health")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unhealthy status code: %d", resp.StatusCode)
	}

	return nil
}
