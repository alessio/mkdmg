package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func TestMkdmgBasicCreation(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("skipping macOS-specific test on", runtime.GOOS)
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
