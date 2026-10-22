---
title: Quickstart
description: Clone Lampo, run the CLI scaffold, and start the documentation site.
order: 1
category: Start
summary: Work on the Lampo scaffold locally while the preview engine is in development.
---

Lampo is in early development. Its LaTeX preview server and PDF export are not available yet; the `hello` command below is a temporary starter example.

## Clone the repository

```bash
git clone https://github.com/amxv/lampo.git
cd lampo
```

Project identity is already initialized. No bootstrap step is needed.

## Install dependencies

Install Go, Node.js, and Bun, then install JavaScript dependencies:

```bash
make docs-install
```

The repository uses Go for the CLI, Node for the npm wrapper, and a self-contained
Astro/ZueDocs workspace under `docs/` for the docs site.

## Run the starter CLI

Use the make targets to validate and build the starter command:

```bash
make check
make build
./dist/lampo --help
./dist/lampo hello
```

The sample command is intentionally small so it is easy to replace.

## Start the docs site

Run the embedded documentation site locally:

```bash
make docs-dev
```

Astro usually serves the site at `http://localhost:4321`.

## First customization pass

Replace the starter command behavior in:

```bash
internal/app/app.go
internal/app/app_test.go
```

Keep the docs open while you edit so the quickstart, command reference, and release notes stay aligned with the actual CLI.
