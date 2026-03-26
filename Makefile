SHELL := /bin/sh

COMPOSE_FILE := deploy/compose/docker-compose.yml
COMPOSE_OVERRIDE := deploy/compose/docker-compose.override.yml
ENV_FILE := .env
COMPOSE := docker compose --env-file $(ENV_FILE) -f $(COMPOSE_FILE) -f $(COMPOSE_OVERRIDE)
GO_ENV := GOCACHE=$(CURDIR)/.cache/go-build GOMODCACHE=$(CURDIR)/.cache/go-mod
COVERAGE_FILE := coverage.out
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -X 'clawbot-server/internal/version.Value=$(VERSION)' \
           -X 'clawbot-server/internal/version.Commit=$(COMMIT)' \
           -X 'clawbot-server/internal/version.BuildDate=$(BUILD_DATE)'

.PHONY: help check-env up down restart ps logs smoke clean lint test coverage coverage-html security compose-validate build run-server migrate-up migrate-down

help: ## Show available targets.
	@awk 'BEGIN {FS = ": ## "}; /^[a-zA-Z0-9_.-]+: ## / {printf "  %-18s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

check-env:
	@./scripts/check-env.sh

up: check-env ## Build and start the foundation stack.
	$(COMPOSE) up -d --build

down: check-env ## Stop the foundation stack.
	$(COMPOSE) down --remove-orphans

restart: down up ## Restart the foundation stack.

ps: check-env ## Show foundation stack containers.
	$(COMPOSE) ps

logs: check-env ## Tail stack logs.
	$(COMPOSE) logs -f --tail=100

smoke: ## Wait for the stack and run the Go smoke checks.
	./scripts/wait-for-stack.sh
	./scripts/smoke.sh

build: ## Build the clawbot-server binary with version metadata.
	@mkdir -p .cache/go-build .cache/go-mod bin
	$(GO_ENV) go build -ldflags "$(LDFLAGS)" -o bin/clawbot-server ./cmd/clawbot-server

run-server: check-env ## Run the control-plane service locally.
	@mkdir -p .cache/go-build .cache/go-mod
	@set -a; . ./.env; set +a; $(GO_ENV) go run -ldflags "$(LDFLAGS)" ./cmd/clawbot-server serve

migrate-up: check-env ## Apply embedded database migrations.
	@mkdir -p .cache/go-build .cache/go-mod
	@set -a; . ./.env; set +a; $(GO_ENV) go run -ldflags "$(LDFLAGS)" ./cmd/clawbot-server migrate up

migrate-down: check-env ## Roll back the latest embedded database migration.
	@mkdir -p .cache/go-build .cache/go-mod
	@set -a; . ./.env; set +a; $(GO_ENV) go run -ldflags "$(LDFLAGS)" ./cmd/clawbot-server migrate down

clean: check-env ## Remove the foundation stack and named volumes.
	$(COMPOSE) down -v --remove-orphans

lint: ## Run local formatting and Go lint checks.
	@mkdir -p .cache/go-build .cache/go-mod
	@fmt_out=$$(find cmd internal -name '*.go' -print | xargs gofmt -l); \
	if [ -n "$$fmt_out" ]; then \
		echo "$$fmt_out"; \
		echo "gofmt reported unformatted files"; \
		exit 1; \
	fi
	$(GO_ENV) go vet ./...
	@if command -v golangci-lint >/dev/null 2>&1; then golangci-lint run ./...; else echo "golangci-lint not installed; skipping"; fi

test: ## Run deterministic Go unit tests.
	@mkdir -p .cache/go-build .cache/go-mod
	$(GO_ENV) go test ./...

coverage: ## Run Go tests with a coverage profile and summary.
	@mkdir -p .cache/go-build .cache/go-mod
	$(GO_ENV) go test -covermode=atomic -coverprofile=$(COVERAGE_FILE) ./...
	go tool cover -func=$(COVERAGE_FILE)

coverage-html: coverage ## Render an HTML coverage report at coverage.html.
	go tool cover -html=$(COVERAGE_FILE) -o coverage.html
	@echo "wrote coverage.html"

security: ## Run local security checks when the tools are installed.
	@if command -v gosec >/dev/null 2>&1; then gosec ./...; else echo "gosec not installed; skipping"; fi
	@if command -v govulncheck >/dev/null 2>&1; then govulncheck ./...; else echo "govulncheck not installed; skipping"; fi
	@if command -v gitleaks >/dev/null 2>&1; then gitleaks detect --no-banner --redact; else echo "gitleaks not installed; skipping"; fi
	@if command -v trivy >/dev/null 2>&1; then trivy fs --exit-code 1 --severity HIGH,CRITICAL .; else echo "trivy not installed; skipping"; fi

compose-validate: check-env ## Validate the rendered Docker Compose configuration.
	$(COMPOSE) config >/dev/null