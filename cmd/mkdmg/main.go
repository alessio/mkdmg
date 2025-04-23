package main

import (
	"io"
	"log"
	"os"

	"github.com/alessio/mkdmg/pkg/hdiutil"
	// /	"github.com/joho/godotenv"
	"github.com/spf13/pflag"
)

var (
	volumeName      string
	size            int64
	verbose         bool
	bless           bool
	signingIdentity string
	apfsFs          bool
	sandboxSafe     bool
	format          string
	simulate        bool

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
		VolumeName:       volumeName,
		VolumeSizeMb:     size,
		SandboxSafe:      sandboxSafe,
		ImageFormat:      format,
		HDIUtilVerbosity: hdiutilVerbosity,
		SigningIdentity:  signingIdentity,
		Simulate:         simulate,
		OutputPath:       pflag.Arg(0),
		SourceDir:        pflag.Arg(1),
	}
	if apfsFs {
		cfg.FileSystem = "APFS"
	}

	runner := hdiutil.New(cfg)
	if err := runner.Setup(); err != nil {
		log.Fatalf("Failed to setup: %v", err)
	}

	checkErr(runner.CreateDstDMG())

	checkErr(runner.AttachDiskImage())
	checkErr(runner.DetachDiskImage())
	checkErr(runner.FinalizeDMG())

	if signingIdentity != "" {
		checkErr(runner.CodesignFinalDMG())
	}

	verboseLog.Printf("DMG created successfully: %s\n", runner.OutputPath)
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
