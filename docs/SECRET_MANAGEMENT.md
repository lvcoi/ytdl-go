# Secret Management and Prevention Guide

This guide explains how to prevent accidentally committing secrets to the repository and what to do if a secret is exposed.

## Table of Contents

- [Prevention](#prevention)
- [What to Do If a Secret is Exposed](#what-to-do-if-a-secret-is-exposed)
- [Git History Cleanup](#git-history-cleanup)
- [Tools and Configuration](#tools-and-configuration)

## Prevention

### 1. Use Environment Variables

**NEVER** hardcode secrets in your code. Always use environment variables:

```go
// ❌ BAD - Never do this
token := "ghp_1234567890abcdefghijklmnopqrstuvwxyz"

// ✅ GOOD - Use environment variables
token := os.Getenv("GITHUB_TOKEN")
if token == "" {
    log.Fatal("GITHUB_TOKEN environment variable is required")
}
```

### 2. Use .gitignore

Ensure sensitive files are excluded from version control. The `.gitignore` file includes:

```
# Environment files
.env
.env.local
.env.*.local

# Secrets and credentials
secrets/
.secrets
*.secrets
credentials/
.credentials
*.credentials
*.key
*.pem
```

### 3. Enable Pre-commit Hooks

This repository includes a pre-commit hook that scans for secrets using Gitleaks.

**Setup (one-time per developer):**

```bash
# Configure Git to use the custom hooks directory
git config core.hooksPath .githooks

# Install gitleaks (macOS)
brew install gitleaks

# Or download binary from https://github.com/gitleaks/gitleaks/releases
```

The hook will automatically scan your commits for secrets before allowing them through.

### 4. Use Placeholder Examples in Documentation

When writing documentation, always use placeholders:

```bash
# ✅ GOOD - Use placeholders
export GITHUB_TOKEN="ghp_..."
export API_KEY="your_api_key_here"

# ❌ BAD - Never include real tokens
export GITHUB_TOKEN="ghp_1234567890abcdefghijklmnopqrstuvwxyz"
```

## What to Do If a Secret is Exposed

If you accidentally commit a secret, follow these steps immediately:

### Step 1: Revoke the Secret

**This is the most important step!** Immediately revoke/rotate the exposed secret:

- **GitHub Personal Access Token**: Go to Settings → Developer settings → Personal access tokens → Revoke
- **API Keys**: Revoke the key in the respective service's dashboard
- **AWS Keys**: Deactivate in IAM console
- **Other credentials**: Follow the service's revocation process

### Step 2: Verify Current Code is Clean

```bash
# Scan the current codebase
gitleaks detect --verbose

# Or use GitHub's secret scanning (if enabled)
```

### Step 3: Clean Up Git History

⚠️ **WARNING**: Rewriting git history can be destructive. Coordinate with your team first!

#### Option A: Using BFG Repo-Cleaner (Recommended)

```bash
# Install BFG
brew install bfg  # macOS
# or download from https://rtyley.github.io/bfg-repo-cleaner/

# Create a backup first!
git clone --mirror https://github.com/lvcoi/ytdl-go.git ytdl-go-backup.git

# Clone the repo fresh
git clone https://github.com/lvcoi/ytdl-go.git ytdl-go-clean
cd ytdl-go-clean

# Create a text file with the secret patterns to remove
echo "ghp_1234567890abcdefghijklmnopqrstuvwxyz" > secrets.txt

# Run BFG to remove the secrets
bfg --replace-text secrets.txt

# Clean up
git reflog expire --expire=now --all
git gc --prune=now --aggressive

# Force push (requires force push permission)
git push --force --all
git push --force --tags
```

#### Option B: Using git-filter-repo

```bash
# Install git-filter-repo
pip install git-filter-repo

# Clone the repo fresh
git clone https://github.com/lvcoi/ytdl-go.git ytdl-go-clean
cd ytdl-go-clean

# Replace the secret with a placeholder
git filter-repo --replace-text <(echo "ghp_1234567890abcdefghijklmnopqrstuvwxyz==>REDACTED")

# Force push
git push --force --all
git push --force --tags
```

#### Option C: For Recent Commits Only

If the secret was only in the last few commits:

```bash
# For the last commit
git reset --soft HEAD~1
# Remove the secret from files
git add .
git commit -m "Remove exposed secret"
git push --force

# For multiple commits (interactive rebase)
git rebase -i HEAD~5  # Adjust number as needed
# In the editor, mark commits to edit
# Remove secrets from files
# Continue the rebase
git push --force
```

### Step 4: Notify Affected Parties

- If this is a public repository, consider notifying users
- If enterprise/team repository, notify team members
- Consider creating a security incident report

### Step 5: Prevent Future Occurrences

1. Enable the pre-commit hooks (see Prevention section)
2. Enable GitHub's secret scanning (for public repos, it's automatic)
3. Review and update `.gitignore`
4. Educate team members on secret management best practices

## Tools and Configuration

### Gitleaks

This repository uses Gitleaks for secret detection.

**Configuration files:**
- `.gitleaks.toml` - Main configuration
- `.gitleaksignore` - False positive exclusions

**Manual scanning:**

```bash
# Scan entire repository history
gitleaks detect --verbose

# Scan only uncommitted changes
gitleaks protect --staged

# Scan specific files
gitleaks detect --source /path/to/file
```

### GitHub Actions

The repository includes a GitHub Actions workflow (`.github/workflows/secret-scanning.yml`) that automatically scans every push and pull request for secrets.

**To view scan results:**
1. Go to the Actions tab in GitHub
2. Click on the "Secret Scanning" workflow
3. Check the run details

### False Positives

If Gitleaks detects a false positive, add it to `.gitleaksignore`:

```
# .gitleaksignore
path/to/file.go:line-with-false-positive
```

Or update `.gitleaks.toml` to add regex patterns to the allowlist.

## Best Practices Summary

1. ✅ **DO** use environment variables for all secrets
2. ✅ **DO** use `.env` files locally (ensure they're in `.gitignore`)
3. ✅ **DO** use placeholders in documentation and examples
4. ✅ **DO** enable pre-commit hooks
5. ✅ **DO** rotate secrets regularly
6. ✅ **DO** use secret management services (AWS Secrets Manager, Azure Key Vault, etc.)
7. ❌ **DON'T** commit `.env` files
8. ❌ **DON'T** hardcode secrets in code
9. ❌ **DON'T** include secrets in commit messages
10. ❌ **DON'T** use `--no-verify` unless absolutely necessary

## Resources

- [Gitleaks Documentation](https://github.com/gitleaks/gitleaks)
- [GitHub Secret Scanning](https://docs.github.com/en/code-security/secret-scanning)
- [BFG Repo-Cleaner](https://rtyley.github.io/bfg-repo-cleaner/)
- [git-filter-repo](https://github.com/newren/git-filter-repo)
- [OWASP Secrets Management Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Secrets_Management_Cheat_Sheet.html)

## Getting Help

If you discover a security vulnerability or have questions about secret management:

1. Review the [SECURITY.md](../SECURITY.md) file
2. Contact the maintainers privately
3. Do NOT open a public issue for active security concerns
