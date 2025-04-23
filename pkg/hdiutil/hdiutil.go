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
	ErrInvSourceDir     = errors.New("invalid source directory")
	ErrInvFormatOpt     = errors.New("invalid image format")
	ErrInvFilesystemOpt = errors.New("invalid image filesystem")
	ErrCreateDir        = errors.New("couldn't create directory")
	ErrImageFileExt     = errors.New("output file must have a .dmg extension")
	ErrMountImage       = errors.New("couldn't attach disk image")
	ErrSignIdNotFound   = errors.New("signing identity not found")
	ErrCodesignFailed   = errors.New("codesign command failed")
	ErrNotarizeFailed   = errors.New("notarization failed")
	ErrSandboxAPFS      = errors.New("creating an APFS disk image that is sandbox safe is not supported")
)

var (
	verboseLog *log.Logger
)

func init() {
	verboseLog = log.New(io.Discard, "hdiutil:", 0)
}

func SetLogWriter(w io.Writer) {
	verboseLog.SetOutput(w)
}

func New(c *Config) *Runner {
	return &Runner{Config: c}
}

func (r *Runner) Setup() error {
	return r.init()
}

type Runner struct {
	*Config

	volNameOpts []string
	formatOpts  []string
	sizeOpts    []string
	fsOpts      []string
	signOpt     string
	notarizeOpt string
	hdiutilOpts []string

	srcDir   string
	tmpDir   string
	mountDir string

	tmpDmg   string
	finalDmg string

	simulate bool

	permFixed    bool
	cleanupFuncs []func()
}

func (r *Runner) CreateDstDMG() error {
	if r.Config.SandboxSafe {
		return r.createTempImageSandboxSafe()
	}

	return r.createTempImage()
}

func (r *Runner) AttachDiskImage() error {
	output, err := r.runHdiutilOutput("attach", "-nobrowse", "-noverify", r.tmpDmg)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrMountImage, output)
	}
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "/Volumes/") {
			fields := strings.Fields(line)
			r.mountDir = fields[len(fields)-1]
			return nil
		}
	}

	return fmt.Errorf("%w: couldn't find mount point: %q", ErrMountImage, output)
}

func (r *Runner) DetachDiskImage() error {
	r.fixPermissions()
	return r.runHdiutil("detach", r.mountDir)
}

func (r *Runner) Bless() error {
	r.fixPermissions()
	if !r.Config.Bless {
		return nil
	}

	if r.SandboxSafe {
		verboseLog.Println("Skipping blessing on sandbox safe images")
		return nil
	}

	return runCommand("bless", "--folder", r.mountDir)
}

func (r *Runner) FinalizeDMG() error {
	return r.runHdiutil(r.setHdiutilVerbosity(slices.Concat(
		[]string{"convert", r.tmpDmg},
		r.formatOpts,
		[]string{"-o", r.finalDmg}),
	)...)
}

func (r *Runner) Codesign() error {
	if len(r.signOpt) == 0 {
		verboseLog.Println("Skipping codesign")
		return nil
	}

	if err := runCommand("codesign", "-s", r.signOpt, r.finalDmg); err != nil {
		return fmt.Errorf("%w: codesign command failed: %v", ErrCodesignFailed, err)
	}

	if err := runCommand("codesign",
		"--verify", "--deep", "--strict", "--verbose=2", r.finalDmg); err != nil {
		return fmt.Errorf("%w: the signature seems invalid: %v", ErrCodesignFailed, err)
	}

	verboseLog.Println("codesign complete")
	return nil
}

func (r *Runner) Notarize() error {
	if len(r.notarizeOpt) == 0 {
		verboseLog.Println("Skipping notarization")
		return nil
	}

	verboseLog.Println("Start notarization")
	if err := runCommand("xcrun", "notarytool", "submit",
		r.finalDmg, "--keychain-profile", r.notarizeOpt,
	); err != nil {
		return fmt.Errorf("%w: notarization failed: %v", ErrNotarizeFailed, err)
	}

	verboseLog.Println("Stapling the notarization ticket")
	if output, err := runCommandOutput(
		"xcrun", "stapler", "staple", r.finalDmg); err != nil {
		return fmt.Errorf("%w: stapler failed: %v", ErrNotarizeFailed, output)
	}

	verboseLog.Println("Notarization complete")

	return nil
}

func (r *Runner) createTempImage() error {
	args := slices.Concat([]string{"create"},
		r.filesystemToArgs(),
		r.sizeOpts,
		[]string{
			"-format", "UDRW", "-volname", r.VolumeName, "-srcfolder", r.srcDir, r.tmpDmg},
	)

	return r.runHdiutil(r.setHdiutilVerbosity(args)...)
}

func (r *Runner) createTempImageSandboxSafe() error {
	args1 := r.setHdiutilVerbosity([]string{"makehybrid",
		"-default-volume-name", r.VolumeName, "-hfs", "-r", r.tmpDmg, r.srcDir})
	if err := r.runHdiutil(args1...); err != nil {
		return err
	}

	args2 := r.setHdiutilVerbosity([]string{"convert",
		"-format", "UDRW", "-ov", "-r", r.tmpDmg, r.tmpDmg})

	return r.runHdiutil(args2...)
}

func (r *Runner) setHdiutilVerbosity(args []string) []string {
	if len(args) == 0 || r.HDIUtilVerbosity == 0 {
		return args
	}

	var val string

	switch r.HDIUtilVerbosity {
	case 1:
		val = "-quiet"
	case 2:
		val = "-verbose"
	default:
		val = "-debug"
	}

	switch args[0] {
	case "create", "makehybrid", "convert":
		return slices.Insert(args, 1, val)
	default:
		return slices.Insert(args, 0, val)
	}
}

func (r *Runner) init() error {
	if len(r.Config.SourceDir) == 0 {
		return ErrInvSourceDir
	}

	r.srcDir = filepath.Clean(r.Config.SourceDir)

	if filepath.Ext(r.Config.OutputPath) != ".dmg" {
		return ErrImageFileExt
	}

	r.finalDmg = r.Config.OutputPath

	// generate a volume name if empty
	if len(r.Config.VolumeName) == 0 {
		vname := strings.TrimSuffix(filepath.Base(r.Config.OutputPath), ".dmg")
		r.volNameOpts = []string{"-volname", vname}
	} else {
		r.volNameOpts = []string{"-volname", r.Config.VolumeName}
	}

	// validate image format
	if v := r.Config.imageFormatToArgs(); len(v) > 0 {
		r.formatOpts = v
	} else {
		return ErrInvFormatOpt
	}

	// validate filesystem
	if v := r.Config.filesystemToArgs(); len(v) > 0 {
		r.fsOpts = v
	} else {
		return ErrInvFilesystemOpt
	}

	// sandbox safe and APFS are mutually exclusive
	if r.Config.SandboxSafe && strings.ToUpper(r.Config.FileSystem) == "APFS" {
		return ErrSandboxAPFS
	}

	// check custom size if it's passed
	if r.Config.VolumeSizeMb > 0 {
		r.sizeOpts = []string{"-size", fmt.Sprintf("%dm", r.Config.VolumeSizeMb)}
	}

	r.cleanupFuncs = []func(){}

	// create a working directory
	tmpDir, err := os.MkdirTemp("", "mkdmg-")
	if err != nil {
		return fmt.Errorf("%v: %w", ErrCreateDir, err)
	}

	r.tmpDir = tmpDir
	r.tmpDmg = filepath.Join(tmpDir, "temp.dmg")
	// signingIdentity
	r.signOpt = r.Config.SigningIdentity
	r.notarizeOpt = r.Config.NotarizeCredentials

	return nil
}

func (r *Runner) fixPermissions() {
	if r.permFixed {
		return
	}

	verboseLog.Println("Fixing permissions")
	if err := runCommand("chmod", []string{
		"-Rf", "go-w", r.mountDir,
	}...); err != nil {
		verboseLog.Printf("chmod failed: %v", err)
	}

	r.permFixed = true
}

func (r *Runner) runHdiutil(args ...string) error {
	if r.Simulate {
		verboseLog.Println("Simulating hdiutil command: ", args)
		return nil
	}
	return runCommand("hdiutil", args...)
}

func (r *Runner) runHdiutilOutput(args ...string) (string, error) {
	if r.Simulate {
		verboseLog.Println("Simulating hdiutil command: ", args)
		return "", nil
	}

	return runCommandOutput("hdiutil", args...)
}

func runCommand(name string, args ...string) error {
	verboseLog.Println("Running '", name, args)
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runCommandOutput(name string, args ...string) (string, error) {
	verboseLog.Println("Running '", name, args)
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}
