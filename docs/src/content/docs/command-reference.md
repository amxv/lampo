---
title: Command reference
description: Lampo CLI options, project detection, and preview behavior.
order: 4
category: Reference
summary: Flags and examples for local LaTeX preview sessions.
---

## Usage

```bash
lampo [file.tex | directory] [options]
```

Without arguments, Lampo previews the current working directory. If the directory contains `main.tex`, that's the default. Otherwise, if exactly one `.tex` file exists, it's selected. For multiple files, Lampo can select the unique file containing `\documentclass`; if ambiguous, pass a specific filename.

| Option | Default | Behavior |
| --- | --- | --- |
| `-p`, `--port NUMBER` | Free port | Set a fixed HTTP port (0 means random) |
| `--host ADDRESS` | `127.0.0.1` | Bind to another network address |
| `--engine NAME` | `pdflatex` | Choose `pdflatex`, `xelatex`, or `lualatex` |
| `--no-open` | Off | Print the URL but don't open a browser |
| `--open` | On | Open the browser when starting |
| `-h`, `--help` | | Show help and usage examples |
| `-v`, `--version` | | Display the current binary version |

Options can be placed before or after the project path.

## Examples

```bash
lampo
lampo /path/to/resume/main.tex
lampo ./research --port 4444
lampo paper.tex --engine lualatex --no-open
```

## Watching and compilation

- Saves to `.tex`, `.bib`, `.sty`, `.cls`, figures and other source asset types trigger builds; new nested directories are watched automatically.
- Short, burst-coalescing delays prevent duplicate rebuilds during editor saves. Superseded builds are canceled.
- Only fully compiled PDFs appear in the browser. If a build fails, the previous PDF remains readable and the error log is available.
- Resizing the browser scales existing PDF layers without re-rendering them. Zoom changes take effect immediately; higher zoom levels sharpen the PDF after interaction settles.
- Auxiliary build files live outside the source directory and are removed when the server stops.

The browser uses bundled Mozilla PDF.js; communication is over local HTTP and server-sent events. Lampo does not require Node.js at runtime and makes no requests to a cloud service.

## Security

The default listener accepts local connections (`127.0.0.1`). `--host 0.0.0.0` exposes the preview and PDF to anyone who can reach that port, so only use it on trusted networks. Lampo does not sandbox LaTeX itself; compile trusted sources.
