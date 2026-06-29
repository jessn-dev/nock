# nock — developer task runner.
# Run `make help` for the list of targets.

BINARY      := nock
PKG         := github.com/jessn-dev/nock
CMD         := ./cmd/nock
BIN_DIR     := bin

VERSION     ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT      ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE        ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS := -s -w \
	-X $(PKG)/internal/version.Version=$(VERSION) \
	-X $(PKG)/internal/version.Commit=$(COMMIT) \
	-X $(PKG)/internal/version.Date=$(DATE)

GO        ?= go
GOFLAGS   ?=

.DEFAULT_GOAL := help

## help: show this help
.PHONY: help
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## //' | awk -F': ' '{printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2}'

## build: compile the nock binary into bin/
.PHONY: build
build:
	@mkdir -p $(BIN_DIR)
	$(GO) build $(GOFLAGS) -ldflags '$(LDFLAGS)' -o $(BIN_DIR)/$(BINARY) $(CMD)

## install: install nock into GOBIN
.PHONY: install
install:
	$(GO) install -ldflags '$(LDFLAGS)' $(CMD)

## run: build and run (ARGS="search web")
.PHONY: run
run:
	$(GO) run $(CMD) $(ARGS)

## test: run unit tests with race detector
.PHONY: test
test:
	$(GO) test -race -count=1 ./...

## cover: run tests and write a coverage report
.PHONY: cover
cover:
	$(GO) test -race -covermode=atomic -coverprofile=coverage.txt ./...
	$(GO) tool cover -func=coverage.txt | tail -1

## vet: run go vet
.PHONY: vet
vet:
	$(GO) vet ./...

## fmt: format all Go code
.PHONY: fmt
fmt:
	$(GO) fmt ./...

## lint: run golangci-lint (install: https://golangci-lint.run)
.PHONY: lint
lint:
	golangci-lint run ./...

## vuln: scan for known vulnerabilities
.PHONY: vuln
vuln:
	$(GO) run golang.org/x/vuln/cmd/govulncheck@latest ./...

## tidy: tidy and verify go.mod/go.sum
.PHONY: tidy
tidy:
	$(GO) mod tidy
	$(GO) mod verify

## check: the full local gate (fmt, vet, lint, test) — run before pushing
.PHONY: check
check: fmt vet lint test

## verify: full CI mirror incl. cross-platform build/lint (scripts/dev-check.sh)
.PHONY: verify
verify:
	./scripts/dev-check.sh

## clean: remove build artifacts
.PHONY: clean
clean:
	rm -rf $(BIN_DIR) dist coverage.txt coverage.html
