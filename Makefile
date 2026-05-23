APP_NAME := export-ua-history
CMD_PATH := ./cmd/$(APP_NAME)
BIN_DIR := $(CURDIR)/bin
BIN_PATH := $(BIN_DIR)/$(APP_NAME)
GO ?= go
GOCACHE ?= $(CURDIR)/.gocache

export GOCACHE

.PHONY: help build test fmt tidy install run version clean

help: ## Show available targets.
	@awk 'BEGIN {FS = ":.*## "}; /^[a-zA-Z_-]+:.*## / {printf "  %-12s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the CLI into ./bin.
	@mkdir -p $(BIN_DIR)
	$(GO) build -o $(BIN_PATH) $(CMD_PATH)

test: ## Run all tests.
	$(GO) test ./...

fmt: ## Format Go source files.
	$(GO) fmt ./...

tidy: ## Tidy module dependencies.
	$(GO) mod tidy

install: ## Install the CLI into GOBIN/bin.
	$(GO) install $(CMD_PATH)

run: ## Run the CLI. Pass ARGS='...'.
	$(GO) run $(CMD_PATH) $(ARGS)

version: ## Print the CLI version.
	$(GO) run $(CMD_PATH) version

clean: ## Remove local build outputs.
	rm -rf $(BIN_DIR)
