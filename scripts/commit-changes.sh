#!/bin/bash

################################################################################
# Git Commit Automation Script for Optimization Rounds 1-3
#
# This script automates the creation of Git commits for all optimization work.
# It creates 3 separate commits, one for each optimization round.
#
# Usage:
#   ./scripts/commit-changes.sh              # Interactive mode
#   ./scripts/commit-changes.sh --all        # Commit all rounds automatically
#   ./scripts/commit-changes.sh --round 1    # Commit specific round only
#   ./scripts/commit-changes.sh --dry-run    # Preview without committing
#
# Author: Optimization Team
# Date: 2025-11-06
################################################################################

set -e  # Exit on error

# Color definitions
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
MAGENTA='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
COMMIT_MSGS_DIR="${PROJECT_ROOT}/docs/commits"

# Configuration
DRY_RUN=false
COMMIT_ALL=false
SPECIFIC_ROUND=""

################################################################################
# Helper Functions
################################################################################

print_header() {
    echo -e "\n${CYAN}============================================================${NC}"
    echo -e "${CYAN}$1${NC}"
    echo -e "${CYAN}============================================================${NC}\n"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ $1${NC}"
}

print_step() {
    echo -e "${MAGENTA}▶ $1${NC}"
}

################################################################################
# Git Status Functions
################################################################################

check_git_status() {
    print_header "Checking Git Status"

    # Check if we're in a git repository
    if ! git rev-parse --git-dir > /dev/null 2>&1; then
        print_error "Not a git repository!"
        exit 1
    fi

    # Check current branch
    current_branch=$(git branch --show-current)
    print_info "Current branch: ${current_branch}"

    # Check for uncommitted changes
    if [[ -z $(git status --porcelain) ]]; then
        print_warning "No changes to commit!"
        echo -e "\n${YELLOW}The working directory is clean. Nothing to commit.${NC}"
        exit 0
    fi

    # Show summary of changes
    echo -e "\n${BLUE}Summary of changes:${NC}"
    git status --short

    # Count changes
    local added=$(git status --porcelain | grep -c "^A" || true)
    local modified=$(git status --porcelain | grep -c "^M" || true)
    local deleted=$(git status --porcelain | grep -c "^D" || true)
    local untracked=$(git status --porcelain | grep -c "^??" || true)

    echo -e "\n${GREEN}Files added:${NC} ${added}"
    echo -e "${YELLOW}Files modified:${NC} ${modified}"
    echo -e "${RED}Files deleted:${NC} ${deleted}"
    echo -e "${CYAN}Files untracked:${NC} ${untracked}"

    print_success "Git status check completed"
}

################################################################################
# Commit Message Functions
################################################################################

get_commit_message() {
    local round=$1
    local msg_file="${COMMIT_MSGS_DIR}/commit-${round}-*.txt"

    if [[ ! -f $(echo $msg_file) ]]; then
        print_error "Commit message file not found for round ${round}"
        exit 1
    fi

    cat $(echo $msg_file)
}

################################################################################
# File Staging Functions
################################################################################

stage_round1_files() {
    print_step "Staging Round 1 files (Code Cleanup and Features)..."

    # Deleted files (if they still show as deleted in git status)
    if git status --porcelain | grep -q "D.*internal/common/context/"; then
        git add -A internal/common/context/ 2>/dev/null || true
    fi

    # Modified files - import fixes
    git add internal/agent-manager/command/handler.go 2>/dev/null || true
    git add internal/cluster/handlers/cluster_handler.go 2>/dev/null || true
    git add internal/cluster/handlers/config_handler.go 2>/dev/null || true
    git add internal/cluster/handlers/kubeconfig_handler.go 2>/dev/null || true
    git add internal/cluster/handlers/metrics_handler.go 2>/dev/null || true
    git add internal/reasoning/handlers/analysis_handler.go 2>/dev/null || true
    git add internal/reasoning/handlers/chat_handler.go 2>/dev/null || true
    git add internal/reasoning/handlers/recommendation_handler.go 2>/dev/null || true

    # Deduplication
    git add internal/orchestrator/event/processor.go 2>/dev/null || true

    # New interface
    git add internal/agent-manager/event/interface.go 2>/dev/null || true
    git add internal/agent-manager/event/processor.go 2>/dev/null || true

    # Documentation
    git add docs/AGENT_MANAGER_CLEANUP_REPORT.md 2>/dev/null || true
    git add docs/code-quality/ROUND_1_OPTIMIZATION_REPORT.md 2>/dev/null || true

    print_success "Round 1 files staged"
}

stage_round2_files() {
    print_step "Staging Round 2 files (Structure Optimization and Tests)..."

    # Agent Manager handler splits
    git add internal/agent-manager/api/handler.go 2>/dev/null || true
    git add internal/agent-manager/api/agent_handlers.go 2>/dev/null || true
    git add internal/agent-manager/api/heartbeat_handlers.go 2>/dev/null || true
    git add internal/agent-manager/api/command_handlers.go 2>/dev/null || true
    git add internal/agent-manager/api/event_handlers.go 2>/dev/null || true
    git add internal/agent-manager/api/metrics_handlers.go 2>/dev/null || true

    # Cluster handler splits
    git add internal/cluster/handlers/cluster_handler.go 2>/dev/null || true
    git add internal/cluster/handlers/cluster_operations.go 2>/dev/null || true
    git add internal/cluster/handlers/cluster_advanced.go 2>/dev/null || true

    # Orchestrator handler splits
    git add internal/orchestrator/api/handler.go 2>/dev/null || true
    git add internal/orchestrator/api/workflow_handlers.go 2>/dev/null || true
    git add internal/orchestrator/api/strategy_handlers.go 2>/dev/null || true
    git add internal/orchestrator/api/execution_handlers.go 2>/dev/null || true

    # Reasoning handler splits
    git add internal/reasoning/handlers/analysis_handler.go 2>/dev/null || true
    git add internal/reasoning/handlers/prediction_handler.go 2>/dev/null || true
    git add internal/reasoning/handlers/recommendation_handler.go 2>/dev/null || true

    # Test files
    git add internal/auth/handlers/auth_handler_test.go 2>/dev/null || true
    git add internal/auth/jwt/manager_test.go 2>/dev/null || true
    git add internal/auth/storage/user_storage_test.go 2>/dev/null || true
    git add internal/cluster/handlers/cluster_handler_test.go 2>/dev/null || true
    git add internal/cluster/handlers/cluster_advanced_test.go 2>/dev/null || true
    git add internal/cluster/storage/cluster_storage_test.go 2>/dev/null || true
    git add internal/orchestrator/api/workflow_handlers_test.go 2>/dev/null || true
    git add internal/orchestrator/api/strategy_handlers_test.go 2>/dev/null || true
    git add internal/orchestrator/api/execution_handlers_test.go 2>/dev/null || true
    git add internal/reasoning/handlers/prediction_handler_test.go 2>/dev/null || true
    git add internal/reasoning/handlers/recommendation_handler_test.go 2>/dev/null || true
    git add internal/reasoning/engine/analyzer_test.go 2>/dev/null || true

    # Documentation
    git add docs/code-quality/ROUND_2_OPTIMIZATION_REPORT.md 2>/dev/null || true
    git add docs/code-quality/TEST_COVERAGE_REPORT.md 2>/dev/null || true
    git add docs/architecture/HANDLER_STRUCTURE.md 2>/dev/null || true

    print_success "Round 2 files staged"
}

stage_round3_files() {
    print_step "Staging Round 3 files (Decorator Pattern and Debt Cleanup)..."

    # Core decorator package
    git add common/decorator/ 2>/dev/null || true

    # Service decorator implementations
    git add internal/agent-manager/api/decorators/ 2>/dev/null || true
    git add internal/cluster/handlers/decorators/ 2>/dev/null || true
    git add internal/orchestrator/api/decorators/ 2>/dev/null || true
    git add internal/reasoning/handlers/decorators/ 2>/dev/null || true
    git add internal/auth/handlers/decorators/ 2>/dev/null || true

    # Modified handlers (decorator application)
    git add internal/agent-manager/api/agent_handlers.go 2>/dev/null || true
    git add internal/agent-manager/api/command_handlers.go 2>/dev/null || true
    git add internal/agent-manager/api/event_handlers.go 2>/dev/null || true
    git add internal/cluster/handlers/cluster_operations.go 2>/dev/null || true
    git add internal/cluster/handlers/cluster_advanced.go 2>/dev/null || true
    git add internal/orchestrator/api/workflow_handlers.go 2>/dev/null || true
    git add internal/orchestrator/api/execution_handlers.go 2>/dev/null || true
    git add internal/reasoning/handlers/analysis_handler.go 2>/dev/null || true
    git add internal/reasoning/handlers/prediction_handler.go 2>/dev/null || true
    git add internal/auth/handlers/auth_handler.go 2>/dev/null || true

    # TODO resolution files
    git add internal/agent-manager/event/processor.go 2>/dev/null || true
    git add internal/orchestrator/workflow/executor.go 2>/dev/null || true
    git add internal/reasoning/engine/analyzer.go 2>/dev/null || true
    git add internal/cluster/client/kubernetes.go 2>/dev/null || true
    git add internal/auth/jwt/manager.go 2>/dev/null || true
    git add internal/agent-manager/command/dispatcher.go 2>/dev/null || true
    git add internal/orchestrator/strategy/matcher.go 2>/dev/null || true
    git add internal/reasoning/knowledge/graph.go 2>/dev/null || true
    git add internal/agent-manager/storage/agent_storage.go 2>/dev/null || true
    git add internal/orchestrator/api/handler.go 2>/dev/null || true
    git add internal/reasoning/llm/client.go 2>/dev/null || true
    git add internal/cluster/storage/cluster_storage.go 2>/dev/null || true
    git add internal/auth/storage/session_storage.go 2>/dev/null || true

    # Config tests
    git add internal/agent-manager/config/config_test.go 2>/dev/null || true
    git add internal/orchestrator/config/config_test.go 2>/dev/null || true
    git add internal/reasoning/config/config_test.go 2>/dev/null || true
    git add internal/cluster/config/config_test.go 2>/dev/null || true
    git add internal/auth/config/config_test.go 2>/dev/null || true

    # Documentation
    git add docs/code-quality/ROUND_3_OPTIMIZATION_REPORT.md 2>/dev/null || true
    git add docs/architecture/DECORATOR_PATTERN.md 2>/dev/null || true
    git add docs/architecture/METRICS_GUIDE.md 2>/dev/null || true
    git add docs/TODO_RESOLUTION_REPORT.md 2>/dev/null || true

    print_success "Round 3 files staged"
}

stage_documentation_files() {
    print_step "Staging documentation files..."

    git add docs/commits/ 2>/dev/null || true
    git add docs/CHANGES_SUMMARY.md 2>/dev/null || true
    git add scripts/commit-changes.sh 2>/dev/null || true

    print_success "Documentation files staged"
}

################################################################################
# Commit Functions
################################################################################

create_commit() {
    local round=$1
    local message_file="${COMMIT_MSGS_DIR}/commit-${round}-*.txt"

    print_header "Creating Commit for Round ${round}"

    # Check if message file exists
    if [[ ! -f $(echo $message_file) ]]; then
        print_error "Commit message file not found for round ${round}"
        return 1
    fi

    # Stage files for this round
    case $round in
        1)
            stage_round1_files
            ;;
        2)
            stage_round2_files
            ;;
        3)
            stage_round3_files
            ;;
        *)
            print_error "Invalid round number: ${round}"
            return 1
            ;;
    esac

    # Show what will be committed
    echo -e "\n${BLUE}Files to be committed:${NC}"
    git diff --cached --name-only --diff-filter=ACMR

    # Get commit message
    local commit_msg=$(cat $(echo $message_file))

    # Preview commit message
    echo -e "\n${YELLOW}Commit message preview (first 10 lines):${NC}"
    echo "$commit_msg" | head -n 10
    echo "..."

    if [[ "$DRY_RUN" == "true" ]]; then
        print_warning "DRY RUN: Commit would be created but not executed"
        git reset HEAD -- . >/dev/null 2>&1 || true
        return 0
    fi

    # Interactive confirmation
    if [[ "$COMMIT_ALL" != "true" ]]; then
        echo -e "\n${YELLOW}Create commit for Round ${round}? (y/n)${NC} "
        read -r response
        if [[ ! "$response" =~ ^[Yy]$ ]]; then
            print_warning "Skipping Round ${round}"
            git reset HEAD -- . >/dev/null 2>&1 || true
            return 0
        fi
    fi

    # Create commit
    if git commit -F <(echo "$commit_msg"); then
        print_success "Commit created for Round ${round}"

        # Show commit hash
        local commit_hash=$(git rev-parse --short HEAD)
        print_info "Commit hash: ${commit_hash}"
    else
        print_error "Failed to create commit for Round ${round}"
        return 1
    fi
}

################################################################################
# Main Workflow
################################################################################

show_usage() {
    cat << EOF
${CYAN}Git Commit Automation Script for Optimization Rounds 1-3${NC}

${YELLOW}Usage:${NC}
  $0 [OPTIONS]

${YELLOW}Options:${NC}
  --all                 Commit all rounds automatically (no prompts)
  --round <1|2|3>      Commit specific round only
  --dry-run            Preview without creating commits
  -h, --help           Show this help message

${YELLOW}Examples:${NC}
  $0                   # Interactive mode (prompts for each round)
  $0 --all            # Commit all rounds automatically
  $0 --round 2        # Commit only Round 2
  $0 --dry-run        # Preview what would be committed

${YELLOW}Commit Strategy:${NC}
  Round 1: Code cleanup and feature implementation
  Round 2: Structure optimization and test coverage
  Round 3: Decorator pattern and technical debt cleanup

${YELLOW}Files:${NC}
  Commit messages: docs/commits/commit-{1,2,3}-*.txt
  Changes summary: docs/CHANGES_SUMMARY.md
  This script:     scripts/commit-changes.sh

EOF
}

main() {
    print_header "Git Commit Automation for Optimization Rounds 1-3"

    # Change to project root
    cd "$PROJECT_ROOT" || exit 1

    # Check git status
    check_git_status

    # Determine which commits to create
    if [[ -n "$SPECIFIC_ROUND" ]]; then
        # Commit specific round
        create_commit "$SPECIFIC_ROUND"
    else
        # Commit all rounds (or interactive)
        if [[ "$COMMIT_ALL" == "true" ]] || [[ "$DRY_RUN" == "true" ]]; then
            create_commit 1
            create_commit 2
            create_commit 3
        else
            # Interactive mode
            echo -e "\n${YELLOW}Select commits to create:${NC}"
            echo "  1) Round 1 only"
            echo "  2) Round 2 only"
            echo "  3) Round 3 only"
            echo "  4) All rounds (1, 2, 3)"
            echo "  5) Custom selection"
            echo "  6) Exit"
            echo -e "\n${YELLOW}Enter choice [1-6]:${NC} "
            read -r choice

            case $choice in
                1)
                    create_commit 1
                    ;;
                2)
                    create_commit 2
                    ;;
                3)
                    create_commit 3
                    ;;
                4)
                    create_commit 1
                    create_commit 2
                    create_commit 3
                    ;;
                5)
                    echo -e "\n${YELLOW}Commit Round 1? (y/n):${NC} "
                    read -r r1
                    [[ "$r1" =~ ^[Yy]$ ]] && create_commit 1

                    echo -e "\n${YELLOW}Commit Round 2? (y/n):${NC} "
                    read -r r2
                    [[ "$r2" =~ ^[Yy]$ ]] && create_commit 2

                    echo -e "\n${YELLOW}Commit Round 3? (y/n):${NC} "
                    read -r r3
                    [[ "$r3" =~ ^[Yy]$ ]] && create_commit 3
                    ;;
                6)
                    print_info "Exiting without creating commits"
                    exit 0
                    ;;
                *)
                    print_error "Invalid choice"
                    exit 1
                    ;;
            esac
        fi
    fi

    # Final summary
    print_header "Commit Summary"

    # Count commits created
    local commits_created=$(git log --oneline --since="1 minute ago" | wc -l)

    if [[ $commits_created -gt 0 ]]; then
        print_success "Successfully created ${commits_created} commit(s)"

        echo -e "\n${BLUE}Recent commits:${NC}"
        git log --oneline --graph --decorate -n 5

        echo -e "\n${GREEN}Next steps:${NC}"
        echo "  1. Review commits: git log --patch"
        echo "  2. Push to remote: git push origin $(git branch --show-current)"
        echo "  3. Create pull request if needed"
    else
        if [[ "$DRY_RUN" == "true" ]]; then
            print_info "Dry run completed - no commits created"
        else
            print_warning "No commits were created"
        fi
    fi
}

################################################################################
# Parse Arguments
################################################################################

parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --all)
                COMMIT_ALL=true
                shift
                ;;
            --round)
                if [[ -z "$2" ]] || [[ ! "$2" =~ ^[1-3]$ ]]; then
                    print_error "Invalid round number. Must be 1, 2, or 3"
                    exit 1
                fi
                SPECIFIC_ROUND="$2"
                shift 2
                ;;
            --dry-run)
                DRY_RUN=true
                shift
                ;;
            -h|--help)
                show_usage
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                show_usage
                exit 1
                ;;
        esac
    done
}

################################################################################
# Script Entry Point
################################################################################

parse_args "$@"
main

exit 0
