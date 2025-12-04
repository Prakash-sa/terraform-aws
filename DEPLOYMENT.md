# Deployment Guide

This guide walks you through deploying the complete production infrastructure and application.

## Prerequisites Checklist

- [ ] AWS Account with admin access
- [ ] AWS CLI installed and configured (`aws configure`)
- [ ] Terraform >= 1.5.0 installed
- [ ] kubectl installed
- [ ] Helm 3.x installed
- [ ] Docker installed
- [ ] Go 1.21+ installed
- [ ] Make installed

## Step-by-Step Deployment

### 1. Clone and Setup

```bash
git clone https://github.com/Prakash-sa/terraform-aws.git
cd terraform-aws

# Set up development dependencies
make dev-setup
```

### 2. Configure AWS

```bash
# Verify AWS credentials
aws sts get-caller-identity

# Set your AWS account ID
export AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
export AWS_REGION=us-east-1
```

### 3. Deploy Infrastructure with Terraform

```bash
cd infra/terraform

# Initialize Terraform
terraform init

# Review what will be created
terraform plan

# Deploy infrastructure (this takes 15-20 minutes)
terraform apply

# Save outputs
terraform output > ../../terraform-outputs.txt
```

**What gets created:**
- VPC with public/private subnets across 3 AZs
- EKS cluster with managed node groups
- ECR repository for Docker images
- S3 bucket for artifacts
- CodeBuild project
- CodePipeline
- IAM roles and policies
- Security groups

### 4. Configure kubectl

```bash
# Update kubeconfig
aws eks update-kubeconfig --region us-east-1 --name go-api-eks-cluster

# Verify connection
kubectl get nodes
kubectl get namespaces
```

### 5. Build and Push Docker Image

```bash
# Return to project root
cd ../..

# Build Docker image
make docker-build

# Log in to ECR
aws ecr get-login-password --region us-east-1 | \
  docker login --username AWS --password-stdin \
  ${AWS_ACCOUNT_ID}.dkr.ecr.us-east-1.amazonaws.com

# Tag and push image
docker tag go-api-app:latest \
  ${AWS_ACCOUNT_ID}.dkr.ecr.us-east-1.amazonaws.com/go-api-app:latest

docker push ${AWS_ACCOUNT_ID}.dkr.ecr.us-east-1.amazonaws.com/go-api-app:latest
```

### 6. Update Helm Values

Edit `deploy/helm/app/values.yaml`:

```yaml
image:
  repository: YOUR_ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/go-api-app
  tag: "latest"

ingress:
  hosts:
    - host: api.yourdomain.com  # Update with your domain
```

### 7. Deploy Application with Helm

```bash
# Install the Helm chart
helm upgrade --install go-api-app ./deploy/helm/app \
  --namespace production \
  --create-namespace \
  --set image.repository=${AWS_ACCOUNT_ID}.dkr.ecr.us-east-1.amazonaws.com/go-api-app \
  --set image.tag=latest \
  --timeout 10m \
  --wait

# Verify deployment
kubectl get pods -n production
kubectl get svc -n production
kubectl get ingress -n production
```

### 8. Verify Application

```bash
# Check pod status
kubectl get pods -n production

# Check logs
kubectl logs -l app.kubernetes.io/name=app -n production

# Port forward to test locally
kubectl port-forward -n production svc/go-api-app 8080:80

# Test endpoints (in another terminal)
curl http://localhost:8080/health
curl http://localhost:8080/ready
curl http://localhost:8080/metrics
```

### 9. Set Up CI/CD (GitHub Actions)

#### Add GitHub Secrets

Go to your repository settings → Secrets → Actions:

1. `AWS_ACCESS_KEY_ID` - Your AWS access key
2. `AWS_SECRET_ACCESS_KEY` - Your AWS secret key

#### Create IAM User for CI/CD

```bash
# Create IAM user
aws iam create-user --user-name github-actions-user

# Attach policies
aws iam attach-user-policy \
  --user-name github-actions-user \
  --policy-arn arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryPowerUser

aws iam attach-user-policy \
  --user-name github-actions-user \
  --policy-arn arn:aws:iam::aws:policy/AmazonEKSClusterPolicy

# Create access key
aws iam create-access-key --user-name github-actions-user
```

Save the Access Key ID and Secret Access Key as GitHub secrets.

### 10. Test CI/CD Pipeline

```bash
# Make a change and push
git add .
git commit -m "feat: trigger CI/CD pipeline"
git push origin main
```

Go to GitHub Actions tab to see the pipeline running.

## Post-Deployment Configuration

### Install NGINX Ingress Controller

```bash
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo update

helm install nginx-ingress ingress-nginx/ingress-nginx \
  --namespace ingress-nginx \
  --create-namespace \
  --set controller.service.type=LoadBalancer
```

### Install cert-manager (for TLS)

```bash
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml

# Create Let's Encrypt issuer
cat <<EOF | kubectl apply -f -
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: your-email@example.com
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
    - http01:
        ingress:
          class: nginx
EOF
```

### Set Up Monitoring (Optional)

```bash
# Install Prometheus Stack
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update

helm install prometheus prometheus-community/kube-prometheus-stack \
  --namespace monitoring \
  --create-namespace

# Access Grafana
kubectl port-forward -n monitoring svc/prometheus-grafana 3000:80
# Default credentials: admin/prom-operator
```

## Validation Checklist

- [ ] All pods are running: `kubectl get pods -n production`
- [ ] Service has external IP: `kubectl get svc -n production`
- [ ] Health check passes: `curl http://<EXTERNAL-IP>/health`
- [ ] Metrics endpoint works: `curl http://<EXTERNAL-IP>/metrics`
- [ ] HPA is configured: `kubectl get hpa -n production`
- [ ] Ingress is configured: `kubectl get ingress -n production`
- [ ] CI/CD pipeline runs successfully
- [ ] Logs are accessible: `kubectl logs -f <pod-name> -n production`

## Troubleshooting

### Pods not starting

```bash
kubectl describe pod <pod-name> -n production
kubectl logs <pod-name> -n production
kubectl get events -n production --sort-by='.lastTimestamp'
```

### Image pull errors

```bash
# Verify ECR repository exists
aws ecr describe-repositories --repository-names go-api-app

# Check if image exists
aws ecr list-images --repository-name go-api-app

# Verify pod can pull images
kubectl get serviceaccount -n production
```

### Ingress not working

```bash
# Check ingress controller
kubectl get pods -n ingress-nginx

# Check ingress resource
kubectl describe ingress -n production

# Check service
kubectl get svc -n production
```

## Cleanup

To tear down everything:

```bash
# Delete Helm release
helm uninstall go-api-app -n production

# Destroy infrastructure
cd infra/terraform
terraform destroy

# Clean local artifacts
make clean
```

## Cost Optimization Tips

1. **Use Spot Instances**: Update node group in EKS module
2. **Reduce Node Count**: Set min/max nodes in terraform variables
3. **Enable Cluster Autoscaler**: Scale down during off-hours
4. **Use Fargate**: For smaller workloads
5. **Clean Up Unused Resources**: Remove old ECR images

## Next Steps

1. Configure custom domain
2. Set up SSL/TLS certificates
3. Configure monitoring and alerting
4. Set up log aggregation
5. Implement backup strategy
6. Configure autoscaling policies
7. Set up disaster recovery
8. Implement security scanning
9. Configure WAF rules
10. Set up cost alerts

## Support

For issues:
- Check logs: `kubectl logs -f <pod-name> -n production`
- Check events: `kubectl get events -n production`
- Review Terraform state: `terraform show`
- Check AWS console for resource status
