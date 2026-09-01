# Go Coding Standards & Conventions for `mkdmg`

1. **Standard Library First**: Prefer Go standard library packages (`flag`, `os`, `path/filepath`, `io`, `fmt`, `log`, `embed`) over introducing external dependencies.
2. **Error Wrapping & Propagation**: Always wrap underlying errors with actionable context using `fmt.Errorf("failed to ...: %w", err)` or `fmt.Errorf("...: %v", err)` where appropriate. Never swallow errors.
3. **Explicit Cleanup**: Always pair resource allocations (open files, created directories, mount points) with a `defer` call or an explicit cleanup handler.
4. **Log Formatting**: Use `log.New(...)` for configurable logging outputs. Direct standard application logs and verbose diagnostics to `os.Stderr`.
5. **No Global Mutation in Production Code**: Keep mutable state isolated. Reset command line flags cleanly during tests.
