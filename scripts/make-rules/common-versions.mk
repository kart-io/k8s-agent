# ==============================================================================
# Versions used by all Makefiles
# This file defines version constants for all development tools
#

# Kubernetes tools
KUSTOMIZE_VERSION ?= v5.3.0
CONTROLLER_TOOLS_VERSION ?= v0.14.0
KIND_VERSION ?= v0.20.0
HELM_VERSION ?= v3.15.1
KUBE_LINTER_VERSION ?= 0.6.8
KUBE_CONFORM_VERSION ?= v0.6.4

# Docker and container tools
HADOLINT_VER ?= v2.12.0

# Go development tools
GOLANGCI_LINT_VERSION ?= v1.55.2
GO_FUMPT_VERSION ?= v0.5.0
GO_IMPORTS_VERSION ?= v0.19.0
GO_MODIFY_TAGS_VERSION ?= v1.16.0
MOCKGEN_VERSION ?= v1.6.0
WIRE_VERSION ?= v0.6.0

# Code generation tools
PROTOC_GEN_GO_VERSION ?= v1.31.0
PROTOC_GEN_GO_GRPC_VERSION ?= v1.3.0
PROTOC_GEN_GRPC_GATEWAY_VERSION ?= v2.18.0
PROTOC_GO_INJECT_TAG_VERSION ?= v1.4.0
BUF_VERSION ?= v1.28.1

# API documentation
GO_SWAGGER_VERSION ?= v0.31.0
SWAGGER_UI_VERSION ?= v5.9.0

# Testing tools
GO_JUNIT_REPORT_VERSION ?= v2.1.0
GO_TESTS_VERSION ?= v1.6.0

# Release and changelog
GIT_CHGLOG_VERSION ?= v0.15.4
GSEMVER_VERSION ?= v0.9.1
UPLIFT_VERSION ?= latest

# License and copyright
ADDLICENSE_VERSION ?= v1.1.1
LICENSE_VERSION ?= v5.0.4

# Development utilities
AIR_VERSION ?= v1.49.0
YQ_VERSION ?= v4.40.5
GRPCURL_VERSION ?= v1.8.9

# Database tools
DB_TO_STRUCT_VERSION ?= v1.0.2

# Code quality
LOGCHECK_VERSION ?= v0.8.1
GO_APIDIFF_VERSION ?= v0.8.2

# Helper function to get version from go.mod
define get_go_version
$(shell go list -m -f '{{.Version}}' $(1) 2>/dev/null || echo "latest")
endef

# Dynamically get versions from go.mod if available
GRPC_GATEWAY_VERSION ?= $(call get_go_version,github.com/grpc-ecosystem/grpc-gateway/v2)
PROTOC_GEN_VALIDATE_VERSION ?= $(call get_go_version,github.com/envoyproxy/protoc-gen-validate)
