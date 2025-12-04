.PHONY: help build test lint docker-build docker-run terraform-init terraform-plan terraform-apply helm-install deploy clean

# Variables
APP_NAME := go-api-app
DOCKER_IMAGE := $(APP_NAME):latest
AWS_REGION := us-east-1
AWS_ACCOUNT_ID := $(shell aws sts get-caller-identity --query Account --output text)
ECR_REPO := $(AWS_ACCOUNT_ID).dkr.ecr.$(AWS_REGION).amazonaws.com/$(APP_NAME)
CLUSTER_NAME := go-api-eks-cluster

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Application targets
build: ## Build Go application
	@echo "Building Go application..."
	cd app && go build -o ../bin/server ./cmd/server/main.go

test: ## Run tests
	@echo "Running tests..."
	cd app && go test -v -race -coverprofile=../coverage.out ./...
	cd app && go tool cover -func=../coverage.out

lint: ## Run linter
	@echo "Running linter..."
	cd app && golangci-lint run

run: ## Run application locally
	@echo "Running application..."
	./bin/server

# Docker targets
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -f build/Dockerfile -t $(DOCKER_IMAGE) .

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	docker run -p 8080:8080 --rm $(DOCKER_IMAGE)

docker-push: ## Push Docker image to ECR
	@echo "Logging into ECR..."
	aws ecr get-login-password --region $(AWS_REGION) | docker login --username AWS --password-stdin $(ECR_REPO)
	@echo "Tagging image..."
	docker tag $(DOCKER_IMAGE) $(ECR_REPO):latest
	@echo "Pushing image..."
	docker push $(ECR_REPO):latest

# Terraform targets
terraform-init: ## Initialize Terraform
	@echo "Initializing Terraform..."
	cd infra/terraform && terraform init

terraform-plan: ## Run Terraform plan
	@echo "Running Terraform plan..."
	cd infra/terraform && terraform plan

terraform-apply: ## Apply Terraform configuration
	@echo "Applying Terraform configuration..."
	cd infra/terraform && terraform apply

terraform-destroy: ## Destroy Terraform infrastructure
	@echo "Destroying Terraform infrastructure..."
	cd infra/terraform && terraform destroy

# Kubernetes targets
kubeconfig: ## Update kubeconfig for EKS
	@echo "Updating kubeconfig..."
	aws eks update-kubeconfig --region $(AWS_REGION) --name $(CLUSTER_NAME)

helm-install: ## Install/upgrade Helm chart
	@echo "Installing Helm chart..."
	helm upgrade --install $(APP_NAME) ./deploy/helm/app \
		--namespace production \
		--create-namespace \
		--set image.repository=$(ECR_REPO) \
		--set image.tag=latest

helm-uninstall: ## Uninstall Helm chart
	@echo "Uninstalling Helm chart..."
	helm uninstall $(APP_NAME) -n production

k8s-status: ## Check Kubernetes deployment status
	@echo "Checking deployment status..."
	kubectl get all -n production
	kubectl get pods -n production
	kubectl logs -l app.kubernetes.io/name=app -n production --tail=50

# Full deployment
deploy: docker-build docker-push helm-install ## Build, push, and deploy

# Development targets
dev-setup: ## Set up development environment
	@echo "Setting up development environment..."
	cd app && go mod download
	cd app && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Cleanup
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -rf app/vendor/
	rm -f coverage.out
	cd infra/terraform && rm -rf .terraform/
	cd infra/terraform && rm -f .terraform.lock.hcl

# Quick checks
check: test lint ## Run tests and linting

# Complete infrastructure setup
infra-up: terraform-init terraform-apply kubeconfig ## Set up complete infrastructure

# Complete teardown
infra-down: helm-uninstall terraform-destroy ## Tear down everything
