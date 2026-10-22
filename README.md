# Lampo ⚡

**Live LaTeX previews, without leaving your browser.**

Lampo is a small, local-first Go CLI that watches a LaTeX project, automatically rebuilds its PDF as you edit, and updates a clean browser preview in place. Designed for people who write `.tex` with their editor or coding agents and want the feedback loop of a web dev server.

> **Pre-release:** Lampo works locally on `main`, but the first npm/GitHub binary release has not shipped yet.

## Try it locally

Install Go 1.26+ and a TeX distribution with `latexmk` and `pdflatex` (MacTeX, TeX Live, TinyTeX or MiKTeX). Then:

```bash
git clone https://github.com/amxv/lampo.git
cd lampo
make install-local

# Anywhere you have a LaTeX project:
cd ~/path/to/latex-project
lampo
```

Lampo opens a browser tab with the PDF preview and prints the local URL in your terminal. Make changes in any editor, or ask your coding agent to edit the project. The browser updates after a successful compilation, without refreshing. Click **Export PDF** to choose a filename and download the result.

```bash
lampo                       # Find main.tex or an unambiguous .tex file here
lampo ./resume/main.tex     # Preview a specific document
lampo ./paper --port 3030   # Use a fixed localhost port
lampo main.tex --no-open    # Just print the URL
lampo main.tex --engine xelatex
lampo --help
```

Lampo runs entirely on your machine: no account, uploads, SaaS, telemetry, CDN, Node.js runtime or frontend build step. It embeds the browser UI and PDF.js inside one Go binary, and stores compilation outputs in a temporary directory, leaving your project clean.

## Built for fast iteration

- **Fast change detection:** Native file events with a short debounce; detects updates to nested `.tex`, `.bib`, styles, images and other source assets, including atomic editor saves.
- **Instant and sharp resizing:** Drag a split-view divider without expensive redraws; after resizing settles, PDF.js quietly sharpens each page at its exact display size and Retina pixel density without blanking the preview.
- **No page reload:** Server-sent events notify the preview when a new PDF is ready. Your browser preserves the preview while the next build runs.
- **Useful build errors:** Compiler error banner and log panel, without discarding the last good PDF.
- **Browser controls:** Fit-to-width, zoom, selectable text, clickable external links, a page counter, and named PDF export.
- **Portable:** Native Go binary for macOS, Windows and Linux; runs against locally installed `latexmk`.

## How it works

```text
Your editor / coding agent -> .tex files -> fsnotify -> latexmk
                                                      |
                                      temporary compiled PDF
                                                      |
Browser preview (bundled PDF.js) <- SSE + local HTTP server
```

LaTeX engines: `pdflatex` (default), `xelatex` and `lualatex`. If the TeX compiler is missing a LaTeX package, install it using your TeX distribution; Lampo displays the compilation error and rebuilds when you save again.

**Privacy:** Lampo binds to `127.0.0.1` on a random free port by default. Only use `--host` to expose it to other machines on a trusted network. As with any LaTeX build, compile only source documents you trust.

## Develop

```bash
make check
make build
make docs-install
make docs-check
make docs-build
```

Source layout: `internal/app/` (CLI), `internal/latex/` (compiler), `internal/watch/` (file watching), `internal/preview/` (HTTP, SSE, embedded UI), `docs/` (Astro/ZueDocs). PDF.js is bundled under `internal/preview/web/vendor/` with its Apache-2.0 license notice.

When the first release is approved, GitHub Actions will publish native binaries and the `@amxv/lampo` npm wrapper. **No release tag has been created yet.**

Documentation: [lampo.ashray.xyz](https://lampo.ashray.xyz) · Source: [github.com/amxv/lampo](https://github.com/amxv/lampo) · License: Apache-2.0.
