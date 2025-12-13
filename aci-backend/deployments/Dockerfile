# Multi-stage Dockerfile for ACI Backend
# Stage 1: Build the Go application
# Stage 2: Create minimal runtime image

###############################################################################
# Stage 1: Builder
###############################################################################
FROM golang:1.24-alpine AS builder

# Install build dependencies
# - git: For go mod download with private repos
# - ca-certificates: For HTTPS requests during build
# - make: For build automation (if using Makefile)
RUN apk add --no-cache git ca-certificates tzdata make

# Set working directory
WORKDIR /build

# Copy go mod files first for better caching
# Docker will cache this layer if go.mod/go.sum haven't changed
COPY go.mod go.sum ./

# Download dependencies
# This layer is cached unless go.mod/go.sum change
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application
# CGO_ENABLED=0: Build static binary without C dependencies
# GOOS=linux: Target Linux OS
# GOARCH=amd64: Target AMD64 architecture
# -ldflags: Linker flags
#   -s: Omit symbol table (reduces binary size)
#   -w: Omit DWARF debugging info (reduces binary size)
#   -X: Set version information from build args
# -trimpath: Remove file system paths from binary
# -o: Output file path
ARG VERSION=dev
ARG BUILD_TIME
ARG GIT_COMMIT

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -trimpath \
    -o /build/aci-backend \
    ./cmd/server

# Optional: Run tests during build (uncomment to enable)
# RUN go test -v ./...

###############################################################################
# Stage 2: Runtime
###############################################################################
FROM alpine:3.19

# Install runtime dependencies
# - ca-certificates: For HTTPS requests
# - tzdata: For timezone support
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user for running the application
# Security best practice: Never run as root
RUN addgroup -g 1000 -S appuser && \
    adduser -u 1000 -S appuser -G appuser

# Create necessary directories with proper permissions
RUN mkdir -p /app/cache /tmp && \
    chown -R appuser:appuser /app /tmp

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder --chown=appuser:appuser /build/aci-backend /app/aci-backend

# Copy configuration files if needed
# COPY --from=builder --chown=appuser:appuser /build/configs /app/configs

# Switch to non-root user
USER appuser

# Expose application port
# This is documentation only; actual port is controlled by SERVER_PORT env var
EXPOSE 8080

# Health check
# Kubernetes will use readiness/liveness probes, but this is useful for Docker Compose
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD ["/app/aci-backend", "health"] || exit 1

# Set entrypoint
ENTRYPOINT ["/app/aci-backend"]

# Default command (can be overridden)
CMD ["serve"]

# Labels for metadata (OCI standard)
LABEL org.opencontainers.image.title="ACI Backend" \
      org.opencontainers.image.description="ACI Backend API Service" \
      org.opencontainers.image.vendor="ACI Platform" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.created="${BUILD_TIME}" \
      org.opencontainers.image.source="https://github.com/yourorg/aci-backend" \
      org.opencontainers.image.licenses="MIT"

###############################################################################
# Build Instructions:
###############################################################################
#
# 1. Build locally:
#    docker build -t aci-backend:latest \
#      --build-arg VERSION=1.0.0 \
#      --build-arg BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ") \
#      --build-arg GIT_COMMIT=$(git rev-parse --short HEAD) \
#      -f deployments/Dockerfile .
#
# 2. Build for Kubernetes (with proper tagging):
#    docker build -t aci-backend:$(git rev-parse --short HEAD) \
#      --build-arg VERSION=$(git describe --tags --always) \
#      --build-arg BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ") \
#      --build-arg GIT_COMMIT=$(git rev-parse HEAD) \
#      -f deployments/Dockerfile .
#
# 3. Build for minikube:
#    eval $(minikube docker-env)
#    docker build -t aci-backend:latest -f deployments/Dockerfile .
#
# 4. Build for kind:
#    docker build -t aci-backend:latest -f deployments/Dockerfile .
#    kind load docker-image aci-backend:latest --name <cluster-name>
#
# 5. Multi-platform build (for ARM/AMD):
#    docker buildx build --platform linux/amd64,linux/arm64 \
#      -t aci-backend:latest \
#      -f deployments/Dockerfile .
#
# 6. Scan for vulnerabilities:
#    docker scan aci-backend:latest
#    # Or use trivy:
#    trivy image aci-backend:latest
#
###############################################################################
