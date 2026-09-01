---
name: code-quality
description: >-
  Maintain code quality, formatting, linting, and security standards for mkdmg.
  Use this skill when formatting code, running golangci-lint, tidying go modules, or checking for security vulnerabilities.
---

# Code Quality & Linting Guide for `mkdmg`

This skill provides step-by-step procedures to maintain code health, formatting, lint compliance, and dependency integrity.

---

## 1. Code Formatting

Format all Go source files according to Go conventions:
```sh
go fmt ./...
```

---

## 2. Module Tidying & Dependency Verification

Prune unused dependencies and verify checksums:
```sh
go mod tidy
go mod verify
```

---

## 3. Running Linter (`golangci-lint`)

Run `golangci-lint` to catch static analysis issues:
```sh
golangci-lint run
```

To run with verbose output:
```sh
golangci-lint run -v
```

---

## 4. Security Scanning & Vulnerability Checks

Run Go's official vulnerability checker:
```sh
govulncheck ./...
```

Check for potential file path injection or insecure argument handling before committing changes.
