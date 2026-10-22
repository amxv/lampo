---
title: CLI reference
description: Every Lampo command option, document selection rule, and runtime detail.
order: 20
category: Reference
summary: Flags, entrypoint detection, LaTeX engines, network access, and how updates work.
---

## Command syntax

```bash
lampo [file.tex | directory] [options]
```

Lampo starts a local HTTP server, watches your source tree, runs `latexmk`, and displays the compiled PDF in your browser. Use **Ctrl+C** to stop it.

| Option | Default | Description |
| --- | --- | --- |
| `-p`, `--port <number>` | Random available port | Choose a fixed TCP port. `0` means any free port. |
| `--host <address>` | `127.0.0.1` | Address on which the preview listens. |
| `--engine <name>` | `pdflatex` | `pdflatex`, `xelatex`, or `lualatex`. |
| `--no-open` | Off | Print the URL without opening a browser tab. |
| `--open` | On | Open the preview in your default browser. |
| `-v`, `--version` | | Print the installed version. |
| `-h`, `--help` | | Print the command reference and examples. |

Options can appear before or after the directory or document path.

```bash
lampo
lampo paper.tex
lampo ./applications --port 3030
lampo ./thesis/main.tex --engine xelatex --no-open
```

## Selecting the document root

Lampo scans the project recursively and recognizes standalone `.tex` files by `\documentclass` or legacy `\documentstyle`. It ignores hidden directories and common generated folders such as `.git`, `node_modules`, and `dist`.

It chooses the initial document in this order: an explicitly provided `.tex` file, `main.tex` at the project root, or the first discovered root alphabetically. For a directory with just one `.tex` file, that file is accepted even without a document declaration.

Subfiles referenced with `\input` and `\include` are watched, but they aren't separate items in the document picker. Entry points in nested directories appear in the dropdown with their relative folder names when relevant.

The browser's **/** shortcut opens the document picker. Choose a file to switch without navigating away from the preview.

## Compilation and cache

- Lampo uses the `latexmk` binary and your selected LaTeX engine from PATH. LaTeX itself is not bundled with Lampo.
- Filesystem events detect edits to `.tex`, `.bib`, `.sty`, `.cls`, images, and other common inputs, including newly created nested directories.
- Saves are briefly coalesced, so an editor's burst of writes doesn't start unnecessary parallel builds.
- Each standalone document gets its own temporary `latexmk` output directory and cached last-good PDF. Switching back to a cached document is immediate until the cache is invalidated by a source change.
- A failed build retains the previous successful preview, and the full compiler output is available from the status indicator.
- Temporary build files are outside the project and are deleted when Lampo stops cleanly.

## Browser preview

The browser UI uses the bundled PDF.js renderer. The server sends build status and revision events over Server-Sent Events (SSE); the browser requests new PDF bytes when a successful revision is published.

Resizing is immediate because the existing rendered layers are transformed in place. After the resize settles, Lampo re-renders pages at the new size and the device's pixel density for sharp text. Text selection, PDF links, zoom, and named downloads remain available.

## Local endpoints

Lampo serves a small HTTP API on the same local port as the preview:

| Endpoint | Behavior |
| --- | --- |
| `GET /` | Embedded browser viewer. |
| `GET /api/status` | Selected document, document list, revision, and build status. |
| `GET /events` | SSE notifications for changes and builds. |
| `POST /api/select` | Switch documents using a discovered document ID. |
| `GET /preview.pdf` | Current document's last successful PDF. |
| `GET /api/export?filename=…` | Download the selected PDF using a chosen filename. |

## Security and privacy

By default, the server binds to `127.0.0.1`. It is accessible only from your own computer. `--host 0.0.0.0` is explicitly for trusted networks: other machines on that network may be able to read your previews and documents. Lampo does not authenticate preview clients or sandbox the LaTeX compiler.

Compile only LaTeX projects you trust. No project content is uploaded by Lampo, and the UI does not depend on a cloud service or external JavaScript CDN.

See [developer reference](/docs/customize-cli) for architecture and testing commands.
