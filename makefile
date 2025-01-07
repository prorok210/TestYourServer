GO_VERSION := $(shell grep "^go " go.mod | cut -d ' ' -f 2)
GO_ENV := $(shell go version | awk '{print $$3}' | sed 's/go//g')
GO_CMD := $(shell command -v go 2>/dev/null)
OS := $(shell uname -s | tr '[:upper:]' '[:lower:]')
ARCH := $(shell uname -m)

MAIN_FILE := cmd/main.go
OUTPUT_DIR := build
PROG_NAME := $(OUTPUT_DIR)/TestYourServer

DEBIAN_PACKAGES := libx11-dev libxext-dev libxinerama-dev libxcursor-dev libxi-dev libxxf86vm-dev
FEDORA_PACKAGES := libX11-devel libXext-devel libXinerama-devel libXcursor-devel libXi-devel libXxf86vm-devel
MACOS_PACKAGES := glfw pkg-config
WINDOWS_PACKAGES := mingw-w64-x86_64-glfw

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

install-deps: check-version
	@echo "Installing dependencies..."
	@if [ "$(OS)" = "linux" ]; then \
		if [ -f /etc/debian_version ]; then \
			echo "Debian-based system detected. Installing dependencies..."; \
			sudo apt-get update && sudo apt-get install -y $(DEBIAN_PACKAGES); \
		elif [ -f /etc/redhat-release ]; then \
			echo "Fedora-based system detected. Installing dependencies..."; \
			sudo dnf install -y $(FEDORA_PACKAGES); \
		else \
			echo "Unknown Linux distribution. Please install dependencies manually."; \
			exit 1; \
		fi \
	elif [ "$(OS)" = "darwin" ]; then \
		echo "macOS system detected. Installing dependencies..."; \
		brew install $(MACOS_PACKAGES); \
	elif [ "$(OS)" = "mingw32" ] || [ "$(OS)" = "mingw64" ]; then \
		echo "Windows system detected. Checking if MSYS2 or MinGW is installed..."; \
		if ! command -v pacman &> /dev/null; then \
			echo "pacman (MSYS2/MinGW) is not installed. Please install MSYS2/MinGW or use vcpkg."; \
			exit 1; \
		fi \
		echo "MSYS2/MinGW detected. Installing dependencies..."; \
		pacman -S --noconfirm $(WINDOWS_PACKAGES); \
	elif [ "$(OS)" = "windows_nt" ]; then \
		echo "Windows NT system detected. Please manually install GLFW using vcpkg or other tools."; \
		exit 1; \
	else \
		echo "Unsupported OS. Please install dependencies manually."; \
		exit 1; \
	fi
	@echo "All dependencies installed."

prepare-dir:
	@echo "Preparing build directory..."
	@mkdir -p $(OUTPUT_DIR)

build-dev: check-version prepare-dir install-deps
	@echo "Building in development mode..."
	@go build -o $(PROG_NAME) -gcflags="all=-N -l" $(MAIN_FILE)

build-prod: check-version prepare-dir install-deps
	@echo "Building in production mode..."
	@go build -o $(PROG_NAME) -ldflags="-s -w" $(MAIN_FILE)

clean:
	@echo "Cleaning..."
	@rm -rf $(OUTPUT_DIR)

.DEFAULT_GOAL := build-prod

.PHONY: check-go check-version prepare-dir build-dev build-prod clean install-deps

