package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/pflag"
)

var (
	volumeName      string
	size            int
	verbose         bool
	signingIdentity string
	apfsFs          bool
	sandboxSafe     bool
	format          string
	// hdiutilVerbose  bool

	ImageKeyArgs []string

	verboseLog *log.Logger
)

type fmtFlag struct{}

var FormatArgs = &fmtFlag{}

func (i *fmtFlag) Validate() bool {
	return len(i.buildArgs()) != 0
}

func (i *fmtFlag) Args() []string {
	return i.buildArgs()
}

func (i *fmtFlag) buildArgs() []string {
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

func init() {
	pflag.CommandLine.SetOutput(os.Stderr)
	log.SetPrefix("go-make-dmg:")
	log.SetFlags(0)
	log.SetOutput(pflag.CommandLine.Output())

	pflag.StringVar(&volumeName, "volname", "", "volume name for the DMG")
	pflag.IntVar(&size, "disk-image-size", 0, "size for the DMG in MB")
	pflag.BoolVarP(&verbose, "verbose", "v", false, "enable verbose mode")
	pflag.StringVar(&signingIdentity, "codesign", "", "signing identity")
	pflag.BoolVar(&apfsFs, "apsf", false, "use APFS as disk image's filesystem (default: HFS+)")
	pflag.BoolVar(&sandboxSafe, "sandbox-safe", false, "use sandbox-safe")
	pflag.StringVar(&format, "format", "", "specify the final disk image format (UDZO|UDBZ|ULFO|ULMO)")
	// pflag.BoolVar(&hdiutilVerbose, "hdiutil-verbose", false, "enable verbose mode")

	verboseLog = log.New(io.Discard, "go-make-dmg: ", 0)
}

func main() {
	pflag.Parse()
	if pflag.NArg() != 2 {
		log.Fatalln("invalid arguments")
	}

	if verbose {
		verboseLog.SetOutput(pflag.CommandLine.Output())
	}

	if sandboxSafe && apfsFs {
		log.Fatalln("creating an APFS disk image that is sandbox safe is not supported")
	}

	if len(imageFormatToArgs()) == 0 {
		log.Fatalln("invalid format:", format)
	}

	finalDMG := pflag.Arg(0)
	sourceFolder := pflag.Arg(1)
	outputFileExt := ".dmg"

	if filepath.Ext(finalDMG) != outputFileExt {
		log.Fatalf("the output disk image must have a .dmg extension: %s", finalDMG)
	}

	if volumeName == "" {
		volumeName = strings.TrimSuffix(filepath.Base(finalDMG), outputFileExt)
	}

	tempDir, err := os.MkdirTemp("", "*-go-make-dmg")
	if err != nil {
		log.Fatalf("couldn't create temp dir: %v", err)
	}

	tempDMG := filepath.Join(tempDir, "temp.dmg")

	verboseLog.Println("Creating temporary DMG...")
	if err := createDstDMG(tempDMG, volumeName, sourceFolder); err != nil {
		log.Fatalf("couldn't create temp DMG: %v", err)
	}

	verboseLog.Println("Mounting temporary DMG...")
	mountPoint := attachDiskImage(tempDMG)
	verboseLog.Printf("DMG mounted at %q\n", mountPoint)

	// Optional: Set background, icons, etc. here (future improvement)

	verboseLog.Println("Unmounting DMG...")
	if err := detachDiskImage(mountPoint); err != nil {
		log.Fatalf("Failed to unmount image: %v", err)
	}

	verboseLog.Println("Converting to final compressed DMG...")
	if err := convertToFinalDMG(tempDMG, finalDMG); err != nil {
		log.Fatalf("Failed to create final DMG: %v", err)
	}

	if signingIdentity != "" {
		if err := codesignFinalDMG(finalDMG); err != nil {
			log.Fatalf("Failed to sign final DMG: %v", err)
		}
	}

	verboseLog.Printf("DMG created successfully: %s\n", finalDMG)
}

func createDstDMG(tempImagePath, volumeName, sourceFolder string) error {
	if sandboxSafe {
		return createTempImageSandboxSafe(tempImagePath, volumeName, sourceFolder)
	}

	return createTempImage(tempImagePath, volumeName, sourceFolder)
}

func createTempImageSandboxSafe(tempImagePath, volumeName, sourceFolder string) error {
	args1 := []string{"makehybrid", "-quiet", "-default-volume-name", volumeName,
		"-hfs", "-o", tempImagePath, sourceFolder}
	if err := runCommand("hdiutil", args1...); err != nil {
		return err
	}

	args2 := []string{"convert", "-format", "UDRW", "-ov", "-o", tempImagePath, tempImagePath}
	return runCommand("hdiutil", args2...)
}

func createTempImage(tempImagePath, volumeName, sourceFolder string) error {
	args := []string{"create"}

	if apfsFs {
		args = append(args, "-fs", "APFS")
	} else {
		args = append(args, "-fs", "HFS+", "-fsargs", "-c c=64,a=16,e=16")
	}

	if size != 0 {
		args = append(args, "-size", fmt.Sprintf("%dm", size))
	}

	args = append(args, "-format", "UDRW", "-volname", volumeName,
		"-quiet", "-srcfolder", sourceFolder, tempImagePath,
	)
	return runCommand("hdiutil", args...)
}

func attachDiskImage(imagePath string) string {
	output, err := runCommandOutput("hdiutil", "attach", "-nobrowse", "-noverify", imagePath)
	if err != nil {
		log.Fatalf("couldn't attach disk image: %s", output)
	}
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "/Volumes/") {
			fields := strings.Fields(line)
			return fields[len(fields)-1]
		}
	}

	log.Fatalf("couldn't find mountpoint: %s", output)

	return ""
}

func detachDiskImage(mountPoint string) error {
	return runCommand("hdiutil", "detach", mountPoint)
}

func convertToFinalDMG(tempImagePath, finalDMGPath string) error {
	args := slices.Concat(
		[]string{"convert", tempImagePath},
		FormatArgs.Args(),
		[]string{"-o", finalDMGPath})

	return runCommand("hdiutil", args...)
}

func codesignFinalDMG(finalDMGPath string) error {
	args := []string{"-s", signingIdentity, finalDMGPath}
	if err := runCommand("codesign", args...); err != nil {
		return fmt.Errorf("codesign command failed: %v", err)
	}

	if err := runCommand("codesign",
		"--verify", "--deep", "--strict", "--verbose=2", finalDMGPath); err != nil {
		return fmt.Errorf("the signature seems invalid: %v", err)
	}

	verboseLog.Println("Codesign complete")
	return nil
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
