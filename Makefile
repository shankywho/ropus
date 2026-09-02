# ==============================================================================
# ROPUS — AI Risk Manager (Razorpay AI Buildathon, Track 02)
# Canonical Developer & Verification Interface
# ==============================================================================

.PHONY: all up down restart test test-backend test-ml test-frontend lint build evaluate demo-webhook docs-check help

all: test

help:
	@echo "ROPUS Buildathon Commands:"
	@echo "  make up              - Start Docker Compose local stack"
	@echo "  make down            - Stop Docker Compose local stack"
	@echo "  make test            - Run all backend, ML, and frontend tests"
	@echo "  make test-backend    - Run Go risk engine test suites"
	@echo "  make test-ml         - Run Python ML & calibration pytest suites"
	@echo "  make lint            - Run frontend TypeScript/ESLint checks"
	@echo "  make build           - Build Go backend and frontend bundles"
	@echo "  make evaluate        - Reproduce canonical frozen holdout metrics"
	@echo "  make demo-webhook    - Run signed Razorpay webhook demonstration"
	@echo "  make docs-check      - Validate documentation files and structure"

up:
	docker compose up -d

down:
	docker compose down

restart: down up

test: test-backend test-ml lint

test-backend:
	@echo "==> Running Go backend test suites..."
	cd backend && go test -race ./...

test-ml:
	@echo "==> Running Python ML & calibration tests..."
	pytest ml-service/tests

lint:
	@echo "==> Running frontend lint..."
	cd frontend && npm run lint

build:
	@echo "==> Building Go binary..."
	cd backend && go build -o /dev/null ./cmd/api
	@echo "==> Building frontend production bundle..."
	cd frontend && npm run build

evaluate:
	@echo "==> Evaluating Champion & Candidate models on frozen 1,200-row holdout set..."
	python3 ml-service/evaluation/evaluate_frozen_holdout.py

demo-webhook:
	@echo "==> Executing Buildathon signed Razorpay webhook demonstration..."
	./scripts/demo_buildathon.sh

docs-check:
	@echo "==> Validating documentation files and Markdown link integrity..."
	python3 scripts/validate_docs.py
