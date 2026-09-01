# mkdmg Development & Agent Directives

Welcome to **mkdmg** (The Fancy Apple Disk Image Creator).
This repository is a Go-based CLI utility that wraps macOS disk image management tools (`hdiutil`, `codesign`, `notarytool`, `stapler`, `bless`).

---

## 🧭 Project Quick Context

- **Module**: `al.essio.dev/cmd/mkdmg`
- **Go Toolchain**: Go 1.26+ (standard library `flag`, `os`, `path/filepath`, `//go:embed`)
- **Core Dependency**: `al.essio.dev/pkg/hdiutil` (`Config`, `Runner`, sentinel errors)
- **Target OS**: macOS (`darwin/amd64`, `darwin/arm64`)
- **Build System**: `Makefile` & GoReleaser

---

## 👥 Multi-Agent Team

When working on this repository, you can adopt or delegate to any of the specialized agent personas located in [`.agents/team/`](file:///Users/alessio/Documents/src/mkdmg/.agents/team/team_manifest.md):

1. **Lead Architect** ([`architect.md`](file:///Users/alessio/Documents/src/mkdmg/.agents/team/architect.md)): Architectural evolution, CLI ergonomics, JSON schema stability, modularity.
2. **Core Go Developer** ([`core_developer.md`](file:///Users/alessio/Documents/src/mkdmg/.agents/team/core_developer.md)): Go CLI implementation, flag handling, error wrapping, performance.
3. **macOS Platform Specialist** ([`macos_specialist.md`](file:///Users/alessio/Documents/src/mkdmg/.agents/team/macos_specialist.md)): `hdiutil`, APFS/HFS+ file systems, sandbox-safe images, codesign, Apple notarization.
4. **QA & Reliability Engineer** ([`qa_engineer.md`](file:///Users/alessio/Documents/src/mkdmg/.agents/team/qa_engineer.md)): Unit tests, mock executors, simulation checks, regression testing, race detection.
5. **Security Auditor** ([`security_auditor.md`](file:///Users/alessio/Documents/src/mkdmg/.agents/team/security_auditor.md)): SAST, argument sanitization, CVE vulnerability scans, secrets protection.
6. **Release & DevSecOps Engineer** ([`release_engineer.md`](file:///Users/alessio/Documents/src/mkdmg/.agents/team/release_engineer.md)): GoReleaser, GitHub Actions CI/CD, release integrity, checksum verification.
7. **Documentation & DX Engineer** ([`doc_engineer.md`](file:///Users/alessio/Documents/src/mkdmg/.agents/team/doc_engineer.md)): Documentation, `README.md`, `docs/index.html`, GoDoc, CLI usage help.

---

## ⚡ Key Workflows & Skills

Skills are located in [`.agents/skills/`](file:///Users/alessio/Documents/src/mkdmg/.agents/skills/):
- **`build-and-test`**: Run code generation, compile binaries, and execute unit/integration tests with race detection.
- **`mock-hdiutil-testing`**: Write isolated tests using mock `CommandExecutor` without needing macOS disk privileges.
- **`macos-signing-notarization`**: Procedures for codesigning, notarizing with `notarytool`, and stapling tickets.
- **`security-audit`**: Perform vulnerability checks (`govulncheck`, `gosec`), verify argument sanitization, and audit workflow security.
- **`release-workflow`**: Run GoReleaser dry-runs, git semantic tagging, and checksum verification.
- **`code-quality`**: Run `golangci-lint`, formatting (`go fmt`), `go mod tidy`, and code standards checks.

---

## 🛑 Golden Rules & Gotchas

1. **Dry-Run & Simulation Precedence**: `hdiutil.Config` parses build settings from JSON, but simulation mode is toggled via CLI flags (`-s`, `--dry-run`) or `runner.SetSimulate(true)`. Tests should either pass `-s` or use mock executors to avoid attempting real `hdiutil` calls.
2. **Sandbox Awareness**: macOS `hdiutil` operations require elevated disk permissions and will fail in restricted sandbox environments. Always use simulation or mocks for sandboxed test suites.
3. **Command Sanitization**: Never bypass `hdiutil.Config.Validate()`. All paths and arguments must be sanitized against null bytes and leading dashes to prevent command injection.
4. **Version Generation**: Always run `go generate ./...` before building or testing to populate `internal/version/version.txt`.

For detailed architecture, refer to [`.agents/memory/project_overview.md`](file:///Users/alessio/Documents/src/mkdmg/.agents/memory/project_overview.md).
