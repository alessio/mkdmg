package hdiutil

import (
	"path/filepath"
	"strings"
)

type Config struct {
	VolumeName          string
	VolumeSizeMb        int64
	SandboxSafe         bool
	Bless               bool
	FileSystem          string
	SigningIdentity     string
	NotarizeCredentials string
	ImageFormat         string

	HDIUtilVerbosity int

	OutputPath string
	SourceDir  string

	Simulate bool
}

func (c *Config) validate() error {
	if len(c.SourceDir) == 0 {
		return ErrInvSourceDir
	}

	if filepath.Ext(c.OutputPath) != ".dmg" {
		return ErrImageFileExt
	}

	if len(c.imageFormatToArgs()) == 0 {
		return ErrInvFormatOpt
	}

	if len(c.filesystemToArgs()) == 0 {
		return ErrInvFilesystemOpt
	}

	// sandbox safe and APFS are mutually exclusive
	if c.SandboxSafe && strings.ToUpper(c.FileSystem) == "APFS" {
		return ErrSandboxAPFS
	}

	return nil
}

func (c *Config) filesystemToArgs() []string {
	switch strings.ToUpper(c.FileSystem) {
	case "", "HFS+":
		return []string{"-fs", "HFS+", "-fsargs", "-c c=64,a=16,e=16"}
	case "APFS":
		return []string{"-fs", "APFS"}
	default:
		return nil
	}
}

func (c *Config) imageFormatToArgs() []string {
	switch c.ImageFormat {
	case "", "UDZO":
		return []string{"-format", "UDZO", "-imagekey", "zlib-level=9"}
	case "UDBZ":
		return []string{"-format", "UDBZ", "-imagekey", "bzip2-level=9"}
	case "ULFO", "ULMO":
		return []string{"-format", c.ImageFormat}
	default:
		return nil
	}
}
