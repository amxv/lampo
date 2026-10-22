# Lampo ⚡

**LaTeX, now in real time.**

Lampo is a tiny local CLI that shows your LaTeX PDF in a browser and updates it automatically whenever you save. Keep your favorite editor or coding agent, see every change, and export the finished PDF when you're ready.

**[Website](https://lampo.ashray.xyz)** · **[Get started](https://lampo.ashray.xyz/docs)** · **[Releases](https://github.com/amxv/lampo/releases)**

## Get started

You need a LaTeX installation with `latexmk` and `pdflatex` (MacTeX, TeX Live, TinyTeX, or MiKTeX).

```bash
npm install -g @amxv/lampo
cd path/to/your/latex-project
lampo
```

Lampo opens your browser to a local live preview. Edit a `.tex` file and save to see it update, without refreshing. Click **Export PDF** to download it with your chosen filename.

## Why Lampo?

- **Live previews:** Changes to source files, chapters, and figures appear automatically.
- **Any editor:** Works alongside your editor, terminal, or coding agent.
- **Sharp at any size:** Instant resizing and crisp PDF text, even in split-screen.
- **Multiple documents:** Switch between standalone `.tex` files in one tab, with cached previews.
- **Built to be reliable:** See compilation errors while keeping the last good PDF visible.
- **Local-first:** Your files never leave your machine. No accounts, uploads, or subscriptions.

## A few useful commands

```bash
lampo                       # Preview the current folder
lampo myfile.tex             # Start on a particular document
lampo --port 3030            # Keep the same local URL
lampo --no-open              # Print the URL without opening a tab
lampo --engine xelatex       # Use XeLaTeX (or lualatex)
```

Lampo is a native Go binary with a bundled browser UI, built for macOS, Linux, and Windows. No separate web server or frontend runtime is needed. See the [guides](https://lampo.ashray.xyz/docs) for everyday use or the [CLI reference](https://lampo.ashray.xyz/docs/command-reference) for more options.

## Contributing

Requires Go 1.26+ to build from source:

```bash
make check
make build
```

PRs and issues are welcome. See [CONTRIBUTORS.md](CONTRIBUTORS.md) for project conventions.

**Apache-2.0** · Built with Go, fsnotify, and Mozilla PDF.js.
