# Makefile
SHELL := /bin/bash

# Configurable vars
BINARY ?= api
PKG ?= ./cmd/api
BUILD_DIR := ./bin
GO ?= go
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -X 'main.version=$(VERSION)'

.PHONY: help all build run test fmt vet lint deps clean

help:
	@printf "Usage: make [target]\n\nTargets:\n  help     Show this help\n  build    Build binary into $(BUILD_DIR)\n  run      Build and run\n  test     Run unit tests\n  fmt      gofmt (via go fmt)\n  vet      go vet\n  lint     run golangci-lint (must be installed)\n  deps     download modules\n  clean    remove build artifacts\n"

all: build

build: deps fmt vet
	@echo "Building $(BINARY) (version: $(VERSION))..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY) $(PKG)

run: build
	@echo "Running $(BINARY)..."
	$(BUILD_DIR)/$(BINARY)

# TODO: make test unit
test:
	$(GO) test ./... -v

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

lint:
	@command -v golangci-lint >/dev/null 2>&1 || (echo "golangci-lint not found. Install: https://golangci-lint.run/usage/install/" && exit 1)
	golangci-lint run

deps:
	$(GO) mod download

clean:
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)
