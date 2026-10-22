---
title: Release & distribution
description: Tag-driven GitHub binaries, npm publication, and installer behavior.
order: 23
category: Reference
summary: Maintainer guide to shipping native binaries and @amxv/lampo on npm.
---

Lampo publishes a single Go CLI for macOS, Linux, and Windows. The `@amxv/lampo` npm package installs the matching prebuilt binary from GitHub Releases.

## First-party release pipeline

The release workflow lives at `.github/workflows/release.yml`. Pushing a `v*` tag starts four jobs:

1. **Quality:** Run Go tests, `go vet`, and Node shim checks.
2. **Build:** Cross-compile native binaries for macOS, Linux, and Windows on amd64 and arm64.
3. **GitHub Release:** Attach all build assets to the tagged release.
4. **npm:** Publish `@amxv/lampo` using the `NPM_TOKEN` GitHub Actions secret.

Release assets use the exact `<binary>_<goos>_<goarch>[.exe]` convention, for example `lampo_darwin_arm64` or `lampo_windows_amd64.exe`. The npm installer and GitHub Action must keep these filenames synchronized.

## Cut a release

Update `docs/src/content/docs/changelog.md` with the actual user-facing behavior of the release. Then run:

```bash
make check
make docs-check
make docs-build
git add .
git commit -m "chore: prepare release"
git push origin main
make release-tag VERSION=0.1.0
```

Replace `0.1.0` with the next intended version. A tag triggers publication, so only push it after the checks pass and the release is approved.

## Verify the release

```bash
gh run list -R amxv/lampo --limit 5
gh release view v0.1.0 -R amxv/lampo
npm view @amxv/lampo version
```

For end-to-end validation, install the published npm package in a clean temporary directory and confirm that `lampo --version` runs. A real LaTeX project also requires a local TeX distribution with `latexmk`.

## Installer fallback

The npm package contains a JavaScript shim and an installation script. When a release binary is available, the postinstall script downloads it into the package's `bin/` directory. If the download fails, it attempts a local Go build when a supported Go toolchain and the source files are present.

The Go module, source files, and PDF.js assets are included in the npm fallback package. The browser UI does not use npm, Node, or an external CDN at runtime.

## Credentials and permissions

The GitHub repository must have `NPM_TOKEN` configured as an Actions secret with publish access to the npm package. GitHub Actions also requires permission to create and upload release assets. Never commit credentials or tokens to the repository.
