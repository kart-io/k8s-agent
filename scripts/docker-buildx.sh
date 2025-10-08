#!/bin/bash

# 多平台 Docker 构建脚本
# 支持 linux/amd64, linux/arm64

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 默认配置
PLATFORMS="${PLATFORMS:-linux/amd64,linux/arm64}"
DOCKER_REGISTRY="${DOCKER_REGISTRY:-docker.io}"
DOCKER_NAMESPACE="${DOCKER_NAMESPACE:-aetherius}"
PUSH="${PUSH:-false}"
BUILD_ARGS="${BUILD_ARGS:-}"

# 显示使用方法
usage() {
    echo "使用方法: $0 <service-name> [options]"
    echo ""
    echo "参数:"
    echo "  service-name              服务名称 (必需)"
    echo ""
    echo "选项:"
    echo "  -v, --version VERSION     镜像版本 (默认: v1.0.0)"
    echo "  -p, --platforms PLATFORMS 目标平台 (默认: linux/amd64,linux/arm64)"
    echo "  -r, --registry REGISTRY   Docker registry (默认: docker.io)"
    echo "  -n, --namespace NAMESPACE Docker namespace (默认: aetherius)"
    echo "  --push                    构建后推送到 registry"
    echo "  -h, --help               显示此帮助信息"
    echo ""
    echo "环境变量:"
    echo "  PLATFORMS                 目标平台列表"
    echo "  DOCKER_REGISTRY          Docker registry"
    echo "  DOCKER_NAMESPACE         Docker namespace"
    echo "  PUSH                     是否推送 (true/false)"
    echo ""
    echo "示例:"
    echo "  $0 collect-agent -v v1.0.0"
    echo "  $0 agent-manager -v v1.1.0 --push"
    echo "  $0 gateway-service -v latest --platforms linux/amd64"
    exit 1
}

# 解析参数
SERVICE_NAME=""
VERSION="v1.0.0"

while [[ $# -gt 0 ]]; do
    case $1 in
        -v|--version)
            VERSION="$2"
            shift 2
            ;;
        -p|--platforms)
            PLATFORMS="$2"
            shift 2
            ;;
        -r|--registry)
            DOCKER_REGISTRY="$2"
            shift 2
            ;;
        -n|--namespace)
            DOCKER_NAMESPACE="$2"
            shift 2
            ;;
        --push)
            PUSH="true"
            shift
            ;;
        -h|--help)
            usage
            ;;
        *)
            if [ -z "$SERVICE_NAME" ]; then
                SERVICE_NAME="$1"
            else
                echo -e "${RED}错误: 未知参数 $1${NC}"
                usage
            fi
            shift
            ;;
    esac
done

# 检查服务名称
if [ -z "$SERVICE_NAME" ]; then
    echo -e "${RED}错误: 必须指定服务名称${NC}"
    usage
fi

# 构建变量
IMAGE_NAME="${DOCKER_REGISTRY}/${DOCKER_NAMESPACE}/${SERVICE_NAME}"
IMAGE_TAG="${VERSION}"

# 服务目录映射
declare -A SERVICE_DIRS=(
    ["collect-agent"]="collect-agent"
    ["agent-manager"]="agent-manager"
    ["orchestrator-service"]="orchestrator-service"
    ["gateway-service"]="gateway-service"
    ["auth-service"]="auth-service"
    ["reasoning-service"]="reasoning-service"
    ["reasoning-service-go"]="reasoning-service-go"
)

SERVICE_DIR="${SERVICE_DIRS[$SERVICE_NAME]}"

if [ -z "$SERVICE_DIR" ]; then
    echo -e "${RED}错误: 未知服务 '$SERVICE_NAME'${NC}"
    echo -e "${YELLOW}支持的服务:${NC}"
    for service in "${!SERVICE_DIRS[@]}"; do
        echo "  - $service"
    done
    exit 1
fi

# 切换到服务目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
SERVICE_PATH="${PROJECT_ROOT}/${SERVICE_DIR}"

if [ ! -d "$SERVICE_PATH" ]; then
    echo -e "${RED}错误: 服务目录不存在: $SERVICE_PATH${NC}"
    exit 1
fi

if [ ! -f "$SERVICE_PATH/Dockerfile" ]; then
    echo -e "${RED}错误: Dockerfile 不存在: $SERVICE_PATH/Dockerfile${NC}"
    exit 1
fi

cd "$SERVICE_PATH"

# 显示构建信息
echo -e "${GREEN}======================================${NC}"
echo -e "${GREEN}多平台 Docker 构建${NC}"
echo -e "${GREEN}======================================${NC}"
echo -e "服务:      ${YELLOW}${SERVICE_NAME}${NC}"
echo -e "版本:      ${YELLOW}${IMAGE_TAG}${NC}"
echo -e "镜像:      ${YELLOW}${IMAGE_NAME}:${IMAGE_TAG}${NC}"
echo -e "平台:      ${YELLOW}${PLATFORMS}${NC}"
echo -e "推送:      ${YELLOW}${PUSH}${NC}"
echo -e "${GREEN}======================================${NC}"
echo ""

# 检查 Docker buildx
if ! docker buildx version >/dev/null 2>&1; then
    echo -e "${RED}错误: Docker buildx 未安装${NC}"
    echo -e "${YELLOW}请运行: docker buildx install${NC}"
    exit 1
fi

# 创建 buildx builder (如果不存在)
BUILDER_NAME="k8s-agent-builder"
if ! docker buildx inspect "$BUILDER_NAME" >/dev/null 2>&1; then
    echo -e "${YELLOW}创建 buildx builder: $BUILDER_NAME${NC}"
    docker buildx create --name "$BUILDER_NAME" --use --bootstrap
else
    echo -e "${GREEN}使用现有 builder: $BUILDER_NAME${NC}"
    docker buildx use "$BUILDER_NAME"
fi

# 构建命令
BUILD_CMD="docker buildx build"
BUILD_CMD="$BUILD_CMD --platform $PLATFORMS"
BUILD_CMD="$BUILD_CMD -t ${IMAGE_NAME}:${IMAGE_TAG}"
BUILD_CMD="$BUILD_CMD -t ${IMAGE_NAME}:latest"

# 添加构建参数
if [ -n "$BUILD_ARGS" ]; then
    BUILD_CMD="$BUILD_CMD $BUILD_ARGS"
fi

# 是否推送
if [ "$PUSH" = "true" ]; then
    BUILD_CMD="$BUILD_CMD --push"
else
    BUILD_CMD="$BUILD_CMD --load"
fi

BUILD_CMD="$BUILD_CMD ."

# 执行构建
echo -e "${GREEN}开始构建...${NC}"
echo -e "${YELLOW}命令: $BUILD_CMD${NC}"
echo ""

if eval "$BUILD_CMD"; then
    echo ""
    echo -e "${GREEN}======================================${NC}"
    echo -e "${GREEN}构建成功!${NC}"
    echo -e "${GREEN}======================================${NC}"
    echo -e "镜像: ${YELLOW}${IMAGE_NAME}:${IMAGE_TAG}${NC}"
    echo -e "平台: ${YELLOW}${PLATFORMS}${NC}"
    if [ "$PUSH" = "true" ]; then
        echo -e "状态: ${YELLOW}已推送到 registry${NC}"
    else
        echo -e "状态: ${YELLOW}已加载到本地${NC}"
    fi
    echo -e "${GREEN}======================================${NC}"
else
    echo ""
    echo -e "${RED}======================================${NC}"
    echo -e "${RED}构建失败!${NC}"
    echo -e "${RED}======================================${NC}"
    exit 1
fi
