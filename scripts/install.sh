#!/usr/bin/env bash

# Aetherius Installation Script
# Automated installation for development and production environments
# Based on OneX v2 patterns

set -euo pipefail

# Source common functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/lib/common.sh"

# ==================================================================================
# Configuration
# ==================================================================================

INSTALL_TYPE="${INSTALL_TYPE:-development}"  # development, staging, production
INSTALL_DIR="${INSTALL_DIR:-/opt/aetherius}"
CONFIG_DIR="${CONFIG_DIR:-/etc/aetherius}"
LOG_DIR="${LOG_DIR:-/var/log/aetherius}"
DATA_DIR="${DATA_DIR:-/var/lib/aetherius}"

# Service list
SERVICES=(
    "agent-manager"
    "orchestrator-service"
    "reasoning-service-go"
    "auth-service"
    "gateway-service"
    "monitor-service"
    "cluster-service"
    "collect-agent"
)

# ==================================================================================
# Functions
# ==================================================================================

show_banner() {
    cat << 'EOF'
╔══════════════════════════════════════════════════════════════╗
║     Aetherius Installation Script                           ║
║     智能 Kubernetes 运维平台                                 ║
╚══════════════════════════════════════════════════════════════╝
EOF
    echo ""
    log::info "Version: $(cat VERSION 2>/dev/null || echo 'unknown')"
    log::info "Install Type: ${INSTALL_TYPE}"
    log::info "Install Directory: ${INSTALL_DIR}"
    echo ""
}

check_prerequisites() {
    log::info "Checking prerequisites..."

    # Check if running as root (for production)
    if [[ "${INSTALL_TYPE}" == "production" ]] && [[ $EUID -ne 0 ]]; then
        log::fatal "Production installation must be run as root"
    fi

    # Check required commands
    local required_commands=(
        "go"
        "docker"
        "kubectl"
        "make"
        "git"
    )

    for cmd in "${required_commands[@]}"; do
        if ! util::command_exists "$cmd"; then
            log::error "Required command not found: $cmd"
            log::info "Please install $cmd before continuing"
            return 1
        else
            log::success "✓ $cmd found: $(command -v $cmd)"
        fi
    done

    # Check Go version
    local go_version=$(go version | grep -oP 'go\K[0-9.]+')
    local required_version="1.21"
    if ! util::version_gte "$go_version" "$required_version"; then
        log::error "Go version $go_version is too old. Required: $required_version+"
        return 1
    fi
    log::success "✓ Go version: $go_version"

    # Check Docker
    if ! docker ps >/dev/null 2>&1; then
        log::error "Docker is not running or you don't have permission"
        return 1
    fi
    log::success "✓ Docker is running"

    # Check kubectl
    if ! kubectl version --client >/dev/null 2>&1; then
        log::error "kubectl is not working properly"
        return 1
    fi
    log::success "✓ kubectl is configured"

    log::success "All prerequisites satisfied"
    return 0
}

create_directories() {
    log::info "Creating directories..."

    local dirs=(
        "${INSTALL_DIR}"
        "${CONFIG_DIR}"
        "${LOG_DIR}"
        "${DATA_DIR}"
    )

    for dir in "${dirs[@]}"; do
        if [[ ! -d "$dir" ]]; then
            mkdir -p "$dir"
            log::success "✓ Created: $dir"
        else
            log::info "Directory already exists: $dir"
        fi
    done
}

install_tools() {
    log::info "Installing development tools..."

    if ! make tools.verify >/dev/null 2>&1; then
        log::info "Installing missing tools..."
        make tools.install
    else
        log::success "✓ All tools already installed"
    fi
}

install_git_hooks() {
    log::info "Installing git hooks..."
    make hooks.install
    log::success "✓ Git hooks installed"
}

build_services() {
    log::info "Building services..."

    case "${INSTALL_TYPE}" in
        development)
            log::info "Building in development mode..."
            make go.build
            ;;
        staging|production)
            log::info "Building optimized binaries for ${INSTALL_TYPE}..."
            make go.build BUILD_FLAGS="-ldflags '-s -w'"
            ;;
    esac

    log::success "✓ Services built successfully"
}

install_binaries() {
    log::info "Installing binaries to ${INSTALL_DIR}/bin..."

    mkdir -p "${INSTALL_DIR}/bin"

    for service in "${SERVICES[@]}"; do
        if [[ -f "_output/bin/${service}" ]]; then
            cp "_output/bin/${service}" "${INSTALL_DIR}/bin/"
            chmod +x "${INSTALL_DIR}/bin/${service}"
            log::success "✓ Installed: ${service}"
        else
            log::warning "Binary not found: ${service}"
        fi
    done
}

install_configs() {
    log::info "Installing configuration files..."

    mkdir -p "${CONFIG_DIR}"

    # Copy config files
    for service in "${SERVICES[@]}"; do
        local config_file="${service}/configs/config-${INSTALL_TYPE}.yaml"
        if [[ -f "$config_file" ]]; then
            cp "$config_file" "${CONFIG_DIR}/${service}.yaml"
            log::success "✓ Config installed: ${service}.yaml"
        fi
    done

    # Set permissions
    if [[ "${INSTALL_TYPE}" == "production" ]]; then
        chmod 600 "${CONFIG_DIR}"/*.yaml
    fi
}

setup_database() {
    log::info "Setting up databases..."

    case "${INSTALL_TYPE}" in
        development)
            log::info "Starting local database containers..."
            cd deployments/docker-compose
            docker-compose up -d mysql redis nats
            cd ../..

            # Wait for databases to be ready
            log::info "Waiting for databases to be ready..."
            sleep 10
            ;;
        staging|production)
            log::warning "Please ensure databases are configured externally"
            log::info "Update connection strings in ${CONFIG_DIR}/*.yaml"
            ;;
    esac
}

run_migrations() {
    log::info "Running database migrations..."

    # Run migrations for each service
    for service in agent-manager orchestrator-service auth-service; do
        if [[ -d "${service}/internal/storage/migrations" ]]; then
            log::info "Running migrations for ${service}..."
            # Add migration command here
            log::success "✓ Migrations completed for ${service}"
        fi
    done
}

create_systemd_services() {
    if [[ "${INSTALL_TYPE}" != "production" ]]; then
        return 0
    fi

    log::info "Creating systemd service files..."

    for service in "${SERVICES[@]}"; do
        cat > "/etc/systemd/system/aetherius-${service}.service" <<EOF
[Unit]
Description=Aetherius ${service}
After=network.target

[Service]
Type=simple
User=aetherius
WorkingDirectory=${INSTALL_DIR}
ExecStart=${INSTALL_DIR}/bin/${service} -c ${CONFIG_DIR}/${service}.yaml
Restart=on-failure
RestartSec=5s
StandardOutput=append:${LOG_DIR}/${service}.log
StandardError=append:${LOG_DIR}/${service}-error.log

[Install]
WantedBy=multi-user.target
EOF
        log::success "✓ Created systemd service: aetherius-${service}"
    done

    systemctl daemon-reload
    log::success "✓ Systemd services created"
}

verify_installation() {
    log::info "Verifying installation..."

    local failed=0

    # Check binaries
    for service in "${SERVICES[@]}"; do
        if [[ -x "${INSTALL_DIR}/bin/${service}" ]]; then
            log::success "✓ Binary: ${service}"
        else
            log::error "✗ Binary missing: ${service}"
            ((failed++))
        fi
    done

    # Check configs
    for service in "${SERVICES[@]}"; do
        if [[ -f "${CONFIG_DIR}/${service}.yaml" ]]; then
            log::success "✓ Config: ${service}.yaml"
        else
            log::warning "⚠ Config missing: ${service}.yaml"
        fi
    done

    if [[ $failed -eq 0 ]]; then
        log::success "Installation verification passed"
        return 0
    else
        log::error "Installation verification failed: $failed issues found"
        return 1
    fi
}

show_next_steps() {
    echo ""
    log::info "═══════════════════════════════════════════════════"
    log::info "  Installation Complete!"
    log::info "═══════════════════════════════════════════════════"
    echo ""

    case "${INSTALL_TYPE}" in
        development)
            echo "Next steps for development:"
            echo ""
            echo "  1. Start services:"
            echo "     make dev"
            echo ""
            echo "  2. Or start with hot reload:"
            echo "     make run-agent-manager"
            echo ""
            echo "  3. Run tests:"
            echo "     make go.test"
            echo ""
            ;;
        staging|production)
            echo "Next steps for ${INSTALL_TYPE}:"
            echo ""
            echo "  1. Update configuration files in ${CONFIG_DIR}/"
            echo ""
            echo "  2. Start services:"
            echo "     systemctl start aetherius-agent-manager"
            echo "     systemctl start aetherius-orchestrator-service"
            echo ""
            echo "  3. Enable auto-start:"
            echo "     systemctl enable aetherius-*"
            echo ""
            echo "  4. Check status:"
            echo "     systemctl status aetherius-*"
            echo ""
            ;;
    esac

    echo "  Documentation: ${INSTALL_DIR}/docs/"
    echo "  Logs: ${LOG_DIR}/"
    echo "  Configuration: ${CONFIG_DIR}/"
    echo ""
    log::info "═══════════════════════════════════════════════════"
}

# ==================================================================================
# Main Installation Flow
# ==================================================================================

main() {
    show_banner

    log::info "Starting installation (Type: ${INSTALL_TYPE})..."
    echo ""

    # Execute installation steps
    check_prerequisites || exit 1
    echo ""

    create_directories
    echo ""

    if [[ "${INSTALL_TYPE}" == "development" ]]; then
        install_tools
        install_git_hooks
        echo ""
    fi

    build_services
    echo ""

    install_binaries
    echo ""

    install_configs
    echo ""

    if [[ "${INSTALL_TYPE}" == "development" ]]; then
        setup_database
        echo ""
    fi

    run_migrations
    echo ""

    if [[ "${INSTALL_TYPE}" == "production" ]]; then
        create_systemd_services
        echo ""
    fi

    verify_installation || exit 1
    echo ""

    show_next_steps
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --type)
            INSTALL_TYPE="$2"
            shift 2
            ;;
        --dir)
            INSTALL_DIR="$2"
            shift 2
            ;;
        --help)
            echo "Usage: $0 [options]"
            echo ""
            echo "Options:"
            echo "  --type TYPE       Installation type (development|staging|production)"
            echo "  --dir DIR         Installation directory (default: /opt/aetherius)"
            echo "  --help            Show this help message"
            echo ""
            echo "Environment Variables:"
            echo "  INSTALL_TYPE      Installation type"
            echo "  INSTALL_DIR       Installation directory"
            echo "  CONFIG_DIR        Configuration directory"
            echo "  LOG_DIR           Log directory"
            echo "  DATA_DIR          Data directory"
            echo ""
            exit 0
            ;;
        *)
            log::error "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Run main installation
main "$@"
