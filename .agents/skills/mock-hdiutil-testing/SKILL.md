---
name: mock-hdiutil-testing
description: >-
  Write and run isolated unit tests for mkdmg using mock executors and simulation modes.
  Use this skill when testing CLI logic, configuration validation, or runner workflows without invoking real macOS hdiutil commands.
---

# Mock HDIUtil Testing Guide

This skill provides patterns and instructions for testing `mkdmg` in environments where real macOS disk mounting / `hdiutil` commands cannot or should not be executed (such as restricted sandboxes, CI Linux containers, or developer machines without root privileges).

---

## 1. Simulation Flag Mode (`-s` / `--dry-run`)

When testing `run()` in `main_test.go`, always ensure that `-s` or `--dry-run` is included in `os.Args`:

```go
func TestCLIWithSimulation(t *testing.T) {
    sourceDir := t.TempDir()
    outputDMG := filepath.Join(t.TempDir(), "test.dmg")
    cfgFile := writeConfigFile(t, map[string]any{
        "output_path": outputDMG,
        "source_dir":  sourceDir,
    })
    
    // Pass -s to ensure Runner.SetSimulate(true) is activated
    resetForTest(t, []string{"mkdmg", "-s", "--config", cfgFile})
    
    err := run()
    if err != nil {
        t.Fatalf("run() failed: %v", err)
    }
}
```

---

## 2. Using `CommandExecutor` Mocks

For unit tests targeting `al.essio.dev/pkg/hdiutil`, use the mock executor pattern:

```go
type mockExecutor struct {
    calls []string
}

func (m *mockExecutor) Hdiutil(args ...string) ([]byte, error) {
    m.calls = append(m.calls, "hdiutil "+strings.Join(args, " "))
    return []byte("created"), nil
}

func (m *mockExecutor) Codesign(args ...string) ([]byte, error) {
    m.calls = append(m.calls, "codesign "+strings.Join(args, " "))
    return nil, nil
}

func (m *mockExecutor) Xcrun(args ...string) ([]byte, error) {
    m.calls = append(m.calls, "xcrun "+strings.Join(args, " "))
    return nil, nil
}

func (m *mockExecutor) Chmod(args ...string) error {
    return nil
}

func (m *mockExecutor) Bless(args ...string) error {
    return nil
}
```

Inject the mock executor via `hdiutil.WithExecutor`:
```go
runner := hdiutil.New(cfg, hdiutil.WithExecutor(mock))
```

---

## 3. Hermetic Test Guidelines

1. **Temporary Directories**: Always use `t.TempDir()` for temporary source files and DMG output targets.
2. **State Reset**: Always restore global state (`flag.CommandLine`, `os.Args`, global flags) with `t.Cleanup()`.
3. **No Network / No Disk Mounts**: Verify that tests run successfully offline and without disk attachment.
