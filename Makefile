# Makefile for MOPS SDK
# ====================
# This Makefile provides code quality checks for the MOPS SDK

# Colors for pretty output
GREEN := \033[0;32m
BLUE := \033[0;34m
YELLOW := \033[1;33m
RED := \033[0;31m
NC := \033[0m # No Color

.PHONY: help deadcode check-deadcode vet lint test

# Default target
help:
	@echo "$(BLUE)MOPS SDK Code Quality Tools$(NC)"
	@echo "============================"
	@echo ""
	@echo "$(YELLOW)Available targets:$(NC)"
	@echo "  $(GREEN)help$(NC)           - Show this help message"
	@echo "  $(GREEN)deadcode$(NC)       - Run deadcode analysis"
	@echo "  $(GREEN)check-deadcode$(NC) - Check if deadcode tool is installed"
	@echo "  $(GREEN)vet$(NC)            - Run go vet"
	@echo "  $(GREEN)lint$(NC)           - Run golint (if available)"
	@echo "  $(GREEN)test$(NC)           - Run tests"
	@echo "  $(GREEN)quality$(NC)        - Run all quality checks"
	@echo "  $(GREEN)unused$(NC)         - Check for unused code (alternative to deadcode)"

# Check if deadcode tool is installed
check-deadcode:
	@echo "$(BLUE)🔍 Checking for deadcode tool...$(NC)"
	@test -x "$$(go env GOPATH)/bin/deadcode" || { \
		echo "$(RED)❌ deadcode tool not found at $$(go env GOPATH)/bin/deadcode$(NC)"; \
		echo "$(YELLOW)💡 Install it with: go install golang.org/x/tools/cmd/deadcode@latest$(NC)"; \
		exit 1; \
	}
	@echo "$(GREEN)✅ deadcode tool is available$(NC)"

# Run deadcode analysis
deadcode: 
	@echo "$(BLUE)🔍 Running deadcode analysis on mops-sdk...$(NC)"
	@echo "$(YELLOW)⚠️  Note: deadcode tool cannot directly analyze library packages$(NC)"
	@echo "$(BLUE)� For SDK libraries, use these alternatives:$(NC)"
	@echo "   • $(GREEN)make unused$(NC) - Uses golangci-lint/staticcheck for unused exports"
	@echo "   • $(GREEN)go vet ./...$(NC) - Basic code analysis"
	@echo ""
	@echo "$(YELLOW)🔍 Checking what deadcode can analyze in this repo...$(NC)"
	@if [ -d "examples" ] && find examples -name "main.go" | head -1 >/dev/null 2>&1; then \
		echo "$(BLUE)✓ Found examples with main packages, analyzing those...$(NC)"; \
		if command -v $$(go env GOPATH)/bin/deadcode >/dev/null 2>&1; then \
			$$(go env GOPATH)/bin/deadcode ./examples/...; \
		else \
			echo "$(RED)❌ deadcode tool not found$(NC)"; \
		fi; \
	else \
		echo "$(YELLOW)⚠️  No main packages found in examples directory$(NC)"; \
	fi
	@echo ""
	@echo "$(BLUE)💡 To detect unused SDK exports, run: $(GREEN)make unused$(NC)"

# Run go vet
vet:
	@echo "$(BLUE)🔍 Running go vet...$(NC)"
	@go vet ./...
	@echo "$(GREEN)✅ go vet completed$(NC)"

# Run golint (if available)
lint:
	@echo "$(BLUE)🔍 Running golint...$(NC)"
	@if command -v golint >/dev/null 2>&1; then \
		golint ./...; \
		echo "$(GREEN)✅ golint completed$(NC)"; \
	else \
		echo "$(YELLOW)⚠️  golint not available, install with: go install golang.org/x/lint/golint@latest$(NC)"; \
	fi

# Run tests
test:
	@echo "$(BLUE)🧪 Running tests...$(NC)"
	@go test ./...
	@echo "$(GREEN)✅ Tests completed$(NC)"

# Check for unused code (better for libraries)
unused:
	@echo "$(BLUE)🔍 Checking for unused exports and code in mops-sdk...$(NC)"
	@echo ""
	@if command -v golangci-lint >/dev/null 2>&1; then \
		echo "$(BLUE)✓ Using golangci-lint for comprehensive analysis...$(NC)"; \
		golangci-lint run --enable=unused,deadcode,varcheck,structcheck,ineffassign,unparam --disable-all; \
	elif command -v staticcheck >/dev/null 2>&1; then \
		echo "$(BLUE)✓ Using staticcheck for unused code detection...$(NC)"; \
		staticcheck ./...; \
	elif command -v $$(go env GOPATH)/bin/ineffassign >/dev/null 2>&1; then \
		echo "$(BLUE)✓ Using ineffassign for unused assignments...$(NC)"; \
		$$(go env GOPATH)/bin/ineffassign ./...; \
	else \
		echo "$(YELLOW)⚠️  No unused code analyzer available$(NC)"; \
		echo "$(BLUE)💡 Install one of these tools:$(NC)"; \
		echo "   • golangci-lint (recommended): curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin"; \
		echo "   • staticcheck: go install honnef.co/go/tools/cmd/staticcheck@latest"; \
		echo "   • ineffassign: go install github.com/gordonklaus/ineffassign@latest"; \
		echo ""
		echo "$(GREEN)Alternative: Manual inspection of exports$(NC)"; \
		echo "Check which exported functions/types might be unused:"; \
		grep -r "^func [A-Z]" *.go | head -10 || true; \
		echo "..."; \
	fi

# Run all quality checks
quality: vet unused test
	@echo "$(GREEN)🎉 All quality checks completed!$(NC)"