# Common variables and functions for Makefile
# Based on OneX and go-protoc project structure

# It's necessary to set this because some environments don't link sh -> bash.
SHELL := /usr/bin/env bash -o errexit -o pipefail +o nounset
.SHELLFLAGS = -ec

# It's necessary to set the errexit flags for the bash shell.
export SHELLOPTS := errexit

# Makefile settings
# We don't need make's built-in rules.
MAKEFLAGS += --no-builtin-rules
ifeq ($(V),1)
  $(warning ***** starting Makefile for goal(s) "$(MAKECMDGOALS)")
  $(warning ***** $(shell date))
else
  # If we're not debugging the Makefile, don't echo recipes.
  MAKEFLAGS += -s --no-print-directory
endif

# Project information
PROJECT_NAME ?= aetherius
ORG := kart-io
REPO := k8s-agent

# Go version requirements
GO := go
GO_MINIMUM_VERSION ?= 1.21

# Version information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "v0.0.0-dev")
GIT_COMMIT := $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
GIT_BRANCH := $(shell git branch --show-current 2>/dev/null || echo "unknown")
GIT_TREE_STATE := "dirty"
ifeq (, $(shell git status --porcelain 2>/dev/null))
    GIT_TREE_STATE := "clean"
endif
BUILD_TIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
GO_VERSION := $(shell go version | awk '{print $$3}')

# Platform information
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
PLATFORM := $(GOOS)/$(GOARCH)

# Directory paths
ROOT_DIR := $(shell pwd)
OUTPUT_DIR := $(ROOT_DIR)/_output
BIN_DIR := $(OUTPUT_DIR)/bin
COVERAGE_DIR := $(OUTPUT_DIR)/coverage
TOOLS_DIR := $(ROOT_DIR)/tools

# Go build flags
# Use external version package: github.com/kart-io/version
LDFLAGS := -s -w \
	-X 'github.com/kart-io/version.gitVersion=$(VERSION)' \
	-X 'github.com/kart-io/version.gitCommit=$(GIT_COMMIT)' \
	-X 'github.com/kart-io/version.gitBranch=$(GIT_BRANCH)' \
	-X 'github.com/kart-io/version.gitTreeState=$(GIT_TREE_STATE)' \
	-X 'github.com/kart-io/version.buildDate=$(BUILD_TIME)' \
	-X 'github.com/kart-io/version.goVersion=$(GO_VERSION)' \
	-X 'github.com/kart-io/version.platform=$(PLATFORM)'

# Go build options
GO_BUILD := $(GO) build
GO_TEST := $(GO) test
GO_VET := $(GO) vet
GO_FMT := gofmt
GO_LINT := golangci-lint
GO_MOD := $(GO) mod

# Build tags
BUILD_TAGS ?=

# Test options
TEST_FLAGS ?= -race -cover
TEST_TIMEOUT ?= 5m

# Docker options
DOCKER := docker
DOCKER_REGISTRY ?= docker.io
DOCKER_NAMESPACE ?= $(ORG)
IMAGE_PREFIX ?= $(DOCKER_REGISTRY)/$(DOCKER_NAMESPACE)

# Linux command settings
FIND := find . ! -path './third_party/*' ! -path './vendor/*' ! -path './_output/*'
XARGS := xargs --no-run-if-empty

# Color output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[0;33m
BLUE := \033[0;34m
CYAN := \033[0;36m
NC := \033[0m # No Color

# Functions
define print_info
	printf "$(GREEN)$(1)$(NC)\n"
endef

define print_error
	printf "$(RED)$(1)$(NC)\n"
endef

define print_warning
	printf "$(YELLOW)$(1)$(NC)\n"
endef

define print_target
	printf "$(BLUE)==> $(1)$(NC)\n"
endef

# Ensure output directories exist
$(shell mkdir -p $(BIN_DIR) $(COVERAGE_DIR))

# Note: help target is defined in root Makefile
