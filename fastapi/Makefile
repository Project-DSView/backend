SHELL := /bin/sh

PY := python
PIP := pip
UVICORN := $(PY) -m uvicorn

APP := app.main:app
HOST := 127.0.0.1
PORT := 8000

.PHONY: help install dev run lint test test-unit test-security test-integration test-coverage test-quick docker-build docker-run compose-up compose-down redis up down

help:
	@echo "Available targets:"
	@echo "  install       Install Python dependencies"
	@echo "  dev           Run FastAPI with reload"
	@echo "  run           Run FastAPI"
	@echo "  lint          Basic syntax check"
	@echo "  test          Run all tests"
	@echo "  test-unit     Run unit tests"
	@echo "  test-security Run security tests"
	@echo "  test-integration Run integration tests"
	@echo "  test-coverage Run tests with coverage report"
	@echo "  test-quick    Run quick health check tests"
	@echo "  docker-build  Build Docker image"
	@echo "  docker-run    Run Docker container"

install:
	$(PIP) install -r requirements.txt

dev:
	$(UVICORN) $(APP) --host $(HOST) --port $(PORT) --reload

run:
	$(UVICORN) $(APP) --host $(HOST) --port $(PORT)

lint:
	$(PY) -m py_compile $$(git ls-files '*.py') || true

test:
	$(PY) run_tests.py

test-unit:
	$(PY) -m pytest tests/test_health.py tests/test_playground.py -v

test-security:
	$(PY) -m pytest tests/test_security.py -v

test-integration:
	$(PY) -m pytest tests/test_integration.py -v

test-coverage:
	$(PY) -m pytest tests/ -v --cov=app --cov-report=html --cov-report=term-missing

test-quick:
	$(PY) -m pytest tests/test_health.py -v

docker-build:
	docker build -t dsview-backend:latest -f dockerfile .

docker-run:
	docker run --rm -p $(PORT):$(PORT) --env-file .env.local dsview-backend:latest


