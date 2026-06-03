
include ./Makefile.Common

SHELL := /bin/sh

ENV ?= local
override ENV := $(shell echo $(ENV) | tr A-Z a-z)

ENV_FILE ?= ./.env.$(ENV)
MIGRATE_BIN ?= migrate
MIGRATE_DIR ?= migrations

DOCKER_REGISTRY_HOST?=local

# -------------------------------------------------------------------
# Tool versions — keep in sync with .gitlab-ci.yml variables block
# -------------------------------------------------------------------
GOLANGCI_LINT_VERSION ?= 2.12.2
GOSEC_VERSION         ?= v2.26.1
GOVULNCHECK_VERSION   ?= v1.3.0
GOIMPORTS_VERSION     ?= v0.45.0

# -------------------------------------------------------------------
# Load env safely into shell
# -------------------------------------------------------------------
define load_env
if [ ! -f "$(ENV_FILE)" ]; then \
  echo "ERROR: env file '$(ENV_FILE)' not found"; exit 1; \
fi; \
set -a; . "$(ENV_FILE)"; set +a; \
if [ -z "$${DB_HOST}" ] && [ -n "$${DATABASE_URL}" ]; then \
  DB_USER=$$(echo "$${DATABASE_URL}" | sed 's|postgres://\([^:]*\):.*|\1|'); \
  DB_PASS=$$(echo "$${DATABASE_URL}" | sed 's|postgres://[^:]*:\([^@]*\)@.*|\1|'); \
  DB_HOST=$$(echo "$${DATABASE_URL}" | sed 's|.*@\([^:]*\):.*|\1|'); \
  DB_PORT=$$(echo "$${DATABASE_URL}" | sed 's|.*:\([0-9]*\)/.*|\1|'); \
  DB_NAME=$$(echo "$${DATABASE_URL}" | sed 's|.*/\([^?]*\).*|\1|'); \
  DB_SCHEMA=$$(echo "$${DATABASE_URL}" | sed 's|.*search_path=\([^&]*\).*|\1|'); \
fi; \
DATABASE_URL="postgres://$${DB_USER}:$${DB_PASS}@$${DB_HOST}:$${DB_PORT}/$${DB_NAME}?sslmode=$${DB_SSLMODE:-disable}&search_path=$${DB_SCHEMA}";
endef

# -------------------------------------------------------------------
# Help
# -------------------------------------------------------------------
.PHONY: help
help:
	@echo ""
	@echo "Usage: make <target> [OPTIONS]"
	@echo ""
	@echo "Run:"
	@echo "  goserver                    Run HTTP server"
	@echo "  goworker                    Run worker"
	@echo ""
	@echo "Test:"
	@echo "  test                        Run all tests"
	@echo "  test-race                   Run tests with race detector"
	@echo "  test-cover                  Run tests with coverage report"
	@echo "  test-failed                 Show only failed tests"
	@echo ""
	@echo "Build:"
	@echo "  build-server                Build HTTP server binary"
	@echo "  build-worker                Build worker binary"
	@echo ""
	@echo "Lint & Format:"
	@echo "  lint                        Run golangci-lint"
	@echo "  fmt                         Format code with gofmt and goimports"
	@echo ""
	@echo "Docs:"
	@echo "  gen-docs                    Generate Swagger documentation"
	@echo ""
	@echo "Tidy:"
	@echo "  tidy                        Run go mod tidy"
	@echo ""
	@echo "Migration:"
	@echo "  migrate-create NAME=<name>  Create a new migration"
	@echo "  migrate-up                  Apply all pending migrations"
	@echo "  migrate-down                Roll back the last migration"
	@echo "  migrate-down-all            Roll back all migrations"
	@echo "  migrate-version             Show current migration version"
	@echo "  migrate-force VERSION=<v>   Force migration to a specific version"
	@echo ""
	@echo "Config:"
	@echo "  ENV          Environment name (default: local) → loads .env.<ENV>"
	@echo "  ENV_FILE     Override env file path directly (e.g. ENV_FILE=.env.ci)"
	@echo "  MIGRATE_BIN  Migration binary (default: migrate)"
	@echo "  MIGRATE_DIR  Migrations directory (default: migrations)"
	@echo ""
	@echo "Examples:"
	@echo "  make migrate-up                  # uses .env.local"
	@echo "  make migrate-up ENV=dev          # uses .env.dev"
	@echo "  make migrate-up ENV=staging      # uses .env.staging"
	@echo "  make migrate-create NAME=create_users_table"
	@echo ""

# -------------------------------------------------------------------
# Check tools version match
# -------------------------------------------------------------------
.PHONY: check-tools
check-tools:
	@echo "Checking tool versions..."

	@INSTALLED=$$(golangci-lint --version 2>&1 | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1); \
	EXPECTED=$(GOLANGCI_LINT_VERSION); \
	if [ "$$INSTALLED" != "$$EXPECTED" ]; then \
		echo "WARN: golangci-lint mismatch — expected $$EXPECTED, got $$INSTALLED"; \
	else \
		echo "✓ golangci-lint $$INSTALLED"; \
	fi

	@INSTALLED=$$(govulncheck -version 2>&1 | grep "Scanner:" | grep -oE 'v[0-9]+\.[0-9]+\.[0-9]+'); \
	EXPECTED=$(GOVULNCHECK_VERSION); \
	if [ "$$INSTALLED" != "$$EXPECTED" ]; then \
		echo "WARN: govulncheck mismatch — expected $$EXPECTED, got $$INSTALLED"; \
	else \
		echo "✓ govulncheck $$INSTALLED"; \
	fi

	@INSTALLED=$$(gosec -version 2>&1 | grep "Version:" | grep -oE 'v[0-9]+\.[0-9]+\.[0-9]+'); \
	EXPECTED=$(GOSEC_VERSION); \
	if [ -z "$$INSTALLED" ]; then \
		echo "WARN: gosec version undetectable — likely installed from source (dev build), expected $$EXPECTED"; \
	elif [ "$$INSTALLED" != "$$EXPECTED" ]; then \
		echo "WARN: gosec mismatch — expected $$EXPECTED, got $$INSTALLED"; \
	else \
		echo "✓ gosec $$INSTALLED"; \
	fi

	@INSTALLED=$$(goimports -version 2>&1 | grep -oE 'v[0-9]+\.[0-9]+\.[0-9]+' | head -1); \
	EXPECTED=$(GOIMPORTS_VERSION); \
	if [ -z "$$INSTALLED" ]; then \
		echo "WARN: goimports version undetectable"; \
	elif [ "$$INSTALLED" != "$$EXPECTED" ]; then \
		echo "WARN: goimports mismatch — expected $$EXPECTED, got $$INSTALLED"; \
	else \
		echo "✓ goimports $$INSTALLED"; \
	fi

# -------------------------------------------------------------------
# Install tools at pinned versions
# -------------------------------------------------------------------
.PHONY: install-tools
install-tools:
	@echo "Installing tools at pinned versions..."
	go install golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION)
	go install github.com/securego/gosec/v2/cmd/gosec@$(GOSEC_VERSION)
	go install golang.org/x/tools/cmd/goimports@$(GOIMPORTS_VERSION)
	curl -sSfL https://github.com/golangci/golangci-lint/releases/download/v$(GOLANGCI_LINT_VERSION)/golangci-lint-$(GOLANGCI_LINT_VERSION)-linux-amd64.tar.gz \
		-o /tmp/golangci-lint.tar.gz
	tar -xzf /tmp/golangci-lint.tar.gz -C /tmp
	mv /tmp/golangci-lint-$(GOLANGCI_LINT_VERSION)-linux-amd64/golangci-lint $$(go env GOPATH)/bin/
	rm -rf /tmp/golangci-lint*
	@echo "✓ All tools installed"

# -------------------------------------------------------------------
# Run
# -------------------------------------------------------------------
.PHONY: goserver
goserver:
	go run ./cmd/api/main.go

.PHONY: goworker
goworker:
	go run ./cmd/worker/main.go

# -------------------------------------------------------------------
# CI Image
# -------------------------------------------------------------------
CI_IMAGE ?= $(DOCKER_REGISTRY_HOST)/ci-tools:latest

.PHONY: build-ci-image
build-ci-image:
	docker build -f Dockerfile.ci \
		--build-arg GOLANGCI_LINT_VERSION=$(GOLANGCI_LINT_VERSION) \
		--build-arg GOSEC_VERSION=$(GOSEC_VERSION) \
		--build-arg GOVULNCHECK_VERSION=$(GOVULNCHECK_VERSION) \
		--build-arg GOIMPORTS_VERSION=$(GOIMPORTS_VERSION) \
		-t $(CI_IMAGE) .

# -------------------------------------------------------------------
# Build
# -------------------------------------------------------------------
.PHONY: build 
build: build-server build-worker

.PHONY: build-server
build-server:
	go build -o bin/http-server ./cmd/api/main.go

.PHONY: build-worker
build-worker:
	go build -o bin/worker ./cmd/worker/main.go

# -------------------------------------------------------------------
# Lint & Format
# -------------------------------------------------------------------
.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: fmt
fmt:
	gofmt -w .
	goimports -w .

# Add a check-only fmt (doesn't write, just exits non-zero if unformatted)
.PHONY: fmt-check
fmt-check:
	@test -z "$$(gofmt -l $$(find . -name '*.go' -not -path './vendor/*'))" || (echo "Unformatted files:"; gofmt -l $$(find . -name '*.go' -not -path './vendor/*'); exit 1)

.PHONY: imports-check  
imports-check:
	@test -z "$$(goimports -l $$(find . -name '*.go' -not -path './vendor/*'))" || (echo "Bad imports:"; goimports -l $$(find . -name '*.go' -not -path './vendor/*'); exit 1)

# -------------------------------------------------------------------
# Docs
# -------------------------------------------------------------------
.PHONY: gen-docs
gen-docs:
	@echo "Generating Swagger docs..."
	swag init -g cmd/api/main.go -o docs --parseInternal --parseDependency
	@echo "Converting to OpenAPI 3.0..."
	go run ./tools -cmd convert-swagger docs/swagger.json
	@echo "Swagger docs generated in docs/"

# -------------------------------------------------------------------
# Security Check 
# -------------------------------------------------------------------
.PHONY: vulncheck
vulncheck:
	./scripts/govulncheck.sh

.PHONY: gosec-json
gosec-json:
	gosec -exclude-dir=cryptoutil -exclude-generated -tests=false -fmt=json -out=gosec.json ./...

.PHONY: gosec
gosec:
	gosec -exclude-dir=cryptoutil -exclude-generated -tests=false -fmt=text ./...

# -------------------------------------------------------------------
# Tidy
# -------------------------------------------------------------------
.PHONY: tidy
tidy:
	go mod tidy

# -------------------------------------------------------------------
# Pre-commit (run locally before pushing)
# -------------------------------------------------------------------
.PHONY: precommit
precommit: check-tools tidy fmt-check imports-check lint test-race
	@echo "✓ All checks passed — safe to commit"

# Slower — run before pushing, not every commit
.PHONY: prepush
prepush: precommit test-race gosec
	@echo "✓ Ready to push"

# -------------------------------------------------------------------
# Migration
# -------------------------------------------------------------------
.PHONY: migrate-create
migrate-create:
	@if [ -z "$(NAME)" ]; then \
		echo "Usage: make migrate-create NAME=create_users_table"; exit 1; \
	fi
	$(MIGRATE_BIN) create -ext sql -dir $(MIGRATE_DIR) -seq $(NAME)

.PHONY: migrate-up
migrate-up:
	@$(load_env) \
	echo "ENV=$(ENV) | File=$(ENV_FILE) | DB=$${DATABASE_URL}"; \
	$(MIGRATE_BIN) -path $(MIGRATE_DIR) -database "$${DATABASE_URL}" up

.PHONY: migrate-down
migrate-down:
	@$(load_env) \
	echo "ENV=$(ENV) | File=$(ENV_FILE) | DB=$${DATABASE_URL}"; \
	$(MIGRATE_BIN) -path $(MIGRATE_DIR) -database "$${DATABASE_URL}" down 1

.PHONY: migrate-down-all
migrate-down-all:
	@$(load_env) \
	echo "ENV=$(ENV) | File=$(ENV_FILE) | DB=$${DATABASE_URL}"; \
	$(MIGRATE_BIN) -path $(MIGRATE_DIR) -database "$${DATABASE_URL}" down -all

.PHONY: migrate-version
migrate-version:
	@$(load_env) \
	echo "ENV=$(ENV) | File=$(ENV_FILE) | DB=$${DATABASE_URL}"; \
	$(MIGRATE_BIN) -path $(MIGRATE_DIR) -database "$${DATABASE_URL}" version

.PHONY: migrate-force
migrate-force:
	@if [ -z "$(VERSION)" ]; then \
		echo "Usage: make migrate-force VERSION=<version>"; exit 1; \
	fi
	@$(load_env) \
	echo "ENV=$(ENV) | File=$(ENV_FILE) | DB=$${DATABASE_URL}"; \
	$(MIGRATE_BIN) -path $(MIGRATE_DIR) -database "$${DATABASE_URL}" force $(VERSION)

# -------------------------------------------------------------------
# MinIO (local dev)
# -------------------------------------------------------------------
MINIO_CONTAINER ?= pim-minio
MINIO_VOLUME    ?= pim-minio-data

.PHONY: minio-start
minio-start:
	@if docker ps -a --format '{{.Names}}' | grep -q '^$(MINIO_CONTAINER)$$'; then \
		echo "Starting existing container $(MINIO_CONTAINER)..."; \
		docker start $(MINIO_CONTAINER); \
	else \
		echo "Creating MinIO container $(MINIO_CONTAINER)..."; \
		docker run -d \
			--name $(MINIO_CONTAINER) \
			-p 9000:9000 \
			-p 9001:9001 \
			-v $(MINIO_VOLUME):/data \
			-e MINIO_ROOT_USER=minioadmin \
			-e MINIO_ROOT_PASSWORD=minioadmin \
			minio/minio:latest server /data --console-address ":9001"; \
	fi
	@echo "MinIO API:     http://localhost:9000"
	@echo "MinIO Console: http://localhost:9001  (minioadmin / minioadmin)"

.PHONY: minio-stop
minio-stop:
	@docker stop $(MINIO_CONTAINER) 2>/dev/null && echo "Stopped $(MINIO_CONTAINER)" || echo "Container not running"

.PHONY: minio-clean
minio-clean:
	@docker rm -f $(MINIO_CONTAINER) 2>/dev/null || true
	@docker volume rm $(MINIO_VOLUME) 2>/dev/null || true
	@echo "Removed container + volume"

# -------------------------------------------------------------------
# PostgreSQL 16 (local dev)
# -------------------------------------------------------------------
POSTGRES_CONTAINER ?= pim-postgres
POSTGRES_VOLUME    ?= pim-postgres-data
POSTGRES_USER      ?= postgres
POSTGRES_PASSWORD  ?= postgres
POSTGRES_DB        ?= mydb
POSTGRES_PORT      ?= 5432

.PHONY: pg-start
pg-start:
	@if docker ps -a --format '{{.Names}}' | grep -q '^$(POSTGRES_CONTAINER)$$'; then \
		echo "Starting existing container $(POSTGRES_CONTAINER)..."; \
		docker start $(POSTGRES_CONTAINER); \
	else \
		echo "Creating PostgreSQL container $(POSTGRES_CONTAINER)..."; \
		docker run -d \
			--name $(POSTGRES_CONTAINER) \
			-p $(POSTGRES_PORT):5432 \
			-v $(POSTGRES_VOLUME):/var/lib/postgresql/data \
			-e POSTGRES_USER=$(POSTGRES_USER) \
			-e POSTGRES_PASSWORD=$(POSTGRES_PASSWORD) \
			-e POSTGRES_DB=$(POSTGRES_DB) \
			postgres:16; \
	fi
	@echo "PostgreSQL running on port $(POSTGRES_PORT)"
	@echo "User: $(POSTGRES_USER), DB: $(POSTGRES_DB)"

.PHONY: pg-stop
pg-stop:
	@docker stop $(POSTGRES_CONTAINER) 2>/dev/null && echo "Stopped $(POSTGRES_CONTAINER)" || echo "Container not running"

.PHONY: pg-clean
pg-clean:
	@docker rm -f $(POSTGRES_CONTAINER) 2>/dev/null || true
	@docker volume rm $(POSTGRES_VOLUME) 2>/dev/null || true
	@echo "Removed container + volume"

.PHONY: pg-init-schemas
pg-init-schemas:
	@docker exec -i $(POSTGRES_CONTAINER) \
		psql -U $(POSTGRES_USER) -d $(POSTGRES_DB) -c "CREATE SCHEMA IF NOT EXISTS pim;"
	@docker exec -i $(POSTGRES_CONTAINER) \
		psql -U $(POSTGRES_USER) -d $(POSTGRES_DB) -c "CREATE SCHEMA IF NOT EXISTS pea;"


DUMP_FILE ?= /tmp/pim_dump.dump

.PHONY: pg-copy-from
pg-copy-from:
	@if [ -z "$(ENV)" ]; then \
		echo "Usage: make pg-copy-from ENV=dev|staging|prod"; \
		exit 1; \
	fi
	@$(load_env) \
	echo "HOST=[$$DB_HOST] PORT=[$$DB_PORT] USER=[$$DB_USER]"; \
	echo "===> Dumping from source ($$DB_HOST)"; \
	PGPASSWORD=$$DB_PASS pg_dump \
		-h $$DB_HOST \
		-p $$DB_PORT \
		-U $$DB_USER \
		-d $$DB_NAME \
		-n $$DB_SCHEMA \
		-Fc \
		-f $(DUMP_FILE); \
	echo "===> Resetting local schema"; \
	PGPASSWORD=$(POSTGRES_PASSWORD) psql \
		-h localhost \
		-p $(POSTGRES_PORT) \
		-U $(POSTGRES_USER) \
		-d $(POSTGRES_DB) \
		-c "DROP SCHEMA IF EXISTS $$DB_SCHEMA CASCADE; CREATE SCHEMA $$DB_SCHEMA;"; \
	echo "===> Restoring into local"; \
	PGPASSWORD=$(POSTGRES_PASSWORD) pg_restore \
		-h localhost \
		-p $(POSTGRES_PORT) \
		-U $(POSTGRES_USER) \
		-d $(POSTGRES_DB) \
		--no-owner \
		--role=$(POSTGRES_USER) \
		-n $$DB_SCHEMA \
		$(DUMP_FILE); \
	echo "===> Done"

# -------------------------------------------------------------------
# Generate DB Struct Model 
# -------------------------------------------------------------------
.PHONY: gen-db
gen-db:
	@$(load_env) \
	echo "ENV=$(ENV) | File=$(ENV_FILE) | DB=$${DATABASE_URL}"; \
	echo "Generating DB struct model from database..."; \
go run ./tools -cmd=generate-models

# -------------------------------------------------------------------
# Seed Role and default admin 
# -------------------------------------------------------------------
.PHONY: seed-role
seed-role:
	@$(load_env) \
	echo "ENV=$(ENV) | File=$(ENV_FILE) | DB=$${DATABASE_URL}"; \
	echo "Seed default role and superadmin"; \
	go run ./tools -cmd=seed-role

# -------------------------------------------------------------------
# Seed basic data for testing
# -------------------------------------------------------------------
.PHONY: seed-test
seed-test:
	@$(load_env) \
	echo "ENV=$(ENV) | File=$(ENV_FILE) | DB=$${DATABASE_URL}"; \
	echo "Seed default role and superadmin"; \
	go run ./tools -cmd=seed-basic-test
