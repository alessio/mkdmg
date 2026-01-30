package hdiutil_test

import (
	"bytes"
	"os"
	"reflect"
	"testing"

	"al.essio.dev/cmd/mkdmg/pkg/hdiutil"
)

func TestConfig_JSON(t *testing.T) {
	t.Parallel()

	original := &hdiutil.Config{
		VolumeName:          "MyVolume",
		VolumeSizeMb:        100,
		SandboxSafe:         true,
		Bless:               true,
		FileSystem:          "HFS+",
		SigningIdentity:     "Developer ID Application: Test",
		NotarizeCredentials: "test-profile",
		ImageFormat:         "UDZO",
		HDIUtilVerbosity:    2,
		OutputPath:          "test.dmg",
		SourceDir:           "src",
		Simulate:            true,
	}

	// Test ToJSON
	var buf bytes.Buffer
	if err := original.ToJSON(&buf); err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	// Test FromJSON
	decoded := &hdiutil.Config{}
	if err := decoded.FromJSON(&buf); err != nil {
		t.Fatalf("FromJSON failed: %v", err)
	}

	// Use reflect.DeepEqual to compare structs, but need to be careful with unexported fields
	// Since original and decoded are fresh, unexported fields should be zeroed/nil in both.
	// However, we should check the exported fields specifically if they matter.

	if !reflect.DeepEqual(original, decoded) {
		t.Errorf("Decoded config does not match original\nOriginal: %+v\nDecoded:  %+v", original, decoded)
	}
}

func TestConfig_FromJSON_Partial(t *testing.T) {
	t.Parallel()

	jsonStr := `{"volume_name": "Test", "output_path": "out.dmg", "source_dir": "src"}`
	buf := bytes.NewBufferString(jsonStr)

	cfg := &hdiutil.Config{}
	if err := cfg.FromJSON(buf); err != nil {
		t.Fatalf("FromJSON failed: %v", err)
	}

	if cfg.VolumeName != "Test" {
		t.Errorf("Expected VolumeName 'Test', got '%s'", cfg.VolumeName)
	}
	if cfg.OutputPath != "out.dmg" {
		t.Errorf("Expected OutputPath 'out.dmg', got '%s'", cfg.OutputPath)
	}
	if cfg.SourceDir != "src" {
		t.Errorf("Expected SourceDir 'src', got '%s'", cfg.SourceDir)
	}
}

func TestLoadConfig(t *testing.T) {
	t.Parallel()

	tmpFile := t.TempDir() + "/config.json"
	jsonStr := `{"volume_name": "TestFile", "output_path": "file.dmg", "source_dir": "src"}`
	if err := os.WriteFile(tmpFile, []byte(jsonStr), 0644); err != nil {
		t.Fatalf("Failed to create temp config file: %v", err)
	}

	cfg, err := hdiutil.LoadConfig(tmpFile)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.VolumeName != "TestFile" {
		t.Errorf("Expected VolumeName 'TestFile', got '%s'", cfg.VolumeName)
	}
}
