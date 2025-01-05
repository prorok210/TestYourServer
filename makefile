# Makefile
# Cross-platform Makefile for building a Go application using Fyne and necessary C libraries.
# Supports building on Linux, macOS, and Windows, with automatic dependency checks and installations.

SHELL := /bin/bash

APP_NAME := TestYourServer
LIB_DIR := lib

PLATFORMS := linux darwin windows
ARCH := amd64

SRC := cmd/main.go

.DEFAULT_GOAL := build

# --- Detect Operating System ---
UNAME_S := $(shell uname -s)

.PHONY: build check-libs clean cross install-libs check-go unsupported

check-go:
	@echo "Checking for Go compiler..."
	@which go > /dev/null 2>&1
	@if [ $$? -ne 0 ]; then \
		echo "Go not found. Installing Go..."; \
		$(MAKE) install-go; \
	else \
		echo "Go is already installed: $$(go version)"; \
	fi

# Target to install Go based on OS
install-go:
ifeq ($(OS), Windows_NT)
	@echo "Installing Go via Chocolatey..."
	@choco install golang -y
	@echo "Verifying Go installation..."
	@which go > /dev/null 2>&1 || { \
		echo "Error: Go installation failed or not found in PATH."; \
		exit 1; \
	}
	@echo "Go is installed: $$(go version)"
else ifeq ($(UNAME_S), Darwin)
	@echo "Installing Go via Homebrew..."
	@brew list go &> /dev/null || { \
		echo "Go not found. Installing Go..."; \
		brew install go; \
	}
	@echo "Verifying Go installation..."
	@which go > /dev/null 2>&1 || { \
		echo "Error: Go installation failed or not found in PATH."; \
		exit 1; \
	}
	@echo "Go is installed: $$(go version)"
else ifeq ($(UNAME_S), Linux)
	@echo "Installing Go via package manager..."
	@if command -v apt-get &> /dev/null; then \
		sudo apt-get update && sudo apt-get install -y golang-go; \
	elif command -v dnf &> /dev/null; then \
		sudo dnf install -y golang; \
	elif command -v yum &> /dev/null; then \
		sudo yum install -y golang; \
	elif command -v pacman &> /dev/null; then \
		sudo pacman -Syu go --noconfirm; \
	else \
		echo "Could not determine package manager to install Go. Please install Go manually."; \
		exit 1; \
	fi
	@echo "Verifying Go installation..."
	@which go > /dev/null 2>&1 || { \
		echo "Error: Go installation failed or not found in PATH."; \
		exit 1; \
	}
	@echo "Go is installed: $$(go version)"
else
	@echo "Error: Unsupported operating system for Go installation."
	@exit 1
endif

# Target to install necessary libraries
install-libs:
	@echo "Checking and installing dependencies..."

ifeq ($(OS), Windows_NT)
	@echo "Detected Windows."
	
	@echo "Checking for Chocolatey..."
	@if ! command -v choco > /dev/null 2>&1; then \
		echo "Chocolatey not found. Installing Chocolatey..."; \
		powershell -NoProfile -ExecutionPolicy Bypass -Command "Set-ExecutionPolicy Bypass -Scope Process -Force; \
		[System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072; \
		iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))"; \
	else \
		echo "Chocolatey is already installed."; \
	fi
	
	@echo "Installing pkgconfiglite via Chocolatey..."
	@choco install pkgconfiglite -y
	
	@echo "Installing MSYS2 via Chocolatey..."
	@choco install msys2 -y
	
	@echo "Updating pacman and installing GLFW..."
	@"C:\msys64\usr\bin\pacman.exe" -Syu --noconfirm
	@"C:\msys64\usr\bin\pacman.exe" -S --noconfirm mingw-w64-x86_64-glfw
	
	@echo "Copying necessary DLLs to $(LIB_DIR)..."
	@mkdir -p $(LIB_DIR)
	@cp /c/msys64/mingw64/bin/*.dll $(LIB_DIR) 2>/dev/null || echo "Failed to copy DLLs. Please ensure the paths are correct."
	
	@echo "GLFW is installed or already present."

else ifeq ($(UNAME_S), Darwin)
	@echo "Detected macOS."

	@echo "Checking for Homebrew..."
	@if ! command -v brew &> /dev/null; then \
		echo "Homebrew not found. Installing Homebrew..."; \
		/bin/bash -c "$$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"; \
		# Add Homebrew to PATH; may require terminal restart \
		echo 'export PATH="/usr/local/bin:$${PATH}"' >> ~/.bash_profile; \
		echo "Homebrew installed."; \
	else \
		echo "Homebrew is already installed."; \
	fi

	@echo "Updating Homebrew..."
	@brew update

	@echo "Checking for pkg-config..."
	@if ! brew list pkg-config &> /dev/null; then \
		echo "pkg-config not found. Installing pkg-config..."; \
		brew install pkg-config; \
	else \
	    echo "pkg-config is already installed."; \
	fi

	@echo "Checking for GLFW via Homebrew..."
	@if ! brew list glfw &> /dev/null; then \
		echo "GLFW not found. Installing GLFW..."; \
		brew install glfw; \
	else \
		echo "GLFW is already installed."; \
	fi

	@echo "Copying necessary libraries to $(LIB_DIR)..."
	@mkdir -p $(LIB_DIR)
	@cp /usr/local/lib/libglfw* $(LIB_DIR) 2>/dev/null || { \
		# For Apple Silicon Macs, Homebrew might use a different path \
		if [ -d "/opt/homebrew/lib" ]; then \
			cp /opt/homebrew/lib/libglfw* $(LIB_DIR) 2>/dev/null || echo "Failed to copy GLFW libraries."; \
		else \
			echo "Failed to copy GLFW libraries."; \
		fi; \
	}

	@echo "GLFW is installed or already present."

else
	@echo "Detected Linux."

	@echo "Checking for pkg-config..."
	@if ! command -v pkg-config &> /dev/null; then \
		echo "pkg-config not found. Attempting to install..."; \
		if command -v apt-get &> /dev/null; then \
			sudo apt-get update && sudo apt-get install -y pkg-config; \
		elif command -v dnf &> /dev/null; then \
			sudo dnf install -y pkg-config; \
		elif command -v yum &> /dev/null; then \
			sudo yum install -y pkgconfig; \
		elif command -v pacman &> /dev/null; then \
			sudo pacman -Syu pkgconf --noconfirm; \
		else \
			echo "Could not determine package manager. Please install pkg-config manually."; \
			exit 1; \
		fi; \
	else \
		echo "pkg-config is already installed."; \
	fi

	@echo "Checking for GLFW via pkg-config..."
	@if ! pkg-config --exists glfw3; then \
		echo "GLFW not found. Attempting to install..."; \
		if command -v apt-get &> /dev/null; then \
			sudo apt-get update && sudo apt-get install -y libglfw3-dev; \
		elif command -v dnf &> /dev/null; then \
			sudo dnf install -y glfw-devel; \
		elif command -v yum &> /dev/null; then \
			sudo yum install -y glfw-devel; \
		elif command -v pacman &> /dev/null; then \
			sudo pacman -Syu mingw-w64-x86_64-glfw --noconfirm; \
		else \
			echo "Could not determine package manager. Please install GLFW manually."; \
			exit 1; \
		fi; \
	else \
		echo "GLFW is confirmed via pkg-config."; \
	fi

	@echo "Checking for specific Linux distributions..."
	@if command -v apt-get &> /dev/null; then \
		echo "Using Debian-based package manager (apt-get)."; \
	elif command -v dnf &> /dev/null; then \
		echo "Using Fedora-based package manager (dnf)."; \
	elif command -v yum &> /dev/null; then \
		echo "Using RHEL-based package manager (yum)."; \
	elif command -v pacman &> /dev/null; then \
		echo "Using Arch-based package manager (pacman)."; \
	else \
		echo "Unsupported Linux distribution."; \
		exit 1; \
	fi

	@echo "Copying necessary libraries to $(LIB_DIR)..."
	@mkdir -p $(LIB_DIR)
	@cp /usr/lib/libglfw* $(LIB_DIR) 2>/dev/null || { \
		# Additional paths for different distributions if necessary \
		if [ -d "/usr/lib64" ]; then \
			cp /usr/lib64/libglfw* $(LIB_DIR) 2>/dev/null || echo "Failed to copy GLFW libraries."; \
		else \
			echo "Failed to copy GLFW libraries."; \
		fi; \
	}

	@echo "GLFW is installed or already present."

endif

# Target to check libraries
check-libs: install-libs
	@echo "Checking for necessary libraries for Fyne..."

	@if ! command -v pkg-config &> /dev/null; then \
		echo "pkg-config not found. Please install it manually."; \
		exit 1; \
	fi

ifeq ($(OS), Windows_NT)
	@echo "Checking for GLFW using pkg-config..."
	@export PKG_CONFIG_PATH=/c/msys64/mingw64/lib/pkgconfig && pkg-config --exists glfw3 || { \
		echo "GLFW not found. Please ensure GLFW is installed via MSYS2."; \
		exit 1; \
	}
else
	@echo "Checking for GLFW via pkg-config..."
	@if ! pkg-config --exists glfw3; then \
		echo "GLFW not found. Please verify the installation."; \
		exit 1; \
	else \
		echo "GLFW is confirmed via pkg-config."; \
	fi
endif

	@echo "All dependencies for Fyne are present or have been installed."

# Target to build the application
build: check-libs
	@echo "Creating Build directory..."
	@mkdir -p Build
	@echo "Building the application in the Build directory..."
	@GO111MODULE=on go build -o Build/$(APP_NAME) $(SRC)
	@echo "Build completed. Output: Build/$(APP_NAME)"

# Target to clean built artifacts
clean:
	@echo "Removing compiled binaries..."
	@rm -rf Build
	@for os in $(PLATFORMS); do \
		rm -f Build/$(APP_NAME)-$$os-$(ARCH); \
	done
	@echo "Cleanup completed."

# Target for cross-compilation
cross: check-libs
	@echo "Starting cross-compilation..."
	@mkdir -p Build
	@for os in $(PLATFORMS); do \
		echo "Building for $$os/$(ARCH)..."; \
		GOOS=$$os GOARCH=$(ARCH) go build -o Build/$(APP_NAME)-$$os-$(ARCH) $(SRC); \
	done
	@echo "Cross-compilation completed."

# Target to handle unsupported systems
unsupported:
	@echo "Error: Unsupported operating system or distribution."
	@exit 1