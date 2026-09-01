# macOS Safety & Argument Sanitization Rules

1. **Strict Input Sanitization**: All user-supplied inputs (`output_path`, `source_dir`, `volume_name`, `signing_identity`, `notarize_credentials`) must pass validation before being passed to any subprocess.
2. **Command Injection Prevention**:
   - Disallow null bytes (`\x00`) in all configuration strings.
   - Disallow paths starting with `-` after `filepath.Clean` to prevent flag injection into macOS utilities.
3. **Format & Filesystem Invariants**:
   - `SandboxSafe` must never be used with `APFS` filesystem (enforced by `ErrSandboxAPFS`).
   - `OutputPath` must always end with `.dmg` (enforced by `ErrImageFileExt`).
4. **Temporary Directory Isolation**: All intermediate files and mounts must be created inside isolated temporary directories (`os.MkdirTemp` / `t.TempDir()`) and removed on completion or failure.
