.PHONY: help install run dev test clean docker-build docker-run

help:
	@echo "Aetherius Reasoning Service - Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  install       - Install Python dependencies"
	@echo "  run           - Run the service"
	@echo "  dev           - Run in development mode with auto-reload"
	@echo "  test          - Run tests"
	@echo "  lint          - Run linting"
	@echo "  format        - Format code"
	@echo "  clean         - Clean temporary files"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-run    - Run in Docker container"

install:
	pip install -r requirements.txt

run:
	python cmd/server/main.py --config=configs/config.yaml

dev:
	uvicorn internal.api.server:app --reload --host 0.0.0.0 --port 8082

test:
	pytest tests/ -v --cov=internal --cov-report=html

lint:
	pylint internal/ pkg/ cmd/

format:
	black internal/ pkg/ cmd/
	isort internal/ pkg/ cmd/

clean:
	find . -type d -name "__pycache__" -exec rm -rf {} + 2>/dev/null || true
	find . -type f -name "*.pyc" -delete
	find . -type d -name "*.egg-info" -exec rm -rf {} + 2>/dev/null || true
	rm -rf .pytest_cache htmlcov .coverage

docker-build:
	docker build -t aetherius-reasoning-service:latest .

docker-run:
	docker run -d \
		--name reasoning-service \
		-p 8082:8082 \
		-e NEO4J_URI=bolt://neo4j:7687 \
		-e NEO4J_USER=neo4j \
		-e NEO4J_PASSWORD=password \
		aetherius-reasoning-service:latest