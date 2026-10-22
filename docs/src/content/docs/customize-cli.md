---
title: Contributing to Lampo
description: Understand the Go compiler/watch loop and contribute improvements.
order: 2
category: Development
summary: Architecture, local tests, and how to iterate on the embedded browser UI.
---

Lampo has a small Go backend and a plain JavaScript/CSS browser frontend. There's no JavaScript bundler in the application runtime or release build.

## Layout

```text
cmd/lampo/             Go CLI entrypoint
internal/app/          Flags, help, version, interrupt handling
internal/latex/        Standalone document discovery, project selection, latexmk
internal/watch/        Recursive filesystem events + source filtering
internal/preview/      HTTP, SSE, per-document build/cache coordinator, PDF downloads
internal/preview/web/  Embedded static HTML, JavaScript, CSS, PDF.js
docs/                 Astro/ZueDocs site (not bundled into CLI)
```

## Run the development version

```bash
go run ./cmd/lampo /path/to/project/main.tex --no-open --port 3030
```

When changing HTML, JS, or CSS, restart the Go process to embed the new assets. Restarting isn't necessary when changing the watched **LaTeX project**.

## Validate changes

```bash
make check
go test -race ./...
make build
make docs-check
make docs-build
```

Run the docs check and docs build serially so Astro doesn't have two processes writing `.astro/` at once.

## Build model

The server discovers LaTeX document roots recursively and gives each one its own temporary `latexmk` directory and cached PDF. A filesystem watcher observes nested source directories and sends coalesced change events. The compiler coordinator cancels superseded builds and publishes a new PDF only after a successful build. The SSE stream carries document lists, selected entrypoint, status and revision IDs; the browser switches without navigation and renders with bundled PDF.js.

Important invariants: never overwrite the user's source files, never serve a partial PDF, retain the last good output on failure, and keep the default server loopback-only.

## Third-party libraries

Lampo uses [fsnotify](https://github.com/fsnotify/fsnotify) for filesystem events and [Mozilla PDF.js](https://mozilla.github.io/pdf.js/) to display PDFs. The prebuilt PDF.js files and its license notice are committed so a released Go binary works completely offline.
