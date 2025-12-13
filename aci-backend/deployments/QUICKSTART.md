# ACI Backend - Quick Start Guide

Get the ACI Backend running in under 5 minutes.

## Option 1: Docker Compose (Fastest - No Kubernetes Required)

### Step 1: Start Services
```bash
cd /Users/phillipboles/Development/n8n-cyber-news/aci-backend/deployments
docker-compose up -d
```

### Step 2: Verify Services
```bash
# Check containers
docker-compose ps

# View logs
docker-compose logs -f aci-backend

# Test API
curl http://localhost:8080/v1/health
```

### Step 3: Done!
Your API is running at: `http://localhost:8080`

**Useful Commands:**
```bash
# Stop services
docker-compose down

# View logs
docker-compose logs -f

# Restart backend
docker-compose restart aci-backend

# Connect to database
docker-compose exec postgres psql -U aci_user -d aci

# Connect to Redis
docker-compose exec redis redis-cli
```

---

## Option 2: Kubernetes (Recommended for Production-Like Testing)

### Prerequisites
Choose and install ONE of these:
- **minikube** (easiest): `brew install minikube`
- **kind**: `brew install kind`
- **k3s**: Linux only

### Step 1: Start Cluster (minikube example)
```bash
# Start minikube
minikube start --cpus=4 --memory=8192

# Enable required addons
minikube addons enable ingress
minikube addons enable metrics-server
```

### Step 2: Deploy
```bash
cd /Users/phillipboles/Development/n8n-cyber-news/aci-backend/deployments

# Automated deployment
./deploy-local.sh minikube

# The script will:
# - Build Docker image
# - Deploy to Kubernetes
# - Configure /etc/hosts
# - Test API
```

### Step 3: Verify
```bash
# Check pods
kubectl get pods -n aci-backend

# View logs
kubectl logs -n aci-backend -l app.kubernetes.io/name=aci-backend -f

# Test API
curl http://aci.local/v1/health
```

### Step 4: Done!
Your API is running at: `http://aci.local`

**Useful Commands:**
```bash
# Port forward (alternative access)
kubectl port-forward -n aci-backend svc/aci-backend 8080:80
curl http://localhost:8080/v1/health

# Scale deployment
kubectl scale deployment aci-backend -n aci-backend --replicas=3

# View all resources
kubectl get all -n aci-backend

# Delete deployment
kubectl delete -k k8s/
```

---

## Option 3: Using Makefile (Most Convenient)

The Makefile provides shortcuts for all common operations.

### View Available Commands
```bash
cd /Users/phillipboles/Development/n8n-cyber-news/aci-backend/deployments
make help
```

### Quick Start with Docker Compose
```bash
# Setup and start
make dev-setup

# View logs
make docker-logs

# Test API
make test-api-local

# Cleanup
make dev-teardown
```

### Quick Start with Kubernetes
```bash
# Deploy to minikube
make deploy-minikube

# Check status
make k8s-status

# View logs
make k8s-logs

# Test API
make test-api

# Cleanup
make undeploy
```

---

## Troubleshooting

### Docker Compose Issues

**Containers won't start:**
```bash
# Check Docker daemon
docker ps

# Check logs for errors
docker-compose logs

# Rebuild from scratch
docker-compose down -v
docker-compose up -d --build
```

**Port already in use:**
```bash
# Stop conflicting services
lsof -ti:8080 | xargs kill -9

# Or change port in docker-compose.yml
# Change "8080:8080" to "8081:8080"
```

### Kubernetes Issues

**Pods won't start:**
```bash
# Check pod status
kubectl get pods -n aci-backend

# Describe pod for details
kubectl describe pod -n aci-backend <pod-name>

# Check events
kubectl get events -n aci-backend --sort-by='.lastTimestamp'
```

**Image not found:**
```bash
# For minikube: Rebuild in minikube's Docker
eval $(minikube docker-env)
docker build -t aci-backend:latest -f Dockerfile ../..

# For kind: Reload image
kind load docker-image aci-backend:latest

# For k3s: Re-import image
docker save aci-backend:latest | sudo k3s ctr images import -
```

**Can't access via aci.local:**
```bash
# Check /etc/hosts
cat /etc/hosts | grep aci.local

# Re-add entry
echo "$(minikube ip) aci.local" | sudo tee -a /etc/hosts

# Or use port-forward instead
kubectl port-forward -n aci-backend svc/aci-backend 8080:80
curl http://localhost:8080/v1/health
```

**Ingress not working:**
```bash
# For minikube: Enable ingress addon
minikube addons enable ingress

# For kind: Install NGINX Ingress
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml

# Check ingress controller
kubectl get pods -n ingress-nginx
```

---

## Next Steps

Once your service is running:

1. **Update Secrets**: Edit `k8s/secret.yaml` with actual values before production
2. **Configure TLS**: Uncomment TLS section in `k8s/ingress.yaml`
3. **Set Up Monitoring**: Add Prometheus ServiceMonitor
4. **Configure Backups**: Set up Velero for disaster recovery
5. **Read Full Documentation**: See `README.md` for comprehensive guide

---

## Common API Endpoints

```bash
# Health check
curl http://localhost:8080/v1/health

# Ready check
curl http://localhost:8080/v1/ready

# API version (example)
curl http://localhost:8080/v1/version
```

---

## Environment Variables

Create a `.env` file for local development:

```bash
# Create .env file
cat > .env <<EOF
ANTHROPIC_API_KEY=sk-ant-your-key-here
N8N_WEBHOOK_SECRET=$(openssl rand -base64 32)
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT=$(git rev-parse --short HEAD)
EOF

# Use with Docker Compose
docker-compose --env-file .env up -d
```

---

## Getting Help

1. Check logs: `docker-compose logs` or `kubectl logs`
2. Review README.md for detailed documentation
3. Check Troubleshooting section above
4. Verify prerequisites are installed correctly

---

## Clean Up

### Docker Compose
```bash
# Stop services
docker-compose down

# Stop and remove data
docker-compose down -v
```

### Kubernetes
```bash
# Delete all resources
kubectl delete -k k8s/

# Delete namespace
kubectl delete namespace aci-backend

# Stop cluster (minikube)
minikube stop
```

### Complete Cleanup
```bash
# Using Makefile
make clean-all
```
