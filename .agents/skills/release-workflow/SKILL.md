---
name: release-workflow
description: >-
  Manage the release process for mkdmg using GoReleaser and git tags.
  Use this skill when preparing a new version release, dry-running GoReleaser, creating semantic tags, or verifying release checksums.
---

# Release Management Runbook for `mkdmg`

This runbook guides the release process for publishing new versions of `mkdmg`.

---

## 1. Pre-Release Checklist

1. **Working Tree Clean**: Ensure all changes are committed and pushed to `main`.
2. **CI Passing**: Confirm all GitHub Actions workflows (`Go`, `golangci-lint`, `codacy`) are green.
3. **Tests Passing**: Run `make check` locally.
4. **Docs Updated**: Ensure `README.md` and `docs/index.html` reflect any new options or features.

---

## 2. GoReleaser Dry-Run Test

Test the build and packaging without publishing:
```sh
goreleaser release --snapshot --clean
```
Inspect artifacts generated in `dist/`:
- `dist/mkdmg_Darwin_x86_64/mkdmg`
- `dist/mkdmg_Darwin_arm64/mkdmg`
- `dist/checksums.txt`

---

## 3. Creating & Pushing a Semantic Version Tag

Determine the next semantic version (e.g. `v0.4.2`):

```sh
# 1. Create annotated git tag
git tag -a v0.4.2 -m "Release v0.4.2"

# 2. Push tag to GitHub to trigger the release workflow
git push origin v0.4.2
```

---

## 4. Post-Release Verification

1. Monitor the GitHub Actions `goreleaser` workflow in `.github/workflows/release.yml`.
2. Verify the release on GitHub Releases:
   - Verify assets: `mkdmg_Darwin_x86_64.tar.gz`, `mkdmg_Darwin_arm64.tar.gz`, `checksums.txt`.
   - Verify generated changelog.
3. Verify binary installation via Go:
   ```sh
   go install al.essio.dev/cmd/mkdmg@v0.4.2
   ```
