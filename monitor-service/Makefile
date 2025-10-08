.PHONY: build run clean test docker-build

SERVICE_NAME=monitor-service
BINARY_NAME=bin/$(SERVICE_NAME)
DOCKER_IMAGE=k8s-agent/$(SERVICE_NAME):latest

build:
	@echo "Building $(SERVICE_NAME)..."
	@mkdir -p bin
	@go build -o $(BINARY_NAME) cmd/server/main.go

run:
	@echo "Running $(SERVICE_NAME)..."
	@go run cmd/server/main.go

clean:
	@echo "Cleaning..."
	@rm -rf bin

test:
	@echo "Running tests..."
	@go test -v ./...

docker-build:
	@echo "Building Docker image..."
	@docker build -t $(DOCKER_IMAGE) .

.DEFAULT_GOAL := build
