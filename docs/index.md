# Welcome to ytdl-go Documentation

<div align="center">

**A powerful, blazing fast YouTube downloader written in Go.**

_Feature-rich, interactive, and dependency-free._

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=for-the-badge&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg?style=for-the-badge)](https://github.com/lvcoi/ytdl-go/LICENSE)
[![Release](https://img.shields.io/github/release/lvcoi/ytdl-go.svg?style=for-the-badge)](https://github.com/lvcoi/ytdl-go/releases)

[Quick Start](user-guide/getting-started/quick-start.md){ .md-button .md-button--primary }
[Installation Guide](user-guide/getting-started/installation.md){ .md-button }
[View on GitHub](https://github.com/lvcoi/ytdl-go){ .md-button }

</div>

---

## 🎯 What is ytdl-go?

ytdl-go is a command-line tool and web application for downloading videos and audio from YouTube. Built with Go for speed and reliability, it offers:

- **High Performance** - Parallel downloads with automatic retries
- **Rich Interactivity** - Beautiful TUI format selector
- **Flexible Output** - Custom templates, metadata embedding, and sidecar files
- **Broad Support** - Videos, audio, playlists, and YouTube Music
- **Zero Dependencies** - Single binary, works out of the box

---

## 📚 Documentation Structure

This documentation is organized to help both users and developers:

### 👤 User Guide

Perfect for end-users who want to download media:

- **[Getting Started](user-guide/getting-started/installation.md)** - Installation and setup
- **[Usage](user-guide/usage/basic-downloads.md)** - How to use ytdl-go effectively
- **[Troubleshooting](user-guide/troubleshooting/common-issues.md)** - Solutions to common problems

### 💻 Developer Guide

For contributors and those interested in the internals:

- **[Architecture](developer-guide/architecture/overview.md)** - System design and structure
- **[API Reference](developer-guide/api-reference/endpoints.md)** - Backend API documentation
- **[Contributing](developer-guide/contributing/getting-started.md)** - How to contribute to ytdl-go

### 📖 Reference

Quick reference materials:

- **[CLI Options](reference/cli-options.md)** - Complete flag reference
- **[Output Placeholders](reference/output-placeholders.md)** - Template variables
- **[Exit Codes](reference/exit-codes.md)** - Error codes and meanings

---

## 🚀 Quick Example

```bash
# Download a video in best quality
ytdl-go https://www.youtube.com/watch?v=dQw4w9WgXcQ

# Download audio only
ytdl-go -audio https://www.youtube.com/watch?v=dQw4w9WgXcQ

# Download entire playlist
ytdl-go https://www.youtube.com/playlist?list=PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf

# Interactive format selection
ytdl-go -list-formats https://www.youtube.com/watch?v=dQw4w9WgXcQ
```

---

## ✨ Key Features

| Feature | Description |
|---------|-------------|
| 🚀 **High Performance** | Parallel downloads, automatic retries, and resume capability |
| 📺 **Broad Support** | Videos, audio, playlists, and YouTube Music URLs |
| 🎨 **Format Control** | Select by quality, container, or exact format ID |
| 🎮 **Interactive TUI** | Visual format selector with real-time preview |
| 🏷️ **Rich Metadata** | ID3 tags, JSON metadata, and sidecar files |
| ⚙️ **Automation Ready** | JSON output mode for scripting and integration |
| 🌐 **Web Interface** | Optional browser-based UI for non-CLI users |
| 🔒 **Cookie Support** | Access age-restricted and private content |

---

## 📦 Installation

Choose your preferred installation method:

=== "Pre-built Binary"

    Download from [GitHub Releases](https://github.com/lvcoi/ytdl-go/releases):

    ```bash
    # Linux/macOS
    curl -L https://github.com/lvcoi/ytdl-go/releases/latest/download/ytdl-go-linux-amd64 -o ytdl-go
    chmod +x ytdl-go
    sudo mv ytdl-go /usr/local/bin/
    ```

=== "From Source"

    Requires Go 1.24+:

    ```bash
    git clone https://github.com/lvcoi/ytdl-go.git
    cd ytdl-go
    ./build.sh
    ```

=== "Build Script"

    Use the one-command build script:

    ```bash
    curl -sSL https://raw.githubusercontent.com/lvcoi/ytdl-go/main/build.sh | bash
    ```

See the [Installation Guide](user-guide/getting-started/installation.md) for detailed instructions.

---

## 🤝 Contributing

We welcome contributions! See our [Contributing Guide](developer-guide/contributing/getting-started.md) to get started.

---

## 📜 License

ytdl-go is released under the [MIT License](reference/legal-license.md).

---

## 🔗 Links

- [GitHub Repository](https://github.com/lvcoi/ytdl-go)
- [Issue Tracker](https://github.com/lvcoi/ytdl-go/issues)
- [Releases](https://github.com/lvcoi/ytdl-go/releases)
- [Go Package Documentation](https://pkg.go.dev/github.com/lvcoi/ytdl-go)
