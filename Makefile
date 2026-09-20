.PHONY: help up down test backend-test frontend-build

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "%-16s %s\n", $$1, $$2}'

up: ## Start local infrastructure
	docker compose up -d

down: ## Stop local infrastructure
	docker compose down

backend-test: ## Run backend tests in a Go container
	docker run --rm -v "$$(pwd)/backend:/src" -w /src golang:1.22-alpine go test ./...

frontend-build: ## Build the frontend
	cd frontend && npm run build

test: backend-test frontend-build ## Run backend tests and frontend build
