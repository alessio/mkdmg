# mkdmg

[![Go](https://github.com/alessio/mkdmg/actions/workflows/go.yml/badge.svg)](https://github.com/alessio/mkdmg/actions/workflows/go.yml)
[![Codacy Badge](https://app.codacy.com/project/badge/Grade/gh/alessio/mkdmg)](https://www.codacy.com/gh/alessio/mkdmg/dashboard?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=alessio/mkdmg&amp;utm_campaign=Badge_Grade)
[![codecov](https://codecov.io/gh/alessio/mkdmg/branch/main/graph/badge.svg)](https://codecov.io/gh/alessio/mkdmg)
[![Go Report Card](https://goreportcard.com/badge/github.com/alessio/mkdmg)](https://goreportcard.com/report/github.com/alessio/mkdmg)
[![License](https://img.shields.io/github/license/alessio/mkdmg.svg)](https://github.com/alessio/mkdmg/blob/main/LICENSE)
[![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/alessio/mkdmg)](https://github.com/alessio/mkdmg/releases)

`mkdmg` is a command-line tool to build fancy Apple Disk Images (`.dmg`) on macOS. It acts as a powerful wrapper around `hdiutil` and other command-line tools to simplify the process of creating, signing, and notarizing DMGs.

## Installation

You can install `mkdmg` using `go install`:
```sh
go install al.essio.dev/cmd/mkdmg@latest
```

## Usage

The basic syntax is:
```sh
mkdmg [OPTIONS]... <output.dmg> <source_directory>
```

Or using a configuration file:
```sh
mkdmg --config config.json
```

### Example

Create a 100MB DMG named "My App.dmg" with the volume name "My App v1.0" from the contents of the `./build` directory:
```sh
mkdmg --volname "My App v1.0" --disk-image-size 100 "My App.dmg" ./build
```

## Options



Here is a list of all available command-line flags:



| Flag                  | Shorthand | Description                                                                 | Default      |

| --------------------- | --------- | --------------------------------------------------------------------------- | ------------ |

| `--config`            |           | Path to a JSON configuration file.                                          | ""           |

| `--volname`           |           | Set the volume name for the DMG.                                            | `<filename>` |

| `--disk-image-size`   |           | Set the size for the DMG in megabytes (MB).                                 | 0            |

| `--codesign`          |           | Provide a signing identity to codesign the final DMG.                       | ""           |

| `--notarize`          |           | Provide keychain-stored credentials to notarize the DMG.                    | ""           |

| `--bless`             |           | Bless the disk image.                                                       | `false`      |

| `--apfs`              |           | Use APFS as the disk image's filesystem.                                    | `false` (HFS+)|

| `--sandbox-safe`      |           | Create a sandbox-safe DMG.                                                  | `false`      |

| `--format`            |           | Specify the final disk image format (`UDZO`, `UDBZ`, `ULFO`, `ULMO`).       | `UDZO`       |

| `--dry-run`           | `-s`      | Simulate the process without creating any files.                            | `false`      |

| `--hdiutil-verbosity` |           | Set `hdiutil` verbosity (0=default, 1=quiet, 2=verbose, 3=debug).             | 0            |

| `--verbose`           | `-v`      | Enable verbose output for `mkdmg`.                                          | `false`      |

| `--version`           | `-V`      | Print version information and exit.                                         | `false`      |

| `--help`              | `-h`      | Display the help message and exit.                                          | `false`      |

## JSON Configuration

`mkdmg` can also be configured using a JSON file.

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


