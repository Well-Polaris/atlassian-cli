# Atlassian CLI Makefile

# Version info
VERSION := $(shell cat VERSION)
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Go build settings
BINARY_NAME := atlassian
GO := go
LDFLAGS := -ldflags "-X github.com/peter/atlassian-cli/internal/version.Version=$(VERSION) \
	-X github.com/peter/atlassian-cli/internal/version.GitCommit=$(GIT_COMMIT) \
	-X github.com/peter/atlassian-cli/internal/version.BuildDate=$(BUILD_DATE)"

.PHONY: all build clean test install version help

all: build

## build: Build the CLI binary
build:
	$(GO) build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/atlassian

## install: Install the CLI to $GOPATH/bin
install:
	$(GO) install $(LDFLAGS) ./cmd/atlassian

## clean: Remove build artifacts
clean:
	rm -f $(BINARY_NAME)
	$(GO) clean

## test: Run tests
test:
	$(GO) test -v ./...

## version: Print version info
version:
	@echo "Version: $(VERSION)"
	@echo "Commit:  $(GIT_COMMIT)"
	@echo "Date:    $(BUILD_DATE)"

## help: Show this help
help:
	@echo "Atlassian CLI v$(VERSION)"
	@echo ""
	@echo "Usage:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'
