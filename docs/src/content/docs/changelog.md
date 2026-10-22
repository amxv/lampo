---
title: Changelog
description: "Release notes for this CLI."
order: 99
category: Distribution
summary: Upcoming release notes. No version has been published yet.
---

Lampo has not shipped its first release. The Go preview CLI is implemented on `main` and is being tested before tagging.

## Unreleased - 0.1.0 candidate

- Local watch/compile loop for TeX projects, with automatic root-file discovery.
- Embedded PDF.js browser preview with live updates, zoom, text selection and PDF export.
- Instant, redraw-free window resizing for smooth browser split-screen layouts.
- Retina-sharp text at any preview width, with asynchronous per-page rendering once resizing stops.
- Compiler error feedback that preserves the last successful PDF.
- Portable Go CLI and npm wrapper, with release publication pending testing.

## Maintainer notes

When adding future entries:

- Keep the newest version at the top.
- Add sections only for version tags that exist.
- Summarize code and product changes from `main` since the previous version tag.
- This is an OSS template, so internal code changes may be included when useful.
- Do not include docs-site-only changes such as site styling, docs package bumps, deploy plumbing, footer/layout changes, or documentation navigation changes.
- Rewrite commit subjects into clear release notes instead of pasting raw commit messages.
- If a release contains only tagging/release metadata, write: `Maintenance release. No direct code behavior changes beyond release preparation.`
