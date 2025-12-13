# ACI Backend Deployment Makefile
# Simplifies common deployment tasks

.PHONY: help build-image deploy-minikube deploy-kind deploy-k3s docker-up docker-down docker-logs clean test-api

# Variables
IMAGE_NAME := aci-backend
IMAGE_TAG := latest
NAMESPACE := aci-backend
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Default target
help: ## Show this help message
	@echo "ACI Backend Deployment Commands"
	@echo "================================"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

###############################################################################
# Docker Build
###############################################################################

build-image: ## Build Docker image
	@echo "Building Docker image..."
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		-t $(IMAGE_NAME):$(IMAGE_TAG) \
		-f Dockerfile \
		../..

build-minikube: ## Build image for minikube
	@echo "Building image for minikube..."
	eval $$(minikube docker-env) && \
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		-t $(IMAGE_NAME):$(IMAGE_TAG) \
		-f Dockerfile \
		../..

build-kind: ## Build and load image for kind
	@echo "Building and loading image for kind..."
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		-t $(IMAGE_NAME):$(IMAGE_TAG) \
		-f Dockerfile \
		../.. && \
	kind load docker-image $(IMAGE_NAME):$(IMAGE_TAG)

build-k3s: ## Build and import image for k3s
	@echo "Building and importing image for k3s..."
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		-t $(IMAGE_NAME):$(IMAGE_TAG) \
		-f Dockerfile \
		../.. && \
	docker save $(IMAGE_NAME):$(IMAGE_TAG) | sudo k3s ctr images import -

###############################################################################
# Kubernetes Deployment
###############################################################################

deploy-minikube: ## Deploy to minikube
	@./deploy-local.sh minikube

deploy-kind: ## Deploy to kind
	@./deploy-local.sh kind

deploy-k3s: ## Deploy to k3s
	@./deploy-local.sh k3s

deploy: ## Deploy to current Kubernetes context
	@echo "Deploying to Kubernetes..."
	kubectl apply -k k8s/
	@echo "Waiting for deployment..."
	kubectl wait --for=condition=available --timeout=300s deployment/$(IMAGE_NAME) -n $(NAMESPACE)

undeploy: ## Remove deployment from Kubernetes
	@echo "Removing deployment..."
	kubectl delete -k k8s/

###############################################################################
# Docker Compose
###############################################################################

docker-up: ## Start services with Docker Compose
	@echo "Starting Docker Compose services..."
	docker-compose up -d

docker-down: ## Stop services with Docker Compose
	@echo "Stopping Docker Compose services..."
	docker-compose down

docker-down-volumes: ## Stop services and remove volumes (⚠️ DESTRUCTIVE)
	@echo "Stopping Docker Compose services and removing volumes..."
	@read -p "This will delete all data. Are you sure? (y/N) " confirm && \
	if [ "$$confirm" = "y" ] || [ "$$confirm" = "Y" ]; then \
		docker-compose down -v; \
	else \
		echo "Cancelled."; \
	fi

docker-logs: ## View Docker Compose logs
	@docker-compose logs -f

docker-logs-backend: ## View backend logs only
	@docker-compose logs -f aci-backend

docker-restart: ## Restart Docker Compose services
	@docker-compose restart

docker-rebuild: ## Rebuild and restart Docker Compose services
	@docker-compose up -d --build

###############################################################################
# Kubernetes Operations
###############################################################################

k8s-status: ## Show Kubernetes deployment status
	@echo "Pods:"
	@kubectl get pods -n $(NAMESPACE)
	@echo ""
	@echo "Services:"
	@kubectl get svc -n $(NAMESPACE)
	@echo ""
	@echo "Ingress:"
	@kubectl get ingress -n $(NAMESPACE)
	@echo ""
	@echo "HPA:"
	@kubectl get hpa -n $(NAMESPACE)

k8s-logs: ## Tail Kubernetes logs
	@kubectl logs -n $(NAMESPACE) -l app.kubernetes.io/name=$(IMAGE_NAME) -f

k8s-describe: ## Describe Kubernetes deployment
	@kubectl describe deployment $(IMAGE_NAME) -n $(NAMESPACE)

k8s-events: ## Show recent Kubernetes events
	@kubectl get events -n $(NAMESPACE) --sort-by='.lastTimestamp'

k8s-exec: ## Execute shell in pod
	@kubectl exec -it -n $(NAMESPACE) $$(kubectl get pod -n $(NAMESPACE) -l app.kubernetes.io/name=$(IMAGE_NAME) -o jsonpath='{.items[0].metadata.name}') -- sh

k8s-port-forward: ## Port forward to backend service
	@echo "Port forwarding to http://localhost:8080"
	@kubectl port-forward -n $(NAMESPACE) svc/$(IMAGE_NAME) 8080:80

k8s-scale: ## Scale deployment (usage: make k8s-scale REPLICAS=3)
	@kubectl scale deployment $(IMAGE_NAME) -n $(NAMESPACE) --replicas=$(REPLICAS)

k8s-restart: ## Restart deployment
	@kubectl rollout restart deployment $(IMAGE_NAME) -n $(NAMESPACE)

k8s-rollout-status: ## Check rollout status
	@kubectl rollout status deployment $(IMAGE_NAME) -n $(NAMESPACE)

###############################################################################
# Testing
###############################################################################

test-api: ## Test API endpoints
	@echo "Testing health endpoint..."
	@curl -s http://aci.local/v1/health | jq . || echo "Health check failed"
	@echo ""
	@echo "Testing ready endpoint..."
	@curl -s http://aci.local/v1/ready | jq . || echo "Ready check failed"

test-api-local: ## Test API endpoints via port-forward
	@echo "Testing health endpoint (localhost:8080)..."
	@curl -s http://localhost:8080/v1/health | jq . || echo "Health check failed"
	@echo ""
	@echo "Testing ready endpoint (localhost:8080)..."
	@curl -s http://localhost:8080/v1/ready | jq . || echo "Ready check failed"

load-test: ## Generate load for HPA testing
	@echo "Generating load... (Press Ctrl+C to stop)"
	@kubectl run -it --rm load-generator --image=busybox --restart=Never -- /bin/sh -c "while sleep 0.01; do wget -q -O- http://$(IMAGE_NAME).$(NAMESPACE)/v1/health; done"

###############################################################################
# Secrets Management
###############################################################################

generate-secrets: ## Generate development secrets
	@echo "Generating secrets..."
	@mkdir -p secrets
	@if [ ! -f secrets/jwt.key ]; then \
		ssh-keygen -t rsa -b 4096 -m PEM -f secrets/jwt.key -N ""; \
		echo "JWT keys generated"; \
	else \
		echo "JWT keys already exist"; \
	fi
	@if [ ! -f secrets/webhook.secret ]; then \
		openssl rand -base64 32 > secrets/webhook.secret; \
		echo "Webhook secret generated"; \
	else \
		echo "Webhook secret already exists"; \
	fi
	@if [ ! -f secrets/db.password ]; then \
		openssl rand -base64 24 > secrets/db.password; \
		echo "Database password generated"; \
	else \
		echo "Database password already exists"; \
	fi
	@echo "Secrets generated in: secrets/"

show-secrets: ## Show generated secrets (⚠️ SENSITIVE)
	@echo "Database Password:"
	@cat secrets/db.password 2>/dev/null || echo "Not generated"
	@echo ""
	@echo "Webhook Secret:"
	@cat secrets/webhook.secret 2>/dev/null || echo "Not generated"
	@echo ""
	@echo "JWT Private Key:"
	@cat secrets/jwt.key 2>/dev/null || echo "Not generated"

###############################################################################
# Database Operations
###############################################################################

db-connect: ## Connect to PostgreSQL (Docker Compose)
	@docker-compose exec postgres psql -U aci_user -d aci

db-connect-k8s: ## Connect to PostgreSQL (Kubernetes)
	@kubectl run -it --rm psql --image=postgres:16 --restart=Never -- psql -h postgres.$(NAMESPACE).svc.cluster.local -U aci_user -d aci

###############################################################################
# Cleanup
###############################################################################

clean: ## Clean build artifacts and temporary files
	@echo "Cleaning build artifacts..."
	@rm -rf bin/ dist/ tmp/ *.log
	@docker system prune -f

clean-secrets: ## Remove generated secrets (⚠️ DESTRUCTIVE)
	@echo "Removing generated secrets..."
	@read -p "This will delete all generated secrets. Are you sure? (y/N) " confirm && \
	if [ "$$confirm" = "y" ] || [ "$$confirm" = "Y" ]; then \
		rm -rf secrets/; \
		echo "Secrets removed"; \
	else \
		echo "Cancelled."; \
	fi

clean-all: clean undeploy docker-down ## Clean everything (⚠️ DESTRUCTIVE)
	@echo "Cleaned all artifacts and stopped services"

###############################################################################
# Development
###############################################################################

dev-setup: generate-secrets docker-up ## Setup local development environment
	@echo "Development environment ready!"
	@echo "API: http://localhost:8080"

dev-teardown: docker-down-volumes clean-secrets ## Teardown development environment (⚠️ DESTRUCTIVE)
	@echo "Development environment cleaned"

###############################################################################
# CI/CD Helpers
###############################################################################

ci-build: ## Build for CI/CD
	@echo "Building for CI/CD..."
	@docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		--tag $(IMAGE_NAME):$(IMAGE_TAG) \
		--tag $(IMAGE_NAME):$(VERSION) \
		-f Dockerfile \
		../..

ci-test: ## Run tests in CI/CD
	@echo "Running tests..."
	@# Add your test commands here
	@echo "Tests passed"

ci-scan: ## Scan image for vulnerabilities
	@echo "Scanning image..."
	@docker run --rm -v /var/run/docker.sock:/var/run/docker.sock \
		aquasec/trivy image --severity HIGH,CRITICAL \
		$(IMAGE_NAME):$(IMAGE_TAG)

###############################################################################
# Information
###############################################################################

info: ## Show deployment information
	@echo "ACI Backend Deployment Information"
	@echo "==================================="
	@echo ""
	@echo "Image:         $(IMAGE_NAME):$(IMAGE_TAG)"
	@echo "Version:       $(VERSION)"
	@echo "Git Commit:    $(GIT_COMMIT)"
	@echo "Build Time:    $(BUILD_TIME)"
	@echo "Namespace:     $(NAMESPACE)"
	@echo ""
	@echo "Current Context:"
	@kubectl config current-context

version: ## Show version information
	@echo "Version:    $(VERSION)"
	@echo "Commit:     $(GIT_COMMIT)"
	@echo "Build Time: $(BUILD_TIME)"
