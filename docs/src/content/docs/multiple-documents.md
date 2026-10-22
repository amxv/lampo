---
title: Multiple documents
description: One folder, many LaTeX documents, and a single browser tab.
order: 3
category: Guides
summary: Switch PDFs instantly from the filename menu, including documents in subfolders.
---

You don't need a separate Lampo process for every document. Start it once in a folder and switch between the PDFs from the preview's filename menu.

## Try it with a folder

For example, suppose your project looks like this:

```text
applications/
├── resume.tex
├── cover-letter.tex
└── chapters/
    └── experience.tex
```

Running `lampo` in `applications/` discovers **resume.tex** and **cover-letter.tex** as standalone documents. `experience.tex` is an included section, so it doesn't clutter the menu.

## Switch without leaving the tab

Click the filename beside the Lampo logo, or press **/**. Choose a document, or type part of its name to find it.

The first time you choose a document, Lampo compiles it. When you switch back, the previous PDF appears immediately from its cache unless that file has changed.

Each document gets its own build files, so two documents named `main.tex` in different folders won't interfere with each other. The dropdown shows the folder path when names repeat.

## Add or remove documents as you go

Create a new standalone `.tex` file and Lampo adds it to the list automatically. Delete one and the picker updates, too. If you remove the document you're viewing, Lampo moves to another available document.

Lampo looks for `\documentclass` to identify a standalone document. Included chapters and other snippets are still watched for edits, but aren't shown as separate PDFs.

## Choose the first document yourself

If you have a preferred starting file, pass it explicitly:

```bash
lampo cover-letter.tex
```

You can still switch to other documents in that folder from the dropdown.

**Next:** [Make the most of the live preview](/docs/your-workflow).
