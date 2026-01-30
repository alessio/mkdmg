# mkdmg

![Build](https://github.com/alessio/mkdmg/workflows/Go/badge.svg)

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

### Example

Create a 100MB DMG named "My App.dmg" with the volume name "My App v1.0" from the contents of the `./build` directory:
```sh
mkdmg --volname "My App v1.0" --disk-image-size 100 "My App.dmg" ./build
```

## Options



Here is a list of all available command-line flags:



| Flag                  | Shorthand | Description                                                                 | Default      |

| --------------------- | --------- | --------------------------------------------------------------------------- | ------------ |

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

| `--verboseMode`       | `-v`      | Enable verbose output for `mkdmg`.                                          | `false`      |

| `--version`           | `-V`      | Print version information and exit.                                         | `false`      |

| `--help`              | `-h`      | Display the help message and exit.                                          | `false`      |


