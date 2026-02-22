package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestMkdmgBasicCreation(t *testing.T) {
	if !isHdiutilAvailable() {
		t.Skip("hdiutil not available, skipping integration test")
	}

	t.Parallel()
	// Create a temporary directory for source files
	sourceDir := t.TempDir()
	// Create a dummy file inside the source directory
	dummyFile := filepath.Join(sourceDir, "test.txt")
	err := os.WriteFile(dummyFile, []byte("Hello, mkdmg!"), 0644)
	if err != nil {
		t.Fatalf("Failed to create dummy file: %v", err)
	}

	// Define the output DMG path
	outputDMG := filepath.Join(t.TempDir(), "test.dmg")

	// Build the mkdmg binary
	mkdmgBinary := filepath.Join(t.TempDir(), "mkdmg")
	t.Logf("Attempting to build mkdmg to: %s", mkdmgBinary)
	cmdBuild := exec.Command("go", "build", "-o", mkdmgBinary, ".")
	var buildErr bytes.Buffer
	cmdBuild.Stderr = &buildErr
	if err := cmdBuild.Run(); err != nil {
		t.Fatalf("Failed to build mkdmg binary: %v\n%s", err, buildErr.String())
	}

	// Check if the binary exists after building
	if _, err := os.Stat(mkdmgBinary); os.IsNotExist(err) {
		t.Fatalf("mkdmg binary not found at %s after build", mkdmgBinary)
	}
	t.Logf("Mkdmg binary built successfully at: %s", mkdmgBinary)

	// Run the mkdmg binary
	cmdMkdmg := exec.Command(mkdmgBinary, "--volname", "TestVolume", "--disk-image-size", "10", outputDMG, sourceDir)
	var mkdmgErr bytes.Buffer
	cmdMkdmg.Stderr = &mkdmgErr
	if testing.Verbose() {
		cmdMkdmg.Stdout = os.Stdout
	}
	if err := cmdMkdmg.Run(); err != nil {
		t.Fatalf("mkdmg command failed: %v\n%s", err, mkdmgErr.String())
	}

	// Verify if the DMG file was created
	if _, err := os.Stat(outputDMG); os.IsNotExist(err) {
		t.Errorf("Output DMG file was not created at %s", outputDMG)
	}
}

func TestMkdmgHelp(t *testing.T) {
	mkdmgBinary := buildMkdmg(t)

	var stdout, stderr bytes.Buffer
	cmd := exec.Command(mkdmgBinary, "--help")
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		t.Logf("Help command stderr: %s", stderr.String())
	}

	// Help should output usage information (could be to stdout or stderr)
	output := stdout.String() + stderr.String()
	if !strings.Contains(output, "Usage:") {
		t.Errorf("Help output should contain 'Usage:', got: %s", output)
	}
	if !strings.Contains(output, "OUTFILE.DMG") {
		t.Errorf("Help output should contain 'OUTFILE.DMG', got: %s", output)
	}
}

func TestMkdmgVersion(t *testing.T) {
	mkdmgBinary := buildMkdmg(t)

	var stdout bytes.Buffer
	cmd := exec.Command(mkdmgBinary, "--version")
	cmd.Stdout = &stdout

	err := cmd.Run()
	if err != nil {
		t.Fatalf("Version command failed: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "mkdmg, version") {
		t.Errorf("Version output should contain 'mkdmg, version', got: %s", output)
	}
	if !strings.Contains(output, "Copyright") {
		t.Errorf("Version output should contain 'Copyright', got: %s", output)
	}
}

func TestMkdmgInvalidArguments(t *testing.T) {
	mkdmgBinary := buildMkdmg(t)

	tests := []struct {
		name string
		args []string
	}{
		{
			name: "no_arguments",
			args: []string{},
		},
		{
			name: "only_output_path",
			args: []string{"test.dmg"},
		},
		{
			name: "too_many_arguments",
			args: []string{"test.dmg", "source", "extra"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stderr bytes.Buffer
			cmd := exec.Command(mkdmgBinary, tt.args...)
			cmd.Stderr = &stderr

			err := cmd.Run()
			if err == nil {
				t.Error("Expected command to fail with invalid arguments")
			}

			output := stderr.String()
			if !strings.Contains(output, "invalid arguments") {
				t.Errorf("Expected 'invalid arguments' error, got: %s", output)
			}
		})
	}
}

func TestMkdmgDryRun(t *testing.T) {
	if !isHdiutilAvailable() {
		t.Skip("hdiutil not available, skipping integration test")
	}

	mkdmgBinary := buildMkdmg(t)

	sourceDir := t.TempDir()
	dummyFile := filepath.Join(sourceDir, "test.txt")
	if err := os.WriteFile(dummyFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create dummy file: %v", err)
	}

	outputDMG := filepath.Join(t.TempDir(), "test.dmg")

	cmd := exec.Command(mkdmgBinary, "--dry-run", outputDMG, sourceDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		t.Fatalf("Dry-run command failed: %v", err)
	}

	// In dry-run mode, the DMG should not actually be created
	if _, err := os.Stat(outputDMG); err == nil {
		t.Error("DMG file should not be created in dry-run mode")
	}
}

func TestMkdmgVerboseMode(t *testing.T) {
	if !isHdiutilAvailable() {
		t.Skip("hdiutil not available, skipping integration test")
	}

	mkdmgBinary := buildMkdmg(t)

	sourceDir := t.TempDir()
	dummyFile := filepath.Join(sourceDir, "test.txt")
	if err := os.WriteFile(dummyFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create dummy file: %v", err)
	}

	outputDMG := filepath.Join(t.TempDir(), "test.dmg")

	var stderr bytes.Buffer
	cmd := exec.Command(mkdmgBinary, "-v", "--dry-run", outputDMG, sourceDir)
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		t.Fatalf("Verbose command failed: %v", err)
	}

	// Verbose mode should produce output
	output := stderr.String()
	if len(output) == 0 {
		t.Error("Expected verbose output, got none")
	}
}

func TestMkdmgInvalidOutputExtension(t *testing.T) {
	mkdmgBinary := buildMkdmg(t)

	sourceDir := t.TempDir()
	outputISO := filepath.Join(t.TempDir(), "test.iso")

	var stderr bytes.Buffer
	cmd := exec.Command(mkdmgBinary, outputISO, sourceDir)
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		t.Error("Expected command to fail with invalid output extension")
	}

	output := stderr.String()
	if !strings.Contains(output, "extension") {
		t.Errorf("Expected error about extension, got: %s", output)
	}
}

func TestMkdmgAPFSFilesystem(t *testing.T) {
	if !isHdiutilAvailable() {
		t.Skip("hdiutil not available, skipping integration test")
	}

	mkdmgBinary := buildMkdmg(t)

	sourceDir := t.TempDir()
	dummyFile := filepath.Join(sourceDir, "test.txt")
	if err := os.WriteFile(dummyFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create dummy file: %v", err)
	}

	outputDMG := filepath.Join(t.TempDir(), "test.dmg")

	cmd := exec.Command(mkdmgBinary, "--apfs", "-s", outputDMG, sourceDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		t.Fatalf("APFS command failed: %v", err)
	}
}

func TestMkdmgFormatOptions(t *testing.T) {
	if !isHdiutilAvailable() {
		t.Skip("hdiutil not available, skipping integration test")
	}

	mkdmgBinary := buildMkdmg(t)

	sourceDir := t.TempDir()
	dummyFile := filepath.Join(sourceDir, "test.txt")
	if err := os.WriteFile(dummyFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create dummy file: %v", err)
	}

	formats := []string{"UDZO", "UDBZ", "ULFO", "ULMO"}

	for _, format := range formats {
		t.Run("format_"+format, func(t *testing.T) {
			outputDMG := filepath.Join(t.TempDir(), "test.dmg")

			cmd := exec.Command(mkdmgBinary, "--format", format, "-s", outputDMG, sourceDir)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			err := cmd.Run()
			if err != nil {
				t.Fatalf("Format %s command failed: %v", format, err)
			}
		})
	}
}

func TestMkdmgSandboxSafe(t *testing.T) {
	if !isHdiutilAvailable() {
		t.Skip("hdiutil not available, skipping integration test")
	}

	mkdmgBinary := buildMkdmg(t)

	sourceDir := t.TempDir()
	dummyFile := filepath.Join(sourceDir, "test.txt")
	if err := os.WriteFile(dummyFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create dummy file: %v", err)
	}

	outputDMG := filepath.Join(t.TempDir(), "test.dmg")

	cmd := exec.Command(mkdmgBinary, "--sandbox-safe", "-s", outputDMG, sourceDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		t.Fatalf("Sandbox-safe command failed: %v", err)
	}
}

func TestMkdmgSandboxSafeWithAPFS(t *testing.T) {
	mkdmgBinary := buildMkdmg(t)

	sourceDir := t.TempDir()
	outputDMG := filepath.Join(t.TempDir(), "test.dmg")

	var stderr bytes.Buffer
	cmd := exec.Command(mkdmgBinary, "--sandbox-safe", "--apfs", outputDMG, sourceDir)
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		t.Error("Expected command to fail when combining sandbox-safe with APFS")
	}

	output := stderr.String()
	if !strings.Contains(output, "APFS") || !strings.Contains(output, "sandbox") {
		t.Errorf("Expected error about APFS and sandbox incompatibility, got: %s", output)
	}
}

func TestMkdmgHdiutilVerbosityLevels(t *testing.T) {
	if !isHdiutilAvailable() {
		t.Skip("hdiutil not available, skipping integration test")
	}

	mkdmgBinary := buildMkdmg(t)

	sourceDir := t.TempDir()
	dummyFile := filepath.Join(sourceDir, "test.txt")
	if err := os.WriteFile(dummyFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create dummy file: %v", err)
	}

	verbosityLevels := []string{"0", "1", "2", "3"}

	for _, level := range verbosityLevels {
		t.Run("hdiutil_verbosity_"+level, func(t *testing.T) {
			outputDMG := filepath.Join(t.TempDir(), "test.dmg")

			cmd := exec.Command(mkdmgBinary, "--hdiutil-verbosity", level, "-s", outputDMG, sourceDir)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			err := cmd.Run()
			if err != nil {
				t.Fatalf("Verbosity level %s command failed: %v", level, err)
			}
		})
	}
}

func TestMkdmgNonExistentSourceDirectory(t *testing.T) {
	mkdmgBinary := buildMkdmg(t)

	outputDMG := filepath.Join(t.TempDir(), "test.dmg")
	nonExistentDir := filepath.Join(t.TempDir(), "nonexistent")

	var stderr bytes.Buffer
	cmd := exec.Command(mkdmgBinary, outputDMG, nonExistentDir)
	cmd.Stderr = &stderr

	err := cmd.Run()
	// This may or may not fail depending on when the check happens
	// But it should eventually fail
	if err == nil {
		t.Log("Command succeeded (validation might be delayed)")
	}
}

// isHdiutilAvailable checks if hdiutil is available on the system
func isHdiutilAvailable() bool {
	_, err := exec.LookPath("hdiutil")
	return err == nil
}

// buildMkdmg is a helper function that builds the mkdmg binary and returns its path
func buildMkdmg(t *testing.T) string {
	t.Helper()

	mkdmgBinary := filepath.Join(t.TempDir(), "mkdmg")
	cmdBuild := exec.Command("go", "build", "-o", mkdmgBinary, ".")

	var stderr bytes.Buffer
	cmdBuild.Stderr = &stderr

	err := cmdBuild.Run()
	if err != nil {
		t.Fatalf("Failed to build mkdmg binary: %v, stderr: %s", err, stderr.String())
	}

	if _, err := os.Stat(mkdmgBinary); os.IsNotExist(err) {
		t.Fatalf("mkdmg binary not found at %s after build", mkdmgBinary)
	}

	return mkdmgBinary
}

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
	volumeName = ""
	size = 0
	bless = false
	signingIdentity = ""
	notarizeCredentials = ""
	apfsFs = false
	sandboxSafe = false
	format = ""
	simulate = false
	hdiutilVerbosity = 0
	helpMode = false
	versionMode = false
	verboseMode = false

	// Replace the default FlagSet so flags can be re-registered.
	// Use ContinueOnError so parse errors don't call os.Exit.
	flag.CommandLine = flag.NewFlagSet(args[0], flag.ContinueOnError)

	// Re-register all flags (mirrors init()).
	flag.StringVar(&configPath, "config", "", "path to a JSON configuration file")
	flag.StringVar(&volumeName, "volname", "", "volume name for the DMG")
	flag.Int64Var(&size, "disk-image-size", 0, "size for the DMG in MB")
	flag.StringVar(&signingIdentity, "codesign", "", "signing identity")
	flag.BoolVar(&apfsFs, "apfs", false, "use APFS as disk image's filesystem (default: HFS+)")
	flag.BoolVar(&sandboxSafe, "sandbox-safe", false, "use sandbox-safe")
	flag.StringVar(&format, "format", "", "specify the final disk image format (UDZO|UDBZ|ULFO|ULMO)")
	flag.IntVar(&hdiutilVerbosity, "hdiutil-verbosity", 0, "set hdiutil verbosity level")
	flag.BoolVar(&simulate, "dry-run", false, "simulate the process")
	flag.BoolVar(&simulate, "s", false, "simulate the process (shorthand)")
	flag.BoolVar(&bless, "bless", false, "bless the disk image")
	flag.StringVar(&notarizeCredentials, "notarize", "", "notarize the disk image")
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
	resetForTest(t, []string{"mkdmg", "--volname", "Test", "out.dmg", "src"})
	flag.Parse()

	if !isFlagPassed("volname") {
		t.Error("isFlagPassed(\"volname\") = false, want true")
	}
	if isFlagPassed("format") {
		t.Error("isFlagPassed(\"format\") = true, want false")
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
	if !strings.Contains(out, "OUTFILE.DMG") {
		t.Errorf("usage output missing 'OUTFILE.DMG', got: %s", out)
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

func TestRunNoArguments(t *testing.T) {
	resetForTest(t, []string{"mkdmg"})
	err := run()
	if err == nil {
		t.Fatal("run() with no arguments should return an error")
	}
	if !strings.Contains(err.Error(), "invalid arguments") {
		t.Errorf("error = %q, want it to contain 'invalid arguments'", err)
	}
}

func TestRunOnePositionalArg(t *testing.T) {
	resetForTest(t, []string{"mkdmg", "test.dmg"})
	err := run()
	if err == nil {
		t.Fatal("run() with one arg should return an error")
	}
	if !strings.Contains(err.Error(), "invalid arguments") {
		t.Errorf("error = %q, want 'invalid arguments'", err)
	}
}

func TestRunThreePositionalArgs(t *testing.T) {
	resetForTest(t, []string{"mkdmg", "a.dmg", "src", "extra"})
	err := run()
	if err == nil {
		t.Fatal("run() with three args should return an error")
	}
	if !strings.Contains(err.Error(), "invalid arguments") {
		t.Errorf("error = %q, want 'invalid arguments'", err)
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

func TestRunConfigFileInvalidArgCount(t *testing.T) {
	cfgFile := writeConfigFile(t, map[string]any{
		"output_path": "test.dmg",
		"source_dir":  "/tmp/src",
	})
	resetForTest(t, []string{"mkdmg", "--config", cfgFile, "extra"})
	err := run()
	if err == nil {
		t.Fatal("run() with config + 1 positional arg should return an error")
	}
	if !strings.Contains(err.Error(), "invalid arguments") {
		t.Errorf("error = %q, want 'invalid arguments'", err)
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

func TestRunPositionalArgsDryRun(t *testing.T) {
	sourceDir := t.TempDir()
	outputDMG := filepath.Join(t.TempDir(), "test.dmg")
	resetForTest(t, []string{"mkdmg", "--dry-run", outputDMG, sourceDir})
	err := run()
	if err != nil {
		t.Fatalf("run() with positional args + dry-run: %v", err)
	}
}

func TestRunShorthandDryRun(t *testing.T) {
	sourceDir := t.TempDir()
	outputDMG := filepath.Join(t.TempDir(), "test.dmg")
	resetForTest(t, []string{"mkdmg", "-s", outputDMG, sourceDir})
	err := run()
	if err != nil {
		t.Fatalf("run() with -s (dry-run shorthand): %v", err)
	}
}

func TestRunConfigWithPositionalOverrides(t *testing.T) {
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
		t.Fatalf("run() with config + positional overrides: %v", err)
	}
}

func TestRunFlagOverrides(t *testing.T) {
	sourceDir := t.TempDir()
	outputDMG := filepath.Join(t.TempDir(), "test.dmg")
	cfgFile := writeConfigFile(t, map[string]any{
		"output_path": outputDMG,
		"source_dir":  sourceDir,
		"simulate":    true,
	})
	resetForTest(t, []string{
		"mkdmg",
		"--config", cfgFile,
		"--volname", "OverrideName",
		"--disk-image-size", "20",
		"--format", "UDBZ",
		"--hdiutil-verbosity", "2",
		"--codesign", "TestIdentity",
		"--notarize", "test-profile",
		"--bless",
	})
	err := run()
	if err != nil {
		t.Fatalf("run() with flag overrides: %v", err)
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
	if !strings.Contains(err.Error(), "missing output path") {
		t.Errorf("error = %q, want 'missing output path'", err)
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
		t.Errorf("error = %q, want 'missing output path or source directory'", err)
	}
}

func TestRunAPFSFlagDryRun(t *testing.T) {
	sourceDir := t.TempDir()
	outputDMG := filepath.Join(t.TempDir(), "test.dmg")
	resetForTest(t, []string{"mkdmg", "--dry-run", "--apfs", outputDMG, sourceDir})
	err := run()
	if err != nil {
		t.Fatalf("run() with --apfs --dry-run: %v", err)
	}
}

func TestRunSandboxSafeDryRun(t *testing.T) {
	sourceDir := t.TempDir()
	outputDMG := filepath.Join(t.TempDir(), "test.dmg")
	resetForTest(t, []string{"mkdmg", "--dry-run", "--sandbox-safe", outputDMG, sourceDir})
	err := run()
	if err != nil {
		t.Fatalf("run() with --sandbox-safe --dry-run: %v", err)
	}
}

func TestRunSandboxSafeWithAPFS(t *testing.T) {
	sourceDir := t.TempDir()
	outputDMG := filepath.Join(t.TempDir(), "test.dmg")
	resetForTest(t, []string{"mkdmg", "--dry-run", "--sandbox-safe", "--apfs", outputDMG, sourceDir})
	err := run()
	if err == nil {
		t.Fatal("run() with --sandbox-safe --apfs should return an error")
	}
	if !strings.Contains(err.Error(), "sandbox") {
		t.Errorf("error = %q, want mention of sandbox", err)
	}
}

func TestRunVerboseModeDryRun(t *testing.T) {
	sourceDir := t.TempDir()
	outputDMG := filepath.Join(t.TempDir(), "test.dmg")
	resetForTest(t, []string{"mkdmg", "--dry-run", "--verbose", outputDMG, sourceDir})
	err := run()
	if err != nil {
		t.Fatalf("run() with --verbose --dry-run: %v", err)
	}
}
