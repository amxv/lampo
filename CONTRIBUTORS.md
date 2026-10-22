# Contributing to Lampo

Thanks for helping make LaTeX feel a little more like modern web development. Lampo is a small, local-first CLI, and the best contributions keep it fast, predictable, and easy to use.

## Start here

You'll need Go 1.26+, Node.js 18+, and Bun for the separate documentation website. To preview real LaTeX, install a TeX distribution with `latexmk`.

```bash
git clone https://github.com/amxv/lampo.git
cd lampo
make check
make build
./dist/lampo --help
```

You can install your local binary with `make install-local`.

## Where things live

| Area | Path |
| --- | --- |
| CLI flags and process lifetime | `internal/app/` |
| Finding documents and compiling LaTeX | `internal/latex/` |
| File changes and rebuilds | `internal/watch/`, `internal/preview/` |
| Browser PDF preview and controls | `internal/preview/web/` |
| Website and user guides | `docs/` |

The CLI embeds its browser interface into the Go binary. Keep runtime dependencies minimal and avoid adding a frontend framework for simple UI behaviors.

## Before opening a PR

Run the checks that apply to your change:

```bash
make check
go test -race ./...
make docs-check
make docs-build
```

Run docs checking and building **serially**, not concurrently. For UI changes, test the live preview at both wide and narrow browser widths. For watch/build changes, test an edit, a compile error, a recovery, and switching between multiple documents.

Please keep private LaTeX content out of public test fixtures. Generated PDFs, LaTeX auxiliary files, and temporary build outputs should never end up in Git.

## Releasing

Release tags are maintained by the repository owner. The full workflow, including GitHub binary assets and npm publication, is in the [release reference](https://lampo.ashray.xyz/docs/release-and-npm).

Lampo is Apache-2.0 licensed. Contributions to its existing codebase follow that license.
