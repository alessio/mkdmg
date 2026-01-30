#!/usr/bin/make -f

BIN := mkdmg
GO := go
PKGS := ./...

.PHONY: all
all: build check

.PHONY: build
build: generate
	$(GO) build -o $(BIN)

.PHONY: install
install: generate
	$(GO) install

.PHONY: check
check: generate
	$(GO) test -v -race $(PKGS)

.PHONY: generate
generate:
	$(GO) generate ./internal/...

.PHONY: lint
lint:
	golangci-lint run

.PHONY: fmt
fmt:
	$(GO) fmt $(PKGS)

.PHONY: mod-tidy
mod-tidy:
	$(GO) mod tidy

.PHONY: clean
clean:
	rm -f $(BIN)
	rm -f coverage.out

.PHONY: distclean
distclean: clean