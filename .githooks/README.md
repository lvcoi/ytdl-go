# Git Hooks

This directory contains Git hooks for the ytdl-go project.

## Setup

To enable these hooks, run:

```bash
./setup-hooks.sh
```

Or manually configure Git:

```bash
git config core.hooksPath .githooks
```

## Available Hooks

### pre-commit

Scans staged files for secrets using Gitleaks before allowing commits.

**What it does:**
- Scans all staged files for potential secrets (API keys, tokens, passwords, etc.)
- Blocks the commit if secrets are detected
- Provides instructions for fixing the issue

**Requirements:**
- [Gitleaks](https://github.com/gitleaks/gitleaks) must be installed

**Bypass (NOT RECOMMENDED):**
```bash
git commit --no-verify
```

## Troubleshooting

### Hook not running

Make sure you've configured Git to use this hooks directory:
```bash
git config core.hooksPath .githooks
```

### "gitleaks: command not found"

Install gitleaks:
```bash
# macOS
brew install gitleaks

# Or download from https://github.com/gitleaks/gitleaks/releases
```

### False positive

If Gitleaks detects a false positive:

1. Add it to `.gitleaksignore`:
   ```
   path/to/file.go:specific-line
   ```

2. Or add a regex pattern to `.gitleaks.toml` allowlist

## More Information

See [docs/SECRET_MANAGEMENT.md](../docs/SECRET_MANAGEMENT.md) for comprehensive secret management guidance.
