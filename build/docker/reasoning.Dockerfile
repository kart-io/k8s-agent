# Build stage
FROM golang:1.24-alpine AS builder

# Build arguments for Go proxy
ARG GOPROXY=https://goproxy.cn,https://goproxy.io,direct

# Set Go proxy to solve network timeout issues
ENV GOPROXY=${GOPROXY}

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build arguments for cross-compilation
ARG TARGETOS
ARG TARGETARCH

# Build binary
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} \
    go build -ldflags="-w -s" -o reasoning-service cmd/server/main.go

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/reasoning-service .

# Copy config
COPY configs configs/

# Create logs directory
RUN mkdir -p logs data

# Expose port
EXPOSE 8082

# Run
CMD ["./reasoning-service", "-config", "configs/config.yaml"]
