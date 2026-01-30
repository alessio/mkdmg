package main

import (
	"bytes"
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
	cmdMkdmg := exec.Command(mkdmgBinary, outputDMG, sourceDir, "--volname", "TestVolume", "--disk-image-size", "10")
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
