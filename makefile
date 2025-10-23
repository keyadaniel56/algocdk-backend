# Makefile
APP_NAME = app
MAIN_FILE = main.go

run: fmt vet lint build exec

fmt:
	@echo "🧹 Formatting code..."
	@go fmt ./...
	@goimports -w .
	@echo "✅ Formatting complete."

vet:
	@echo "🔍 Running go vet..."
	@go vet ./...
	@echo "✅ Vet check passed."

lint:
	@echo "🧠 Linting code..."
	@if ! command -v golangci-lint >/dev/null; then \
		echo "⚠️  golangci-lint not found. Installing..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	@golangci-lint run ./...
	@echo "✅ Linting complete."

build:
	@echo "🏗️  Building $(APP_NAME)..."
	@go build -o $(APP_NAME) $(MAIN_FILE)
	@echo "✅ Build successful."

exec:
	@echo "🚀 Running $(APP_NAME)..."
	@./$(APP_NAME)

clean:
	@echo "🧽 Cleaning up..."
	@rm -f $(APP_NAME)
	@echo "✅ Clean complete."
