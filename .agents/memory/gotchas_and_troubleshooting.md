# Gotchas, Pitfalls, & Troubleshooting Guide

## 1. Simulation Mode: Flag vs Config Mismatch

> [!WARNING]
> In `main.go`, simulation mode is controlled by the CLI flag `-s` / `--dry-run`:
> ```go
> runner := hdiutil.New(cfg)
> runner.SetSimulate(simulate)
> ```
> `hdiutil.Config` does **not** have a `Simulate` JSON field. If a test or user passes `{"simulate": true}` in a JSON config without passing `-s` on the CLI, `runner.SetSimulate(false)` will run and attempt to execute real `hdiutil` commands.
>
> **Best Practice**: When testing CLI behavior, always include `-s` or `--dry-run` in the argument slice (e.g. `[]string{"mkdmg", "-s", "--config", cfgFile}`).

---

## 2. Sandbox Execution & `hdiutil: Operation not permitted`

- **Symptom**: When running `go test ./...` in a sandboxed terminal or container, tests fail with `hdiutil: create failed - Operation not permitted`.
- **Cause**: macOS sandbox restricts low-level disk manipulation (`/dev/disk*` ioctls and disk mounting).
- **Remedy**:
  - For unit tests: Ensure `-s` (dry-run) or mock executors are used.
  - For actual DMG generation: Run the command in standard environment (outside sandbox or with elevated privileges).

---

## 3. Go Toolchain & Version Requirements

- **Toolchain**: Go 1.26+ (specified in `go.mod` with toolchain directive `go1.26.4`).
- **Path Resolution**: On macOS systems, ensure Go is in PATH:
  ```sh
  export PATH="/opt/homebrew/bin:/usr/local/bin:/usr/local/go/bin:$HOME/go/bin:$PATH"
  ```

---

## 4. `go:generate` Prerequisite

- **Symptom**: Build or tests fail with missing or stale version numbers.
- **Cause**: `internal/version/version.go` embeds `version.txt`. If `go generate ./...` is not run, `version.txt` may remain at `v0.0.0-UNKNOWN`.
- **Remedy**: Always run `make generate` or `go generate ./...` prior to `go build` or `go test`.

---

## 5. Apple Notarization Prerequisites

- **Requirements**:
  - `signing_identity`: Must be a valid "Developer ID Application" certificate installed in macOS Keychain.
  - `notarize_credentials`: Must be a stored Keychain profile created via:
    ```sh
    xcrun notarytool store-credentials "PROFILE_NAME" --apple-id "USER@DOMAIN.COM" --team-id "TEAMID" --password "APP-SPECIFIC-PASSWORD"
    ```
- **Behavior**: If either `signing_identity` or `notarize_credentials` is empty in the config, `Runner.Codesign()` and `Runner.Notarize()` safely act as no-ops.
