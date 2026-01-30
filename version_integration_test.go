package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestVersionEmbedding(t *testing.T) {
	// Ensure we can run git
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not found, skipping version embedding test")
	}

	// 1. Run go generate to create version.txt
	target := "./internal/version"
	genCmd := exec.Command("go", "generate", target)
	genCmd.Stdout = os.Stdout
	genCmd.Stderr = os.Stderr
	if err := genCmd.Run(); err != nil {
		t.Fatalf("go generate failed: %v", err)
	}

	// 2. Build the binary
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "mkdmg-version-test")
	
	buildCmd := exec.Command("go", "build", "-o", binaryPath, ".")
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("go build failed: %v", err)
	}

	// 3. Run the binary with --version
	versionCmd := exec.Command(binaryPath, "--version")
	var out bytes.Buffer
	versionCmd.Stdout = &out
	versionCmd.Stderr = os.Stderr
	
	if err := versionCmd.Run(); err != nil {
		t.Fatalf("failed to run produced binary with --version: %v", err)
	}

	output := out.String()
	t.Logf("Binary version output:\n%s", output)

	// 4. Verify the output
	// Expected format: "mkdmg, version <version_string>"
	if !strings.Contains(output, "mkdmg, version") {
		t.Errorf("Output does not contain 'mkdmg, version'. Got:\n%s", output)
	}
	
	// Check that it's not unknown or empty if git is working
	lines := strings.Split(output, "\n")
	if len(lines) > 0 {
		firstLine := lines[0]
		parts := strings.Split(firstLine, "version ")
		if len(parts) < 2 {
			t.Errorf("Could not parse version from line: %s", firstLine)
		} else {
			v := strings.TrimSpace(parts[1])
			if v == "" {
				t.Error("Version string is empty")
			}
		}
	}
}
