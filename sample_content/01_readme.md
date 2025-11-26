---
title: README - Getting Started with DSBG
description: How to install and use Dead Simple Blog Generator.
created: 2025 01 03
cover_image: 01_dsbg_logo.webp
---

# DSBG: Dead Simple Blog Generator

![Demo](https://github.com/tesserato/DSBG/blob/main/art/DemoSmall.gif?raw=true)

DSBG is a **minimalist, single-binary static blog generator** written in Go.
Point it at a folder of Markdown/HTML files, run one command, and get a clean, modern, fast blog.

---

# Why DSBG?

* **Zero Setup**: No runtime dependencies. Just one binary.
* **Fast by Design**: Concurrent processing, instant rebuilds in watch mode.
* **Clean, Modern Output**: Responsive layout, dark/light themes, code highlighting, MathJax.
* **Smart Content Handling**: Frontmatter, tag extraction, date parsing, automatic resource copying:DSBG only copies resources (images, PDFs, videos) that are explicitly referenced in your Markdown or HTML. Unused files in your content folder are not moved to the output.
* **Full-Text Search**: Built-in Lunr-powered search index.
* **Social Sharing Ready**: Add your own share buttons with URL templates.
* **RSS Included**: Automatic, standards-compliant feed.
* **Pages & Posts**: Tag `PAGE` to add top-level navigation pages. **Note:** HTML files tagged as `PAGE` should live in their own dedicated subfolders; DSBG copies the entire parent folder to preserve local scripts and assets.
* **Easy Customization**: Themes, custom CSS/JS, custom favicon, publisher metadata, and more.
* **SEO-Ready Out of the Box**: Open Graph, JSON-LD schema, canonical/share URL overrides, publisher logo.
* **Smarter Index Page**: Tag filters, fuzzy full-text search with snippets, and one-click `Copy Markdown` sharing.
* **Flexible Input & Dates**: Markdown or HTML, auto-extracted tags/metadata, and date parsing from filenames or file mtimes.


---

# Quick Start

## Install

**Binary Releases:**
Download the latest release from GitHub.

**Or via Go:**

```bash
go install github.com/tesserato/DSBG@latest
```
or

```bash
go install github.com/tesserato/DSBG@v0.1.4
```


## Generate a Site

```bash
dsbg -input content/ -output public/ -title "My Blog" -watch
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

My first post with DSBG!
```

---

# Configuration & Flags

DSBG exposes many options for themes, sorting, metadata, sharing, custom assets, and more.

**Run `dsbg -h` for the full list of commands, available themes, and flags.**


---

# Deploying

The generated folder is 100% static: deploy anywhere:

* GitHub Pages
* Netlify
* Vercel
* Cloudflare Pages
* Any static hosting service

---

# Notes

## 1. URLs & File Structure
*   **"Clean URL" Transformation:** DSBG converts every input file into a directory containing an `index.html`.
    *   *Input:* `content/posts/my-cool-story.md`
    *   *Output:* `public/posts/my-cool-story/index.html`
    *   *Result:* Your URL becomes `domain.com/posts/my-cool-story/` (trailing slash).
*   **URL Sanitization:** URLs are aggressive sanitized. Non-alphanumeric characters are removed, and spaces/underscores become dashes (e.g., `C# for C++/CLI` becomes `csharp-for-cpluspluscli`).
*   **Dates in URLs:** By default, date patterns (e.g., `2024-11-03-`) are stripped from filenames and URLs. Use `-keep-date-in-paths` to preserve them.

## 2. Dates & Sorting
*   **Date Hierarchy:** The creation date is determined in this priority order:
    1.  Frontmatter/Meta Tag (`created: YYYY-MM-DD`).
    2.  Filename Pattern (`2023-10-05-my-post.md`).
    3.  File Modification Time (Warning: This may change if you clone the repo to a new machine).
*   **Date Formats:** Dates must be in recognizable numeric formats (e.g., `2024-03-03 10:00`) for the parser to detect them.
*   **Sorting:** The `-sort` flag is strict and only accepts specific values like `date-created`, `reverse-date-created`, `title`, etc.

## 3. Tags & Organization
*   **Folders = Tags:** By default, folder names become tags. A file in `content/linux/kernel/` gets `linux` and `kernel` tags automatically. Disable with `-ignore-tags-from-paths`.
*   **`PAGE` Tag:** Articles tagged with `PAGE` are removed from the main feed and added to the top navigation bar.
*   **HTML `PAGE` Isolation:** If an HTML file is tagged as `PAGE`, its **entire parent folder** is copied recursively to the output. **Isolate HTML pages in their own subfolders** to avoid duplicating unrelated files.

## 4. Resource Handling
*   **Smart Copying:** DSBG only copies resources (images, PDFs, videos) explicitly referenced in your content. Unreferenced files are ignored.
*   **Relative Paths:** Root-relative paths (e.g., `/img/logo.png`) are treated as relative to the **article's directory**, not the site root.
*   **Strict Validation:** By default, the build **fails** if a referenced resource is missing. Use `-ignore-errors` to log warnings instead.

## 5. SEO & Social Features
*   **Base URL Required:** For production builds, you **must** set `-base-url https://yourdomain.com`. Without it, RSS feeds, Sitemaps, and Social Sharing preview cards (Open Graph) will point to `localhost`.
*   **Canonical URLs:** You can override the auto-generated canonical URL per post using the `canonical_url` frontmatter field.
*   **Share URL:** The `share_url` field overrides the link used by share buttons, useful for "link blogs" where the post title should link to an external site.

## 6. Theming & Customization
*   **Highlighting Detection:** Syntax highlighting (Light/Dark) is automatically chosen by scanning your theme CSS for `color-scheme: dark;`.
*   **Custom CSS:** Providing a `-css-path` **replaces** the built-in theme entirely.
*   **Custom JS:** Providing a `-js-path` **appends** your script to the default functionality (search, copy buttons, etc. remain active).
*   **Table Styling:** All Markdown tables are wrapped in `<div class="table-wrapper">` to allow for responsive scrolling and styling.

## 7. Watch Mode
*   **Port:** Default server port is `666`.
*   **Live Reload:** The browser automatically opens on start. Content, assets, and custom CSS/JS are watched for changes.
*   **Cache Busting:** A version query string (`?v=TIMESTAMP`) is appended to assets on every rebuild to ensure you always see the latest changes.

# Contributing

PRs and feature suggestions are welcome!

GitHub: `https://github.com/tesserato/DSBG`

# Star History

[![Star History Chart](https://api.star-history.com/svg?repos=tesserato/DSBG&type=Date)](https://star-history.com/#tesserato/DSBG&Date)

<a href="https://www.producthunt.com/posts/dead-simple-blog-generator?embed=true&utm_source=badge-featured&utm_medium=badge&utm_source=badge-dead&#0045;simple&#0045;blog&#0045;generator" target="_blank"><img src="https://api.producthunt.com/widgets/embed-image/v1/featured.svg?post_id=912370&theme=light&t=1740653394729" alt="Dead&#0032;Simple&#0032;Blog&#0032;Generator - Static&#0032;Site&#0032;Generator&#0032;That&#0032;Fast&#0045;Tracks&#0032;Your&#0032;Digital&#0032;Presence | Product Hunt" style="width: 250px; height: 54px;" width="250" height="54" /></a>