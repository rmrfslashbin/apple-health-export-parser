# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOVET=$(GOCMD) vet

# Binary name
BINARY_NAME=apple-health-export-parser

# Build directory
BUILD_DIR=build

# Project paths
PACKAGE_PATH=./...
 
# Build flags
BUILD_FLAGS=-v

# Version information
VERSION ?= $(shell git describe --tags --always --dirty)
COMMIT=$(shell git rev-parse --short HEAD)
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildTime=$(BUILD_TIME)"

# Platforms
PLATFORMS=linux/amd64 linux/arm64 darwin/amd64 darwin/arm64

# Make all directories that don't exist
$(shell mkdir -p $(BUILD_DIR))

.PHONY: all build clean test coverage deps lint vet fmt help run build-all

all: help

## build: Build the application for current platform
build: deps
	$(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(PACKAGE_PATH)

## build-all: Build for all platforms (Linux x86/ARM, MacOS x86/ARM)
build-all: deps
	$(foreach platform,$(PLATFORMS),\
		echo "Building for $(platform)..." && \
		GOOS=$(shell echo $(platform) | cut -d"/" -f1) \
		GOARCH=$(shell echo $(platform) | cut -d"/" -f2) \
		$(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) \
		-o $(BUILD_DIR)/$(BINARY_NAME)_$(shell echo $(platform) | tr "/" "_") \
		$(PACKAGE_PATH) && \
	) true

## build-linux-amd64: Build for Linux x86_64
build-linux-amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) \
		-o $(BUILD_DIR)/$(BINARY_NAME)_linux_amd64 $(PACKAGE_PATH)

## build-linux-arm64: Build for Linux ARM64
build-linux-arm64:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) \
		-o $(BUILD_DIR)/$(BINARY_NAME)_linux_arm64 $(PACKAGE_PATH)

## build-darwin-amd64: Build for MacOS x86_64
build-darwin-amd64:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) \
		-o $(BUILD_DIR)/$(BINARY_NAME)_darwin_amd64 $(PACKAGE_PATH)

## build-darwin-arm64: Build for MacOS ARM64 (Apple Silicon)
build-darwin-arm64:
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) \
		-o $(BUILD_DIR)/$(BINARY_NAME)_darwin_arm64 $(PACKAGE_PATH)

## clean: Clean build directory
clean:
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)

## test: Run unit tests
test:
	$(GOTEST) -v ./...

## coverage: Run tests with coverage
coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

## deps: Download and tidy dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy
	$(GOMOD) verify

## lint: Run linter
lint:
	golangci-lint run

## vet: Run go vet
vet:
	$(GOVET) ./...

## fmt: Run go fmt
fmt:
	$(GOCMD) fmt ./...

## run: Build and run the application
run: build
	./$(BUILD_DIR)/$(BINARY_NAME)

## help: Show this help message
help:
	@echo "Usage:"
	@echo
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'
	@echo
	@echo "Build Targets:"
	@echo " build              - Build for current platform"
	@echo " build-all          - Build for all platforms"
	@echo " build-linux-amd64  - Build for Linux x86_64"
	@echo " build-linux-arm64  - Build for Linux ARM64"
	@echo " build-darwin-amd64 - Build for MacOS x86_64"
	@echo " build-darwin-arm64 - Build for MacOS ARM64 (Apple Silicon)"
	@echo
	@echo "Development Targets:"
	@echo " clean         - Clean build directory"
	@echo " test         - Run unit tests"
	@echo " coverage     - Run tests with coverage"
	@echo " deps         - Download and tidy dependencies"
	@echo " lint         - Run linter"
	@echo " vet          - Run go vet"
	@echo " fmt          - Run go fmt"
	@echo " run          - Build and run the application"

.DEFAULT_GOAL := help