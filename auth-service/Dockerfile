# Multi-stage Dockerfile for auth-service
# Supports multi-platform builds (linux/amd64, linux/arm64)

# Build stage
FROM golang:1.24-alpine AS builder

# Build arguments for Go proxy
ARG GOPROXY=https://goproxy.cn,https://goproxy.io,direct

# Set Go proxy to solve network timeout issues
ENV GOPROXY=${GOPROXY}

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build arguments for cross-compilation
ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

# Build the binary
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -ldflags="-w -s" -o auth-service ./cmd/server

# Runtime stage - using scratch for zero vulnerabilities
FROM scratch

# Copy CA certificates from builder for HTTPS
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy timezone data from builder
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy binary and configs from builder
COPY --from=builder /build/auth-service /auth-service
COPY --from=builder /build/configs /configs

# Expose port
EXPOSE 8090

# Run as non-root user (UID 65534 = nobody)
USER 65534:65534

# Run the binary
ENTRYPOINT ["/auth-service"]
CMD ["--config=/configs/config.yaml"]
