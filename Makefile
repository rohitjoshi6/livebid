.PHONY: help up down test

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "%-16s %s\n", $$1, $$2}'

up: ## Start local infrastructure
	docker compose up -d

down: ## Stop local infrastructure
	docker compose down

test: ## Run all tests
	@echo "Tests will be added as backend and frontend applications are implemented."

