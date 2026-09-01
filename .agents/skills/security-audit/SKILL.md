---
name: security-audit
description: >-
  Perform security audits, vulnerability scans, and input sanitization reviews for mkdmg.
  Use this skill when auditing command argument injection risks, running govulncheck/gosec, checking GitHub Actions permissions, or verifying code signing security.
---

# Security Audit Runbook for `mkdmg`

This skill defines the procedures for conducting comprehensive security audits, static analysis checks (SAST), dependency scans, and input sanitization reviews on `mkdmg`.

---

## 1. Vulnerability Scanning (`govulncheck`)

Scan all Go dependencies and code paths for known Common Vulnerabilities and Exposures (CVEs):

```sh
govulncheck ./...
```

*If `govulncheck` is not installed:*
```sh
go install golang.org/x/vuln/cmd/govulncheck@latest
```

---

## 2. Static Security Analysis (`gosec`)

Run `gosec` to identify common security flaws in Go source code (such as file inclusion via variables `G304`, command injection `G204`, unhandled errors):

```sh
gosec -exclude-generated ./...
```

### Auditing Justified `#nosec` Directives
Inspect all `#nosec` annotations in the codebase (e.g. in `main.go` for `G304`) to verify that the target path is sanitized with `filepath.Clean` and validated before use:
```sh
grep -rn "#nosec" .
```

---

## 3. Command Injection & Sanitization Audit

Verify that user-supplied configuration values cannot manipulate underlying subprocess arguments:

1. **Null Byte Check**: Ensure all string inputs (`output_path`, `source_dir`, `volume_name`, `signing_identity`, `notarize_credentials`) reject null characters (`\x00`).
2. **Flag Injection Check**: Ensure paths starting with a leading dash (`-`) after `filepath.Clean` are rejected.
3. **No Raw Shell Interpolation**: Verify that commands are executed using parameter slices (`exec.Command(binary, arg1, arg2...)`) and never via shell string expansion (`sh -c` or `bash -c`).

---

## 4. CI/CD & Supply Chain Security Audit

1. **GitHub Workflow Permissions**: Verify that all workflows under `.github/workflows/` declare minimal `permissions:` blocks (e.g., `contents: read`).
2. **Action Pinning**: Ensure third-party GitHub Actions are pinned to stable major or SHA versions.
3. **Dependency Review**: Confirm `.github/workflows/dependency-review.yml` and `.github/workflows/codacy.yml` are active on pull requests.

---

## 5. macOS Code Signing & Notarization Audit

1. **Runtime Hardening**: Confirm that code signing enforces the hardened runtime flag (`--options runtime`).
2. **Strict Verification**: Confirm that signature verification uses `--deep --strict`.
3. **Keychain Credential Protection**: Verify that Apple ID passwords are never passed as CLI flags or environment variables, but retrieved securely via `xcrun notarytool --keychain-profile`.
