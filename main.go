package main

import (
	"io"
	"log"
	"os"

	"al.essio.dev/cmd/mkdmg/pkg/hdiutil"

	"github.com/spf13/pflag"
)

var (
	volumeName          string
	size                int64
	verbose             bool
	bless               bool
	signingIdentity     string
	notarizeCredentials string
	apfsFs              bool
	sandboxSafe         bool
	format              string
	simulate            bool

	hdiutilVerbosity int

	verboseLog *log.Logger
)

func init() {
	pflag.CommandLine.SetOutput(os.Stderr)
	log.SetPrefix("mkdmg: ")
	log.SetFlags(0)
	log.SetOutput(pflag.CommandLine.Output())

	pflag.StringVar(&volumeName, "volname", "", "volume name for the DMG")
	pflag.Int64Var(&size, "disk-image-size", 0, "size for the DMG in MB")
	pflag.BoolVarP(&verbose, "verbose", "v", false, "enable verbose mode")
	pflag.StringVar(&signingIdentity, "codesign", "", "signing identity")
	pflag.BoolVar(&apfsFs, "apfs", false, "use APFS as disk image's filesystem (default: HFS+)")
	pflag.BoolVar(&sandboxSafe, "sandbox-safe", false, "use sandbox-safe")
	pflag.StringVar(&format, "format", "", "specify the final disk image format (UDZO|UDBZ|ULFO|ULMO)")
	pflag.IntVarP(&hdiutilVerbosity, "hdiutil-verbosity", "V", 0, "set hdiutil verbosity level (0=default - 1=quiet - 2=verbose - 3=debug)")
	pflag.BoolVarP(&simulate, "dry-run", "s", false, "simulate the process")
	pflag.BoolVar(&bless, "bless", false, "bless the disk image")
	pflag.StringVar(&notarizeCredentials, "notarize", "", "notarize the disk image (waits and staples) with the keychain stored credentials")

	verboseLog = log.New(io.Discard, "mkdmg: ", 0)
}

func main() {
	pflag.Parse()
	if pflag.NArg() != 2 {
		log.Fatalln("invalid arguments")
	}

	if verbose {
		verboseLog.SetOutput(pflag.CommandLine.Output())
		hdiutil.SetLogWriter(os.Stderr)
	}

	cfg := &hdiutil.Config{
		VolumeName:          volumeName,
		VolumeSizeMb:        size,
		SandboxSafe:         sandboxSafe,
		ImageFormat:         format,
		HDIUtilVerbosity:    hdiutilVerbosity,
		SigningIdentity:     signingIdentity,
		NotarizeCredentials: notarizeCredentials,
		Simulate:            simulate,
		Bless:               bless,
		OutputPath:          pflag.Arg(0),
		SourceDir:           pflag.Arg(1),
	}
	if apfsFs {
		cfg.FileSystem = "APFS"
	}

	runner := hdiutil.New(cfg)
	if err := runner.Setup(); err != nil {
		log.Fatalf("Failed to setup: %v", err)
	}
	defer runner.Cleanup()

	checkErr(runner.Start())

	checkErr(runner.AttachDiskImage())
	checkErr(runner.Bless())
	checkErr(runner.DetachDiskImage())
	checkErr(runner.FinalizeDMG())

	checkErr(runner.Codesign())
	checkErr(runner.Notarize())

	verboseLog.Printf("DMG created successfully: %s\n", runner.OutputPath)
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
