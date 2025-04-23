package hdiutil

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

var (
	verboseLog *log.Logger
	inst       *optsType
)

func init() {
	verboseLog = log.New(io.Discard, "hdiutil:", 0)
	inst = nil
}

func SetLogWriter(w io.Writer) {
	verboseLog.SetOutput(w)
}

func Init(c *Config) error {
	o := &optsType{Config: c}

	// Validate destination pathname
	if filepath.Ext(c.OutputPath) != ".dmg" {
		return ErrImageFileExt
	}

	o.finalDmg = c.OutputPath

	// generate a volume name if empty
	if len(c.VolumeName) == 0 {
		vname := strings.TrimSuffix(filepath.Base(c.OutputPath), ".dmg")
		o.volNameOpts = []string{"-volname", vname}
	} else {
		o.volNameOpts = []string{"-volname", c.VolumeName}
	}

	// validate image format
	if v := c.imageFormatToArgs(); len(v) > 0 {
		o.formatOpts = v
	} else {
		return ErrInvFormatOpt
	}

	// validate filesystem
	if v := c.filesystemToArgs(); len(v) > 0 {
		o.fsOpts = v
	} else {
		return ErrInvFilesystemOpt
	}

	// check custom size if it's passed
	if c.VolumeSizeMb > 0 {
		o.sizeOpts = []string{"-size", fmt.Sprintf("%dm", c.VolumeSizeMb)}
	}

	// signingIdentity
	o.signOpt = c.SigningIdentity

	// create working directory
	tmpDir, err := os.MkdirTemp("", "mkdmg-")
	if err != nil {
		return fmt.Errorf("%v: %w", ErrCreateDir, err)
	}
	o.tmpDir = tmpDir
	o.tmpDmg = filepath.Join(tmpDir, "temp.dng")

	inst = o
	return nil
}

func (o *optsType) createTempImage() error {
	args := slices.Concat([]string{"create"}, o.filesystemToArgs(), o.sizeOpts)

	args = append(args, "-format", "UDRW", "-volname", o.VolumeName,
		"-quiet", "-srcfolder", o.SourceDir, o.tmpDir,
	)

	return runCommand("hdiutil", args...)
}

func (o *optsType) createTempImageSandboxSafe() error {
	args1 := []string{"makehybrid", "-quiet", "-default-volume-name", o.VolumeName,
		"-hfs", "-o", o.tmpDmg, o.SourceDir}
	if err := runCommand("hdiutil", args1...); err != nil {
		return err
	}

	args2 := []string{"convert", "-format", "UDRW", "-ov", "-o", o.tmpDmg, o.tmpDmg}
	return runCommand("hdiutil", args2...)
}

func (o *optsType) CreateDstDMG() error {
	if o.Config.SandboxSafe {
		return o.createTempImageSandboxSafe()
	}

	return o.createTempImage()
}

func (o *optsType) AttachDiskImage() (string, error) {
	output, err := runCommandOutput("hdiutil", "attach", "-nobrowse", "-noverify", o.tmpDmg)
	if err != nil {
		return "",
			fmt.Errorf("%w: %s", ErrMountImage, output)
	}
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "/Volumes/") {
			fields := strings.Fields(line)
			return fields[len(fields)-1], nil
		}
	}

	return "", fmt.Errorf("%w: couldn't find mount point: %q", ErrMountImage, output)
}

func (o *optsType) DetachDiskImage(mountPoint string) error {
	return runCommand("hdiutil", "detach", mountPoint)
}

func runCommand(name string, args ...string) error {
	verboseLog.Println("Running '", name, args)
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runCommandOutput(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

var (
	ErrInvFormatOpt     = errors.New("invalid image format")
	ErrInvFilesystemOpt = errors.New("invalid image filesystem")
	ErrCreateDir        = errors.New("couldn't create directory")
	ErrImageFileExt     = errors.New("output file must have a .dmg extension")
	ErrMountImage       = errors.New("couldn't attach disk image")
)
