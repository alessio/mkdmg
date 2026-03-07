package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ---- Unit Tests (in-process) ----

// resetForTest creates a fresh flag.CommandLine and resets all global variables,
// allowing run() to be called repeatedly with different arguments.
func resetForTest(t *testing.T, args []string) {
	t.Helper()

	savedArgs := os.Args
	savedFlags := flag.CommandLine
	t.Cleanup(func() {
		os.Args = savedArgs
		flag.CommandLine = savedFlags
	})

	// Reset all global config variables to zero values.
	configPath = ""
	simulate = false
	helpMode = false
	versionMode = false
	verboseMode = false

	// Replace the default FlagSet so flags can be re-registered.
	// Use ContinueOnError so parse errors don't call os.Exit.
	flag.CommandLine = flag.NewFlagSet(args[0], flag.ContinueOnError)

	// Re-register all flags (mirrors init()).
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

	os.Args = args
}

// captureStdout redirects os.Stdout to a pipe, runs fn, and returns
// everything written to stdout.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stdout = w
	fn()
	if err := w.Close(); err != nil {
		t.Fatalf("w.Close: %v", err)
	}
	os.Stdout = old
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("io.Copy: %v", err)
	}
	return buf.String()
}

// writeConfigFile creates a temporary JSON config file.
func writeConfigFile(t *testing.T, cfg map[string]any) string {
	t.Helper()
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	f := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(f, data, 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	return f
}

// --- loadConfig ---

func TestLoadConfigValid(t *testing.T) {
	cfgFile := writeConfigFile(t, map[string]any{
		"output_path": "test.dmg",
		"source_dir":  "/tmp/src",
		"volume_name": "TestVol",
	})
	cfg, err := loadConfig(cfgFile)
	if err != nil {
		t.Fatalf("loadConfig: %v", err)
	}
	if cfg.OutputPath != "test.dmg" {
		t.Errorf("OutputPath = %q, want %q", cfg.OutputPath, "test.dmg")
	}
	if cfg.SourceDir != "/tmp/src" {
		t.Errorf("SourceDir = %q, want %q", cfg.SourceDir, "/tmp/src")
	}
	if cfg.VolumeName != "TestVol" {
		t.Errorf("VolumeName = %q, want %q", cfg.VolumeName, "TestVol")
	}
}

func TestLoadConfigInvalidJSON(t *testing.T) {
	f := filepath.Join(t.TempDir(), "bad.json")
	if err := os.WriteFile(f, []byte("{not valid json"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := loadConfig(f)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestLoadConfigMissingFile(t *testing.T) {
	_, err := loadConfig(filepath.Join(t.TempDir(), "nonexistent.json"))
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

// --- isFlagPassed ---

func TestIsFlagPassed(t *testing.T) {
	resetForTest(t, []string{"mkdmg", "--config", "test.json"})
	flag.Parse()

	if !isFlagPassed("config") {
		t.Error("isFlagPassed(\"config\") = false, want true")
	}
	if isFlagPassed("dry-run") {
		t.Error("isFlagPassed(\"dry-run\") = true, want false")
	}
}

// --- usage ---

func TestUsageOutput(t *testing.T) {
	var buf bytes.Buffer
	resetForTest(t, []string{"mkdmg"})
	flag.CommandLine.SetOutput(&buf)
	binBasename = "mkdmg"

	usage()

	out := buf.String()
	if !strings.Contains(out, "Usage: mkdmg") {
		t.Errorf("usage output missing 'Usage: mkdmg', got: %s", out)
	}
	if !strings.Contains(out, "-config") {
		t.Errorf("usage output missing '-config', got: %s", out)
	}
}

// --- printVersion ---

func TestPrintVersionOutput(t *testing.T) {
	out := captureStdout(t, printVersion)
	if !strings.Contains(out, "mkdmg, version") {
		t.Errorf("printVersion output missing 'mkdmg, version', got: %s", out)
	}
	if !strings.Contains(out, "Copyright") {
		t.Errorf("printVersion output missing 'Copyright', got: %s", out)
	}
}

// --- run() ---

func TestRunHelpMode(t *testing.T) {
	var buf bytes.Buffer
	resetForTest(t, []string{"mkdmg", "--help"})
	flag.CommandLine.SetOutput(&buf)

	err := run()
	if err != nil {
		t.Fatalf("run() with --help returned error: %v", err)
	}
	if !strings.Contains(buf.String(), "Usage:") {
		t.Errorf("help output missing 'Usage:', got: %s", buf.String())
	}
}

func TestRunHelpShorthand(t *testing.T) {
	var buf bytes.Buffer
	resetForTest(t, []string{"mkdmg", "-h"})
	flag.CommandLine.SetOutput(&buf)

	err := run()
	if err != nil {
		t.Fatalf("run() with -h returned error: %v", err)
	}
	if !strings.Contains(buf.String(), "Usage:") {
		t.Errorf("help output missing 'Usage:', got: %s", buf.String())
	}
}

func TestRunVersionMode(t *testing.T) {
	resetForTest(t, []string{"mkdmg", "--version"})
	var runErr error
	out := captureStdout(t, func() {
		runErr = run()
	})
	if runErr != nil {
		t.Fatalf("run() with --version returned error: %v", runErr)
	}
	if !strings.Contains(out, "mkdmg, version") {
		t.Errorf("version output missing 'mkdmg, version', got: %s", out)
	}
}

func TestRunVersionShorthand(t *testing.T) {
	resetForTest(t, []string{"mkdmg", "-V"})
	var runErr error
	out := captureStdout(t, func() {
		runErr = run()
	})
	if runErr != nil {
		t.Fatalf("run() with -V returned error: %v", runErr)
	}
	if !strings.Contains(out, "mkdmg, version") {
		t.Errorf("version output missing 'mkdmg, version', got: %s", out)
	}
}

func TestRunDefaultConfig(t *testing.T) {
	resetForTest(t, []string{"mkdmg"})
	err := run()
	if err == nil {
		t.Fatal("run() with default config (mkdmg.json not present) should return an error")
	}
	if !strings.Contains(err.Error(), "failed to load config") {
		t.Errorf("error = %q, want it to contain 'failed to load config'", err)
	}
}

func TestRunTooManyPositionalArgs(t *testing.T) {
	cfgFile := writeConfigFile(t, map[string]any{
		"simulate": true,
	})
	resetForTest(t, []string{"mkdmg", "--config", cfgFile, "a.dmg", "src", "extra"})
	err := run()
	if err == nil {
		t.Fatal("run() with 3 positional args should return an error")
	}
	if !strings.Contains(err.Error(), "too many positional arguments") {
		t.Errorf("error = %q, want 'too many positional arguments'", err)
	}
}

func TestRunPositionalOutputAndSource(t *testing.T) {
	sourceDir := t.TempDir()
	outputDMG := filepath.Join(t.TempDir(), "test.dmg")
	cfgFile := writeConfigFile(t, map[string]any{
		"simulate": true,
	})
	resetForTest(t, []string{"mkdmg", "--config", cfgFile, outputDMG, sourceDir})
	err := run()
	if err != nil {
		t.Fatalf("run() with positional output and source: %v", err)
	}
}

func TestRunPositionalOutputOnly(t *testing.T) {
	outputDMG := filepath.Join(t.TempDir(), "test.dmg")
	sourceDir := t.TempDir()
	cfgFile := writeConfigFile(t, map[string]any{
		"source_dir": sourceDir,
		"simulate":   true,
	})
	resetForTest(t, []string{"mkdmg", "--config", cfgFile, outputDMG})
	err := run()
	if err != nil {
		t.Fatalf("run() with positional output only: %v", err)
	}
}

func TestRunPositionalOverridesConfig(t *testing.T) {
	sourceDir := t.TempDir()
	outputDMG := filepath.Join(t.TempDir(), "override.dmg")
	cfgFile := writeConfigFile(t, map[string]any{
		"output_path": "original.dmg",
		"source_dir":  "/original/src",
		"simulate":    true,
	})
	resetForTest(t, []string{"mkdmg", "--config", cfgFile, outputDMG, sourceDir})
	err := run()
	if err != nil {
		t.Fatalf("run() with positional overrides: %v", err)
	}
}

func TestRunConfigFileMissing(t *testing.T) {
	resetForTest(t, []string{"mkdmg", "--config", "/nonexistent/config.json"})
	err := run()
	if err == nil {
		t.Fatal("run() with missing config should return an error")
	}
	if !strings.Contains(err.Error(), "failed to load config") {
		t.Errorf("error = %q, want 'failed to load config'", err)
	}
}

func TestRunConfigFileDryRun(t *testing.T) {
	sourceDir := t.TempDir()
	outputDMG := filepath.Join(t.TempDir(), "test.dmg")
	cfgFile := writeConfigFile(t, map[string]any{
		"output_path": outputDMG,
		"source_dir":  sourceDir,
		"simulate":    true,
	})
	resetForTest(t, []string{"mkdmg", "--config", cfgFile})
	err := run()
	if err != nil {
		t.Fatalf("run() with valid config (dry-run): %v", err)
	}
}

func TestRunDryRunFlag(t *testing.T) {
	sourceDir := t.TempDir()
	outputDMG := filepath.Join(t.TempDir(), "test.dmg")
	cfgFile := writeConfigFile(t, map[string]any{
		"output_path": outputDMG,
		"source_dir":  sourceDir,
	})
	resetForTest(t, []string{"mkdmg", "--config", cfgFile, "--dry-run"})
	err := run()
	if err != nil {
		t.Fatalf("run() with --dry-run flag: %v", err)
	}
}

func TestRunShorthandDryRun(t *testing.T) {
	sourceDir := t.TempDir()
	outputDMG := filepath.Join(t.TempDir(), "test.dmg")
	cfgFile := writeConfigFile(t, map[string]any{
		"output_path": outputDMG,
		"source_dir":  sourceDir,
	})
	resetForTest(t, []string{"mkdmg", "--config", cfgFile, "-s"})
	err := run()
	if err != nil {
		t.Fatalf("run() with -s (dry-run shorthand): %v", err)
	}
}

func TestRunMissingOutputPath(t *testing.T) {
	cfgFile := writeConfigFile(t, map[string]any{
		"source_dir": "/tmp/src",
		"simulate":   true,
	})
	resetForTest(t, []string{"mkdmg", "--config", cfgFile})
	err := run()
	if err == nil {
		t.Fatal("run() with missing output_path should return an error")
	}
	if !strings.Contains(err.Error(), "missing output path or source directory") {
		t.Errorf("error = %q, want 'missing output_path or source_dir'", err)
	}
}

func TestRunMissingSourceDir(t *testing.T) {
	outputDMG := filepath.Join(t.TempDir(), "test.dmg")
	cfgFile := writeConfigFile(t, map[string]any{
		"output_path": outputDMG,
		"simulate":    true,
	})
	resetForTest(t, []string{"mkdmg", "--config", cfgFile})
	err := run()
	if err == nil {
		t.Fatal("run() with missing source_dir should return an error")
	}
	if !strings.Contains(err.Error(), "missing output path or source directory") {
		t.Errorf("error = %q, want 'missing output_path or source_dir'", err)
	}
}

func TestRunVerboseModeDryRun(t *testing.T) {
	sourceDir := t.TempDir()
	outputDMG := filepath.Join(t.TempDir(), "test.dmg")
	cfgFile := writeConfigFile(t, map[string]any{
		"output_path": outputDMG,
		"source_dir":  sourceDir,
		"simulate":    true,
	})
	resetForTest(t, []string{"mkdmg", "--config", cfgFile, "--verbose"})
	err := run()
	if err != nil {
		t.Fatalf("run() with --verbose --dry-run: %v", err)
	}
}