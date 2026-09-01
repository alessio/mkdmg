# `mkdmg` Project Overview & Architecture Memory

## 1. Executive Summary

**mkdmg** is a fast, modern CLI wrapper around macOS disk image and security utilities (`hdiutil`, `codesign`, `notarytool`, `stapler`, `bless`). It was designed to replace complex bash scripts and streamline the packaging, signing, and notarization of macOS applications into distribution-ready `.dmg` files.

- **Repository**: `github.com/alessio/mkdmg`
- **Go Module**: `al.essio.dev/cmd/mkdmg`
- **Author**: Alessio Treglia <alessio@debian.org>
- **License**: MIT

---

## 2. Codebase Architecture

```text
mkdmg/
├── .github/
│   ├── workflows/             # CI/CD Workflows
│   │   ├── go.yml             # Go build, test with race detector, Codecov upload
│   │   ├── lint.yml           # golangci-lint
│   │   ├── release.yml        # GoReleaser automated tag releases
│   │   ├── codacy.yml         # Security & code quality scanning
│   │   ├── dependency-review.yml # Dependency review
│   │   └── static.yml         # GitHub Pages hosting for docs/
│   └── CODEOWNERS             # Ownership definition
├── docs/
│   └── index.html             # Web landing page & documentation
├── internal/
│   └── version/               # Version embedding package
│       ├── generate_version.sh# Generates version.txt from git describe
│       ├── version.go         # //go:embed version.txt and Version() function
│       ├── version.txt        # Embedded version file
│       └── version_test.go    # Unit tests for version package
├── .goreleaser.yaml           # GoReleaser configuration for universal macOS binaries
├── Makefile                   # Developer shortcuts (build, test, lint, clean)
├── go.mod                     # Go module definitions (Go 1.26+)
├── go.sum                     # Checksums for external dependencies
├── main.go                    # CLI entrypoint, flag parsing, config loading, runner orchestration
├── main_test.go               # Unit and integration tests for CLI flags and configuration
└── version_integration_test.go# Integration test verifying version generation and binary embedding
```

---

## 3. Core Engine Dependency (`al.essio.dev/pkg/hdiutil`)

`mkdmg` relies on `al.essio.dev/pkg/hdiutil` for all low-level DMG operations.

### Key Types & APIs
- **`hdiutil.Config`**: Struct representing the JSON configuration file (`mkdmg.json`).
  - `VolumeName` (`string`): Name of the mounted volume (defaults to output filename without extension).
  - `VolumeSizeMb` (`int64`): Volume size in MB (0 = automatic size calculation by `hdiutil`).
  - `SandboxSafe` (`bool`): Uses `hdiutil makehybrid` + `convert` for sandbox compliance (incompatible with APFS).
  - `Bless` (`bool`): Runs `bless --folder <mount> --openfolder <mount>` to auto-open folder when mounted.
  - `FileSystem` (`string`): `"HFS+"` (default with tuned `-fsargs -c c=64,a=16,e=16`) or `"APFS"`.
  - `SigningIdentity` (`string`): Identity for `codesign --sign <identity> --deep --strict <dmg>`.
  - `NotarizeCredentials` (`string`): Keychain profile name for `xcrun notarytool submit --keychain-profile <profile> --wait <dmg>`.
  - `ImageFormat` (`string`): Compression format (`"UDZO"` zlib default, `"UDBZ"` bzip2, `"ULFO"` lzfse, `"ULMO"` lzma).
  - `HDIUtilVerbosity` (`int`): 0 = normal, 1 = quiet, 2 = verbose, 3+ = debug.
  - `OutputPath` (`string`): Destination path (must end with `.dmg`).
  - `SourceDir` (`string`): Source directory to package.
- **`hdiutil.Runner`**: Lifecycle coordinator executing:
  1. `Setup()`: Validates configuration, checks inputs, creates temporary staging directory.
  2. `Start()`: Creates initial writable disk image from `SourceDir`.
  3. `AttachDiskImage()`: Mounts writable image in temporary mount point.
  4. `Bless()`: Applies folder blessing if configured.
  5. `DetachDiskImage()`: Detaches/unmounts disk image safely.
  6. `FinalizeDMG()`: Converts writable image to target compressed format (`UDZO`, `ULFO`, etc.).
  7. `Codesign()`: Code signs DMG if `SigningIdentity` is provided.
  8. `Notarize()`: Submits DMG to Apple notary service and staples ticket if `NotarizeCredentials` is provided.
  9. `Cleanup()`: Cleans up temporary working directories.

---

## 4. CLI Execution & Flag Precedence

- **Flag Parsing**: Standard Go `flag` library.
  - `--config <path>`: Path to JSON configuration file (default `mkdmg.json`).
  - `--dry-run`, `-s`: Toggles simulation mode without creating files.
  - `--verbose`, `-v`: Enables verbose logging to `stderr`.
  - `--help`, `-h`: Shows usage message and defaults.
  - `--version`, `-V`: Outputs version information and copyright.
- **Positional Arguments**:
  - `mkdmg [OUTFILE.DMG [DIRECTORY]]`
  - Positional arguments override the values specified in the configuration file.
  - Specifying more than 2 positional arguments produces an error.

---

## 5. Build & CI/CD Pipeline

- **Local Build**:
  - `make build` -> runs `go generate ./internal/...` then `go build -o mkdmg`.
  - `make check` -> runs `go generate ./internal/...` then `go test -v -race ./...`.
  - `make lint` -> runs `golangci-lint run`.
  - `make clean` -> removes binary, coverage files, and resets `version.txt`.
- **Release Matrix**:
  - GoReleaser builds universal macOS binaries: `darwin/amd64` and `darwin/arm64`.
  - Releases generate `checksums.txt` and package tarballs (`mkdmg_Darwin_x86_64.tar.gz`, `mkdmg_Darwin_arm64.tar.gz`).
