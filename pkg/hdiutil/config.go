package hdiutil

import (
	"fmt"
	"path/filepath"
	"strings"
)

// OptFn is a function type that returns a value of type T when called.
// It is used to lazily compute configuration options after validation.
type OptFn[T string | []string] func() T

// Config holds the configuration for creating a DMG disk image.
type Config struct {
	// VolumeName is the name of the mounted volume. If empty, it defaults to the output filename without extension.
	VolumeName string
	// VolumeSizeMb specifies the volume size in megabytes. If zero, hdiutil determines the size automatically.
	VolumeSizeMb int64
	// SandboxSafe enables sandbox-safe mode. Cannot be used with APFS filesystem.
	SandboxSafe bool
	// Bless marks the volume as bootable.
	Bless bool
	// FileSystem specifies the filesystem type (e.g., "HFS+", "APFS"). Defaults to "HFS+".
	FileSystem string
	SigningIdentity string
	// NotarizeCredentials contains credentials for Apple notarization.
	NotarizeCredentials string
	// ImageFormat specifies the DMG format (e.g., "UDZO", "UDBZ", "ULFO", "ULMO"). Defaults to "UDZO".
	ImageFormat string

	// HDIUtilVerbosity controls the verbosity level of hdiutil output.
	HDIUtilVerbosity int

	// OutputPath is the destination path for the created DMG file. Must have .dmg extension.
	OutputPath string
	// SourceDir is the directory containing files to include in the DMG.
	SourceDir string

	// Simulate enables dry-run mode without actually creating the DMG.
	Simulate bool

	valid bool

	// FilesystemOpts returns the hdiutil arguments for the configured filesystem.
	// Only available after calling Validate.
	FilesystemOpts OptFn[[]string]
	// ImageFormatOpts returns the hdiutil arguments for the configured image format.
	// Only available after calling Validate.
	ImageFormatOpts OptFn[[]string]
	// VolumeSizeOpts returns the hdiutil arguments for the configured volume size.
	// Only available after calling Validate.
	VolumeSizeOpts OptFn[[]string]
	// VolumeNameOpt returns the resolved volume name.
	// Only available after calling Validate.
	VolumeNameOpt OptFn[string]
}

// Validate checks the configuration for errors and initializes the option functions.
// It must be called before using FilesystemOpts, ImageFormatOpts, VolumeSizeOpts, or VolumeNameOpt.
// Returns an error if:
//   - SourceDir is empty
//   - OutputPath does not have a .dmg extension
//   - ImageFormat is invalid
//   - FileSystem is invalid
//   - SandboxSafe is enabled with APFS filesystem
func (c *Config) Validate() error {
	c.valid = false
	if len(c.SourceDir) == 0 {
		return ErrInvSourceDir
	}

	if filepath.Ext(c.OutputPath) != ".dmg" {
		return ErrImageFileExt
	}

	if len(c.imageFormatToOpts()) == 0 {
		return ErrInvFormatOpt
	}

	if len(c.filesystemToOpts()) == 0 {
		return ErrInvFilesystemOpt
	}

	// sandbox safe and APFS are mutually exclusive
	if c.SandboxSafe && strings.ToUpper(c.FileSystem) == "APFS" {
		return ErrSandboxAPFS
	}

	c.valid = true

	c.FilesystemOpts = c.validWrapper(c.filesystemToOpts)
	c.ImageFormatOpts = c.validWrapper(c.imageFormatToOpts)
	c.VolumeSizeOpts = c.validWrapper(c.volumeSizeToOpts)
	c.VolumeNameOpt = c.validWrapperStr(c.volumeNameToOpt)

	return nil
}

// volumeNameToOpt returns the volume name, defaulting to the output filename without extension.
func (c *Config) volumeNameToOpt() string {
	if len(c.VolumeName) == 0 {
		return strings.TrimSuffix(filepath.Base(c.OutputPath), ".dmg")
	} else {
		return c.VolumeName
	}
}

// validWrapper wraps a function to ensure Validate has been called before execution.
// Panics if called before validation.
func (c *Config) validWrapper(fn func() []string) OptFn[[]string] {
	return func() []string {
		if !c.valid {
			panic("state is corrupted, Validate() must be called first")
		}
		return fn()
	}
}

// validWrapperStr wraps a string-returning function to ensure Validate has been called before execution.
// Panics if called before validation.
func (c *Config) validWrapperStr(fn func() string) OptFn[string] {
	return func() string {
		if !c.valid {
			panic("state is corrupted, Validate() must be called first")
		}
		return fn()
	}
}

// filesystemToOpts returns the hdiutil arguments for the configured filesystem.
// Supports "HFS+" (default) and "APFS". Returns nil for unsupported filesystems.
func (c *Config) filesystemToOpts() []string {
	switch strings.ToUpper(c.FileSystem) {
	case "", "HFS+":
		return []string{"-fs", "HFS+", "-fsargs", "-c c=64,a=16,e=16"}
	case "APFS":
		return []string{"-fs", "APFS"}
	default:
		return nil
	}
}

// imageFormatToOpts returns the hdiutil arguments for the configured image format.
// Supports "UDZO" (default), "UDBZ", "ULFO", and "ULMO". Returns nil for unsupported formats.
func (c *Config) imageFormatToOpts() []string {
	format := strings.ToUpper(c.ImageFormat)
	switch format {
	case "", "UDZO":
		return []string{"-format", "UDZO", "-imagekey", "zlib-level=9"}
	case "UDBZ":
		return []string{"-format", "UDBZ", "-imagekey", "bzip2-level=9"}
	case "ULFO", "ULMO":
		return []string{"-format", format}
	default:
		return nil
	}
}

// volumeSizeToOpts returns the hdiutil arguments for the configured volume size.
// Returns nil if VolumeSizeMb is zero or negative.
func (c *Config) volumeSizeToOpts() []string {
	if c.VolumeSizeMb > 0 {
		return []string{"-size", fmt.Sprintf("%dm", c.VolumeSizeMb)}
	}

	return nil
}
