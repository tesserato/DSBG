---
title: README - Getting Started with DSBG
description: How to install and use Dead Simple Blog Generator.
created: 2025 01 03
cover_image: 01_dsbg_logo.webp
---

# DSBG: Dead Simple Blog Generator

![DSBG Screenshot](https://raw.githubusercontent.com/tesserato/DSBG/refs/heads/main/art/logo.webp)

DSBG is a **minimalist, single-binary static blog generator** written in Go.
Point it at a folder of Markdown/HTML files, run one command, and get a clean, modern, fast blog.

---

## Why DSBG?

* **Zero Setup**: No runtime dependencies. Just one binary.
* **Fast by Design**: Concurrent processing, instant rebuilds in watch mode.
* **Clean, Modern Output**: Responsive layout, dark/light themes, code highlighting, MathJax.
* **Smart Content Handling**: Frontmatter, automatic resource copying, tag extraction, date parsing.
* **Full-Text Search**: Built-in Lunr-powered search index.
* **Social Sharing Ready**: Add your own share buttons with URL templates.
* **RSS Included**: Automatic, standards-compliant feed.
* **Pages & Posts**: Tag `PAGE` to add top-level navigation pages.
* **Easy Customization**: Themes, custom CSS/JS, custom favicon, publisher metadata, and more.

---

## Quick Start

### Install

**Binary Releases:**
Download the latest release from GitHub.

**Or via Go:**

```bash
go install github.com/tesserato/DSBG@latest
```

### Generate a Site

```bash
dsbg -input-path content/ -output-path public/ -title "My Blog"
```

### Live Preview

```bash
dsbg -input-path content/ -watch
```

Serves your site at:
`http://localhost:666`

---

## Writing Content

DSBG uses standard Markdown with YAML frontmatter.

To get a **quick start template** with the current date pre-filled, simply run:

```bash
dsbg -h
```

Copy the output from the **TEMPLATE EXAMPLE** section into a new `.md` file.

Example structure:

```yaml
---
title: My New Post
description: A short summary of the post.
created: 2025 11 22
tags: Technology, Go
coverImagePath: image.webp
---

# Hello World
...
```

---

## Configuration & Flags

DSBG exposes many options for themes, sorting, metadata, sharing, custom assets, and more.

**Run `dsbg -h` for the full list of commands, available themes, and flags.**

### Common Flags

*   `-theme`: Choose a built-in style (e.g., `default`, `dark`, `clean`, `paper`, `industrial`, `black`, `colorful`, `terminal`).
*   `-publisher`: Name of the publisher/organization (for JSON-LD metadata).
*   `-author-name`: Default author name for posts.
*   `-sort`: Sort order for the index page (e.g., `date-created`, `reverse-date-updated`).
*   `-share`: Add custom share buttons (see help for format).

---

## Deploying

The generated `public/` folder is 100% static: deploy anywhere:

* GitHub Pages
* Netlify
* Vercel
* Cloudflare Pages
* Any static hosting service

---

## Contributing

PRs and feature suggestions are welcome!

GitHub: `https://github.com/tesserato/dsbg`

# Star History

[![Star History Chart](https://api.star-history.com/svg?repos=tesserato/DSBG&type=Date)](https://star-history.com/#tesserato/DSBG&Date)

<a href="https://www.producthunt.com/posts/dead-simple-blog-generator?embed=true&utm_source=badge-featured&utm_medium=badge&utm_souce=badge-dead&#0045;simple&#0045;blog&#0045;generator" target="_blank"><img src="https://api.producthunt.com/widgets/embed-image/v1/featured.svg?post_id=912370&theme=light&t=1740653394729" alt="Dead&#0032;Simple&#0032;Blog&#0032;Generator - Static&#0032;Site&#0032;Generator&#0032;That&#0032;Fast&#0045;Tracks&#0032;Your&#0032;Digital&#0032;Presence | Product Hunt" style="width: 250px; height: 54px;" width="250" height="54" /></a>