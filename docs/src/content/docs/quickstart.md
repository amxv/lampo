---
title: Quickstart
description: Run a browser preview of your LaTeX project in one command.
order: 1
category: Start
summary: Install Lampo locally, open your preview, edit files, and export a PDF.
---

Lampo is a **local live-preview server** for LaTeX. It watches your source files, invokes your installed LaTeX compiler, and shows the output in your browser. No accounts or cloud compilation.

> Lampo is available in the public source repository for testing; the first npm release hasn't shipped yet.

## Prerequisites

- [Go](https://go.dev/dl/) 1.26 or newer, for building from source during pre-release testing.
- A local TeX distribution with `latexmk` and `pdflatex` on PATH (MacTeX/TeX Live, TinyTeX or MiKTeX). XeLaTeX and LuaLaTeX are optional.

## Install from source

```bash
git clone https://github.com/amxv/lampo.git
cd lampo
make install-local
```

The installer puts the binary at `~/.local/bin/lampo`. Make sure that directory is on PATH. The binary contains the entire UI, so there's nothing else to install for the browser.

## Preview your project

```bash
cd ~/path/to/your/latex-project
lampo
```

Lampo discovers standalone `.tex` documents in the current folder and its subfolders. It starts with `main.tex` if present, otherwise the first discovered entrypoint. To switch documents, click the current filename in the preview toolbar, search for a document, or press `/` to open the searchable picker. Switching keeps you in the **same browser tab** and reuses cached PDFs.

To start on a particular document instead, specify its filename:

```bash
lampo thesis.tex
```

The terminal prints a URL on `127.0.0.1`, and the browser opens. Edit your `.tex` file and save it. Lampo rebuilds and replaces the preview without reloading the tab. Click **Export PDF**, enter your filename, and download the finished PDF.

**What appears in the picker?** Standalone files declaring `\documentclass` (or legacy `\documentstyle`). Files containing only an included chapter or reusable LaTeX macro are watched but omitted from the menu. Newly added or removed documents appear automatically without restarting Lampo. If only one `.tex` file exists, it is still selectable even without a document declaration.

## Useful commands

```bash
lampo mydoc.tex --no-open       # Show the URL without opening a tab
lampo . --port 3030            # Stable URL for split-screen setups
lampo ./paper --engine xelatex # Use XeLaTeX
```

The view stays open as long as the command runs. Use **Ctrl+C** to stop it. The temporary TeX build directory is cleaned up automatically.

## Troubleshooting

**Compiler not found?** Verify both `latexmk` and your chosen engine are on PATH (`latexmk -v`, `pdflatex --version`).

**Missing `.sty` or language files?** Your TeX distribution is missing a dependency used by your document. Install that package through TeX Live's `tlmgr` or your distribution's package manager, then save the file again.

**Build failed?** Click the status indicator or **View log** in the error banner. Lampo retains your last successful preview until a clean build finishes.

**No document found?** Add a standalone `.tex` file with a `\documentclass` declaration, or run `lampo path/to/file.tex` to explicitly select a source file. A folder containing multiple snippets with no entrypoint cannot compile as a document.
