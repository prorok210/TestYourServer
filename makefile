GO_VERSION := $(shell grep "^go " go.mod | cut -d ' ' -f 2)
GO_ENV := $(shell go version | awk '{print $$3}' | sed 's/go//g')
GO_CMD := $(shell command -v go 2>/dev/null)
OS := $(shell uname -s | tr '[:upper:]' '[:lower:]')
ARCH := $(shell uname -m)

MAIN_FILE := cmd/main.go
OUTPUT_DIR := build
PROG_NAME := $(OUTPUT_DIR)/TestYourServer

check-go:
	@echo "Checking Go compiler..."
	@if [ -z "$(GO_CMD)" ]; then \
		echo "Go compiler not found"; \
		exit 1; \
	fi
	@echo "Go compiler found: $(GO_CMD)"

check-version: check-go
	@echo "Checking Go version..."
	@if [ "$(GO_ENV)" \< "$(GO_VERSION)" ]; then \
		echo "Go version $(GO_VERSION) or higher is required, current version: $(GO_ENV)"; \
		exit 1; \
	fi

prepare-dir:
	@echo "Preparing build directory..."
	@mkdir -p $(OUTPUT_DIR)

build-dev: check-version prepare-dir
	@echo "Building in development mode..."
	@go build -o $(PROG_NAME) -gcflags="all=-N -l" $(MAIN_FILE)

build-prod: check-version prepare-dir
	@echo "Building in production mode..."
	@go build -o $(PROG_NAME) -ldflags="-s -w" $(MAIN_FILE)

clean:
	@echo "Cleaning..."
	@rm -rf $(OUTPUT_DIR)

.DEFAULT_GOAL := build-prod

.PHONY: check-go check-version prepare-dir build-dev build-prod clean
