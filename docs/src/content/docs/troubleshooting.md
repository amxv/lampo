---
title: Quick fixes
description: Solutions to the most common LaTeX preview problems.
order: 4
category: Guides
summary: Missing compiler, build errors, slow updates, and documents not appearing.
---

Most Lampo issues are actually LaTeX compilation issues. The good news is you can usually fix them without restarting Lampo.

## It says the compiler is missing

Lampo uses a LaTeX installation already on your computer. Install one if needed, then confirm these commands work in a terminal:

```bash
latexmk -v
pdflatex --version
```

MacTeX and TeX Live are common options. On Windows, MiKTeX also works. Restart your terminal after installing LaTeX so the commands are available on your PATH.

## My document doesn't compile

Click **Build failed** or **View log** in the preview. The error usually includes the file and line that needs attention. Fix the LaTeX source and save again.

The previous successful PDF stays on-screen while you're working on the fix.

If the error mentions a missing `.sty` package or language file, install that dependency using your LaTeX distribution.

## My document doesn't appear in the menu

The dropdown shows standalone documents, usually files with `\documentclass` near the top. A file included with `\input` isn't a separate document.

If you have a nonstandard entrypoint, launch Lampo with the exact `.tex` path:

```bash
lampo path/to/document.tex
```

## Changes aren't showing up

Make sure you've saved the file. The status indicator should briefly show **Updating**. If it says **Build failed**, open the compiler log. If the browser was disconnected, try reloading the tab.

You can also stop Lampo with **Ctrl+C** and run `lampo` again in the project folder.

## The browser doesn't open by itself

Copy the `http://127.0.0.1:...` address from your terminal and open it in your browser. Lampo doesn't require the window to open automatically. If you prefer that behavior:

```bash
lampo --no-open
```

## Still stuck?

Find more options in the [CLI reference](/docs/command-reference), or [open an issue on GitHub](https://github.com/amxv/lampo/issues) with the command you ran and the relevant build error.
