.PHONY: lint
lint: ## Run golangci-lint linter.
	@echo "Running golangci-lint..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint cache clean && golangci-lint run; \
	else \
		echo "Error: golangci-lint not found. Please install it:"; \
		echo " go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		exit 1; \
	fi
