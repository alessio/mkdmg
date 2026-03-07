package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"al.essio.dev/cmd/mkdmg/internal/version"
	"al.essio.dev/pkg/hdiutil"
)

var (
	configPath string
	simulate   bool

	helpMode    bool
	versionMode bool
	verboseMode bool
	binBasename string

	verboseLog *log.Logger
)

func init() {
	binBasename = filepath.Base(os.Args[0])

	flag.CommandLine.SetOutput(os.Stderr)

	flag.StringVar(&configPath, "config", "mkdmg.json", "path to a JSON configuration file")
	flag.BoolVar(&simulate, "dry-run", false, "simulate the process")
	flag.BoolVar(&simulate, "s", false, "simulate the process (shorthand)")
	flag.BoolVar(&helpMode, "help", false, "display this help and exit.")
	flag.BoolVar(&helpMode, "h", false, "display this help and exit (shorthand)")
	flag.BoolVar(&versionMode, "version", false, "output version information and exit.")
	flag.BoolVar(&versionMode, "V", false, "output version information and exit (shorthand)")
	flag.BoolVar(&verboseMode, "verbose", false, "enable verbose output")
	flag.BoolVar(&verboseMode, "v", false, "enable verbose output (shorthand)")
	flag.Usage = usage

	verboseLog = log.New(io.Discard, "mkdmg: ", 0)
}

func main() {
	log.SetFlags(0)
	log.SetPrefix(fmt.Sprintf("%s: ", binBasename))
	log.SetOutput(os.Stderr)
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	flag.Parse()

	if helpMode {
		usage()
		return nil
	}

	if versionMode {
		printVersion()
		return nil
	}

	if flag.NArg() > 2 {
		return fmt.Errorf("too many positional arguments")
	}

	cfg, err := loadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	switch flag.NArg() {
	case 2:
		cfg.OutputPath = flag.Arg(0)
		cfg.SourceDir = flag.Arg(1)
	case 1:
		cfg.OutputPath = flag.Arg(0)
	}

	if cfg.OutputPath == "" || cfg.SourceDir == "" {
		return fmt.Errorf("missing output path or source directory")
	}

	if verboseMode {
		verboseLog.SetOutput(os.Stderr)
		hdiutil.SetLogWriter(os.Stderr)
	}

	runner := hdiutil.New(cfg)
	runner.SetSimulate(simulate)
	if err := runner.Setup(); err != nil {
		return fmt.Errorf("failed to setup: %v", err)
	}
	defer runner.Cleanup()

	if err := runner.Start(); err != nil {
		return err
	}
	if err := runner.AttachDiskImage(); err != nil {
		return err
	}
	if err := runner.Bless(); err != nil {
		return err
	}
	if err := runner.DetachDiskImage(); err != nil {
		return err
	}
	if err := runner.FinalizeDMG(); err != nil {
		return err
	}
	if err := runner.Codesign(); err != nil {
		return err
	}
	if err := runner.Notarize(); err != nil {
		return err
	}

	verboseLog.Printf("DMG created successfully: %s\n", runner.OutputPath)
	return nil
}

func usage() {
	w := flag.CommandLine.Output()
	_, _ = fmt.Fprintf(w, "Usage: %s [OPTION]... [OUTFILE.DMG [DIRECTORY]]\n", binBasename)
	flag.PrintDefaults()
}

func printVersion() {
	fmt.Println("mkdmg, version", version.Version())
	fmt.Println("Copyright (C) 2025,2026 Alessio Treglia <alessio@debian.org>")
}

func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

// loadConfig reads the configuration from a JSON file.
func loadConfig(path string) (*hdiutil.Config, error) {
	// Clean the path to ensure it is normalized.
	// G304: Potential file inclusion via variable.
	// This is a CLI tool and the user is expected to provide a path to the config file.
	// #nosec G304
	f, err := os.Open(filepath.Clean(path))
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = f.Close()
	}()

	cfg := &hdiutil.Config{}
	if err := cfg.FromJSON(f); err != nil {
		return nil, err
	}

	return cfg, nil
}
