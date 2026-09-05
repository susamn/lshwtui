# lshwtui — build and development tasks.

BINARY  := lshwtui
PREFIX  ?= $(HOME)/.local
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)

# GOROOT is unset for every Go invocation: a stale GOROOT pointing at a
# different toolchain than the `go` on PATH makes the compiler refuse to build
# the standard library. Harmless when GOROOT is already correct or absent.
GO := env -u GOROOT go

.DEFAULT_GOAL := build

.PHONY: build
build: ## Build the binary into ./$(BINARY)
	$(GO) build -trimpath -ldflags '$(LDFLAGS)' -o $(BINARY) .

.PHONY: run
run: build ## Build and run against the live hardware tree
	./$(BINARY)

.PHONY: test
test: ## Run the test suite
	$(GO) test ./...

.PHONY: vet
vet: ## Run go vet
	$(GO) vet ./...

.PHONY: fmt
fmt: ## Format all Go sources
	gofmt -w .

.PHONY: fmt-check
fmt-check: ## Fail if any source is unformatted
	@out=$$(gofmt -l .); \
	if [ -n "$$out" ]; then echo "unformatted files:"; echo "$$out"; exit 1; fi

.PHONY: check
check: fmt-check vet test ## Everything CI should run

.PHONY: install
install: build ## Install to $(PREFIX)/bin
	install -Dm755 $(BINARY) $(PREFIX)/bin/$(BINARY)

.PHONY: uninstall
uninstall: ## Remove the installed binary
	rm -f $(PREFIX)/bin/$(BINARY)

.PHONY: snapshot
snapshot: ## Capture a fresh lshw JSON snapshot as root
	sudo lshw -json > lshw-$$(hostname)-$$(date +%Y%m%d).json
	@echo "wrote lshw-$$(hostname)-$$(date +%Y%m%d).json"

.PHONY: clean
clean: ## Remove build artifacts
	rm -f $(BINARY)
	$(GO) clean

.PHONY: help
help: ## List available targets
	@grep -hE '^[a-z-]+:.*?## ' $(MAKEFILE_LIST) \
		| awk -F':.*?## ' '{printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'
