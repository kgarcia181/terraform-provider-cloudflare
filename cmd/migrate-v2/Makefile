# Makefile for migrate-v2

BINARY_NAME=migrate-v2
GO=go
GOFLAGS=-v
LDFLAGS=

.PHONY: setup
setup:
	@echo "Setting up submodules..."
	@./setup.sh

.PHONY: build
build:
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) ./cmd

.PHONY: clean
clean:
	rm -f $(BINARY_NAME)

.PHONY: run
run: build
	./$(BINARY_NAME) -help

.PHONY: test
test:
	$(GO) test $(GOFLAGS) ./...

.PHONY: fmt
fmt:
	$(GO) fmt ./...

.PHONY: vet
vet:
	$(GO) vet ./...

.PHONY: all
all: fmt vet test build

# Example targets for common use cases
.PHONY: example-dryrun
example-dryrun: build
	./$(BINARY_NAME) -dryrun -config ./examples -state ./examples/terraform.tfstate

.PHONY: example-config
example-config: build
	./$(BINARY_NAME) -config ./examples

.PHONY: example-state
example-state: build
	./$(BINARY_NAME) -state ./examples/terraform.tfstate

.DEFAULT_GOAL := build