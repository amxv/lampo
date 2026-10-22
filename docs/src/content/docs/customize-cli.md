---
title: Developer reference
description: Source layout, compiler coordination, browser frontend, and verification commands.
order: 21
category: Reference
summary: Architecture and local development of the Go CLI and embedded PDF viewer.
---

Lampo is intentionally small: a Go CLI runs the local server and the browser frontend is plain JavaScript, CSS, and bundled PDF.js. There is no SPA framework or application bundler in the runtime.

## Source layout

```text
cmd/lampo/             CLI entrypoint
internal/app/          Flags, help, signal handling
internal/latex/        Root discovery, compile with latexmk
internal/watch/        Recursive filesystem event watching
internal/preview/      HTTP, SSE, build coordination, PDF cache
internal/preview/web/  Embedded HTML, JS, CSS, PDF.js assets
docs/                 Astro/ZueDocs website (not in CLI binary)
```

## Run from source

With Go 1.26+ installed:

```bash
go run ./cmd/lampo ~/Documents/my-paper/main.tex --port 3030 --no-open
```

Changing the watched LaTeX source triggers a build as usual. Changing embedded frontend assets requires rebuilding and restarting the Go server.

## Build and test

```bash
make check           # Formatting, tests, vet, script checks
go test -race ./...  # Race detection
make build           # Local native binary
make build-all       # Cross-platform native binaries
make install-local   # Install to ~/.local/bin/lampo

make docs-install
make docs-check
make docs-build
```

Run `make docs-check` and `make docs-build` serially because both share Astro-generated files.

## Runtime design

The build coordinator owns the selected document, active `latexmk` subprocess, and per-document cache. Source edits are debounced, superseded compiles are canceled, and only fully compiled PDF bytes reach the viewer. Each document receives its own temporary output directory to avoid auxiliary-file collisions.

The status snapshot includes available documents, the selected relative ID, a sequence number, revision, error text, build log, and last-known PDF availability. Browser clients receive snapshots over SSE and load new PDF bytes from the local server.

The viewer changes zoom and layout immediately using CSS transforms, then re-renders sharp pages when interaction settles. Keep this instant-resize behavior when changing the PDF pipeline.

## Safety invariants

- Never touch or overwrite LaTeX source files or store private projects in the public repository.
- Validate document switching against the discovered source list. The HTTP API must not resolve arbitrary filesystem paths.
- Retain last-good PDFs on compiler failure, and don't publish a stale PDF for a different selected document.
- Keep the server loopback-only by default. It is not an authenticated web service.
- Test the complete save → recompile → browser update → error → recovery sequence with a synthetic fixture before exercising private projects.

## Third-party dependencies

Go uses [fsnotify](https://github.com/fsnotify/fsnotify) for platform-specific filesystem events. The frontend bundles [Mozilla PDF.js](https://mozilla.github.io/pdf.js/) and its Apache-2.0 license notice in `internal/preview/web/vendor/`.

The separate docs workspace uses Astro and ZueDocs. Deployment and maintenance are covered in [Docs website reference](/docs/docs-site).
