# Incident Management API - Deployment Checklist

## Pre-Deployment

- [ ] **API Keys Obtained**
  - [ ] OpenAI API key (https://platform.openai.com/api-keys)
  - [ ] OR Anthropic API key (https://console.anthropic.com/keys)
  - [ ] Keys are valid and have appropriate permissions

- [ ] **Code Review**
  - [ ] All tests pass: `go test ./...`
  - [ ] Code builds successfully: `go build ./cmd/server/main.go`
  - [ ] No security issues in dependencies: `go mod verify`
  - [ ] Linting passes: `golangci-lint run` (if configured)

- [ ] **Environment Preparation**
  - [ ] Kubernetes cluster available
  - [ ] Helm 3+ installed
  - [ ] kubectl configured to correct context
  - [ ] Docker registry credentials configured (if using private registry)

- [ ] **Documentation Review**
  - [ ] Reviewed INCIDENT_API.md
  - [ ] Reviewed INCIDENT_API_QUICKSTART.md
  - [ ] Reviewed IMPLEMENTATION_SUMMARY.md
  - [ ] Team trained on API usage

## Local Testing

- [ ] **Run Locally**
  - [ ] Set environment variables for AI provider
  - [ ] Start server: `go run ./cmd/server/main.go`
  - [ ] Test health endpoint: `curl http://localhost:8080/health`
  - [ ] Create test incident: `curl -X POST http://localhost:8080/api/v1/incidents -d '...'`
  - [ ] Verify AI analysis works (if AI configured)

- [ ] **Integration Testing**
  - [ ] Test all CRUD operations
  - [ ] Test incident filtering
  - [ ] Test analysis endpoint
  - [ ] Test RCA generation
  - [ ] Test log summarization
  - [ ] Verify error handling

- [ ] **Docker Testing**
  - [ ] Build Docker image: `docker build -f build/Dockerfile -t incident-api:test .`
  - [ ] Run container: `docker run -p 8080:8080 -e AI_PROVIDER=openai -e OPENAI_API_KEY=... incident-api:test`
  - [ ] Test endpoints via HTTP

## Kubernetes Preparation

- [ ] **Create Namespace**
  ```bash
  kubectl create namespace incident-api
  ```

- [ ] **Create Secrets**
  ```bash
  kubectl create secret generic ai-secrets \
    --from-literal=OPENAI_API_KEY=sk-your-key \
    -n incident-api
  ```
  OR
  ```bash
  kubectl create secret generic ai-secrets \
    --from-literal=ANTHROPIC_API_KEY=sk-ant-your-key \
    -n incident-api
  ```

- [ ] **Verify Secret Creation**
  ```bash
  kubectl get secrets -n incident-api
  kubectl describe secret ai-secrets -n incident-api
  ```

- [ ] **Configure Helm Values**
  - [ ] Update image repository in values.yaml
  - [ ] Configure AI provider selection
  - [ ] Set resource limits appropriately
  - [ ] Configure ingress if needed
  - [ ] Review all values: `helm template incident-api deploy/helm/app -f values.yaml`

## Helm Deployment

- [ ] **Dry Run**
  ```bash
  helm install incident-api deploy/helm/app \
    -n incident-api \
    --dry-run --debug
  ```

- [ ] **Install Release**
  ```bash
  helm install incident-api deploy/helm/app \
    -n incident-api \
    --values deploy/helm/app/custom-values.yaml
  ```

- [ ] **Verify Deployment**
  ```bash
  kubectl get pods -n incident-api
  kubectl get svc -n incident-api
  kubectl get ingress -n incident-api (if enabled)
  ```

- [ ] **Check Pod Status**
  ```bash
  kubectl describe pod <pod-name> -n incident-api
  kubectl logs <pod-name> -n incident-api
  ```

## Post-Deployment Validation

- [ ] **Pod Health**
  - [ ] All pods running: `kubectl get pods -n incident-api`
  - [ ] No pending or crash loop states
  - [ ] Ready status shows 1/1
  - [ ] Age shows recent startup

- [ ] **Service Connectivity**
  ```bash
  kubectl port-forward svc/incident-api 8080:80 -n incident-api
  curl http://localhost:8080/health
  curl http://localhost:8080/ready
  ```

- [ ] **API Endpoints**
  - [ ] Health check: `curl http://localhost:8080/health`
  - [ ] Readiness: `curl http://localhost:8080/ready`
  - [ ] Metrics: `curl http://localhost:8080/metrics`
  - [ ] Create incident: Test POST /api/v1/incidents
  - [ ] List incidents: Test GET /api/v1/incidents
  - [ ] Analyze: Test POST /api/v1/incidents/{id}/analyze
  - [ ] RCA: Test POST /api/v1/incidents/{id}/rca/generate
  - [ ] Logs: Test POST /api/v1/logs/summarize

- [ ] **AI Functionality (if configured)**
  - [ ] Create incident with AI auto-classification
  - [ ] Verify severity classification works
  - [ ] Run analysis and verify findings
  - [ ] Generate RCA and verify content quality
  - [ ] Summarize logs and check insights

- [ ] **Logs and Monitoring**
  ```bash
  kubectl logs -f deployment/incident-api -n incident-api
  kubectl top pods -n incident-api
  kubectl top nodes
  ```

- [ ] **Prometheus Metrics (if monitoring configured)**
  - [ ] Verify metrics endpoint responding
  - [ ] Check http_requests_total counter
  - [ ] Check http_request_duration_seconds histogram
  - [ ] Check active_connections gauge

## Configuration Verification

- [ ] **Environment Variables**
  ```bash
  kubectl exec -it <pod-name> -n incident-api -- env | grep -E "(AI_|OPENAI_|ANTHROPIC_)"
  ```

- [ ] **Secret Mounting**
  ```bash
  kubectl exec -it <pod-name> -n incident-api -- env | grep -E "API_KEY"
  ```

- [ ] **Application Configuration**
  - [ ] Verify AI provider is set correctly
  - [ ] Verify timeout settings
  - [ ] Verify model selection
  - [ ] Verify log level

## Scaling and High Availability

- [ ] **HPA Configuration**
  - [ ] Verify HPA created: `kubectl get hpa -n incident-api`
  - [ ] Check current replicas vs desired
  - [ ] Monitor scaling events: `kubectl describe hpa incident-api -n incident-api`

- [ ] **Pod Disruption Budget**
  - [ ] Verify PDB created: `kubectl get pdb -n incident-api`
  - [ ] Check disruptions allowed: `kubectl describe pdb incident-api -n incident-api`

- [ ] **Network Policy (if enabled)**
  - [ ] Verify policy applied: `kubectl get networkpolicies -n incident-api`
  - [ ] Test ingress/egress as expected

- [ ] **Pod Anti-Affinity**
  - [ ] Verify pods spread across nodes: `kubectl get pods -o wide -n incident-api`
  - [ ] Check node distribution

## Backup and Disaster Recovery

- [ ] **Helm Release Backup**
  ```bash
  helm get values incident-api -n incident-api > incident-api-values-backup.yaml
  helm get manifest incident-api -n incident-api > incident-api-manifest-backup.yaml
  ```

- [ ] **Secret Backup** (encrypted)
  ```bash
  kubectl get secret ai-secrets -n incident-api -o yaml | sops -e > ai-secrets-encrypted.yaml
  ```

- [ ] **Configuration Backup**
  - [ ] Document all environment variables used
  - [ ] Document custom values in values.yaml
  - [ ] Store backups in secure location

- [ ] **Recovery Plan**
  - [ ] Document recovery procedure
  - [ ] Test recovery in test environment
  - [ ] Create runbook for team

## Monitoring and Alerting

- [ ] **Prometheus Rules (if applicable)**
  - [ ] Pod crashes/restarts
  - [ ] High error rate
  - [ ] API latency > threshold
  - [ ] AI API failures

- [ ] **Log Aggregation (if applicable)**
  - [ ] Logs collected to ELK/Loki
  - [ ] Error logs monitored
  - [ ] Search configured for troubleshooting

- [ ] **Alerting**
  - [ ] PagerDuty/Slack alerts configured
  - [ ] Alert recipients identified
  - [ ] Escalation policy established

## Security Review

- [ ] **Secret Management**
  - [ ] Secrets not in version control
  - [ ] Secrets stored in encrypted etcd
  - [ ] Secret rotation policy documented
  - [ ] Access limited to service account

- [ ] **RBAC**
  - [ ] Service account created
  - [ ] Role/ClusterRole assigned appropriately
  - [ ] Pod security policy applied
  - [ ] Network policy restricts access

- [ ] **Container Security**
  - [ ] Running as non-root user
  - [ ] Read-only root filesystem
  - [ ] No privileged containers
  - [ ] Security scanning passed (if enabled)

- [ ] **API Security**
  - [ ] HTTPS enabled (via ingress)
  - [ ] Rate limiting configured
  - [ ] Authentication implemented (if required)
  - [ ] CORS policies set appropriately

## Rollback Plan

- [ ] **Rollback Procedure Documented**
  ```bash
  helm rollback incident-api -n incident-api
  ```

- [ ] **Rollback Testing**
  - [ ] Tested rollback in staging
  - [ ] Verified previous version restores
  - [ ] Confirmed data integrity

- [ ] **Communication**
  - [ ] Rollback decision makers identified
  - [ ] Communication channels established
  - [ ] Stakeholders notified of changes

## Post-Deployment Handoff

- [ ] **Documentation**
  - [ ] Deployment runbook finalized
  - [ ] Troubleshooting guide completed
  - [ ] Architecture diagram updated
  - [ ] API examples documented

- [ ] **Training**
  - [ ] On-call team trained
  - [ ] Support team understands API
  - [ ] Monitoring team configured
  - [ ] Development team trained

- [ ] **Support**
  - [ ] Support channel established
  - [ ] Escalation path defined
  - [ ] Known issues documented
  - [ ] FAQ created

## Go-Live

- [ ] **Final Verification**
  - [ ] All checklist items completed
  - [ ] No outstanding issues
  - [ ] Team ready for deployment
  - [ ] Stakeholders approved

- [ ] **Deployment Execution**
  - [ ] Execute deployment during planned window
  - [ ] Monitor closely for first hour
  - [ ] Have rollback ready
  - [ ] Document any issues encountered

- [ ] **Post Go-Live**
  - [ ] Monitor metrics closely (24 hours)
  - [ ] Respond to any issues quickly
  - [ ] Gather feedback from users
  - [ ] Document lessons learned

## Success Criteria

- [ ] All pods running and healthy
- [ ] All API endpoints responding correctly
- [ ] AI integration functioning (if configured)
- [ ] No error spikes in logs
- [ ] Metrics showing normal behavior
- [ ] Users able to create and analyze incidents
- [ ] No security concerns
- [ ] Team confident with operation
