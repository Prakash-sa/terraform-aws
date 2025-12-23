# Incident Management API Documentation

## Overview

The Incident Management API provides automated incident triage, analysis, and Root Cause Analysis (RCA) documentation using AI providers (OpenAI GPT-4 or Anthropic Claude). The API enables organizations to reduce manual incident response overhead through intelligent analysis of incidents and log data.

## Features

- **Automated Severity Classification**: AI-powered incident severity assessment
- **Incident Analysis**: Generate findings, root causes, and recommended actions
- **RCA Generation**: Comprehensive Root Cause Analysis documents
- **Log Summarization**: Extract key insights and alerts from logs
- **Multi-Provider Support**: OpenAI and Anthropic Claude with seamless switching
- **Graceful Degradation**: Service continues with basic functionality if AI is unavailable
- **Thread-Safe Operations**: Concurrent incident management with in-memory storage
- **Comprehensive Lifecycle Management**: Track incidents from creation through resolution

## Architecture

### Data Models

#### Incident
The core data model for tracking incidents with structured metadata:

```json
{
  "id": "INC-1703001234-1",
  "title": "Database Connection Pool Exhaustion",
  "description": "Critical database connection pool exhaustion on prod-main",
  "source": "prometheus",
  "severity": "critical",
  "status": "open",
  "logs": ["ERROR: Connection pool exhausted", "WARN: Requests queuing"],
  "tags": ["database", "critical", "production"],
  "metadata": {
    "service": "api-gateway",
    "region": "us-east-1",
    "affected_users": "10000"
  },
  "created_at": "2024-01-01T10:00:00Z",
  "updated_at": "2024-01-01T10:15:00Z",
  "resolved_at": null,
  "assigned_to": "on-call-engineer@company.com",
  "ai_analysis": { /* AI-generated analysis */ },
  "rca_document": { /* AI-generated RCA */ }
}
```

#### Severity Levels
- `critical`: Service down, data loss, or security breach
- `high`: Service degraded or functional issue
- `medium`: Performance issue or non-critical error
- `low`: Information or minor issue
- `unknown`: Not yet classified

#### Incident Status
- `open`: Recently created incident
- `in_progress`: Actively being investigated
- `resolved`: Root cause identified and fixed
- `closed`: Post-resolution review complete

### AI Integration

#### Supported Providers
- **OpenAI**: GPT-4 (default model)
- **Anthropic**: Claude 3.5 Sonnet

#### Response Processing
- Handles markdown-wrapped JSON responses
- Graceful fallback for parsing failures
- 60-second timeout on AI API calls

## REST API Endpoints

### Incidents

#### Create Incident
```
POST /api/v1/incidents
```

Creates a new incident with optional AI severity classification.

**Request Body:**
```json
{
  "title": "CPU spike on prod-1",
  "description": "CPU at 95%, memory pressure detected",
  "source": "prometheus",
  "severity": "high",
  "logs": ["ERROR: OOM warning", "WARN: GC pressure"],
  "tags": ["performance", "production"],
  "assigned_to": "engineer@company.com"
}
```

**Response:** `201 Created`
```json
{
  "id": "INC-1703001234-1",
  "title": "CPU spike on prod-1",
  "description": "CPU at 95%, memory pressure detected",
  "source": "prometheus",
  "severity": "high",
  "status": "open",
  "logs": ["ERROR: OOM warning", "WARN: GC pressure"],
  "tags": ["performance", "production"],
  "created_at": "2024-01-01T10:00:00Z",
  "updated_at": "2024-01-01T10:00:00Z"
}
```

**Notes:**
- If `severity` is omitted, AI automatically classifies it
- `source` field helps track incident origin (prometheus, manual, logs, etc.)

#### Get Incident
```
GET /api/v1/incidents/{id}
```

Retrieves a specific incident with all analysis and RCA data.

**Response:** `200 OK`
```json
{
  "id": "INC-1703001234-1",
  "title": "CPU spike on prod-1",
  "description": "CPU at 95%, memory pressure detected",
  "source": "prometheus",
  "severity": "high",
  "status": "open",
  "logs": ["ERROR: OOM warning", "WARN: GC pressure"],
  "tags": ["performance", "production"],
  "created_at": "2024-01-01T10:00:00Z",
  "updated_at": "2024-01-01T10:00:00Z",
  "ai_analysis": {
    "summary": "Memory pressure on production instance causing CPU spikes",
    "findings": [
      "JVM heap usage exceeding 85%",
      "Garbage collection pauses increasing"
    ],
    "root_causes": [
      "Memory leak in cache module",
      "Inadequate heap size configuration"
    ],
    "recommended_actions": [
      "Increase JVM heap size to 2GB",
      "Deploy cache memory leak fix",
      "Enable memory profiling"
    ],
    "severity_suggestion": "high",
    "generated_at": "2024-01-01T10:05:00Z",
    "model": "gpt-4",
    "provider": "openai"
  }
}
```

#### List Incidents
```
GET /api/v1/incidents?status=open&severity=critical
```

Lists all incidents with optional filtering.

**Query Parameters:**
- `status` (optional): Filter by status (open, in_progress, resolved, closed)
- `severity` (optional): Filter by severity (critical, high, medium, low, unknown)

**Response:** `200 OK`
```json
[
  {
    "id": "INC-1703001234-1",
    "title": "CPU spike on prod-1",
    "severity": "critical",
    "status": "open",
    "created_at": "2024-01-01T10:00:00Z",
    ...
  }
]
```

#### Update Incident
```
PUT /api/v1/incidents/{id}
```

Updates incident details and status.

**Request Body:**
```json
{
  "title": "Updated Title",
  "description": "Updated Description",
  "severity": "critical",
  "status": "in_progress",
  "assigned_to": "different-engineer@company.com",
  "tags": ["performance", "production", "urgent"]
}
```

**Response:** `200 OK`

**Notes:**
- All fields are optional
- When `status` changes to `resolved`, `resolved_at` is automatically set

#### Delete Incident
```
DELETE /api/v1/incidents/{id}
```

Deletes an incident and all associated data.

**Response:** `204 No Content`

### Analysis & RCA

#### Analyze Incident
```
POST /api/v1/incidents/{id}/analyze
```

Generates AI-powered analysis for an incident including findings, root causes, and recommended actions.

**Request Body:** (empty)

**Response:** `200 OK`
```json
{
  "id": "INC-1703001234-1",
  "title": "CPU spike on prod-1",
  "ai_analysis": {
    "summary": "Memory pressure on production instance causing CPU spikes",
    "findings": [
      "JVM heap usage exceeding 85%",
      "Garbage collection pauses increasing"
    ],
    "root_causes": [
      "Memory leak in cache module",
      "Inadequate heap size configuration"
    ],
    "recommended_actions": [
      "Increase JVM heap size to 2GB",
      "Deploy cache memory leak fix",
      "Enable memory profiling"
    ],
    "severity_suggestion": "high",
    "generated_at": "2024-01-01T10:05:00Z",
    "model": "gpt-4",
    "provider": "openai"
  },
  "updated_at": "2024-01-01T10:05:00Z"
}
```

**Notes:**
- Requires valid incident ID
- Returns incident even if analysis fails, with error in response
- AI analysis is cached in the incident

#### Generate RCA Document
```
POST /api/v1/incidents/{id}/rca/generate
```

Generates comprehensive Root Cause Analysis document.

**Request Body:** (empty)

**Response:** `200 OK`
```json
{
  "id": "INC-1703001234-1",
  "title": "CPU spike on prod-1",
  "rca_document": {
    "timeline": "2024-01-01 10:00 - Incident detected via alerting. 10:05 - Engineering team notified. 10:15 - Root cause identified as memory leak. 10:30 - Hotfix deployed.",
    "root_cause": "Memory leak in cache eviction logic causing unbounded growth of in-memory object cache",
    "impact": "Production service degradation for 30 minutes affecting 10,000 users. P99 latency increased from 100ms to 2s.",
    "immediate_resolution": "Deployment of hotfix with corrected cache eviction logic and manual cache flush",
    "preventive_measures": [
      "Implement memory leak detection in CI/CD pipeline",
      "Add cache size monitoring with alerts at 70% threshold",
      "Implement automatic cache eviction on memory pressure",
      "Add unit tests for cache boundary conditions"
    ],
    "lessons_learned": [
      "Need better memory profiling in staging environment",
      "Should have caught memory leak in code review",
      "Consider cache middleware library with built-in safeguards"
    ],
    "generated_at": "2024-01-01T10:45:00Z",
    "model": "gpt-4",
    "provider": "openai"
  },
  "updated_at": "2024-01-01T10:45:00Z"
}
```

### Log Analysis

#### Summarize Logs
```
POST /api/v1/logs/summarize
```

Extracts key insights and alerts from log collections.

**Request Body:**
```json
{
  "logs": [
    "2024-01-01T10:00:00Z ERROR: Connection pool exhausted",
    "2024-01-01T10:01:00Z WARN: Requests queuing detected",
    "2024-01-01T10:02:00Z ERROR: Database connection timeout",
    "2024-01-01T10:03:00Z ERROR: Query execution failed after 30s"
  ],
  "context": {
    "service": "api-gateway",
    "environment": "production"
  }
}
```

**Response:** `200 OK`
```json
{
  "summary": "Database connection pool exhaustion caused cascading timeouts affecting API responses",
  "key_insights": [
    "Connection pool completely exhausted around 10:00",
    "Timeout threshold reached for 100% of queries",
    "Recovery would require either pool increase or traffic reduction",
    "No signs of database-side issues - connection limit is the bottleneck"
  ],
  "alerts": [
    "Critical: All database queries timing out",
    "Warning: Connection pool growth exceeded projections",
    "Action: Increase pool size or implement circuit breaker"
  ],
  "generated_at": "2024-01-01T10:05:00Z"
}
```

## Configuration

### Environment Variables

#### AI Provider Configuration
```bash
# AI Provider selection
AI_PROVIDER=openai              # or "anthropic"

# OpenAI Configuration
OPENAI_API_KEY=sk-xxx...        # Required if using OpenAI
OPENAI_MODEL=gpt-4              # Default model

# Anthropic Configuration
ANTHROPIC_API_KEY=sk-ant-xxx... # Required if using Anthropic
ANTHROPIC_MODEL=claude-3-5-sonnet-20241022  # Default model

# Common AI Settings
AI_TIMEOUT=60                   # Seconds
AI_TEMPERATURE=0.7              # 0.0-1.0, controls randomness
AI_MAX_TOKENS=2000              # Maximum response length
```

#### Server Configuration
```bash
PORT=8080
ENVIRONMENT=production
LOG_LEVEL=info
APP_VERSION=1.0.0
```

### Kubernetes Secrets

Create secrets for API keys:

```bash
# Create OpenAI secret
kubectl create secret generic ai-secrets \
  --from-literal=OPENAI_API_KEY=sk-xxx... \
  --from-literal=ANTHROPIC_API_KEY=sk-ant-xxx...

# Or using base64
kubectl apply -f - <<EOF
apiVersion: v1
kind: Secret
metadata:
  name: ai-secrets
type: Opaque
data:
  OPENAI_API_KEY: $(echo -n 'sk-xxx...' | base64)
  ANTHROPIC_API_KEY: $(echo -n 'sk-ant-xxx...' | base64)
EOF
```

### Helm Values

In `values.yaml`:

```yaml
aiSecret:
  create: true
  name: "ai-secrets"
  data:
    OPENAI_API_KEY: "base64-encoded-key"
    ANTHROPIC_API_KEY: "base64-encoded-key"

env:
  - name: AI_PROVIDER
    value: "openai"
  - name: OPENAI_MODEL
    value: "gpt-4"
  - name: AI_TIMEOUT
    value: "60"
```

## Usage Examples

### cURL Examples

#### Create Incident with Auto-Classification
```bash
curl -X POST http://localhost:8080/api/v1/incidents \
  -H "Content-Type: application/json" \
  -d '{
    "title": "CPU spike on prod-1",
    "description": "CPU at 95%, memory pressure detected",
    "source": "prometheus",
    "logs": ["ERROR: OOM warning", "WARN: GC pressure"]
  }'
```

#### Create Incident with Explicit Severity
```bash
curl -X POST http://localhost:8080/api/v1/incidents \
  -H "Content-Type: application/json" \
  -d '{
    "title": "API latency increase",
    "description": "P99 latency increased to 2000ms",
    "source": "datadog",
    "severity": "high",
    "tags": ["performance"],
    "assigned_to": "on-call@company.com"
  }'
```

#### Get AI Analysis
```bash
curl -X POST http://localhost:8080/api/v1/incidents/INC-1703001234-1/analyze
```

#### Generate RCA Document
```bash
curl -X POST http://localhost:8080/api/v1/incidents/INC-1703001234-1/rca/generate
```

#### Summarize Logs
```bash
curl -X POST http://localhost:8080/api/v1/logs/summarize \
  -H "Content-Type: application/json" \
  -d '{
    "logs": [
      "ERROR: Connection pool exhausted",
      "WARN: Requests queuing",
      "ERROR: Database timeout"
    ]
  }'
```

#### List Critical Open Incidents
```bash
curl http://localhost:8080/api/v1/incidents?status=open&severity=critical
```

#### Update Incident Status
```bash
curl -X PUT http://localhost:8080/api/v1/incidents/INC-1703001234-1 \
  -H "Content-Type: application/json" \
  -d '{
    "status": "resolved",
    "severity": "high"
  }'
```

### Python Integration

```python
import requests
import json

BASE_URL = "http://localhost:8080/api/v1"

# Create incident
response = requests.post(
    f"{BASE_URL}/incidents",
    json={
        "title": "Database Connection Failure",
        "description": "Connection pool exhausted on prod-db-1",
        "source": "prometheus",
        "logs": ["ERROR: Connection timeout", "WARN: Queue backlog"],
        "tags": ["database", "critical"]
    }
)
incident = response.json()
incident_id = incident["id"]

# Get AI analysis
response = requests.post(f"{BASE_URL}/incidents/{incident_id}/analyze")
analysis = response.json()
print(f"Analysis: {analysis['ai_analysis']['summary']}")

# Generate RCA
response = requests.post(f"{BASE_URL}/incidents/{incident_id}/rca/generate")
rca = response.json()
print(f"Root Cause: {rca['rca_document']['root_cause']}")

# Summarize logs
response = requests.post(
    f"{BASE_URL}/logs/summarize",
    json={"logs": ["ERROR log 1", "ERROR log 2"]}
)
summary = response.json()
print(f"Summary: {summary['summary']}")
```

## Error Handling

### Standard Error Response
```json
{
  "error": "Descriptive error message",
  "timestamp": "2024-01-01T10:00:00Z"
}
```

### HTTP Status Codes
- `200 OK`: Successful request
- `201 Created`: Incident successfully created
- `204 No Content`: Deletion successful
- `400 Bad Request`: Invalid request payload
- `404 Not Found`: Resource not found
- `500 Internal Server Error`: Server error

### Graceful Degradation
If AI provider is not configured:
- API continues to function for incident CRUD operations
- Analysis and RCA generation return a no-op response
- Service includes warning messages in logs

## Performance Considerations

- **AI Timeout**: 60 seconds (configurable)
- **Temperature**: 0.7 (balanced creativity/determinism)
- **Max Tokens**: 2000 (configurable)
- **Concurrent Requests**: Thread-safe with sync.RWMutex
- **Storage**: In-memory (recommended for production: PostgreSQL/MongoDB)

## Security Best Practices

1. **API Keys**: Store in Kubernetes secrets, never in code
2. **HTTPS**: Always use HTTPS in production
3. **Authentication**: Implement JWT or OAuth2 (future enhancement)
4. **Rate Limiting**: Configure per endpoint (future enhancement)
5. **Audit Logging**: All incident modifications are logged

## Deployment

### Docker
```bash
docker build -f build/Dockerfile -t incident-api:latest .
docker run -e AI_PROVIDER=openai \
  -e OPENAI_API_KEY=sk-xxx \
  -p 8080:8080 \
  incident-api:latest
```

### Kubernetes with Helm
```bash
helm install incident-api deploy/helm/app \
  --set aiSecret.data.OPENAI_API_KEY="base64-key" \
  --set env[0].value="gpt-4"
```

## Future Enhancements

- PostgreSQL/MongoDB backend for persistent storage
- Webhook integration for incident notifications
- Slack/Teams message support
- RBAC and API authentication
- Advanced filtering and search
- Metrics and analytics dashboard
- Batch processing for bulk log analysis
- Custom AI prompts per organization
- Integration with incident tracking systems (Jira, ServiceNow)
