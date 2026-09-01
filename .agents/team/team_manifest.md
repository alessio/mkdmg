# `mkdmg` Multi-Agent Engineering Squad Manifest

This document outlines the specialized team of AI agents designed to build, maintain, test, and release the **mkdmg** project. Each agent has dedicated responsibilities, tools, and collaboration protocols.

---

## 👥 Squad Overview

```text
                               ┌───────────────────────────┐
                               │     Lead Architect        │
                               │        @architect         │
                               └─────────────┬─────────────┘
                                             │
                     ┌───────────────────────┼───────────────────────┐
                     │                       │                       │
       ┌─────────────▼─────────────┐ ┌───────▼───────────┐ ┌─────────▼─────────────┐
       │     Core Go Developer     │ │ macOS Specialist  │ │ QA & Reliability Eng  │
       │        @core-dev          │ │  @macos-platform  │ │      @qa-tester       │
       └─────────────┬─────────────┘ └───────┬───────────┘ └─────────┬─────────────┘
                     │                       │                       │
                     └───────────────────────┼───────────────────────┘
                                             │
                     ┌───────────────────────┼───────────────────────┐
                     │                       │                       │
       ┌─────────────▼─────────────┐ ┌───────▼───────────┐ ┌─────────▼─────────────┐
       │ Release & DevSecOps Eng   │ │ Security Auditor  │ │ Documentation & DX Eng│
       │       @release-ops        │ │   @sec-auditor    │ │       @doc-engineer   │
       └───────────────────────────┘ └───────────────────┘ └───────────────────────┘
```

---

## 📋 Agent Roster & Quick Reference

| Role | Alias | Focus Area | Detailed Persona |
| :--- | :--- | :--- | :--- |
| **Lead Architect** | `@architect` | Architecture, schema design, CLI ergonomics, ADRs | [architect.md](./architect.md) |
| **Core Go Developer** | `@core-dev` | Go CLI implementation, flag parsing, runner wiring | [core_developer.md](./core_developer.md) |
| **macOS Platform Specialist** | `@macos-platform` | `hdiutil`, APFS/HFS+, codesign, notarytool, bless | [macos_specialist.md](./macos_specialist.md) |
| **QA & Reliability Engineer** | `@qa-tester` | Mock testing, unit/integration test suite, race checks | [qa_engineer.md](./qa_engineer.md) |
| **Security Auditor** | `@sec-auditor` | SAST, argument sanitization, CVE scans, secrets protection | [security_auditor.md](./security_auditor.md) |
| **Release & DevSecOps Engineer** | `@release-ops` | GoReleaser, GitHub Actions, Dependabot, release artifacts | [release_engineer.md](./release_engineer.md) |
| **Documentation & DX Engineer** | `@doc-engineer` | README, website docs, manpages, GoDoc, examples | [doc_engineer.md](./doc_engineer.md) |

---

## 🔄 Collaboration & Routing Matrix

### 1. New Feature / CLI Option Request
1. `@architect` evaluates CLI ergonomics, backward compatibility, and JSON schema.
2. `@macos-platform` validates macOS command compatibility (`hdiutil` flags, APFS/HFS+ behavior).
3. `@core-dev` implements the flag in `main.go` and hooks it into `hdiutil.Runner`.
4. `@sec-auditor` reviews the input paths and parameters for argument injection vulnerabilities.
5. `@qa-tester` writes unit tests with dry-run/mock executors and verifies race conditions.
6. `@doc-engineer` updates `README.md`, `docs/index.html`, and usage help text.

### 2. Security Audit & Vulnerability Review
1. `@sec-auditor` runs `govulncheck` and `gosec` static analysis scans.
2. `@sec-auditor` reviews GitHub Actions workflow permissions and Dependabot alerts.
3. `@core-dev` patches any vulnerable dependencies or implements stricter input validation.
4. `@qa-tester` adds regression tests to verify that invalid/malformed inputs are rejected.

### 3. Release & Version Bump
1. `@sec-auditor` and `@release-ops` verify dependency health and CI status.
2. `@release-ops` runs GoReleaser dry run (`goreleaser release --snapshot --clean`).
3. `@release-ops` tags the release (`git tag -a vX.Y.Z -m "Release vX.Y.Z"`).
4. `@doc-engineer` verifies CHANGELOG and release notes.
