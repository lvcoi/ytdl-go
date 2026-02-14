# ytdl-go Documentation

This directory contains the HUGO-based documentation for ytdl-go.

## Building Locally

### Prerequisites

- [HUGO Extended](https://gohugo.io/installation/) v0.121.0 or later
- Git (for cloning the theme)

### Setup

1. Clone the hugo-book theme:
   ```bash
   git clone --branch v13 --depth 1 https://github.com/alex-shpak/hugo-book themes/hugo-book
   ```

2. Run the HUGO development server:
   ```bash
   hugo server -D
   ```

3. Open your browser to `http://localhost:1313/ytdl-go/`

### Building for Production

```bash
hugo --gc --minify
```

The site will be generated in the `public/` directory.

## Structure

```
content/
├── _index.md              # Home page
└── docs/
    ├── _index.md          # Docs section index
    └── user-guide/
        ├── _index.md      # User guide index
        ├── getting-started/
        │   ├── _index.md
        │   ├── installation.md
        │   ├── quick-start.md
        │   └── configuration.md
        └── usage/
            ├── _index.md
            ├── basic-downloads.md
            ├── playlists.md
            ├── audio-only.md
            ├── output-templates.md
            ├── format-selection.md
            └── metadata-sidecars.md
```

## Adding New Pages

1. Create a markdown file in the appropriate directory
2. Add HUGO front matter:
   ```yaml
   ---
   title: "Page Title"
   weight: 10
   ---
   ```
3. Write your content using standard markdown
4. Add an entry in the section's `_index.md` if needed

## Front Matter

Each page should have front matter with at least:

- `title`: The page title shown in navigation
- `weight`: Controls ordering (10, 20, 30, etc. - lower numbers appear first)

## Deployment

The documentation is automatically deployed to GitHub Pages when changes are pushed to the `main` branch via the GitHub Actions workflow in `.github/workflows/docs.yml`.

The site will be available at: `https://lvcoi.github.io/ytdl-go/`

## Theme

We use the [hugo-book](https://github.com/alex-shpak/hugo-book) theme, which is optimized for documentation sites.

## Writing Guidelines

- Use clear, concise language
- Include code examples where appropriate
- Use proper markdown formatting
- Remember: ytdl-go uses SINGLE dash flags (-flag), not double dash (--flag)
- Cross-reference related pages when helpful
- Use blockquotes for notes, tips, and warnings:
  ```markdown
  > **Note:** Important information here
  > **Tip:** Helpful suggestion here
  > **Warning:** Caution message here
  ```
