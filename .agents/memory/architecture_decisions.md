# Architecture Decision Records (ADRs) & Design Memory

## ADR 1: Separation of CLI Frontend (`mkdmg`) and Engine (`hdiutil`)
- **Context**: DMG creation involves many macOS-specific edge cases, format flags, and security rules.
- **Decision**: Keep `mkdmg` strictly as a lightweight CLI frontend that parses flags, reads JSON config, and delegates all disk orchestration to the `al.essio.dev/pkg/hdiutil` package.
- **Benefits**: Simplifies CLI maintenance, enables independent testing of the library, and allows programmatic reuse of `hdiutil` in other Go applications.

---

## ADR 2: Security Sanitization & Argument Injection Prevention
- **Context**: The underlying runner invokes macOS command line utilities (`hdiutil`, `codesign`, `xcrun`). Malformed or malicious paths in user configs could result in argument injection.
- **Decision**: Enforce strict sanitization inside `hdiutil.Config.Validate()`:
  1. Reject any string fields containing null bytes (`\x00`).
  2. Reject paths (`SourceDir`, `OutputPath`) that begin with `-` after `filepath.Clean`.
  3. Reject unsafe combinations (e.g. `SandboxSafe` with `APFS`).
- **Consequence**: Users cannot pass raw flags through path parameters. All options must be declared through structured JSON fields or CLI flags.

---

## ADR 3: Version Embedding via `go:generate` and `git describe`
- **Context**: Need exact semantic release versions and development build commit hashes embedded in the binary without requiring manual edits.
- **Decision**: Use a shell script (`generate_version.sh`) triggered via `//go:generate bash generate_version.sh` to write `git describe --abbrev=6` into `version.txt`, which is embedded via Go's `//go:embed` directive into `internal/version`.
- **Fallback**: Defaults to `v0.0.0-UNKNOWN` if git is not present or not in a git repository.

---

## ADR 4: Mock Executor Pattern for Testing
- **Context**: `hdiutil` requires macOS root/disk privileges and cannot run inside restricted sandboxes or Linux build containers.
- **Decision**: `al.essio.dev/pkg/hdiutil` provides a `CommandExecutor` interface with typed methods (`Hdiutil`, `Codesign`, `Xcrun`, `Chmod`, `Bless`).
- **Guidelines**:
  - Unit tests that verify command construction should use `WithExecutor(mockExecutor)`.
  - CLI tests in `mkdmg` should use `--dry-run`/`-s` or mock executors to prevent invoking real `hdiutil` in restricted test environments.

---

## ADR 5: Two-Stage Sandbox-Safe DMG Generation
- **Context**: Sandboxed macOS applications require disk images created with ISO9660/HFS hybrid structures rather than pure HFS+ disk images.
- **Decision**: When `SandboxSafe` is enabled, `hdiutil` generates the image using `hdiutil makehybrid -hfs -iso -joliet` and converts it with `hdiutil convert -format UDZO`.
- **Constraint**: `SandboxSafe` is strictly incompatible with `APFS` and will return `ErrSandboxAPFS` if both are requested.
