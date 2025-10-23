#!/usr/bin/env bash

# Environment Setup and Validation Script
# Sets up development environment and validates configuration
# Based on OneX v2 patterns

set -euo pipefail

# Source common functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/lib/common.sh"

# ==================================================================================
# Configuration
# ==================================================================================

ENV="${ENV:-development}"  # development, staging, production, testing
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

# ==================================================================================
# Functions
# ==================================================================================

show_banner() {
    cat << 'EOF'
╔══════════════════════════════════════════════════════════════╗
║     Aetherius Environment Setup                             ║
╚══════════════════════════════════════════════════════════════╝
EOF
    echo ""
    log::info "Environment: ${ENV}"
    log::info "Project Root: ${PROJECT_ROOT}"
    echo ""
}

check_go_environment() {
    log::info "Checking Go environment..."

    # Check Go version
    if ! util::command_exists go; then
        log::error "Go is not installed"
        log::info "Please install Go 1.21+ from https://golang.org/dl/"
        return 1
    fi

    local go_version=$(go version | grep -oP 'go\K[0-9.]+')
    local required_version=$(cat .go-version 2>/dev/null || echo "1.21")

    log::success "✓ Go version: ${go_version} (required: ${required_version}+)"

    # Check GOPATH
    if [[ -z "${GOPATH:-}" ]]; then
        log::warning "GOPATH is not set, using default: $(go env GOPATH)"
        export GOPATH=$(go env GOPATH)
    else
        log::success "✓ GOPATH: ${GOPATH}"
    fi

    # Check GOBIN
    if [[ -z "${GOBIN:-}" ]]; then
        export GOBIN="${GOPATH}/bin"
        log::info "GOBIN set to: ${GOBIN}"
    else
        log::success "✓ GOBIN: ${GOBIN}"
    fi

    # Add GOBIN to PATH if not already there
    if [[ ":${PATH}:" != *":${GOBIN}:"* ]]; then
        export PATH="${GOBIN}:${PATH}"
        log::info "Added ${GOBIN} to PATH"
    fi

    return 0
}

check_docker_environment() {
    log::info "Checking Docker environment..."

    if ! util::command_exists docker; then
        log::error "Docker is not installed"
        log::info "Please install Docker from https://docs.docker.com/get-docker/"
        return 1
    fi

    log::success "✓ Docker: $(docker version --format '{{.Client.Version}}')"

    # Check if Docker daemon is running
    if ! docker ps >/dev/null 2>&1; then
        log::error "Docker daemon is not running"
        return 1
    fi
    log::success "✓ Docker daemon is running"

    # Check Docker Compose
    if util::command_exists docker-compose; then
        log::success "✓ Docker Compose: $(docker-compose version --short)"
    else
        log::warning "Docker Compose not found (optional)"
    fi

    return 0
}

check_kubernetes_environment() {
    log::info "Checking Kubernetes environment..."

    if ! util::command_exists kubectl; then
        log::warning "kubectl is not installed (optional for development)"
        return 0
    fi

    log::success "✓ kubectl: $(kubectl version --client --short 2>/dev/null | head -1)"

    # Check if kubectl can connect to a cluster
    if kubectl cluster-info >/dev/null 2>&1; then
        log::success "✓ Connected to Kubernetes cluster"
        kubectl get nodes 2>/dev/null | head -5
    else
        log::warning "Not connected to a Kubernetes cluster (optional for development)"
    fi

    return 0
}

check_required_tools() {
    log::info "Checking required development tools..."

    local tools=(
        "make:GNU Make:required"
        "git:Git:required"
        "curl:cURL:required"
        "jq:jq JSON processor:optional"
    )

    local missing_required=0

    for tool_spec in "${tools[@]}"; do
        IFS=':' read -r cmd name requirement <<< "$tool_spec"

        if util::command_exists "$cmd"; then
            log::success "✓ ${name}: $(command -v $cmd)"
        else
            if [[ "$requirement" == "required" ]]; then
                log::error "✗ ${name} is required but not installed"
                ((missing_required++))
            else
                log::warning "⚠ ${name} is optional but not installed"
            fi
        fi
    done

    if [[ $missing_required -gt 0 ]]; then
        log::error "Missing $missing_required required tools"
        return 1
    fi

    return 0
}

check_project_dependencies() {
    log::info "Checking project dependencies..."

    cd "${PROJECT_ROOT}"

    # Check if go.mod exists
    if [[ ! -f "go.mod" ]]; then
        log::error "go.mod not found"
        return 1
    fi
    log::success "✓ go.mod exists"

    # Verify dependencies
    log::info "Verifying Go dependencies..."
    if go mod verify >/dev/null 2>&1; then
        log::success "✓ Go dependencies verified"
    else
        log::warning "Go dependencies verification failed, running 'go mod tidy'..."
        go mod tidy
    fi

    # Download dependencies if needed
    log::info "Ensuring all dependencies are downloaded..."
    go mod download

    log::success "✓ All dependencies ready"

    return 0
}

setup_git_config() {
    log::info "Setting up Git configuration..."

    cd "${PROJECT_ROOT}"

    # Check if git hooks are installed
    if [[ -f ".git/hooks/pre-commit" ]] && [[ -f ".git/hooks/commit-msg" ]]; then
        log::success "✓ Git hooks are installed"
    else
        log::info "Installing Git hooks..."
        make hooks.install >/dev/null 2>&1 || true
        log::success "✓ Git hooks installed"
    fi

    # Configure git to use LF line endings
    git config core.autocrlf input 2>/dev/null || true

    return 0
}

create_output_directories() {
    log::info "Creating output directories..."

    local dirs=(
        "_output/bin"
        "_output/logs"
        "_output/coverage"
        "_output/tmp"
    )

    for dir in "${dirs[@]}"; do
        mkdir -p "${PROJECT_ROOT}/${dir}"
    done

    log::success "✓ Output directories created"

    return 0
}

setup_environment_file() {
    log::info "Setting up environment configuration..."

    local env_file="${PROJECT_ROOT}/.env.${ENV}"

    if [[ ! -f "$env_file" ]]; then
        log::info "Creating ${env_file}..."

        cat > "$env_file" <<EOF
# Aetherius Environment Configuration - ${ENV}
# Generated on $(date)

# Environment
ENVIRONMENT=${ENV}

# Go Configuration
GOPATH=${GOPATH:-$(go env GOPATH)}
GOBIN=${GOBIN:-${GOPATH}/bin}

# Database Configuration
DB_HOST=localhost
DB_PORT=3306
DB_USER=aetherius
DB_PASSWORD=aetherius_pwd
DB_NAME=aetherius

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# NATS Configuration
NATS_URL=nats://localhost:4222

# Neo4j Configuration
NEO4J_URI=bolt://localhost:7687
NEO4J_USER=neo4j
NEO4J_PASSWORD=aetherius_pwd

# Service Ports
AGENT_MANAGER_PORT=8080
ORCHESTRATOR_PORT=8081
REASONING_PORT=8082
AUTH_PORT=8083

# Logging
LOG_LEVEL=debug
LOG_FORMAT=json

# Development Options (for development environment)
ENABLE_HOT_RELOAD=true
ENABLE_DEBUG=true
ENABLE_PROFILING=true
EOF

        log::success "✓ Created ${env_file}"
        log::info "Please review and update the configuration"
    else
        log::success "✓ Environment file already exists: ${env_file}"
    fi

    return 0
}

validate_configuration() {
    log::info "Validating configuration files..."

    cd "${PROJECT_ROOT}"

    local errors=0

    # Check required configuration files
    local required_files=(
        ".golangci.yml"
        ".air.toml"
        "VERSION"
        "go.mod"
        "Makefile"
    )

    for file in "${required_files[@]}"; do
        if [[ -f "$file" ]]; then
            log::success "✓ ${file}"
        else
            log::error "✗ ${file} is missing"
            ((errors++))
        fi
    done

    # Validate YAML files
    log::info "Validating YAML configuration files..."
    for yaml_file in $(find . -maxdepth 1 -name "*.yaml" -o -name "*.yml" | grep -v ".golangci.yml"); do
        if [[ -f "$yaml_file" ]]; then
            # Basic YAML validation (check if file is readable)
            if cat "$yaml_file" > /dev/null 2>&1; then
                log::success "✓ Valid YAML: ${yaml_file}"
            else
                log::error "✗ Invalid YAML: ${yaml_file}"
                ((errors++))
            fi
        fi
    done

    if [[ $errors -eq 0 ]]; then
        log::success "Configuration validation passed"
        return 0
    else
        log::error "Configuration validation failed with $errors errors"
        return 1
    fi
}

show_environment_info() {
    echo ""
    log::info "═══════════════════════════════════════════════════"
    log::info "  Environment Information"
    log::info "═══════════════════════════════════════════════════"
    echo ""
    echo "Environment:     ${ENV}"
    echo "Project Root:    ${PROJECT_ROOT}"
    echo "Go Version:      $(go version | awk '{print $3}')"
    echo "Go Path:         ${GOPATH}"
    echo "Go Bin:          ${GOBIN}"
    echo "Docker Version:  $(docker version --format '{{.Client.Version}}' 2>/dev/null || echo 'N/A')"
    echo "Make Version:    $(make --version | head -1)"
    echo "Git Version:     $(git --version)"
    echo ""
    log::info "═══════════════════════════════════════════════════"
    echo ""
}

show_next_steps() {
    echo ""
    log::info "═══════════════════════════════════════════════════"
    log::info "  Setup Complete!"
    log::info "═══════════════════════════════════════════════════"
    echo ""
    echo "Next steps:"
    echo ""
    echo "  1. Review environment configuration:"
    echo "     cat .env.${ENV}"
    echo ""
    echo "  2. Install development tools:"
    echo "     make tools.install"
    echo ""
    echo "  3. Build the project:"
    echo "     make go.build"
    echo ""
    echo "  4. Run tests:"
    echo "     make go.test"
    echo ""
    echo "  5. Start development:"
    echo "     make dev"
    echo ""
    log::info "═══════════════════════════════════════════════════"
}

# ==================================================================================
# Main Flow
# ==================================================================================

main() {
    show_banner

    local failed=0

    # Check all environments
    check_go_environment || ((failed++))
    echo ""

    check_docker_environment || ((failed++))
    echo ""

    check_kubernetes_environment || true  # Optional, don't fail
    echo ""

    check_required_tools || ((failed++))
    echo ""

    if [[ $failed -gt 0 ]]; then
        log::error "Environment check failed with $failed critical errors"
        exit 1
    fi

    # Setup project
    check_project_dependencies || exit 1
    echo ""

    create_output_directories
    echo ""

    setup_git_config
    echo ""

    setup_environment_file
    echo ""

    validate_configuration || exit 1
    echo ""

    show_environment_info
    show_next_steps
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --env)
            ENV="$2"
            shift 2
            ;;
        --help)
            echo "Usage: $0 [options]"
            echo ""
            echo "Options:"
            echo "  --env ENV     Environment (development|staging|production|testing)"
            echo "  --help        Show this help message"
            echo ""
            echo "Environment Variables:"
            echo "  ENV           Environment name"
            echo ""
            exit 0
            ;;
        *)
            log::error "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Run main setup
main "$@"
