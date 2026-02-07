---
title: "Troubleshooting"
weight: 30
---

# Troubleshooting

Having issues with ytdl-go? This section provides solutions to common problems, error code explanations, and frequently asked questions.

## Quick Links

- **[Common Issues](common-issues)** - Solutions to frequently encountered problems
- **[Error Codes](error-codes)** - Understanding ytdl-go exit codes and what they mean
- **[FAQ](faq)** - Frequently asked questions and answers

## Getting Help

If you can't find a solution here:

1. **Check the logs** - Run with `-json` flag for detailed output
2. **Search existing issues** - Visit [GitHub Issues](https://github.com/lvcoi/ytdl-go/issues)
3. **Report a bug** - [Open a new issue](https://github.com/lvcoi/ytdl-go/issues/new) with:
   - ytdl-go version (`ytdl-go -version`)
   - Full command you ran
   - Complete error message
   - Operating system and Go version

## Debug Mode

For detailed debugging information:

```bash
# JSON output mode for machine-readable logs
ytdl-go -json URL

# Increase timeout for slow connections
ytdl-go -timeout 300 URL

# Test without downloading
ytdl-go -info URL
```

## Common Error Categories

ytdl-go categorizes errors to help you understand what went wrong:

| Category | Exit Code | Meaning |
|----------|-----------|---------|
| `invalid_url` | 2 | URL is malformed or not supported |
| `unsupported` | 3 | Feature or format not supported |
| `restricted` | 4 | Content is restricted (age-gated, private, etc.) |
| `network` | 5 | Network connectivity issues |
| `filesystem` | 6 | File system errors (permissions, disk space, etc.) |
| `unknown` | 1 | General error |

See [Error Codes](error-codes) for detailed explanations.
