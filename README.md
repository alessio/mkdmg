<div align="center">

# 💿 mkdmg

**The Fancy Apple Disk Image Creator**

[![Go](https://github.com/alessio/mkdmg/actions/workflows/go.yml/badge.svg)](https://github.com/alessio/mkdmg/actions/workflows/go.yml)
[![GoDoc](https://godoc.org/al.essio.dev/cmd/mkdmg?status.svg)](https://pkg.go.dev/al.essio.dev/cmd/mkdmg)
[![Go Report Card](https://goreportcard.com/badge/github.com/alessio/mkdmg)](https://goreportcard.com/report/github.com/alessio/mkdmg)
[![License](https://img.shields.io/github/license/alessio/mkdmg.svg)](https://github.com/alessio/mkdmg/blob/main/LICENSE)
[![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/alessio/mkdmg)](https://github.com/alessio/mkdmg/releases)

<p align="center">
  <br />
  <b>mkdmg</b> is a powerful, modern CLI wrapper around <code>hdiutil</code> designed to make creating, signing, and notarizing macOS Disk Images (<code>.dmg</code>) effortless.
  <br />
</p>

</div>

---

## ✨ Features

- 🚀 **Simple:** Create DMGs with a single command.
- ⚙️ **Configurable:** JSON configuration support for reproducible builds.
- 📦 **Formats:** Supports multiple DMG formats (`UDZO`, `UDBZ`, `ULFO`, `ULMO`).
- 🔐 **Security:** Integrated codesigning and notarization workflow.
- 🖥️ **Filesystems:** Support for both HFS+ and APFS.
- 🛡️ **Sandbox:** Create sandbox-safe disk images.

## 📦 Installation

### Pre-built Binaries

You can download the latest pre-built binaries for macOS (Darwin) from the [GitHub Releases page](https://github.com/alessio/mkdmg/releases).

1.  Visit the [releases page](https://github.com/alessio/mkdmg/releases).
2.  Download the archive matching your architecture (`x86_64` or `arm64`).
3.  Extract the archive and move the `mkdmg` binary to a directory in your `PATH` (e.g., `/usr/local/bin`).

### From Source

Requires Go 1.26 or later.

```sh
go install al.essio.dev/cmd/mkdmg@latest
```

To build from a local checkout:

```sh
make build
```

### Verification

To verify the integrity of the downloaded binary, you can use the `checksums.txt` file provided in the [GitHub Releases](https://github.com/alessio/mkdmg/releases).

1. Download the binary archive and the `checksums.txt` file.
2. Run the following command to verify the checksum:

```sh
sha256sum -c checksums.txt --ignore-missing
```

## 🚀 Usage

The basic syntax is intuitive:

```sh
mkdmg [OPTION]... OUTFILE.DMG DIRECTORY
```

> **Note:** All flags must be specified **before** the positional arguments.

Or make it reproducible using a configuration file:

```sh
mkdmg --config config.json
```

You can also combine a configuration file with positional arguments to override the output path and source directory:

```sh
mkdmg --config config.json OUTFILE.DMG DIRECTORY
```

### Example

Create a 100MB DMG named "My App.dmg" with the volume name "My App v1.0" from the contents of the `./build` directory:

```sh
mkdmg \
  --volname "My App v1.0" \
  --disk-image-size 100 \
  "My App.dmg" \
  ./build
```

## ⚙️ Options

Here is a list of all available command-line flags:

| Flag | Shorthand | Description | Default |
| :--- | :---: | :--- | :--- |
| `--config` | | Path to a JSON configuration file. | `""` |
| `--volname` | | Set the volume name for the DMG. | `<filename>` |
| `--disk-image-size` | | Set the size for the DMG in megabytes (MB). | `0` |
| `--codesign` | | Provide a signing identity to codesign the final DMG. | `""` |
| `--notarize` | | Provide keychain-stored credentials to notarize the DMG. | `""` |
| `--bless` | | Bless the disk image. | `false` |
| `--apfs` | | Use APFS as the disk image's filesystem. | `false` (HFS+) |
| `--sandbox-safe` | | Create a sandbox-safe DMG. | `false` |
| `--format` | | Specify the final disk image format (`UDZO`, `UDBZ`, `ULFO`, `ULMO`). | `UDZO` |
| `--dry-run` | `-s` | Simulate the process without creating any files. | `false` |
| `--hdiutil-verbosity` | | Set `hdiutil` verbosity (0=default, 1=quiet, 2=verbose, 3=debug). | `0` |
| `--verbose` | `-v` | Enable verbose output for `mkdmg`. | `false` |
| `--version` | `-V` | Print version information and exit. | `false` |
| `--help` | `-h` | Display the help message and exit. | `false` |

## 📄 JSON Configuration

`mkdmg` can also be fully configured using a JSON file.

### Example `config.json`

```json
{
  "volume_name": "MyApplication",
  "volume_size_mb": 0,
  "sandbox_safe": false,
  "bless": false,
  "filesystem": "HFS+",
  "signing_identity": "",
  "notarize_credentials": "",
  "image_format": "UDZO",
  "hdiutil_verbosity": 0,
  "output_path": "./dist/MyApplication.dmg",
  "source_dir": "./build/Release",
  "simulate": false
}
```

### Configuration Reference

| Field | Type | Default | Description |
| :--- | :--- | :--- | :--- |
| `volume_name` | `string` | *(Filename)* | Name of the mounted volume. Defaults to `output_path` filename if empty. |
| `volume_size_mb` | `number` | `0` | Volume size in megabytes. If `0`, `hdiutil` calculates the minimum size automatically. |
| `sandbox_safe` | `boolean` | `false` | Enables sandbox-safe mode. **Incompatible with `APFS`**. |
| `bless` | `boolean` | `false` | Blesses the folder/volume (makes it bootable/auto-open). |
| `filesystem` | `string` | `"HFS+"` | Filesystem type. Options: `"HFS+"`, `"APFS"`. |
| `signing_identity` | `string` | `""` | Name or hash of the code signing identity to use. |
| `notarize_credentials` | `string` | `""` | Profile name for `notarytool` credentials. |
| `image_format` | `string` | `"UDZO"` | DMG format. Options: `"UDZO"` (zlib), `"UDBZ"` (bzip2), `"ULFO"` (lzfse), `"ULMO"` (lzma). |
| `hdiutil_verbosity` | `number` | `0` | Verbosity level for the underlying `hdiutil` command. |
| `output_path` | `string` | *(Required)* | Destination path for the `.dmg` file. |
| `source_dir` | `string` | *(Required)* | Directory containing files to package. |
| `simulate` | `boolean` | `false` | If `true`, prints commands without executing them. |

---

<div align="center">
  Made with ❤️ by <a href="https://github.com/alessio">Alessio Treglia</a>
</div>