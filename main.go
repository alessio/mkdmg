package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

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

	var cfg *hdiutil.Config
	var err error

	if isFlagPassed("config") {
		cfg, err = loadConfig(configPath)
		if err != nil {
			return fmt.Errorf("failed to load config: %v", err)
		}
	} else {
		cfg, err = loadConfig(configPath)
		if err != nil {
			if os.IsNotExist(err) {
				cfg = &hdiutil.Config{}
			} else {
				return fmt.Errorf("failed to load config: %v", err)
			}
		}
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

	var (
		mu       sync.Mutex
		attached bool
	)

	// withRunner serializes a runner operation against the signal handler's
	// shutdown sequence, so temporary files are never removed while an
	// external command is still using them.
	withRunner := func(op func() error) error {
		mu.Lock()
		defer mu.Unlock()
		return op()
	}

	// detach unmounts the disk image if it is still attached. The caller must
	// hold mu.
	detach := func() error {
		if !attached {
			return nil
		}
		if err := runner.DetachDiskImage(); err != nil {
			return err
		}
		attached = false
		return nil
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	// Cleanup only removes the temporary directory, so the disk image has to be
	// detached explicitly on every path out of run().
	defer func() {
		_ = withRunner(detach)
	}()

	done := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		select {
		case <-sigChan:
			verboseLog.Println("Caught interrupt signal, cleaning up...")
			mu.Lock()
			_ = detach()
			runner.Cleanup()
			mu.Unlock()
			os.Exit(130)
		case <-done:
			return
		}
	}()
	// Wait for the listener to return so it can never run concurrently with the
	// deferred detach and Cleanup below.
	defer func() {
		close(done)
		wg.Wait()
	}()

	if err := withRunner(runner.Start); err != nil {
		return fmt.Errorf("failed to start: %v", err)
	}

	if err := withRunner(func() error {
		if err := runner.AttachDiskImage(); err != nil {
			return err
		}
		attached = true
		return nil
	}); err != nil {
		return fmt.Errorf("failed to attach disk image: %v", err)
	}

	if err := withRunner(runner.Bless); err != nil {
		return fmt.Errorf("failed to bless: %v", err)
	}

	if err := withRunner(detach); err != nil {
		return fmt.Errorf("failed to detach disk image: %v", err)
	}

	if err := withRunner(runner.FinalizeDMG); err != nil {
		return fmt.Errorf("failed to finalize dmg: %v", err)
	}
	if err := withRunner(runner.Codesign); err != nil {
		return fmt.Errorf("failed to sign: %v", err)
	}
	if err := withRunner(runner.Notarize); err != nil {
		return fmt.Errorf("failed to notarize: %v", err)
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
