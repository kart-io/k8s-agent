# ==============================================================================
# This file includes all other makefiles to centralize rule management.
# All rules are organized by category for easy management and maintenance.
# ==============================================================================

# Common variables and version constants
include scripts/make-rules/common.mk
include scripts/make-rules/common-versions.mk

# Core build and development rules
include scripts/make-rules/golang.mk
include scripts/make-rules/docker.mk
include scripts/make-rules/image.mk
include scripts/make-rules/proto.mk

# Tools and utilities
include scripts/make-rules/tools.mk
include scripts/make-rules/hooks.mk

# Kubernetes and deployment
include scripts/make-rules/k8s.mk
include scripts/make-rules/deploy.mk

# Version and release management
include scripts/make-rules/version.mk
include scripts/make-rules/release.mk

# Code generation and quality
include scripts/make-rules/gen.mk
include scripts/make-rules/lint.mk
include scripts/make-rules/copyright.mk

# API documentation
include scripts/make-rules/swagger.mk
