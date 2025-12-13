# ACI Backend Deployment Guide

Production-ready Kubernetes deployment manifests for the ACI Backend service with support for local development using minikube, k3s, or kind.

## Table of Contents

- [Quick Start](#quick-start)
- [Prerequisites](#prerequisites)
- [Local Development (Docker Compose)](#local-development-docker-compose)
- [Kubernetes Deployment](#kubernetes-deployment)
- [Configuration](#configuration)
- [Security](#security)
- [Monitoring](#monitoring)
- [Troubleshooting](#troubleshooting)

## Quick Start

### Option 1: Docker Compose (Fastest)

```bash
# Start all services
docker-compose -f aci-backend/deployments/docker-compose.yml up -d

# Check health
curl http://localhost:8080/v1/health

# View logs
docker-compose -f aci-backend/deployments/docker-compose.yml logs -f
```

### Option 2: Kubernetes (minikube)

```bash
# Build and deploy
./aci-backend/deployments/deploy-local.sh minikube

# Check status
kubectl get pods -n aci-backend

# Test API
curl http://aci.local/v1/health
```

## Prerequisites

### For Docker Compose

- Docker Engine 20.10+
- Docker Compose 2.0+

### For Kubernetes

Choose one:

#### minikube
```bash
# Install minikube
brew install minikube  # macOS
# or download from https://minikube.sigs.k8s.io/

# Start cluster
minikube start --cpus=4 --memory=8192 --driver=docker

# Enable addons
minikube addons enable ingress
minikube addons enable metrics-server
```

#### k3s
```bash
# Install k3s (Linux only)
curl -sfL https://get.k3s.io | sh -

# Get kubeconfig
export KUBECONFIG=/etc/rancher/k3s/k3s.yaml
```

#### kind
```bash
# Install kind
brew install kind  # macOS
# or download from https://kind.sigs.k8s.io/

# Create cluster with ingress support
cat <<EOF | kind create cluster --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "ingress-ready=true"
  extraPortMappings:
  - containerPort: 80
    hostPort: 80
  - containerPort: 443
    hostPort: 443
EOF

# Install NGINX Ingress
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml
```

## Local Development (Docker Compose)

### First-Time Setup

1. **Generate secrets**:

```bash
cd aci-backend/deployments

# Generate JWT keys
ssh-keygen -t rsa -b 4096 -m PEM -f jwt.key -N ""

# Generate webhook secret
openssl rand -base64 32 > webhook.secret

# Generate database password
openssl rand -base64 24 > db.password
```

2. **Create .env file**:

```bash
cat > .env <<EOF
ANTHROPIC_API_KEY=sk-ant-your-key-here
N8N_WEBHOOK_SECRET=$(cat webhook.secret)
JWT_PRIVATE_KEY=$(cat jwt.key)
JWT_PUBLIC_KEY=$(cat jwt.key.pub)
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT=$(git rev-parse --short HEAD)
EOF
```

3. **Start services**:

```bash
docker-compose --env-file .env up -d
```

### Common Operations

```bash
# View all logs
docker-compose logs -f

# View backend logs only
docker-compose logs -f aci-backend

# Restart backend
docker-compose restart aci-backend

# Rebuild after code changes
docker-compose up -d --build aci-backend

# Execute commands
docker-compose exec aci-backend sh
docker-compose exec postgres psql -U aci_user -d aci
docker-compose exec redis redis-cli

# Stop all services
docker-compose down

# Stop and remove data (⚠️ DESTRUCTIVE)
docker-compose down -v
```

## Kubernetes Deployment

### Method 1: Using Deploy Script (Recommended)

```bash
# Deploy to minikube
./aci-backend/deployments/deploy-local.sh minikube

# Deploy to kind
./aci-backend/deployments/deploy-local.sh kind

# Deploy to k3s
./aci-backend/deployments/deploy-local.sh k3s
```

### Method 2: Manual Deployment

#### Step 1: Build Docker Image

**For minikube**:
```bash
# Use minikube's Docker daemon
eval $(minikube docker-env)

# Build image
docker build -t aci-backend:latest \
  --build-arg VERSION=1.0.0 \
  --build-arg BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ") \
  --build-arg GIT_COMMIT=$(git rev-parse --short HEAD) \
  -f aci-backend/deployments/Dockerfile .
```

**For kind**:
```bash
# Build image
docker build -t aci-backend:latest \
  --build-arg VERSION=1.0.0 \
  --build-arg BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ") \
  --build-arg GIT_COMMIT=$(git rev-parse --short HEAD) \
  -f aci-backend/deployments/Dockerfile .

# Load into kind cluster
kind load docker-image aci-backend:latest
```

**For k3s**:
```bash
# Build and save image
docker build -t aci-backend:latest \
  --build-arg VERSION=1.0.0 \
  --build-arg BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ") \
  --build-arg GIT_COMMIT=$(git rev-parse --short HEAD) \
  -f aci-backend/deployments/Dockerfile .

# Import into k3s
docker save aci-backend:latest | sudo k3s ctr images import -
```

#### Step 2: Update Secrets

```bash
# Generate secrets
cd aci-backend/deployments/k8s
ssh-keygen -t rsa -b 4096 -m PEM -f jwt.key -N ""

# Edit secret.yaml with actual values
# Replace REPLACE_WITH_ACTUAL_* placeholders
```

#### Step 3: Deploy with kubectl

```bash
# Apply all manifests
kubectl apply -f aci-backend/deployments/k8s/

# Or use kustomize
kubectl apply -k aci-backend/deployments/k8s/
```

#### Step 4: Configure /etc/hosts

```bash
# For minikube
echo "$(minikube ip) aci.local" | sudo tee -a /etc/hosts

# For kind/k3s
echo "127.0.0.1 aci.local" | sudo tee -a /etc/hosts
```

#### Step 5: Verify Deployment

```bash
# Check pods
kubectl get pods -n aci-backend

# Check services
kubectl get svc -n aci-backend

# Check ingress
kubectl get ingress -n aci-backend

# View logs
kubectl logs -n aci-backend -l app.kubernetes.io/name=aci-backend -f

# Test API
curl http://aci.local/v1/health
```

## Configuration

### ConfigMap Parameters

Edit `k8s/configmap.yaml` to customize:

| Parameter | Default | Description |
|-----------|---------|-------------|
| SERVER_PORT | 8080 | HTTP server port |
| LOG_LEVEL | info | Logging level (debug/info/warn/error) |
| JWT_ACCESS_TOKEN_EXPIRY | 15m | Access token lifetime |
| JWT_REFRESH_TOKEN_EXPIRY | 168h | Refresh token lifetime (7 days) |
| DATABASE_HOST | postgres.aci-backend.svc.cluster.local | PostgreSQL hostname |
| DATABASE_MAX_CONNECTIONS | 25 | Max DB connections |
| REDIS_HOST | redis.aci-backend.svc.cluster.local | Redis hostname |
| RATE_LIMIT_REQUESTS_PER_MINUTE | 60 | Rate limit per IP |

### Secret Parameters

Edit `k8s/secret.yaml` with actual values:

| Parameter | Required | Description |
|-----------|----------|-------------|
| DATABASE_PASSWORD | Yes | PostgreSQL password |
| JWT_PRIVATE_KEY | Yes | RSA private key for JWT signing |
| JWT_PUBLIC_KEY | Yes | RSA public key for JWT verification |
| N8N_WEBHOOK_SECRET | Yes | Shared secret for n8n webhooks |
| ANTHROPIC_API_KEY | Yes | Anthropic API key for Claude |

### Resource Limits

Edit `k8s/deployment.yaml` to adjust resources:

```yaml
resources:
  requests:
    cpu: 250m      # Guaranteed CPU
    memory: 128Mi  # Guaranteed memory
  limits:
    cpu: 500m      # Maximum CPU
    memory: 256Mi  # Maximum memory
```

### Autoscaling

Edit `k8s/hpa.yaml` to adjust scaling:

```yaml
minReplicas: 2          # Minimum pods
maxReplicas: 10         # Maximum pods
targetCPUUtilization: 70%     # Scale at 70% CPU
targetMemoryUtilization: 80%  # Scale at 80% memory
```

## Security

### 1. Secret Management

**Development**: Use Kubernetes Secrets with base64 encoding

**Production**: Use one of these options:

#### Sealed Secrets
```bash
# Install sealed-secrets controller
kubectl apply -f https://github.com/bitnami-labs/sealed-secrets/releases/download/v0.24.0/controller.yaml

# Seal secrets
kubeseal -f secret.yaml -w sealed-secret.yaml

# Deploy sealed secret
kubectl apply -f sealed-secret.yaml
```

#### External Secrets Operator
```bash
# Install operator
helm repo add external-secrets https://charts.external-secrets.io
helm install external-secrets external-secrets/external-secrets -n external-secrets-system --create-namespace

# Use with AWS Secrets Manager, GCP Secret Manager, Azure Key Vault, or HashiCorp Vault
```

### 2. Network Policies

Create `k8s/networkpolicy.yaml`:

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: aci-backend-netpol
  namespace: aci-backend
spec:
  podSelector:
    matchLabels:
      app.kubernetes.io/name: aci-backend
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: ingress-nginx
    ports:
    - protocol: TCP
      port: 8080
  egress:
  - to:
    - podSelector:
        matchLabels:
          app: postgres
    ports:
    - protocol: TCP
      port: 5432
  - to:
    - podSelector:
        matchLabels:
          app: redis
    ports:
    - protocol: TCP
      port: 6379
  - to:
    - namespaceSelector: {}
    ports:
    - protocol: TCP
      port: 53
    - protocol: UDP
      port: 53
```

### 3. RBAC

Create service account with minimal permissions:

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: aci-backend
  namespace: aci-backend
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: aci-backend
  namespace: aci-backend
rules:
- apiGroups: [""]
  resources: ["configmaps", "secrets"]
  verbs: ["get", "list"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: aci-backend
  namespace: aci-backend
subjects:
- kind: ServiceAccount
  name: aci-backend
  namespace: aci-backend
roleRef:
  kind: Role
  name: aci-backend
  apiGroup: rbac.authorization.k8s.io
```

### 4. Image Scanning

```bash
# Scan with Trivy
trivy image aci-backend:latest

# Scan with Grype
grype aci-backend:latest

# Fail build on HIGH/CRITICAL vulnerabilities
trivy image --severity HIGH,CRITICAL --exit-code 1 aci-backend:latest
```

## Monitoring

### Prometheus Metrics

Add ServiceMonitor for Prometheus:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: aci-backend
  namespace: aci-backend
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: aci-backend
  endpoints:
  - port: http
    path: /metrics
    interval: 30s
```

### Logging

View logs:

```bash
# All pods
kubectl logs -n aci-backend -l app.kubernetes.io/name=aci-backend -f

# Specific pod
kubectl logs -n aci-backend <pod-name> -f

# Previous container (after crash)
kubectl logs -n aci-backend <pod-name> -p

# With timestamps
kubectl logs -n aci-backend <pod-name> -f --timestamps

# Last 100 lines
kubectl logs -n aci-backend <pod-name> --tail=100
```

### Health Checks

```bash
# Readiness probe
curl http://aci.local/v1/ready

# Liveness probe
curl http://aci.local/v1/health

# Check from inside cluster
kubectl run -it --rm debug --image=curlimages/curl --restart=Never -- curl http://aci-backend.aci-backend/v1/health
```

## Troubleshooting

### Pod Not Starting

```bash
# Check pod status
kubectl get pods -n aci-backend

# Describe pod
kubectl describe pod -n aci-backend <pod-name>

# Check events
kubectl get events -n aci-backend --sort-by='.lastTimestamp'

# Check logs
kubectl logs -n aci-backend <pod-name>
```

### Image Pull Errors

```bash
# For minikube: Rebuild in minikube's Docker
eval $(minikube docker-env)
docker build -t aci-backend:latest -f aci-backend/deployments/Dockerfile .

# For kind: Reload image
docker build -t aci-backend:latest -f aci-backend/deployments/Dockerfile .
kind load docker-image aci-backend:latest

# For k3s: Re-import image
docker build -t aci-backend:latest -f aci-backend/deployments/Dockerfile .
docker save aci-backend:latest | sudo k3s ctr images import -
```

### Database Connection Issues

```bash
# Check if PostgreSQL is running
kubectl get pods -n aci-backend -l app=postgres

# Test database connection
kubectl run -it --rm psql --image=postgres:16 --restart=Never -- psql -h postgres.aci-backend.svc.cluster.local -U aci_user -d aci

# Check DNS resolution
kubectl run -it --rm dns-debug --image=busybox --restart=Never -- nslookup postgres.aci-backend.svc.cluster.local
```

### Ingress Not Working

```bash
# Check ingress controller
kubectl get pods -n ingress-nginx  # For NGINX
kubectl get pods -n kube-system -l app=traefik  # For Traefik (k3s)

# Check ingress
kubectl describe ingress -n aci-backend aci-backend

# Check /etc/hosts
cat /etc/hosts | grep aci.local

# For minikube: Get IP
minikube ip

# Test without ingress
kubectl port-forward -n aci-backend svc/aci-backend 8080:80
curl http://localhost:8080/v1/health
```

### HPA Not Scaling

```bash
# Check metrics server
kubectl top nodes
kubectl top pods -n aci-backend

# Check HPA status
kubectl get hpa -n aci-backend
kubectl describe hpa -n aci-backend aci-backend

# Generate load for testing
kubectl run -it --rm load-generator --image=busybox --restart=Never -- /bin/sh -c "while sleep 0.01; do wget -q -O- http://aci-backend.aci-backend/v1/health; done"
```

### Secret Decoding

```bash
# View secret
kubectl get secret -n aci-backend aci-backend-secrets -o yaml

# Decode specific value
kubectl get secret -n aci-backend aci-backend-secrets -o jsonpath='{.data.DATABASE_PASSWORD}' | base64 -d
```

## Cleanup

```bash
# Delete all resources
kubectl delete -k aci-backend/deployments/k8s/

# Or
kubectl delete namespace aci-backend

# Remove from /etc/hosts
sudo sed -i '' '/aci.local/d' /etc/hosts

# Stop minikube
minikube stop

# Delete minikube cluster
minikube delete

# Delete kind cluster
kind delete cluster
```

## Production Checklist

Before deploying to production:

- [ ] Replace all placeholder secrets with actual values
- [ ] Use external secret management (Sealed Secrets, External Secrets Operator, Vault)
- [ ] Configure TLS/SSL with cert-manager or manual certificates
- [ ] Set up proper logging aggregation (ELK, Loki)
- [ ] Configure monitoring with Prometheus and Grafana
- [ ] Set up alerting for critical metrics
- [ ] Implement NetworkPolicies for pod-to-pod security
- [ ] Configure RBAC with minimal permissions
- [ ] Scan images for vulnerabilities in CI/CD
- [ ] Set appropriate resource limits based on load testing
- [ ] Configure PodDisruptionBudget for high availability
- [ ] Set up database backups (Velero, cloud-native solutions)
- [ ] Test disaster recovery procedures
- [ ] Document rollback procedures
- [ ] Set up GitOps workflow (Argo CD, FluxCD)
- [ ] Configure ingress with rate limiting and WAF rules
- [ ] Enable audit logging
- [ ] Implement circuit breakers and retry logic in application
- [ ] Load test with production-like data
- [ ] Create runbooks for common incidents

## Support

For issues or questions:
1. Check logs: `kubectl logs -n aci-backend -l app.kubernetes.io/name=aci-backend`
2. Check events: `kubectl get events -n aci-backend`
3. Review this README's troubleshooting section
4. Open an issue in the project repository
