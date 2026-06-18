# VIDI Makefile
# Project: VIDI (Validate, Integrate, Data, Input)
# Author: Emperor42
# Description: Build, run, and manage the VIDI middleware server.

# --- Configuration ---
APP_NAME := vidi-server
VERSION := 1.0.0
GO := go
BUILD_DIR := ./bin
CMD_PATH := ./cmd/server
PKG_PATH := ./pkg
STATIC_PATH := ./static
DATA_PATH := ./data

# --- Colors for Output (Optional, works in most terminals) ---
GREEN  := \033[0;32m
YELLOW := \033[0;33m
BLUE   := \033[0;34m
NC     := \033[0m # No Color

# --- Default Target ---
.PHONY: help
help:
	@echo "$$(BLUE)VIDI Build System$$(NC)"
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  $$(GREEN)build$$(NC)      - Compile the binary to $$(BUILD_DIR)/$$(APP_NAME)"
	@echo "  $$(GREEN)run$$(NC)        - Run the server locally (port 8080)"
	@echo "  $$(GREEN)clean$$(NC)      - Remove binaries and generated data"
	@echo "  $$(GREEN)test$$(NC)       - Run all unit tests"
	@echo "  $$(GREEN)deps$$(NC)       - Download and tidy dependencies"
	@echo "  $$(GREEN)fmt$$(NC)        - Format Go code"
	@echo "  $$(GREEN)lint$$(NC)       - Run linter (if installed)"
	@echo "  $$(GREEN)init$$(NC)       - Initialize directories (data, bin)"
	@echo "  $$(GREEN)help$$(NC)       - Show this help message"

# --- Initialization ---
.PHONY: init
init:
	@echo "$$(YELLOW)Initializing project directories...$$(NC)"
	@mkdir -p $(BUILD_DIR)
	@mkdir -p $(DATA_PATH)
	@mkdir -p $(STATIC_PATH)
	@echo "$$(GREEN)Done.$$(NC)"

# --- Build ---
.PHONY: build
build: init
	@echo "$(BLUE)Building $$(APP_NAME) v$$(VERSION)...$(NC)"
	@$(GO) build -ldflags="-s -w" -o $$(BUILD_DIR)/$$(APP_NAME) $(CMD_PATH)
	@echo "$(GREEN)Build successful: $$(BUILD_DIR)/$$(APP_NAME)$(NC)"

# --- Run ---
.PHONY: run
run: init
	@echo "$$(BLUE)Starting VIDI server on http://localhost:8080...$$(NC)"
	@echo "$$(YELLOW)Press Ctrl+C to stop.$$(NC)"
	@$(GO) run $(CMD_PATH)/main.go

# --- Clean ---
.PHONY: clean
clean:
	@echo "$$(YELLOW)Cleaning up...$$(NC)"
	@rm -rf $(BUILD_DIR)
	@rm -rf $(DATA_PATH)
	@rm -f *.log
	@echo "$$(GREEN)Cleanup complete.$$(NC)"

# --- Test ---
.PHONY: test
test:
	@echo "$$(BLUE)Running tests...$$(NC)"
	@$(GO) test -v -race -cover ./...

# --- Dependencies ---
.PHONY: deps
deps:
	@echo "$$(BLUE)Fetching dependencies...$$(NC)"
	@$(GO) mod download
	@$(GO) mod tidy
	@echo "$$(GREEN)Dependencies updated.$$(NC)"

# --- Formatting ---
.PHONY: fmt
fmt:
	@echo "$$(BLUE)Formatting code...$$(NC)"
	@$(GO) fmt ./...
	@echo "$$(GREEN)Code formatted.$$(NC)"

# --- Linting (Optional) ---
.PHONY: lint
lint:
	@if command -v golangci-lint > /dev/null; then \
		echo "$$(BLUE)Running golangci-lint...$$(NC)"; \
		golangci-lint run; \
	else \
		echo "$$(YELLOW)golangci-lint not found. Skipping linting.$$(NC)"; \
	fi

# --- Docker (Optional Future Extension) ---
# Uncomment if you plan to add Docker support later
# .PHONY: docker-build
# docker-build:
# 	@docker build -t emperor42/vidi:$(VERSION) .