# Security Policy

## Supported Versions

Only the latest version is actively supported with security updates.

```bash
go install github.com/lvcoi/ytdl-go@latest
```

| Version | Supported |
| --- | --- |
| v1.x | Yes |
| < v1.0 | No |

## Reporting a Vulnerability

1. **Do not** open a public GitHub issue for security vulnerabilities
2. Use GitHub's private vulnerability reporting feature or email the maintainer directly
3. Include a clear description and steps to reproduce

**Response times:**

- Initial response: within 48 hours
- Status update: within 7 days
- Fix timeline: typically 14-30 days depending on severity

## Security Scope

Relevant security concerns:

- Path traversal in output templates or filenames
- Command injection via crafted URLs or metadata
- Dependency vulnerabilities in third-party Go modules
- Unsafe file operations that could overwrite unintended files

### Out of Scope

- Ability to download copyrighted content (user responsibility)
- YouTube API rate limiting or blocking
- Issues in upstream dependencies (report those upstream)

## Security Design

- **No credential storage** — does not store passwords, tokens, or cookies
- **No browser automation** — does not interact with browsers or extract session data
- **No DRM circumvention** — refuses encrypted/protected content
- **Filesystem sanitization** — output filenames are sanitized to prevent path traversal
- **No code execution** — downloaded content is never executed
