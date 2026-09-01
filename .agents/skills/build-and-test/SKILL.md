---
name: build-and-test
description: >-
  Build and test mkdmg locally. Use this skill when the user asks to compile the binary,
  generate version strings, run unit tests, or execute the test suite with race detection.
---

# Build & Test Procedure for `mkdmg`

This runbook describes how to build, generate assets, and run tests for `mkdmg`.

---

## 1. Prerequisites & Environment

Ensure Go 1.26+ is available in `PATH`:
```sh
export PATH="/opt/homebrew/bin:/usr/local/bin:/usr/local/go/bin:$HOME/go/bin:$PATH"
```

---

## 2. Code Generation (Required Step)

`mkdmg` embeds version information generated from git:
```sh
go generate ./...
```
*Note: This executes `internal/version/generate_version.sh` and populates `internal/version/version.txt`.*

---

## 3. Compiling the Binary

To compile the `mkdmg` executable in the root directory:
```sh
go build -o mkdmg .
```

To verify the compiled binary:
```sh
./mkdmg --version
./mkdmg --help
```

---

## 4. Running the Test Suite

### Running Unit & Integration Tests
```sh
go test -v -race ./...
```

### Running with Coverage
```sh
go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...
go tool cover -func=coverage.txt
```

### Running Specific Tests
```sh
# Run version package tests
go test -v ./internal/version

# Run CLI integration tests
go test -v -run TestVersionEmbedding
```
