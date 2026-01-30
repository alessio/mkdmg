package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestMkdmgBasicCreation(t *testing.T) {
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
	cmdBuild.Stdout = os.Stdout
	cmdBuild.Stderr = os.Stderr
	err = cmdBuild.Run()
	if err != nil {
		t.Fatalf("Failed to build mkdmg binary: %v", err)
	}

	// Check if the binary exists after building
	if _, err := os.Stat(mkdmgBinary); os.IsNotExist(err) {
		t.Fatalf("mkdmg binary not found at %s after build", mkdmgBinary)
	}
	t.Logf("Mkdmg binary built successfully at: %s", mkdmgBinary)

	// Run the mkdmg binary
	cmdMkdmg := exec.Command(mkdmgBinary, outputDMG, sourceDir, "--volname", "TestVolume", "--disk-image-size", "10")
	cmdMkdmg.Stdout = os.Stdout
	cmdMkdmg.Stderr = os.Stderr
	err = cmdMkdmg.Run()
	if err != nil {
		t.Fatalf("mkdmg command failed: %v", err)
	}

	// Verify if the DMG file was created
	if _, err := os.Stat(outputDMG); os.IsNotExist(err) {
		t.Errorf("Output DMG file was not created at %s", outputDMG)
	}
}
