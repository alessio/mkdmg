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
	ErrCodesignFailed   = errors.New("codesign command failed")
	ErrNotarizeFailed   = errors.New("notarization failed")
	ErrSandboxAPFS      = errors.New("creating an APFS disk image that is sandbox safe is not supported")
	ErrNeedInit         = errors.New("runner not properly initialized, call Setup() first")
)

var (
	verboseLog *log.Logger
)

func init() {
	verboseLog = log.New(io.Discard, "hdiutil: ", 0)
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

	formatOpts  []string
	sizeOpts    []string
	fsOpts      []string
	volNameOpt  string
	signOpt     string
	notarizeOpt string

	srcDir   string
	tmpDir   string
	mountDir string

	tmpDmg   string
	finalDmg string

	permFixed bool

	cleanupFuncs []func()
}

func (r *Runner) Cleanup() {
	for _, f := range r.cleanupFuncs {
		f()
	}
}

func (r *Runner) Start() error {
	if r.tmpDir == "" || r.tmpDmg == "" {
		return ErrNeedInit
	}

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

	return r.runCommand("bless", "--folder", r.mountDir)
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

	if err := r.runCommand("codesign", "-s", r.signOpt, r.finalDmg); err != nil {
		return fmt.Errorf("%w: codesign command failed: %v", ErrCodesignFailed, err)
	}

	if err := r.runCommand("codesign",
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
	if err := r.runCommand("xcrun", "notarytool", "submit",
		r.finalDmg, "--keychain-profile", r.notarizeOpt,
	); err != nil {
		return fmt.Errorf("%w: notarization failed: %v", ErrNotarizeFailed, err)
	}

	verboseLog.Println("Stapling the notarization ticket")
	if output, err := r.runCommandOutput(
		"xcrun", "stapler", "staple", r.finalDmg); err != nil {
		return fmt.Errorf("%w: stapler failed: %v", ErrNotarizeFailed, output)
	}

	verboseLog.Println("Notarization complete")

	return nil
}

func (r *Runner) createTempImage() error {
	args := slices.Concat([]string{"create"},
		r.fsOpts,
		r.sizeOpts,
		[]string{"-format", "UDRW", "-volname", r.volNameOpt, "-srcfolder", r.srcDir, r.tmpDmg},
	)

	return r.runHdiutil(r.setHdiutilVerbosity(args)...)
}

func (r *Runner) createTempImageSandboxSafe() error {
	args1 := r.setHdiutilVerbosity([]string{"makehybrid",
		"-default-volume-name", r.volNameOpt, "-hfs", "-r", r.tmpDmg, r.srcDir})
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
	if err := r.Config.validate(); err != nil {
		return err
	}

	r.srcDir = filepath.Clean(r.Config.SourceDir)
	r.finalDmg = r.Config.OutputPath

	// generate a volume name if empty
	if len(r.Config.VolumeName) == 0 {
		vname := strings.TrimSuffix(filepath.Base(r.Config.OutputPath), ".dmg")
		r.volNameOpt = vname
	} else {
		r.volNameOpt = r.Config.VolumeName
	}

	r.formatOpts = r.Config.imageFormatToArgs()
	r.fsOpts = r.Config.filesystemToArgs()

	// check custom size if it's passed
	if r.Config.VolumeSizeMb > 0 {
		r.sizeOpts = []string{"-size", fmt.Sprintf("%dm", r.Config.VolumeSizeMb)}
	}

	// create a working directory
	tmpDir, err := os.MkdirTemp("", "mkdmg-")
	if err != nil {
		return fmt.Errorf("%v: %w", ErrCreateDir, err)
	}
	r.tmpDir = tmpDir

	r.cleanupFuncs = []func(){}
	r.cleanupFuncs = append(r.cleanupFuncs, func() {
		if r.tmpDir != "" {
			verboseLog.Println("Removing temporary directory: ", r.tmpDir)
			_ = os.RemoveAll(r.tmpDir)
		}
	})

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
	if err := r.runCommand("chmod", []string{
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
	return r.runCommand("hdiutil", args...)
}

func (r *Runner) runHdiutilOutput(args ...string) (string, error) {
	if r.Simulate {
		verboseLog.Println("Simulating hdiutil command: ", args)
		return "", nil
	}

	return r.runCommandOutput("hdiutil", args...)
}

func (r *Runner) runCommand(name string, args ...string) error {
	verboseLog.Println("Running '", name, args)
	if r.Simulate {
		return nil
	}
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (r *Runner) runCommandOutput(name string, args ...string) (string, error) {
	verboseLog.Println("Running '", name, args)
	if r.Simulate {
		return "", nil
	}
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}
