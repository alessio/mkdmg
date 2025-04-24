#!/usr/bin/make -f

all: mkdmg check

mkdmg: generate
	go build

check:
	go test -v -race ./...

generate: mod-tidy
	go generate ./internal/...

mod-tidy: go.mod
	go mod tidy

distclean: clean
clean:
	rm -fv mkdmg

.PHONY: all check clean distclean mkdmg mod-tidy generate
