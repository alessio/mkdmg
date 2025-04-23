package hdiutil

import (
	"fmt"
	"strings"
)

type FileSystemType string

const (
	HFSPlus FileSystemType = "HFS+"
	APFS    FileSystemType = "APFS"
)

type Config struct {
	VolumeName      string
	VolumeSizeMb    int64
	SandboxSafe     bool
	FileSystem      string
	SigningIdentity string
	ImageFormat     string

	HDIUtilVerbose bool
	HDIUtilQuiet   bool

	OutputPath string
	SourceDir  string

	Verbose bool
}

type optsType struct {
	*Config

	volNameOpts []string
	formatOpts  []string
	sizeOpts    []string
	fsOpts      []string

	tmpDir   string
	tmpDmg   string
	finalDmg string

	signOpt string

	hdiutilOpts []string
}

func (o *optsType) codesignFinalDMG(finalDMGPath string) error {
	args := []string{"-s", o.signOpt, finalDMGPath}
	if err := runCommand("codesign", args...); err != nil {
		return fmt.Errorf("codesign command failed: %v", err)
	}

	if err := runCommand("codesign",
		"--verify", "--deep", "--strict", "--verbose=2", finalDMGPath); err != nil {
		return fmt.Errorf("the signature seems invalid: %v", err)
	}

	verboseLog.Println("codesign complete")
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
