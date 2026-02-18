# Security Policy

## Supported Versions

Only the latest version is actively supported with security updates. We recommend always running:

```bash
go install github.com/lvcoi/ytdl-go@latest
```

| Version | Supported          |
| ------- | ------------------ |
| v1.x    | :white_check_mark: |
| < v1.0  | :x:                |

## Reporting a Vulnerability

If you discover a security vulnerability in ytdl-go, please report it responsibly:

1. **Do not** open a public GitHub issue for security vulnerabilities
2. Email the maintainer directly or use GitHub's private vulnerability reporting feature
3. Include a clear description of the vulnerability and steps to reproduce

You can expect:

- **Initial response**: Within 48 hours
- **Status update**: Within 7 days
- **Fix timeline**: Depends on severity, typically within 14-30 days

## Security Scope

ytdl-go is a CLI tool that downloads publicly accessible YouTube content. Relevant security concerns include:

- **Path traversal** in output templates or filenames
- **Command injection** via crafted URLs or metadata
- **Dependency vulnerabilities** in third-party Go modules
- **Unsafe file operations** that could overwrite unintended files
- **Exposed secrets** in code, configuration, or git history

### Out of Scope

The following are **not** security vulnerabilities:

- Ability to download copyrighted content (this is a user responsibility)
- YouTube API rate limiting or blocking
- Issues in upstream dependencies (report those upstream)

## Security Design

ytdl-go follows these security principles:

- **No credential storage**: Does not store passwords, tokens, or cookies
- **No browser automation**: Does not interact with browsers or extract session data
- **No DRM circumvention**: Refuses encrypted/protected content
- **Filesystem sanitization**: Output filenames are sanitized to prevent path traversal
- **No code execution**: Downloaded content is never executed
- **Secret scanning**: Automated scanning for exposed secrets in code and commits
- **Pre-commit hooks**: Optional local validation to prevent committing secrets

## Secret Management

See the [Secret Management Guide](docs/SECRET_MANAGEMENT.md) for comprehensive information on:

- How to prevent accidentally committing secrets
- What to do if a secret is exposed
- Git history cleanup procedures
- Tool configuration and usage

**Quick prevention tips:**
- Always use environment variables for secrets (never hardcode)
- Enable the pre-commit hooks: `git config core.hooksPath .githooks`
- Use `.env` files locally (they're in `.gitignore`)
- Check the secret scanning workflow results in GitHub Actions
