#!/usr/bin/env bash

# Version management script for k8s-agent
# Based on OneX project patterns

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

# Source common utilities
# shellcheck source=./lib/common.sh
source "$SCRIPT_DIR/lib/common.sh"

# Version file
VERSION_FILE="$ROOT_DIR/VERSION"

# Usage information
usage() {
    cat <<EOF
Version Management Script

Usage:
    $0 <command> [options]

Commands:
    get                     Get current version
    bump <major|minor|patch>  Bump version
    set <version>           Set specific version
    tag                     Create git tag for current version
    validate <version>      Validate version format

Examples:
    $0 get
    $0 bump patch
    $0 set v1.2.3
    $0 tag
    $0 validate v1.2.3

EOF
}

# Get current version
get_version() {
    if [[ -f "$VERSION_FILE" ]]; then
        cat "$VERSION_FILE"
    else
        echo "v0.0.0"
    fi
}

# Parse version string
parse_version() {
    local version="$1"

    # Remove 'v' prefix if present
    version="${version#v}"

    # Extract major, minor, patch
    local major minor patch
    major=$(echo "$version" | cut -d. -f1)
    minor=$(echo "$version" | cut -d. -f2)
    patch=$(echo "$version" | cut -d. -f3 | cut -d- -f1)

    echo "$major $minor $patch"
}

# Validate version format
validate_version() {
    local version="$1"

    # Version should match semver pattern
    if [[ ! "$version" =~ ^v?[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9.-]+)?(\+[a-zA-Z0-9.-]+)?$ ]]; then
        log::error "Invalid version format: $version"
        log::info "Expected format: v<major>.<minor>.<patch>[-prerelease][+build]"
        log::info "Examples: v1.0.0, v1.2.3-alpha.1, v2.0.0+20231201"
        return 1
    fi

    log::success "Version format is valid: $version"
    return 0
}

# Bump version
bump_version() {
    local bump_type="$1"
    local current_version
    current_version=$(get_version)

    log::info "Current version: $current_version"

    # Parse current version
    read -r major minor patch <<< "$(parse_version "$current_version")"

    # Bump version
    case "$bump_type" in
        major)
            major=$((major + 1))
            minor=0
            patch=0
            ;;
        minor)
            minor=$((minor + 1))
            patch=0
            ;;
        patch)
            patch=$((patch + 1))
            ;;
        *)
            log::error "Invalid bump type: $bump_type"
            log::info "Valid types: major, minor, patch"
            return 1
            ;;
    esac

    local new_version="v${major}.${minor}.${patch}"

    log::info "New version: $new_version"

    # Confirm
    if ! util::confirm "Set version to $new_version?" "y"; then
        log::warning "Version bump cancelled"
        return 1
    fi

    # Write new version
    echo "$new_version" > "$VERSION_FILE"
    log::success "Version bumped to $new_version"

    # Update changelog
    update_changelog "$new_version"
}

# Set specific version
set_version() {
    local version="$1"

    # Validate format
    if ! validate_version "$version"; then
        return 1
    fi

    # Ensure 'v' prefix
    if [[ ! "$version" =~ ^v ]]; then
        version="v${version}"
    fi

    log::info "Setting version to: $version"

    # Confirm
    if ! util::confirm "Set version to $version?" "y"; then
        log::warning "Version change cancelled"
        return 1
    fi

    # Write version
    echo "$version" > "$VERSION_FILE"
    log::success "Version set to $version"

    # Update changelog
    update_changelog "$version"
}

# Create git tag
create_tag() {
    local version
    version=$(get_version)

    log::section "Creating Git Tag"

    log::info "Version: $version"

    # Check if tag already exists
    if git rev-parse "$version" >/dev/null 2>&1; then
        log::error "Tag $version already exists"
        return 1
    fi

    # Check for uncommitted changes
    if [[ -n $(git status --porcelain) ]]; then
        log::error "You have uncommitted changes"
        log::info "Please commit or stash your changes first"
        return 1
    fi

    # Confirm
    if ! util::confirm "Create tag $version?" "y"; then
        log::warning "Tag creation cancelled"
        return 1
    fi

    # Create annotated tag
    git tag -a "$version" -m "Release $version"
    log::success "Created tag: $version"

    log::info ""
    log::info "To push the tag to remote:"
    log::info "  git push origin $version"
    log::info ""
    log::info "Or push all tags:"
    log::info "  git push --tags"
}

# Update CHANGELOG.md
update_changelog() {
    local version="$1"
    local changelog="$ROOT_DIR/CHANGELOG.md"
    local date
    date=$(date +%Y-%m-%d)

    if [[ ! -f "$changelog" ]]; then
        log::warning "CHANGELOG.md not found, skipping"
        return 0
    fi

    log::info "Updating CHANGELOG.md"

    # Create temporary file
    local temp_file
    temp_file=$(mktemp)

    # Add new version section
    {
        echo "# Changelog"
        echo ""
        echo "## [$version] - $date"
        echo ""
        echo "### Added"
        echo "- TODO: Add new features"
        echo ""
        echo "### Changed"
        echo "- TODO: Add changes"
        echo ""
        echo "### Fixed"
        echo "- TODO: Add bug fixes"
        echo ""

        # Append existing content (skip first line if it's "# Changelog")
        tail -n +2 "$changelog"
    } > "$temp_file"

    mv "$temp_file" "$changelog"
    log::success "Updated CHANGELOG.md"
    log::warning "Please edit CHANGELOG.md to add release notes"
}

# Main function
main() {
    if [[ $# -eq 0 ]]; then
        usage
        exit 0
    fi

    local command="$1"
    shift

    case "$command" in
        get)
            get_version
            ;;
        bump)
            if [[ $# -eq 0 ]]; then
                log::error "Bump type required"
                usage
                exit 1
            fi
            bump_version "$1"
            ;;
        set)
            if [[ $# -eq 0 ]]; then
                log::error "Version required"
                usage
                exit 1
            fi
            set_version "$1"
            ;;
        tag)
            create_tag
            ;;
        validate)
            if [[ $# -eq 0 ]]; then
                log::error "Version required"
                usage
                exit 1
            fi
            validate_version "$1"
            ;;
        help|--help|-h)
            usage
            ;;
        *)
            log::error "Unknown command: $command"
            usage
            exit 1
            ;;
    esac
}

main "$@"
