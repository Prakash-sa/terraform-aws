# Incident Management API - File Reference Guide

## Quick File Lookup

### 📦 Core Packages

#### Models (`app/pkg/models/incident.go`)
Data structures for the incident management system.

**Key Types:**
- `Incident` - Core incident entity with lifecycle tracking
- `Severity` - Enum: critical, high, medium, low, unknown
- `IncidentStatus` - Enum: open, in_progress, resolved, closed
- `AIAnalysis` - AI-generated analysis results
- `RCADocument` - Root Cause Analysis structure
- Request/Response DTOs

**Usage:**
```go
import "github.com/Prakash-sa/terraform-aws/app/pkg/models"

incident := &models.Incident{
    ID: "INC-123",
    Title: "API Down",
    Severity: models.SeverityCritical,
    Status: models.StatusOpen,
}
```

---

#### AI Client (`app/pkg/ai/`)

**client.go** - Provider-agnostic AI interface
- `Client` interface for AI operations
- `NewClient()` factory function
- `NoOpClient` for graceful degradation
- Request/Response types

**openai.go** - OpenAI/GPT-4 implementation
- `OpenAIClient` struct
- Chat completion API integration
- Response parsing

**anthropic.go** - Anthropic Claude implementation
- `AnthropicClient` struct
- Messages API integration
- Response parsing

**parsing.go** - Utilities for response parsing
- `extractJSON()` - Extract JSON from markdown
- `parseAnalysisResponse()` - Parse analysis
- `parseRCAResponse()` - Parse RCA
- `parseSummarizeResponse()` - Parse summaries

**Usage:**
```go
import "github.com/Prakash-sa/terraform-aws/app/pkg/ai"

client, err := ai.NewClient(ai.ClientConfig{
    Provider: ai.ProviderOpenAI,
    APIKey: "sk-xxx",
    Model: "gpt-4",
})

analysis, err := client.AnalyzeIncident(ctx, request)
```

---

#### Service (`app/pkg/service/incident.go`)
Business logic and CRUD operations.

**Key Types:**
- `IncidentStore` - Thread-safe in-memory storage
- `IncidentService` - Business logic layer

**Key Methods:**
- `CreateIncident()` - Create with auto-classification
- `GetIncident()` - Retrieve by ID
- `ListIncidents()` - List with optional filtering
- `UpdateIncident()` - Update fields
- `DeleteIncident()` - Delete incident
- `AnalyzeIncident()` - AI analysis
- `GenerateRCA()` - RCA generation
- `SummarizeLogs()` - Log analysis

**Usage:**
```go
import "github.com/Prakash-sa/terraform-aws/app/pkg/service"

store := service.NewIncidentStore()
svc := service.NewIncidentService(store, aiClient, logger)

incident, err := svc.CreateIncident(&models.CreateIncidentRequest{
    Title: "API Down",
    Description: "...",
})
```

**Tests:** `incident_test.go`
- 13+ unit tests
- Mock AI client
- Concurrent operation tests

---

#### Handlers (`app/pkg/handlers/incident.go`)
HTTP request/response handling.

**Key Types:**
- `IncidentHandler` - HTTP handler struct
- `APIResponse` - Standard response wrapper

**Key Methods:**
- `RegisterRoutes()` - Register all endpoints
- `CreateIncident()` - POST /incidents
- `GetIncident()` - GET /incidents/{id}
- `ListIncidents()` - GET /incidents
- `UpdateIncident()` - PUT /incidents/{id}
- `DeleteIncident()` - DELETE /incidents/{id}
- `AnalyzeIncident()` - POST /incidents/{id}/analyze
- `GenerateRCA()` - POST /incidents/{id}/rca/generate
- `SummarizeLogs()` - POST /logs/summarize

**Usage:**
```go
import "github.com/Prakash-sa/terraform-aws/app/pkg/handlers"

handler := handlers.NewIncidentHandler(service, logger)
handler.RegisterRoutes(muxRouter)
```

**Tests:** `incident_test.go`
- 8+ handler tests
- HTTP status verification
- Request/response validation

---

#### Configuration (`app/pkg/config/config.go`)
Environment-based configuration.

**Key Types:**
- `Config` - Main configuration
- `AIConfig` - AI provider configuration
- `ServerConfig` - Server settings
- `LogConfig` - Logging settings

**Key Functions:**
- `LoadConfig()` - Load from environment
- `CreateAIClient()` - Initialize AI client
- `Validate()` - Validate configuration

**Environment Variables:**
```
AI_PROVIDER=openai|anthropic
OPENAI_API_KEY=sk-...
ANTHROPIC_API_KEY=sk-ant-...
AI_TIMEOUT=60
AI_TEMPERATURE=0.7
AI_MAX_TOKENS=2000
PORT=8080
LOG_LEVEL=info|debug|warn|error
```

---

### 🚀 Entry Point

#### Main (`app/cmd/server/main.go`)
Application entry point with incident service integration.

**Integration Points:**
- Loads AI configuration
- Creates incident service
- Registers API routes
- Sets up middleware and metrics
- Handles graceful shutdown

**Key Changes:**
- Added incident handler registration
- Added AI client initialization
- Added service layer setup
- Enhanced logging for AI configuration

---

### 🐳 Deployment

#### Dockerfile (`build/Dockerfile`)
Multi-stage Docker build for production.

**Stages:**
1. **Builder** - Go 1.21 alpine, compile binary
2. **Runtime** - Scratch image, non-root user

**Features:**
- Build-time version injection
- Multi-stage optimization
- Health check configuration
- Non-root user (UID 1000)
- CA certificates included

---

#### Helm Chart (`deploy/helm/app/`)

**values.yaml** - Configuration values
- Updated with AI environment variables
- Added AI secret configuration
- Security contexts and policies
- Resource limits and autoscaling
- Pod anti-affinity

**deployment.yaml** - Kubernetes deployment
- Updated to include AI secret injection
- Environment variable configuration
- Resource specifications
- Security context

**ai-secret.yaml** - AI credentials secret template
- Separate secret for API keys
- Support for OpenAI and Anthropic
- Base64-encoded values
- Secure handling

---

### 📚 Documentation

#### INCIDENT_API.md
**Comprehensive API reference (500+ lines)**

Sections:
- Overview and features
- Data models with examples
- REST API endpoints with curl examples
- Configuration guide
- Error handling
- Security best practices
- Performance considerations
- Future enhancements

**Best for:** API integration, endpoint reference

---

#### INCIDENT_API_QUICKSTART.md
**Quick start guide (300+ lines)**

Sections:
- Prerequisites
- Local development setup
- Docker deployment
- Kubernetes deployment
- Quick API tests with examples
- Running tests
- Troubleshooting
- Configuration reference

**Best for:** Getting started quickly, basic testing

---

#### IMPLEMENTATION_SUMMARY.md
**Technical implementation details (400+ lines)**

Sections:
- Components implemented
- Technical highlights
- Security features
- Performance characteristics
- Integration points
- Production considerations
- Files created/modified
- Next steps

**Best for:** Understanding architecture, integration planning

---

#### DEPLOYMENT_CHECKLIST.md
**Production deployment guide (350+ lines)**

Sections:
- Pre-deployment checklist
- Local testing
- Kubernetes preparation
- Helm deployment
- Post-deployment validation
- Configuration verification
- Scaling and HA
- Backup and disaster recovery
- Monitoring and alerting
- Security review
- Go-live execution

**Best for:** Production deployment planning and execution

---

#### INCIDENT_MANAGEMENT_README.md
**Main project README (400+ lines)**

Sections:
- Executive summary
- Quick start
- Documentation index
- Architecture overview
- API endpoints summary
- Configuration options
- Testing instructions
- Deployment options
- Security features
- Performance characteristics
- Features and capabilities
- Integration points
- Production considerations
- Troubleshooting
- Support resources
- Next steps

**Best for:** Project overview, navigation

---

## How to Use This Guide

### For Development
1. Start with **INCIDENT_API_QUICKSTART.md**
2. Review **Models** section in this guide
3. Study **Service** and **Handlers** implementation
4. Read test files for examples

### For Deployment
1. Follow **DEPLOYMENT_CHECKLIST.md** step by step
2. Reference **values.yaml** for Helm configuration
3. Use **INCIDENT_API.md** for API validation
4. Check **Dockerfile** for build customization

### For Integration
1. Read **INCIDENT_API.md** for endpoint details
2. Reference **AI Client** section for provider details
3. Check **Configuration** section for setup
4. Review curl/Python examples in API docs

### For Troubleshooting
1. Check **INCIDENT_API_QUICKSTART.md** troubleshooting section
2. Review **Configuration** in this guide
3. Check environment variables in **config.go**
4. Review test files for expected behavior

---

## File Statistics

| File | Type | Lines | Purpose |
|------|------|-------|---------|
| incident.go (models) | Go | 150 | Data structures |
| client.go | Go | 120 | AI interface |
| openai.go | Go | 200 | OpenAI impl |
| anthropic.go | Go | 200 | Anthropic impl |
| parsing.go | Go | 100 | JSON parsing |
| incident.go (service) | Go | 300 | Business logic |
| incident_test.go (service) | Go | 400 | Unit tests |
| incident.go (handlers) | Go | 250 | HTTP handlers |
| incident_test.go (handlers) | Go | 350 | Handler tests |
| config.go | Go | 150 | Configuration |
| main.go | Go | 50 | Integration |
| **Total Go** | | **2,470** | |
| Dockerfile | Docker | 50 | Build |
| values.yaml | Yaml | 200 | Helm values |
| ai-secret.yaml | Yaml | 20 | Helm secret |
| deployment.yaml | Yaml | 30 | Helm deployment |
| **Total Config** | | **300** | |
| INCIDENT_API.md | Markdown | 550 | API reference |
| INCIDENT_API_QUICKSTART.md | Markdown | 300 | Quick start |
| IMPLEMENTATION_SUMMARY.md | Markdown | 400 | Details |
| DEPLOYMENT_CHECKLIST.md | Markdown | 350 | Deployment |
| INCIDENT_MANAGEMENT_README.md | Markdown | 400 | Overview |
| **Total Docs** | | **2,000** | |
| **GRAND TOTAL** | | **4,770** | |

---

## Index by Use Case

### "I want to create an incident"
→ INCIDENT_API.md (POST /incidents section)

### "I want to analyze incident with AI"
→ INCIDENT_API.md (Analyze Incident section)

### "I want to deploy to Kubernetes"
→ DEPLOYMENT_CHECKLIST.md

### "I want to understand the code"
→ IMPLEMENTATION_SUMMARY.md

### "I want to run tests"
→ INCIDENT_API_QUICKSTART.md (Testing section)

### "I want to configure AI provider"
→ INCIDENT_API_QUICKSTART.md (Configuration section)

### "I want to troubleshoot"
→ INCIDENT_API_QUICKSTART.md (Troubleshooting section)

### "I want integration examples"
→ INCIDENT_API.md (Usage Examples section)

---

## Key Concepts Quick Reference

| Term | Definition | File |
|------|-----------|------|
| **Incident** | Core entity tracking issue lifecycle | models/incident.go |
| **Severity** | Classification level (critical-low) | models/incident.go |
| **Status** | Lifecycle state (open-closed) | models/incident.go |
| **AI Provider** | Service providing analysis (OpenAI/Anthropic) | ai/client.go |
| **Analysis** | AI-generated findings and recommendations | models/incident.go |
| **RCA** | Root Cause Analysis document | models/incident.go |
| **Service** | Business logic and CRUD operations | service/incident.go |
| **Handler** | HTTP request/response processing | handlers/incident.go |

---

## Common Commands

```bash
# Development
go run ./cmd/server/main.go
go test ./...
go test ./pkg/service/... -v
go test ./pkg/handlers/... -v

# Docker
docker build -f build/Dockerfile -t incident-api:latest .
docker run -e AI_PROVIDER=openai -e OPENAI_API_KEY=sk-... -p 8080:8080 incident-api

# Kubernetes
kubectl apply -f deploy/helm/app/
helm install incident-api deploy/helm/app -n incident-api
kubectl get pods -n incident-api
kubectl logs -f deployment/incident-api -n incident-api
```

---

Last Updated: December 22, 2024
