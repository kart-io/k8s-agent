# Build stage
FROM golang:1.24-alpine AS builder

# Build arguments for cross-compilation
ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH
ARG GOPROXY=https://goproxy.cn,https://goproxy.io,direct

# Set Go proxy to solve network timeout issues
ENV GOPROXY=${GOPROXY}

# Install necessary packages
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} \
    go build -ldflags='-w -s' -o collect-agent ./main.go

# Final stage - using scratch for zero vulnerabilities (no OS packages)
FROM scratch

# Copy CA certificates from builder for HTTPS
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy timezone data from builder
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy binary from builder stage
COPY --from=builder /app/collect-agent /collect-agent

# Note: scratch image has no shell, user management, or package manager
# This provides the smallest attack surface (zero OS vulnerabilities)
# Health checks must be done via Kubernetes liveness/readiness probes

# Expose health check port
EXPOSE 8080

# Run as non-root user (UID 65534 = nobody)
USER 65534:65534

# Default command
ENTRYPOINT ["/collect-agent"]
CMD ["--config=/etc/aetherius/config.yaml"]