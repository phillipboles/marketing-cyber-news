#!/bin/bash

###############################################################################
# ACI Backend Local Deployment Script
#
# Automated deployment script for local Kubernetes clusters
# Supports: minikube, kind, k3s
#
# Usage:
#   ./deploy-local.sh [minikube|kind|k3s]
#
# Example:
#   ./deploy-local.sh minikube
###############################################################################

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
NAMESPACE="aci-backend"
IMAGE_NAME="aci-backend"
IMAGE_TAG="latest"
CLUSTER_TYPE="${1:-minikube}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."

    # Check kubectl
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl is not installed. Please install kubectl first."
        exit 1
    fi

    # Check Docker
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed. Please install Docker first."
        exit 1
    fi

    # Check cluster-specific tools
    case "${CLUSTER_TYPE}" in
        minikube)
            if ! command -v minikube &> /dev/null; then
                log_error "minikube is not installed. Please install minikube first."
                exit 1
            fi
            ;;
        kind)
            if ! command -v kind &> /dev/null; then
                log_error "kind is not installed. Please install kind first."
                exit 1
            fi
            ;;
        k3s)
            if ! command -v k3s &> /dev/null; then
                log_error "k3s is not installed. Please install k3s first."
                exit 1
            fi
            ;;
        *)
            log_error "Unsupported cluster type: ${CLUSTER_TYPE}"
            log_info "Supported types: minikube, kind, k3s"
            exit 1
            ;;
    esac

    log_success "All prerequisites met"
}

# Ensure cluster is running
ensure_cluster() {
    log_info "Ensuring ${CLUSTER_TYPE} cluster is running..."

    case "${CLUSTER_TYPE}" in
        minikube)
            if ! minikube status &> /dev/null; then
                log_warning "minikube is not running. Starting minikube..."
                minikube start --cpus=4 --memory=8192 --driver=docker
                minikube addons enable ingress
                minikube addons enable metrics-server
            else
                log_success "minikube is already running"
            fi
            ;;
        kind)
            if ! kind get clusters | grep -q "kind"; then
                log_warning "kind cluster not found. Creating cluster..."
                kind create cluster
            else
                log_success "kind cluster is running"
            fi
            ;;
        k3s)
            if ! systemctl is-active --quiet k3s; then
                log_error "k3s is not running. Please start k3s first: sudo systemctl start k3s"
                exit 1
            fi
            log_success "k3s is running"
            ;;
    esac
}

# Build Docker image
build_image() {
    log_info "Building Docker image..."

    local build_args=(
        "--build-arg" "VERSION=dev"
        "--build-arg" "BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
        "--build-arg" "GIT_COMMIT=$(git -C "${PROJECT_ROOT}" rev-parse --short HEAD 2>/dev/null || echo 'unknown')"
        "-t" "${IMAGE_NAME}:${IMAGE_TAG}"
        "-f" "${SCRIPT_DIR}/Dockerfile"
        "${PROJECT_ROOT}"
    )

    case "${CLUSTER_TYPE}" in
        minikube)
            log_info "Using minikube's Docker daemon..."
            eval "$(minikube docker-env)"
            docker build "${build_args[@]}"
            ;;
        kind)
            docker build "${build_args[@]}"
            log_info "Loading image into kind cluster..."
            kind load docker-image "${IMAGE_NAME}:${IMAGE_TAG}"
            ;;
        k3s)
            docker build "${build_args[@]}"
            log_info "Importing image into k3s..."
            docker save "${IMAGE_NAME}:${IMAGE_TAG}" | sudo k3s ctr images import -
            ;;
    esac

    log_success "Image built successfully"
}

# Generate secrets if not exist
generate_secrets() {
    log_info "Checking secrets..."

    local secrets_dir="${SCRIPT_DIR}/secrets"
    mkdir -p "${secrets_dir}"

    # Generate JWT keys if not exist
    if [[ ! -f "${secrets_dir}/jwt.key" ]]; then
        log_warning "JWT keys not found. Generating..."
        ssh-keygen -t rsa -b 4096 -m PEM -f "${secrets_dir}/jwt.key" -N "" -q
        log_success "JWT keys generated"
    fi

    # Generate webhook secret if not exist
    if [[ ! -f "${secrets_dir}/webhook.secret" ]]; then
        log_warning "Webhook secret not found. Generating..."
        openssl rand -base64 32 > "${secrets_dir}/webhook.secret"
        log_success "Webhook secret generated"
    fi

    # Generate database password if not exist
    if [[ ! -f "${secrets_dir}/db.password" ]]; then
        log_warning "Database password not found. Generating..."
        openssl rand -base64 24 > "${secrets_dir}/db.password"
        log_success "Database password generated"
    fi

    log_info "Secrets are ready at: ${secrets_dir}"
    log_warning "IMPORTANT: Update k8s/secret.yaml with actual secret values before production deployment"
}

# Deploy to Kubernetes
deploy_to_k8s() {
    log_info "Deploying to Kubernetes..."

    # Create namespace if not exists
    if ! kubectl get namespace "${NAMESPACE}" &> /dev/null; then
        log_info "Creating namespace ${NAMESPACE}..."
        kubectl create namespace "${NAMESPACE}"
    fi

    # Apply manifests using kustomize
    log_info "Applying Kubernetes manifests..."
    kubectl apply -k "${SCRIPT_DIR}/k8s/"

    log_success "Manifests applied"
}

# Wait for deployment to be ready
wait_for_deployment() {
    log_info "Waiting for deployment to be ready..."

    if kubectl wait --for=condition=available --timeout=300s \
        deployment/aci-backend -n "${NAMESPACE}" &> /dev/null; then
        log_success "Deployment is ready"
    else
        log_error "Deployment failed to become ready"
        log_info "Checking pod status..."
        kubectl get pods -n "${NAMESPACE}"
        log_info "Recent events:"
        kubectl get events -n "${NAMESPACE}" --sort-by='.lastTimestamp' | tail -10
        exit 1
    fi
}

# Configure /etc/hosts
configure_hosts() {
    log_info "Configuring /etc/hosts for aci.local..."

    local cluster_ip

    case "${CLUSTER_TYPE}" in
        minikube)
            cluster_ip=$(minikube ip)
            ;;
        kind|k3s)
            cluster_ip="127.0.0.1"
            ;;
    esac

    # Check if entry already exists
    if grep -q "aci.local" /etc/hosts; then
        log_warning "aci.local entry already exists in /etc/hosts"
        log_info "Current entry:"
        grep "aci.local" /etc/hosts
        read -p "Do you want to update it? (y/N) " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            sudo sed -i.bak '/aci.local/d' /etc/hosts
            echo "${cluster_ip} aci.local" | sudo tee -a /etc/hosts > /dev/null
            log_success "Updated /etc/hosts"
        fi
    else
        echo "${cluster_ip} aci.local" | sudo tee -a /etc/hosts > /dev/null
        log_success "Added aci.local to /etc/hosts"
    fi
}

# Display deployment info
display_info() {
    log_success "Deployment completed successfully!"
    echo ""
    echo "====================================================================="
    echo "  ACI Backend Deployment Info"
    echo "====================================================================="
    echo ""
    echo "Namespace:     ${NAMESPACE}"
    echo "Image:         ${IMAGE_NAME}:${IMAGE_TAG}"
    echo "Cluster Type:  ${CLUSTER_TYPE}"
    echo ""
    echo "---------------------------------------------------------------------"
    echo "  Quick Commands"
    echo "---------------------------------------------------------------------"
    echo ""
    echo "Check pods:"
    echo "  kubectl get pods -n ${NAMESPACE}"
    echo ""
    echo "View logs:"
    echo "  kubectl logs -n ${NAMESPACE} -l app.kubernetes.io/name=${IMAGE_NAME} -f"
    echo ""
    echo "Check services:"
    echo "  kubectl get svc -n ${NAMESPACE}"
    echo ""
    echo "Check ingress:"
    echo "  kubectl get ingress -n ${NAMESPACE}"
    echo ""
    echo "Test API:"
    echo "  curl http://aci.local/v1/health"
    echo ""
    echo "Port forward (alternative to ingress):"
    echo "  kubectl port-forward -n ${NAMESPACE} svc/aci-backend 8080:80"
    echo "  curl http://localhost:8080/v1/health"
    echo ""
    echo "Scale deployment:"
    echo "  kubectl scale deployment aci-backend -n ${NAMESPACE} --replicas=3"
    echo ""
    echo "Delete deployment:"
    echo "  kubectl delete -k ${SCRIPT_DIR}/k8s/"
    echo ""
    echo "====================================================================="
    echo ""

    # Test API
    log_info "Testing API endpoint..."
    sleep 5  # Give ingress a moment to update

    if curl -s -o /dev/null -w "%{http_code}" http://aci.local/v1/health | grep -q "200"; then
        log_success "API is responding! Try: curl http://aci.local/v1/health"
    else
        log_warning "API is not responding yet. It may take a few moments for ingress to update."
        log_info "Try: kubectl port-forward -n ${NAMESPACE} svc/aci-backend 8080:80"
        log_info "Then: curl http://localhost:8080/v1/health"
    fi
}

# Main deployment flow
main() {
    echo ""
    echo "====================================================================="
    echo "  ACI Backend Local Deployment"
    echo "====================================================================="
    echo ""
    echo "Cluster Type: ${CLUSTER_TYPE}"
    echo "Namespace:    ${NAMESPACE}"
    echo "Image:        ${IMAGE_NAME}:${IMAGE_TAG}"
    echo ""

    check_prerequisites
    ensure_cluster
    build_image
    generate_secrets
    deploy_to_k8s
    wait_for_deployment
    configure_hosts
    display_info
}

# Run main function
main
