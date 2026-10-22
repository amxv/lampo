# Lampo

Lampo is an upcoming lightweight CLI for live LaTeX previews in the browser, designed for fast editing sessions with coding agents. The goal is to watch `.tex` files and export a finished PDF without leaving the browser.

**Status: scaffold only.** The real-time renderer, preview server, and PDF export are not implemented yet. The command is a Go starter with a sample `hello` subcommand. No version has been published to npm.

## Development

```bash
git clone https://github.com/amxv/lampo.git
cd lampo
make check
make build
./dist/lampo --help
./dist/lampo hello
```

## Documentation site

```bash
make docs-install
make docs-dev
make docs-check
make docs-build
```

The site lives in `docs/`, uses Astro and ZueDocs, and is hosted at [lampo.ashray.xyz](https://lampo.ashray.xyz) on Vercel with Cloudflare DNS.

## Distribution (not released yet)

The scaffold includes a JavaScript npm shim, a native Go binary installer, and a GitHub Actions release workflow for `@amxv/lampo`. The workflow builds binaries and publishes to npm when a `v*` tag is pushed. Do not tag a release until the actual product and changelog are ready.

## Project layout

- `cmd/lampo/`: Go entrypoint
- `internal/app/`: starter command logic and tests
- `bin/lampo.js`: npm shim
- `scripts/postinstall.js`: native binary installer
- `.github/workflows/release.yml`: release and npm automation
- `docs/`: Astro/ZueDocs website
- `project.config.json`: generated project identity

Bootstrapped from [`amxv/go-cli-template`](https://github.com/amxv/go-cli-template). See [developer documentation](https://lampo.ashray.xyz/docs) and `CONTRIBUTORS.md` for maintenance and release instructions.

## License

Apache-2.0
