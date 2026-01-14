package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"al.essio.dev/cmd/mkdmg/pkg/hdiutil"
	"al.essio.dev/cmd/mkdmg/internal/version"

	flag "github.com/spf13/pflag"
)

var (
	volumeName          string
	size                int64
	bless               bool
	signingIdentity     string
	notarizeCredentials string
	apfsFs              bool
	sandboxSafe         bool
	format              string
	simulate            bool
	hdiutilVerbosity    int

	helpMode    bool
	versionMode bool
	verboseMode bool
	binBasename string

	verboseLog *log.Logger
)

func init() {
	binBasename = filepath.Base(os.Args[0])

	flag.CommandLine.SetOutput(os.Stderr)

	flag.StringVar(&volumeName, "volname", "", "volume name for the DMG")
	flag.Int64Var(&size, "disk-image-size", 0, "size for the DMG in MB")
	flag.StringVar(&signingIdentity, "codesign", "", "signing identity")
	flag.BoolVar(&apfsFs, "apfs", false, "use APFS as disk image's filesystem (default: HFS+)")
	flag.BoolVar(&sandboxSafe, "sandbox-safe", false, "use sandbox-safe")
	flag.StringVar(&format, "format", "", "specify the final disk image format (UDZO|UDBZ|ULFO|ULMO)")
	flag.IntVarP(&hdiutilVerbosity, "hdiutil-verbosity", "V", 0, "set hdiutil verbosity level (0=default - 1=quiet - 2=verboseMode - 3=debug)")
	flag.BoolVarP(&simulate, "dry-run", "s", false, "simulate the process")
	flag.BoolVar(&bless, "bless", false, "bless the disk image")
	flag.StringVar(&notarizeCredentials, "notarize", "", "notarize the disk image (waits and staples) with the keychain stored credentials")
	flag.BoolVarP(&helpMode, "help", "h", false, "display this help and exit.")
	flag.BoolVarP(&versionMode, "version", "V", false, "output version information and exit.")
	flag.BoolVarP(&verboseMode, "verboseMode", "v", false, "enable verboseMode mode")
	flag.Usage = usage
	flag.ErrHelp = nil

	verboseLog = log.New(io.Discard, "mkdmg: ", 0)

	flag.CommandLine.SetOutput(os.Stderr)
}

func main() {
	log.SetFlags(0)
	log.SetPrefix(fmt.Sprintf("%s: ", binBasename))
	log.SetOutput(os.Stderr)
	flag.Parse()

	if helpMode {
		usage()
		return
	}

	if versionMode {
		printVersion()
		return
	}

	if flag.NArg() != 2 {
		log.Fatalln("invalid arguments")
	}

	if verboseMode {
		verboseLog.SetOutput(os.Stderr)
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
		OutputPath:          flag.Arg(0),
		SourceDir:           flag.Arg(1),
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

func usage() {
	fmt.Printf("Usage: %s [OPTION]... OUTFILE.DMG DIRECTORY\n", binBasename)
	fmt.Print(flag.CommandLine.FlagUsages())
}

func printVersion() {
	fmt.Println("mkdmg, version", version.Version)
	fmt.Println("Copyright (C) 2025 Alessio Treglia <alessio@debian.org>")
}
