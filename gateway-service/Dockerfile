# 构建阶段
FROM golang:1.24-alpine AS builder

# Build arguments for Go proxy
ARG GOPROXY=https://goproxy.cn,https://goproxy.io,direct

# Set Go proxy to solve network timeout issues
ENV GOPROXY=${GOPROXY}

# 安装必要的工具
RUN apk add --no-cache git make

# 设置工作目录
WORKDIR /build

# 复制 go mod 文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# Build arguments for cross-compilation
ARG TARGETOS
ARG TARGETARCH

# 编译
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} \
    go build -ldflags="-w -s" -o gateway-service cmd/server/main.go

# 运行阶段 - using scratch for zero vulnerabilities
FROM scratch

# Copy CA certificates from builder for HTTPS
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy timezone data from builder
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# 从构建阶段复制二进制文件
COPY --from=builder /build/gateway-service /gateway-service

# 复制配置文件
COPY configs /configs

# 暴露端口
EXPOSE 8080

# Run as non-root user (UID 65534 = nobody)
USER 65534:65534

# 运行
ENTRYPOINT ["/gateway-service"]
