# Atlassian CLI Makefile

# Version info — the latest release tag, falling back to the VERSION file.
VERSION := $(shell git describe --tags --abbrev=0 2>/dev/null | sed 's/^v//' | grep . || cat VERSION)
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Go build settings
BINARY_NAME := atlassian
GO := go
LDFLAGS := -ldflags "-X github.com/Well-Polaris/atlassian-cli/internal/version.Version=$(VERSION) \
	-X github.com/Well-Polaris/atlassian-cli/internal/version.GitCommit=$(GIT_COMMIT) \
	-X github.com/Well-Polaris/atlassian-cli/internal/version.BuildDate=$(BUILD_DATE)"

.PHONY: all build clean test install version bump-patch bump-minor bump-major release help

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

## bump-patch: Increment the patch version (x.y.Z)
bump-patch:
	@scripts/bump-version.sh patch

## bump-minor: Increment the minor version (x.Y.0)
bump-minor:
	@scripts/bump-version.sh minor

## bump-major: Increment the major version (X.0.0)
bump-major:
	@scripts/bump-version.sh major

## release: Build, commit the current VERSION and create its git tag
release: build
	@VERSION=$$(cat VERSION); \
	git add VERSION; \
	git commit -m "Release v$$VERSION"; \
	git tag -a "v$$VERSION" -m "Release v$$VERSION"; \
	echo "Tagged v$$VERSION — push with: git push --follow-tags"

## help: Show this help
help:
	@echo "Atlassian CLI v$(VERSION)"
	@echo ""
	@echo "Usage:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'
