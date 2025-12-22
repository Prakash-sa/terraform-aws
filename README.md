# Production-Ready DevOps Project

Full-stack production-ready DevOps project with Go API, Kubernetes deployment on AWS EKS, complete CI/CD pipeline, and infrastructure as code.

## 🚀 Features

### Go API Application
- ✅ Production-ready REST API with Gorilla Mux
- ✅ Prometheus metrics endpoint (`/metrics`)
- ✅ Health check endpoint (`/health`)
- ✅ Readiness probe endpoint (`/ready`)
- ✅ Structured logging with Zap
- ✅ Graceful shutdown
- ✅ Request/response middleware
- ✅ Sample API endpoints

### AI-Powered Incident Automation 🤖
- ✅ Automated incident creation and tracking
- ✅ AI-powered incident analysis using OpenAI/Claude
- ✅ Automated severity classification and triage
- ✅ Log summarization and pattern detection
- ✅ Root Cause Analysis (RCA) document generation
- ✅ Recommended actions and preventive measures
- ✅ REST API for incident management

### Docker
- ✅ Multi-stage build for minimal image size
- ✅ Scratch-based final image
- ✅ Non-root user
- ✅ Health check configuration
- ✅ Security hardening

### Kubernetes (Helm)
- ✅ Deployment with best practices
- ✅ Service (ClusterIP)
- ✅ Ingress with TLS
- ✅ Horizontal Pod Autoscaler (HPA)
- ✅ ConfigMap for configuration
- ✅ Secret management
- ✅ ServiceAccount with RBAC
- ✅ PodDisruptionBudget
- ✅ ServiceMonitor for Prometheus
- ✅ Resource limits and requests
- ✅ Security contexts
- ✅ Liveness and readiness probes

### AWS Infrastructure (Terraform)
- ✅ EKS cluster with managed node groups
- ✅ VPC with public/private subnets
- ✅ ECR for container images
- ✅ S3 for artifacts
- ✅ CodeBuild for CI
- ✅ CodePipeline for CD
- ✅ IAM roles and policies
- ✅ Security groups
- ✅ Auto-scaling configuration

### CI/CD Pipeline (GitHub Actions)
- ✅ Automated testing
- ✅ Code linting (golangci-lint)
- ✅ Security scanning (Trivy)
- ✅ Docker build and push to ECR
- ✅ Helm deployment to EKS
- ✅ Deployment verification
- ✅ Smoke tests
- ✅ Multi-environment support

## 🛠️ Quick Start

### 1. Deploy Infrastructure

```bash
cd infra/terraform

# Initialize Terraform
terraform init

# Review the plan
terraform plan

# Deploy infrastructure
terraform apply

# Configure kubectl
aws eks update-kubeconfig --region us-east-1 --name go-api-eks-cluster
```

### 2. Build and Run Locally

```bash
# Build the application
cd app
go mod download
go build -o ../bin/server ./cmd/server/main.go

# Run locally
../bin/server

# Test endpoints
curl http://localhost:8080/health
curl http://localhost:8080/metrics
```

### 3. Build Docker Image

```bash
# Build image
docker build -f build/Dockerfile -t go-api-app:latest .

# Run container
docker run -p 8080:8080 go-api-app:latest

# Test
curl http://localhost:8080/health
```

### 4. Deploy with Helm

```bash
# Update image repository in values.yaml
# Replace ACCOUNT_ID with your AWS account ID

# Install/upgrade the chart
helm upgrade --install go-api-app ./deploy/helm/app \
  --namespace production \
  --create-namespace \
  --set image.repository=<ACCOUNT_ID>.dkr.ecr.us-east-1.amazonaws.com/go-api-app \
  --set image.tag=latest

# Verify deployment
kubectl get pods -n production
kubectl get svc -n production
```

### 5. Configure GitHub Actions

Set up the following secrets in your GitHub repository:
- `AWS_ACCESS_KEY_ID`
- `AWS_SECRET_ACCESS_KEY`

Push to `main` branch to trigger the CI/CD pipeline.

## 🔍 API Endpoints

### Core Endpoints
| Endpoint | Method | Description |
|----------|--------|-------------|
| `/` | GET | Home endpoint |
| `/health` | GET | Health check |
| `/ready` | GET | Readiness check |
| `/metrics` | GET | Prometheus metrics |
| `/api/v1/data` | GET | Sample data endpoint |
| `/api/v1/echo` | POST | Echo endpoint |

### AI-Powered Incident Management Endpoints 🤖
| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/incidents` | POST | Create a new incident |
| `/api/v1/incidents` | GET | List all incidents |
| `/api/v1/incidents/{id}` | GET | Get incident details |
| `/api/v1/incidents/{id}/analyze` | POST | Analyze incident with AI |
| `/api/v1/incidents/{id}/analysis` | GET | Get AI analysis results |
| `/api/v1/incidents/{id}/rca/generate` | POST | Generate RCA document |
| `/api/v1/incidents/{id}/rca` | GET | Get RCA document |
| `/api/v1/logs/summarize` | POST | Summarize logs with AI |

### Using the Incident Automation API 🤖

**1. Create an Incident:**
```bash
curl -X POST http://localhost:8080/api/v1/incidents \
  -H "Content-Type: application/json" \
  -d '{
    "title": "High CPU Usage Alert",
    "description": "CPU usage exceeded 90% on production servers",
    "source": "prometheus",
    "alert_data": "cpu_usage{instance=\"prod-1\"} > 90",
    "logs": [
      "2024-12-22 10:15:23 ERROR: CPU at 92%",
      "2024-12-22 10:15:45 WARNING: Memory pressure detected"
    ]
  }'
```

**2. Analyze Incident with AI:**
```bash
curl -X POST http://localhost:8080/api/v1/incidents/{incident-id}/analyze
```

**3. Get AI Analysis:**
```bash
curl http://localhost:8080/api/v1/incidents/{incident-id}/analysis
```

**4. Generate RCA Document:**
```bash
curl -X POST http://localhost:8080/api/v1/incidents/{incident-id}/rca/generate
```

**5. Summarize Logs:**
```bash
curl -X POST http://localhost:8080/api/v1/logs/summarize \
  -H "Content-Type: application/json" \
  -d '{
    "logs": [
      "Error: Connection timeout",
      "Warning: Retry attempt 1",
      "Error: Max retries exceeded"
    ]
  }'
```

## 📊 Monitoring

### Prometheus Metrics
The application exposes the following metrics:
- `http_requests_total` - Total HTTP requests by method, endpoint, and status
- `http_request_duration_seconds` - HTTP request duration histogram
- `active_connections` - Number of active connections

### Access Metrics
```bash
curl http://localhost:8080/metrics
```

## 🧪 Testing

```bash
# Unit tests
cd app
go test -v ./...

# With coverage
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Linting
golangci-lint run
```

## 🔄 CI/CD Workflow

1. **Push Code** → Triggers GitHub Actions
2. **Test** → Run unit tests and linting
3. **Security Scan** → Trivy scans code and dependencies
4. **Build** → Build Docker image
5. **Push** → Push to Amazon ECR
6. **Deploy** → Helm upgrade on EKS
7. **Verify** → Run smoke tests
8. **Notify** → Send deployment status

## 📝 Configuration

### Environment Variables
- `PORT` - Server port (default: 8080)
- `ENVIRONMENT` - Environment name (dev/staging/production)
- `APP_VERSION` - Application version
- `LOG_LEVEL` - Logging level

### AI Service Configuration 🤖
- `AI_PROVIDER` - AI service provider: `openai` or `anthropic` (default: openai)
- `OPENAI_API_KEY` - OpenAI API key (required for OpenAI)
- `OPENAI_MODEL` - OpenAI model to use (default: gpt-4)
- `ANTHROPIC_API_KEY` - Anthropic API key (required for Claude)
- `ANTHROPIC_MODEL` - Anthropic model to use (default: claude-3-5-sonnet-20241022)

#### Setting up AI API Keys

**For OpenAI:**
1. Get your API key from https://platform.openai.com/api-keys
2. Base64 encode it: `echo -n 'sk-your-key' | base64`
3. Add to `deploy/helm/app/values.yaml` under `secret.data.OPENAI_API_KEY`

**For Anthropic (Claude):**
1. Get your API key from https://console.anthropic.com/
2. Base64 encode it: `echo -n 'sk-ant-your-key' | base64`
3. Add to `deploy/helm/app/values.yaml` under `secret.data.ANTHROPIC_API_KEY`

**Note:** API keys are optional. The service will work without them but AI-powered features will be disabled.

### Terraform Variables
See `infra/terraform/variables.tf` for all configurable options.

### Helm Values
See `deploy/helm/app/values.yaml` for all configurable options.

## 🐛 Troubleshooting

### Check Pod Status
```bash
kubectl get pods -n production
kubectl describe pod <pod-name> -n production
kubectl logs <pod-name> -n production
```

### Check Deployment
```bash
kubectl rollout status deployment/go-api-app -n production
kubectl get events -n production --sort-by='.lastTimestamp'
```

### Debug Service
```bash
kubectl get svc -n production
kubectl describe svc go-api-app -n production
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests
5. Submit a pull request

## 📄 License

MIT License - see LICENSE file for details

## 🔗 Additional Resources

- [AWS EKS Documentation](https://docs.aws.amazon.com/eks/)
- [Helm Documentation](https://helm.sh/docs/)
- [Terraform AWS Provider](https://registry.terraform.io/providers/hashicorp/aws/latest/docs)
- [Prometheus Documentation](https://prometheus.io/docs/)
- [Go Best Practices](https://golang.org/doc/effective_go)
