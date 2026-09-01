# Testing Standards for `mkdmg`

1. **Simulation Precedence**: When testing CLI invocation (`run()`), always include `-s` or `--dry-run` in `os.Args` to prevent invoking unprivileged `hdiutil` commands.
2. **Mock Executor Pattern**: Unit tests verifying command strings or process return values should inject a mock `hdiutil.CommandExecutor` using `hdiutil.WithExecutor()`.
3. **Hermetic State**: Tests must never rely on pre-existing disk files or modify user files. Use `t.TempDir()` and cleanup handlers.
4. **Race Detection**: All tests must pass cleanly under `go test -race`.
5. **No Network Dependencies in Unit Tests**: Unit tests must not perform network calls (e.g. contacting Apple Notary service).
