.PHONY: build run test coverage integration-test lint fmt vet security reports reports-zip health-check docker-build docker-run compose-up compose-down migrate clean

APP_NAME := opstrack
BIN_DIR := bin

build: ## Build the server binary
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(APP_NAME) ./cmd/server

run: ## Run the server locally (requires a reachable Postgres, see .env.example)
	go run ./cmd/server

test: ## Run unit tests
	go test ./... -v

coverage: ## Run unit tests with coverage report
	go test ./... -coverprofile=coverage.out -covermode=atomic
	go tool cover -func=coverage.out

integration-test: ## Run integration tests against a real Postgres instance
	go test -tags=integration ./tests/... -v

fmt: ## Check formatting
	gofmt -l .

vet: ## Run go vet
	go vet ./...

lint: ## Run golangci-lint (install separately: https://golangci-lint.run)
	golangci-lint run ./... || echo "golangci-lint not installed locally; CI will still run it"

security: ## Run gosec/trivy/snyk locally where installed (CI always runs all of them)
	./scripts/run-security-scans.sh

reports: ## (Re)build the reports/index.html dashboard from whatever is in reports/
	./scripts/generate-report-dashboard.sh reports

reports-zip: reports ## Build the dashboard and package reports/ into opstrack-ci-reports.zip
	./scripts/generate-zip.sh reports reports/opstrack-ci-reports.zip

health-check: ## Poll /health on a running instance (default: http://localhost:8080)
	./scripts/health-check.sh

docker-build: ## Build the Docker image
	docker build -t $(APP_NAME):local .

docker-run: ## Run the Docker image standalone
	docker run --rm -p 8080:8080 --env-file .env $(APP_NAME):local

compose-up: ## Start the full stack with Docker Compose
	docker compose up --build

compose-down: ## Stop the Docker Compose stack
	docker compose down -v

clean: ## Remove build artifacts and generated reports
	rm -rf $(BIN_DIR) coverage.out
	find reports -mindepth 1 ! -name '.gitkeep' -delete 2>/dev/null || true
