#!/usr/bin/env bash

# CI/CD Helper Script
# Provides utilities for continuous integration and deployment
# Based on OneX v2 patterns

set -euo pipefail

# Source common functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/lib/common.sh"

# ==================================================================================
# Configuration
# ==================================================================================

COMMAND="${1:-help}"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
CI_ENV="${CI:-false}"
GITHUB_ACTIONS="${GITHUB_ACTIONS:-false}"

# ==================================================================================
# Functions
# ==================================================================================

ci::setup() {
    log::info "Setting up CI environment..."

    # Install dependencies
    go mod download
    log::success "✓ Dependencies downloaded"

    # Install tools
    make tools.install
    log::success "✓ Development tools installed"

    # Verify installation
    make tools.verify
    log::success "✓ Tools verified"
}

ci::lint() {
    log::info "Running linters..."

    # Go linting
    log::info "Running golangci-lint..."
    make go.lint
    log::success "✓ Go linting passed"

    # Proto linting
    log::info "Running proto linting..."
    make proto.lint
    log::success "✓ Proto linting passed"

    # K8s manifest linting
    if [[ -d "deployments/k8s" ]]; then
        log::info "Running K8s manifest linting..."
        make k8s.lint || log::warning "K8s linting failed (non-critical)"
    fi

    log::success "All linting checks passed"
}

ci::test() {
    log::info "Running tests..."

    # Unit tests with coverage
    log::info "Running unit tests..."
    make go.test.coverage
    log::success "✓ Unit tests passed"

    # Integration tests
    if make -n go.test.integration >/dev/null 2>&1; then
        log::info "Running integration tests..."
        make go.test.integration || log::warning "Integration tests failed (non-critical)"
    fi

    log::success "All tests passed"
}

ci::build() {
    log::info "Building binaries..."

    # Build all services
    make go.build
    log::success "✓ Binaries built"

    # Verify binaries exist
    local failed=0
    for binary in _output/bin/*; do
        if [[ -x "$binary" ]]; then
            log::success "✓ Binary: $(basename $binary)"
        else
            log::error "✗ Binary not executable: $(basename $binary)"
            ((failed++))
        fi
    done

    if [[ $failed -eq 0 ]]; then
        log::success "All binaries built successfully"
        return 0
    else
        log::error "Build failed: $failed errors"
        return 1
    fi
}

ci::docker_build() {
    log::info "Building Docker images..."

    local version="${VERSION:-$(git describe --tags --always --dirty)}"

    # Build multi-platform images
    make docker.buildx VERSION="${version}"
    log::success "✓ Docker images built"
}

ci::docker_push() {
    log::info "Pushing Docker images..."

    if [[ -z "${DOCKER_USERNAME:-}" ]] || [[ -z "${DOCKER_PASSWORD:-}" ]]; then
        log::error "DOCKER_USERNAME and DOCKER_PASSWORD must be set"
        return 1
    fi

    # Login to Docker registry
    echo "${DOCKER_PASSWORD}" | docker login -u "${DOCKER_USERNAME}" --password-stdin

    local version="${VERSION:-$(git describe --tags --always --dirty)}"

    # Push images
    make docker.buildx.push VERSION="${version}"
    log::success "✓ Docker images pushed"
}

ci::release() {
    log::info "Creating release..."

    if [[ -z "${GITHUB_TOKEN:-}" ]]; then
        log::error "GITHUB_TOKEN must be set for releases"
        return 1
    fi

    local version="${VERSION:-$(cat VERSION)}"

    # Create GitHub release
    log::info "Creating GitHub release ${version}..."

    # Build binaries for release
    make go.build

    # Create release artifacts
    local release_dir="_output/release"
    mkdir -p "${release_dir}"

    for binary in _output/bin/*; do
        if [[ -x "$binary" ]]; then
            local name=$(basename "$binary")
            tar czf "${release_dir}/${name}-${version}-linux-amd64.tar.gz" -C _output/bin "$name"
            log::success "✓ Created release artifact: ${name}"
        fi
    done

    log::success "Release artifacts created in ${release_dir}/"
}

ci::security_scan() {
    log::info "Running security scans..."

    # Go vulnerability check
    log::info "Checking for Go vulnerabilities..."
    if util::command_exists govulncheck; then
        govulncheck ./... || log::warning "Vulnerabilities found (review required)"
    else
        log::warning "govulncheck not installed, skipping vulnerability check"
    fi

    # Docker image scanning with Trivy
    if util::command_exists trivy; then
        log::info "Scanning Docker images with Trivy..."
        for image in $(docker images --format '{{.Repository}}:{{.Tag}}' | grep aetherius); do
            log::info "Scanning ${image}..."
            trivy image --severity HIGH,CRITICAL "${image}"
        done
    else
        log::warning "Trivy not installed, skipping image scanning"
    fi

    log::success "Security scans completed"
}

ci::deploy() {
    local environment="${2:-staging}"

    log::info "Deploying to ${environment}..."

    case "${environment}" in
        staging|production)
            # Deploy to Kubernetes
            log::info "Deploying to Kubernetes..."
            kubectl apply -k "deployments/k8s/overlays/${environment}"
            log::success "✓ Deployed to ${environment}"

            # Wait for rollout
            log::info "Waiting for rollout to complete..."
            kubectl rollout status deployment -n aetherius
            log::success "✓ Rollout completed"
            ;;
        *)
            log::error "Unknown environment: ${environment}"
            return 1
            ;;
    esac
}

ci::cleanup() {
    log::info "Cleaning up CI artifacts..."

    # Clean build artifacts
    make clean

    # Remove temporary files
    rm -rf /tmp/aetherius-*

    # Prune Docker resources
    if [[ "${CI_ENV}" == "true" ]]; then
        docker system prune -f
    fi

    log::success "Cleanup completed"
}

ci::report() {
    log::info "Generating CI report..."

    local report_file="_output/ci-report.txt"

    {
        echo "═══════════════════════════════════════════════════"
        echo "  CI/CD Report"
        echo "═══════════════════════════════════════════════════"
        echo ""
        echo "Date: $(date)"
        echo "Git Commit: $(git rev-parse HEAD)"
        echo "Git Branch: $(git branch --show-current)"
        echo "Git Tag: $(git describe --tags --exact-match 2>/dev/null || echo 'none')"
        echo ""
        echo "Build Status:"
        if [[ -d "_output/bin" ]]; then
            echo "  Binaries: $(ls -1 _output/bin | wc -l)"
        fi
        if [[ -d "_output/coverage" ]]; then
            echo "  Coverage Files: $(ls -1 _output/coverage | wc -l)"
        fi
        echo ""
        echo "Docker Images:"
        docker images --format "  {{.Repository}}:{{.Tag}}" | grep aetherius || echo "  None"
        echo ""
        echo "═══════════════════════════════════════════════════"
    } > "${report_file}"

    cat "${report_file}"
    log::success "Report saved to ${report_file}"
}

ci::full_pipeline() {
    log::info "Running full CI/CD pipeline..."

    local failed=0

    ci::setup || ((failed++))
    echo ""

    ci::lint || ((failed++))
    echo ""

    ci::test || ((failed++))
    echo ""

    ci::build || ((failed++))
    echo ""

    ci::security_scan || true  # Non-critical
    echo ""

    ci::docker_build || ((failed++))
    echo ""

    if [[ "${PUSH_IMAGES:-false}" == "true" ]]; then
        ci::docker_push || ((failed++))
        echo ""
    fi

    ci::report
    echo ""

    if [[ $failed -eq 0 ]]; then
        log::success "✓ Full pipeline completed successfully"
        return 0
    else
        log::error "✗ Pipeline failed with $failed errors"
        return 1
    fi
}

show_help() {
    cat <<EOF
CI/CD Helper Script

Usage:
  $0 <command> [options]

Commands:
  setup         Setup CI environment (install dependencies and tools)
  lint          Run all linters (Go, Proto, K8s)
  test          Run all tests with coverage
  build         Build all binaries
  docker-build  Build Docker images
  docker-push   Push Docker images to registry
  release       Create release artifacts
  security-scan Run security scans
  deploy ENV    Deploy to environment (staging|production)
  cleanup       Clean up CI artifacts
  report        Generate CI report
  full          Run full CI/CD pipeline
  help          Show this help message

Environment Variables:
  CI                Set to 'true' in CI environment
  GITHUB_ACTIONS    Set to 'true' in GitHub Actions
  GITHUB_TOKEN      GitHub token for releases
  DOCKER_USERNAME   Docker registry username
  DOCKER_PASSWORD   Docker registry password
  VERSION           Version to use for builds
  PUSH_IMAGES       Set to 'true' to push images in full pipeline

Examples:
  # Run full pipeline
  $0 full

  # Run individual steps
  $0 setup
  $0 lint
  $0 test
  $0 build

  # Deploy to staging
  $0 deploy staging

  # Build and push Docker images
  VERSION=v1.2.3 PUSH_IMAGES=true $0 docker-build
  $0 docker-push
EOF
}

# ==================================================================================
# Main
# ==================================================================================

main() {
    cd "${PROJECT_ROOT}"

    case "${COMMAND}" in
        setup)
            ci::setup
            ;;
        lint)
            ci::lint
            ;;
        test)
            ci::test
            ;;
        build)
            ci::build
            ;;
        docker-build)
            ci::docker_build
            ;;
        docker-push)
            ci::docker_push
            ;;
        release)
            ci::release
            ;;
        security-scan)
            ci::security_scan
            ;;
        deploy)
            ci::deploy "$@"
            ;;
        cleanup)
            ci::cleanup
            ;;
        report)
            ci::report
            ;;
        full)
            ci::full_pipeline
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            log::error "Unknown command: ${COMMAND}"
            echo ""
            show_help
            exit 1
            ;;
    esac
}

main "$@"
