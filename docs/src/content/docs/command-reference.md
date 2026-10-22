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

Without arguments, Lampo previews the current working directory. It discovers standalone `.tex` entrypoints recursively, ignores helper snippets, and prefers root-level `main.tex` as the initial preview. If there are several roots and no `main.tex`, it chooses the first alphabetically. All detected documents remain available from the toolbar's searchable filename dropdown. An explicit `.tex` path selects that file first.

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
- New standalone documents appear in the dropdown automatically, including roots in subfolders. Two documents with the same filename are distinguished by their relative directories. Removing the selected document switches back to `main.tex` or the next available root.
- Switching selects the correct cached PDF instantly, or compiles the file once if necessary. Every root has a separate auxiliary build directory to avoid output collisions.
- Press `/` to search the document menu, use arrow keys to navigate, and press Enter to switch. Escape closes the menu.
- Short, burst-coalescing delays prevent duplicate rebuilds during editor saves. Superseded builds are canceled.
- Only fully compiled PDFs appear in the browser. If a build fails, the previous PDF remains readable and the error log is available.
- Window resizing and zoom respond immediately by transforming existing layers. After interaction settles, PDF.js renders fresh, crisp text at the final scale and display's pixel density, replacing one completed page at a time without blanking or moving the preview. Intermediate resize frames never trigger expensive redraws.
- Auxiliary build files live outside the source directory and are removed when the server stops.

The browser uses bundled Mozilla PDF.js; communication is over local HTTP and server-sent events. Lampo does not require Node.js at runtime and makes no requests to a cloud service.

## Security

The default listener accepts local connections (`127.0.0.1`). `--host 0.0.0.0` exposes the preview and PDF to anyone who can reach that port, so only use it on trusted networks. Lampo does not sandbox LaTeX itself; compile trusted sources.
