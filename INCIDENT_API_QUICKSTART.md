# Incident Management API - Quick Start Guide

## Prerequisites

- Go 1.21+
- Docker (optional)
- Kubernetes/Helm (for production deployment)
- OpenAI API key OR Anthropic API key

## Local Development Setup

### 1. Clone and Navigate
```bash
cd app
```

### 2. Download Dependencies
```bash
go mod download
```

### 3. Configure Environment
```bash
# Choose your AI provider (openai or anthropic)
export AI_PROVIDER=openai
export OPENAI_API_KEY=sk-your-key-here
# OR for Anthropic:
# export AI_PROVIDER=anthropic
# export ANTHROPIC_API_KEY=sk-ant-your-key-here

# Optional: Configure other settings
export AI_TIMEOUT=60
export AI_TEMPERATURE=0.7
export AI_MAX_TOKENS=2000
export LOG_LEVEL=info
export PORT=8080
```

### 4. Run Server
```bash
go run ./cmd/server/main.go
```

Server will be available at `http://localhost:8080`

### 5. Verify Health
```bash
curl http://localhost:8080/health
```

## Docker Deployment

### Build Image
```bash
docker build -f build/Dockerfile -t incident-api:latest .
```

### Run Container
```bash
docker run -d \
  -p 8080:8080 \
  -e AI_PROVIDER=openai \
  -e OPENAI_API_KEY=sk-your-key \
  incident-api:latest
```

## Kubernetes Deployment

### 1. Create Secrets
```bash
kubectl create namespace incident-api

kubectl create secret generic ai-secrets \
  --from-literal=OPENAI_API_KEY=sk-your-key \
  --from-literal=ANTHROPIC_API_KEY=sk-ant-your-key \
  -n incident-api
```

### 2. Install Helm Chart
```bash
helm install incident-api deploy/helm/app \
  -n incident-api \
  --set image.repository=your-registry/incident-api \
  --set image.tag=latest
```

### 3. Verify Deployment
```bash
kubectl get pods -n incident-api
kubectl logs -f deployment/incident-api -n incident-api
```

## Quick API Tests

### Create an Incident
```bash
curl -X POST http://localhost:8080/api/v1/incidents \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Database Connection Pool Exhausted",
    "description": "Production database showing signs of connection pool exhaustion",
    "source": "prometheus",
    "logs": [
      "ERROR: Unable to acquire connection",
      "WARN: Connection pool queue exceeded"
    ],
    "tags": ["database", "critical"]
  }'
```

Response:
```json
{
  "id": "INC-1703001234-1",
  "title": "Database Connection Pool Exhausted",
  "description": "Production database showing signs of connection pool exhaustion",
  "source": "prometheus",
  "severity": "critical",
  "status": "open",
  "logs": [...],
  "tags": ["database", "critical"],
  "created_at": "2024-01-01T10:00:00Z",
  "updated_at": "2024-01-01T10:00:00Z"
}
```

**Note**: The API automatically classified this as "critical" because the title and description contain keywords like "database" and "pool exhausted".

### Get AI Analysis
```bash
# Use the incident ID from above
INCIDENT_ID="INC-1703001234-1"

curl -X POST http://localhost:8080/api/v1/incidents/$INCIDENT_ID/analyze
```

Response includes:
```json
{
  "ai_analysis": {
    "summary": "...",
    "findings": [...],
    "root_causes": [...],
    "recommended_actions": [...],
    "severity_suggestion": "critical",
    "generated_at": "2024-01-01T10:05:00Z",
    "provider": "openai",
    "model": "gpt-4"
  }
}
```

### Generate RCA Document
```bash
curl -X POST http://localhost:8080/api/v1/incidents/$INCIDENT_ID/rca/generate
```

Response includes comprehensive RCA with timeline, root cause, impact analysis, and lessons learned.

### Summarize Logs
```bash
curl -X POST http://localhost:8080/api/v1/logs/summarize \
  -H "Content-Type: application/json" \
  -d '{
    "logs": [
      "ERROR: Unable to acquire connection",
      "WARN: Connection pool queue exceeded",
      "ERROR: Query timeout after 30s"
    ]
  }'
```

### Update Incident Status
```bash
curl -X PUT http://localhost:8080/api/v1/incidents/$INCIDENT_ID \
  -H "Content-Type: application/json" \
  -d '{
    "status": "resolved",
    "severity": "high"
  }'
```

### List All Incidents
```bash
curl http://localhost:8080/api/v1/incidents
```

### Filter Incidents
```bash
# Get all critical open incidents
curl http://localhost:8080/api/v1/incidents?status=open&severity=critical
```

## Running Tests

### Unit Tests
```bash
go test ./pkg/service/... -v
go test ./pkg/handlers/... -v
```

### Integration Tests
```bash
go test ./... -v -run Integration
```

### Test Coverage
```bash
go test ./... -cover
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Troubleshooting

### AI Provider Not Configured
If you see: `Warning: AI provider not configured, using no-op client`

**Solution**: Set AI provider environment variables
```bash
export OPENAI_API_KEY=sk-your-actual-key
# OR
export ANTHROPIC_API_KEY=sk-ant-your-actual-key
```

### Invalid API Key
If analysis returns errors about authentication:

**Solution**: Verify your API key is correct
```bash
# For OpenAI
export OPENAI_API_KEY=sk-... (should start with sk-)

# For Anthropic
export ANTHROPIC_API_KEY=sk-ant-... (should start with sk-ant-)
```

### Connection Refused
If you get connection errors:

**Solution**: Ensure server is running
```bash
# Check if server is listening
lsof -i :8080

# Restart if needed
go run ./cmd/server/main.go
```

### Database Connection Issues
Current implementation uses in-memory storage. For production, implement:
- PostgreSQL persistence
- MongoDB support
- Redis caching

## Configuration Reference

### Environment Variables

```bash
# AI Provider (required for analysis features)
AI_PROVIDER=openai|anthropic

# OpenAI Settings
OPENAI_API_KEY=sk-...
OPENAI_MODEL=gpt-4 (default)

# Anthropic Settings
ANTHROPIC_API_KEY=sk-ant-...
ANTHROPIC_MODEL=claude-3-5-sonnet-20241022 (default)

# Common AI Settings
AI_TIMEOUT=60 (seconds)
AI_TEMPERATURE=0.7 (0.0-1.0)
AI_MAX_TOKENS=2000

# Server Settings
PORT=8080
ENVIRONMENT=production|development
LOG_LEVEL=debug|info|warn|error
APP_VERSION=1.0.0
```

## Next Steps

1. **Integrate with your monitoring system**: Send incidents from Prometheus, Datadog, etc.
2. **Set up webhooks**: For incident notifications to Slack/Teams
3. **Configure persistence**: Move to PostgreSQL/MongoDB for production
4. **Add authentication**: Implement JWT or OAuth2
5. **Set rate limiting**: Protect API from abuse
6. **Custom prompts**: Tailor AI analysis to your organization

## Support

For issues or questions:
1. Check logs: `docker logs incident-api` or `kubectl logs -f deployment/incident-api`
2. Review API documentation: See [INCIDENT_API.md](./INCIDENT_API.md)
3. Check environment configuration: `env | grep -E "(AI_|OPENAI_|ANTHROPIC_)"`

## References

- [Incident API Documentation](./INCIDENT_API.md)
- [OpenAI API Documentation](https://platform.openai.com/docs)
- [Anthropic API Documentation](https://docs.anthropic.com)
- [Go Module Documentation](https://go.dev/doc/modules)
