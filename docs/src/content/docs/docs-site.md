---
title: Docs website
description: Development, structure, and deployment of lampo.ashray.xyz.
order: 22
category: Reference
summary: Maintain the Astro docs site and ship updates to Vercel.
---

Lampo's website is a separate static Astro project in `docs/`. It is not part of the CLI runtime or npm binary. The docs pages use the shared ZueDocs shell; the homepage has its own purpose-built design.

## Local website development

From the repository root:

```bash
make docs-install
make docs-dev
```

Astro normally serves at `http://localhost:4321`.

## Main files

```text
docs/src/pages/index.astro        Custom homepage and preview illustration
docs/src/styles/landing.css       Homepage typography, colors, layout, motion
docs/src/pages/docs/index.astro   Friendly documentation hub
docs/src/pages/docs/[...slug].astro  Individual guide layout
docs/src/styles/docs-theme.css    Visual styling for the ZueDocs shell
docs/src/content/docs/*.md       User guides and technical reference
docs/src/data/docs.ts            Site title, navigation, categories
docs/astro.config.mjs            Astro and ZueDocs integration
```

Each guide declares its title, description, ordering, and category in the Markdown frontmatter. The categories are **Get started**, **Guides**, and **Reference**. Keep basic task-focused explanations in the first two categories, and move maintainers' and implementation details into Reference.

## Quality checks

```bash
make docs-check
make docs-build
```

Run them one after another rather than in parallel. The build creates static HTML in `docs/dist/`, including raw Markdown docs routes for compatible readers and agent clients.

## Deploy to Vercel

The Vercel project uses `docs` as its root directory. Its `vercel.json` already defines installation, build, and output settings.

Deploy through the Vercel Git integration when configured, or from the linked docs directory:

```bash
cd docs
vercel --prod
```

The canonical domain is [lampo.ashray.xyz](https://lampo.ashray.xyz). After deployment, verify the homepage, guide navigation, mobile layout, and links to GitHub/npm.

The repository's [release workflow](/docs/release-and-npm) is separate from website deployment. A docs-only change does not require a new CLI version.
