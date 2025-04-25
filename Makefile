GO ?= go
EXECUTABLE := github2gitea

# Get all Go files (faster and cross-platform)
GOFILES := $(shell find . -type f -name "*.go")
TAGS ?=
VERSION ?= $(if $(DRONE_TAG),$(DRONE_TAG),$(shell git describe --tags 2>/dev/null || echo "dev-$(shell git rev-parse --short HEAD)"))
COMMIT ?= $(shell git rev-parse --short HEAD)

# Unified ldflags handling
LDFLAGS_BASE := -X 'main.Version=$(VERSION)' \
	-X 'main.Commit=$(COMMIT)' \
	-X 'main.BuildTime=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)' \
	-s -w
ifeq ($(shell uname), Darwin)
	EXTLDFLAGS :=
else
	EXTLDFLAGS := -extldflags "-static"
endif
LDFLAGS := $(EXTLDFLAGS) $(LDFLAGS_BASE)

.PHONY: all
all: $(EXECUTABLE) ## Default target
	@echo "Build $(EXECUTABLE) with version $(VERSION) and commit $(COMMIT)"

.PHONY: help
help: ## Show help message
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

$(EXECUTABLE): $(GOFILES)
	$(GO) build -v -tags '$(TAGS)' -ldflags '$(LDFLAGS)' -o bin/$@ ./cmd/$(EXECUTABLE)

.PHONY: install
install: $(GOFILES) ## Install binary
	$(GO) install -v -tags '$(TAGS)' -ldflags '$(LDFLAGS)' ./cmd/$(EXECUTABLE)
	@echo "Installed $(EXECUTABLE) to $(GOPATH)/bin/$(EXECUTABLE)"

.PHONY: test
test: ## Run tests
	@$(GO) test -v -cover -coverprofile coverage.txt ./... && echo "\n==>\033[32m Ok\033[m\n" || exit 1

.PHONY: clean
clean: ## Clean build artifacts
	rm -rf bin/ coverage.txt

.PHONY: lint
lint: ## Static code analysis
	golangci-lint run -v --timeout 5m --config .golangci.yml ./...
