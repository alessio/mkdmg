package main

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

// parseGoMod reads go.mod and returns a map of directive -> value for simple
// single-value directives (go, toolchain, module).
func parseGoMod(t *testing.T) map[string]string {
	t.Helper()

	data, err := os.ReadFile("go.mod")
	if err != nil {
		t.Fatalf("failed to read go.mod: %v", err)
	}

	result := make(map[string]string)
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		parts := strings.Fields(line)
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}
	return result
}

// TestGoModToolchainDirectiveExists verifies that the toolchain directive is
// present in go.mod.
func TestGoModToolchainDirectiveExists(t *testing.T) {
	directives := parseGoMod(t)
	if _, ok := directives["toolchain"]; !ok {
		t.Error("go.mod is missing the 'toolchain' directive")
	}
}

// TestGoModToolchainVersion verifies that the toolchain is set to the version
// introduced by this PR (go1.26.4).
func TestGoModToolchainVersion(t *testing.T) {
	directives := parseGoMod(t)
	got, ok := directives["toolchain"]
	if !ok {
		t.Fatal("go.mod is missing the 'toolchain' directive")
	}
	const want = "go1.26.4"
	if got != want {
		t.Errorf("toolchain = %q, want %q", got, want)
	}
}

// TestGoModToolchainVersionFormat verifies that the toolchain value follows the
// expected goMAJOR.MINOR.PATCH format.
func TestGoModToolchainVersionFormat(t *testing.T) {
	directives := parseGoMod(t)
	toolchain, ok := directives["toolchain"]
	if !ok {
		t.Fatal("go.mod is missing the 'toolchain' directive")
	}

	// Must match goN.N.N (with optional rc/beta suffix such as go1.26.0rc1).
	pattern := regexp.MustCompile(`^go\d+\.\d+(\.\d+)?(rc\d+|beta\d+)?$`)
	if !pattern.MatchString(toolchain) {
		t.Errorf("toolchain %q does not match expected format goMAJOR.MINOR[.PATCH]", toolchain)
	}
}

// TestGoModToolchainNotDowngraded is a regression test ensuring the toolchain
// is never set to a version older than go1.26.4.
func TestGoModToolchainNotDowngraded(t *testing.T) {
	directives := parseGoMod(t)
	toolchain, ok := directives["toolchain"]
	if !ok {
		t.Fatal("go.mod is missing the 'toolchain' directive")
	}

	// Simple lexicographic comparison works for same-major same-minor versions.
	const minimum = "go1.26.4"
	if toolchain < minimum {
		t.Errorf("toolchain %q is older than the minimum required %q", toolchain, minimum)
	}
}

// TestGoModToolchainConsistentWithGoDirective verifies the toolchain version is
// not older than the 'go' language directive in go.mod.
func TestGoModToolchainConsistentWithGoDirective(t *testing.T) {
	directives := parseGoMod(t)

	goVersion, ok := directives["go"]
	if !ok {
		t.Fatal("go.mod is missing the 'go' directive")
	}

	toolchain, ok := directives["toolchain"]
	if !ok {
		t.Fatal("go.mod is missing the 'toolchain' directive")
	}

	// toolchain must be >= go directive (strip the "go" prefix for comparison).
	goVerNum := strings.TrimPrefix(goVersion, "")    // e.g. "1.26"
	tcVerNum := strings.TrimPrefix(toolchain, "go")  // e.g. "1.26.4"

	// The toolchain version should start with (or be newer than) the go directive.
	if !strings.HasPrefix(tcVerNum, goVerNum) && tcVerNum < goVerNum {
		t.Errorf("toolchain %q appears older than the go directive %q", toolchain, goVersion)
	}
}
