---
title: Get started
description: From install to your first live PDF in a couple of minutes.
order: 1
category: Get started
summary: Install Lampo and see your first live LaTeX preview.
---

Lampo lets you edit LaTeX in your favorite editor while the PDF updates automatically in your browser. There's nothing to upload, no account to create, and no new editor to learn.

## 1. Install Lampo

Open your terminal and run:

```bash
npm install -g @amxv/lampo
```

You'll also need a LaTeX installation on your computer, such as MacTeX, TeX Live, TinyTeX, or MiKTeX. If you can already compile your `.tex` files locally, you're probably set.

> Not using npm? Download a binary for your computer from the [GitHub Releases page](https://github.com/amxv/lampo/releases), or see the [CLI reference](/docs/command-reference) for other install options.

## 2. Open your project

Go to the folder containing your LaTeX files, then start Lampo:

```bash
cd ~/Documents/my-paper
lampo
```

Lampo opens a browser tab with your PDF. If it doesn't open automatically, use the localhost address shown in your terminal.

That's it. **Leave Lampo running and keep writing.**

## 3. See your changes

Edit a sentence in your `.tex` file and save it. Your browser preview updates automatically. You never need to press Refresh or run a LaTeX build command yourself.

Put the browser beside your editor, or beside a chat with a coding agent that's editing the same files. Lampo works with either workflow.

## 4. Export the PDF

When everything looks right, select **Export PDF** in the top-right corner of the preview. Enter a filename and save your finished document.

Lampo only downloads when you ask. The source files in your project stay where they are.

## What's next?

- [Your live preview](/docs/your-workflow) shows you zoom, split-screen viewing, and build errors.
- [Multiple documents](/docs/multiple-documents) explains switching between a resume, cover letter, and other `.tex` files in one tab.

If something isn't working, see [Quick fixes](/docs/troubleshooting).
