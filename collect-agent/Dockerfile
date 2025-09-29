# Build stage
FROM golang:1.25-alpine AS builder

# Install necessary packages
RUN apk add --no-cache git ca-certificates tzdata kubectl

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o collect-agent \
    ./main.go

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates curl kubectl && \
    rm -rf /var/cache/apk/*

# Create non-root user
RUN addgroup -g 65534 -S appgroup && \
    adduser -u 65534 -S appuser -G appgroup

# Create directories
RUN mkdir -p /etc/aetherius /tmp && \
    chown -R appuser:appgroup /etc/aetherius /tmp

WORKDIR /

# Copy binary from builder stage
COPY --from=builder /app/collect-agent /usr/local/bin/collect-agent

# Copy default config (optional)
COPY --from=builder /app/manifests/03-configmap.yaml /etc/aetherius/default-config.yaml

# Make binary executable
RUN chmod +x /usr/local/bin/collect-agent

# Use non-root user
USER appuser

# Expose health check port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=30s --retries=3 \
    CMD curl -f http://localhost:8080/health/live || exit 1

# Default command
ENTRYPOINT ["/usr/local/bin/collect-agent"]
CMD ["--config=/etc/aetherius/config.yaml"]