# Git History Cleanup - Exposed API Key Remediation

This document provides step-by-step instructions for cleaning up the git history after an API key was exposed.

## ⚠️ IMPORTANT NOTICES

1. **Revoke the key FIRST**: Before cleaning history, ensure the exposed key has been revoked/rotated
2. **Coordinate with team**: History rewriting affects all contributors
3. **Backup first**: Create a backup before proceeding
4. **Force push required**: This requires force-push permissions
5. **One-time operation**: Once force-pushed, all contributors must re-clone or reset their local repos

## Status

- ✅ Current code is clean (no exposed secrets in working tree)
- ✅ API key has been revoked by repository owner
- ⚠️ Git history may still contain the exposed key (requires cleanup)

## Option 1: BFG Repo-Cleaner (Recommended)

BFG is faster and simpler than git-filter-branch for removing sensitive data.

### Step 1: Install BFG

```bash
# macOS
brew install bfg

# Or download from https://rtyley.github.io/bfg-repo-cleaner/
```

### Step 2: Create a backup

```bash
# Clone a backup of the repository
git clone --mirror https://github.com/lvcoi/ytdl-go.git ytdl-go-backup.git

# Archive the backup
tar -czf ytdl-go-backup-$(date +%Y%m%d).tar.gz ytdl-go-backup.git
```

### Step 3: Clone a fresh copy

```bash
git clone https://github.com/lvcoi/ytdl-go.git ytdl-go-clean
cd ytdl-go-clean
```

### Step 4: Create secrets file

Create a text file with the patterns to remove (one per line):

```bash
cat > /tmp/secrets-to-remove.txt << 'EOF'
ghp_1234567890abcdefghijklmnopqrstuvwxyz
AIzaSyD1234567890abcdefghijklmnopqrstuvwx
AKIAIOSFODNN7EXAMPLE
EOF
```

**Note**: Replace the above with your actual exposed secrets.

### Step 5: Run BFG

```bash
# Remove the secrets from history
bfg --replace-text /tmp/secrets-to-remove.txt

# Alternative: if the secrets are in specific files only
# bfg --replace-text /tmp/secrets-to-remove.txt --no-blob-protection
```

### Step 6: Clean up

```bash
# Expire reflog
git reflog expire --expire=now --all

# Garbage collect
git gc --prune=now --aggressive
```

### Step 7: Verify

```bash
# Scan the cleaned history
gitleaks detect --verbose

# Manually search for patterns
git log --all -p | grep -i "ghp_\|AIza\|AKIA" || echo "No secrets found"
```

### Step 8: Force push

```bash
# Push all branches
git push --force --all

# Push all tags
git push --force --tags
```

## Option 2: git-filter-repo

More powerful but requires Python.

### Step 1: Install

```bash
pip install git-filter-repo
```

### Step 2: Create expressions file

```bash
cat > /tmp/expressions.txt << 'EOF'
regex:ghp_[A-Za-z0-9]{36}==>***REMOVED_GITHUB_TOKEN***
regex:AIza[A-Za-z0-9_-]{35}==>***REMOVED_API_KEY***
literal:specific-secret-value==>***REMOVED_SECRET***
EOF
```

### Step 3: Run filter-repo

```bash
# Clone fresh copy
git clone https://github.com/lvcoi/ytdl-go.git ytdl-go-clean
cd ytdl-go-clean

# Run the filter
git filter-repo --replace-text /tmp/expressions.txt

# Force push
git push --force --all
```

## Option 3: Interactive Rebase (For Recent Commits Only)

If the secret was only in the last few commits (e.g., last 5-10 commits):

### Step 1: Start interactive rebase

```bash
# Rebase last N commits
git rebase -i HEAD~5  # Adjust number as needed
```

### Step 2: Edit commits

In the editor that opens, change `pick` to `edit` for commits that contain secrets:

```
edit abc1234 Add feature X
pick def5678 Update documentation
edit ghi9012 Fix bug Y
```

### Step 3: Remove secrets

For each commit marked as `edit`:

```bash
# Edit the files to remove secrets
vim path/to/file-with-secret.go

# Stage the changes
git add path/to/file-with-secret.go

# Amend the commit
git commit --amend --no-edit

# Continue the rebase
git rebase --continue
```

### Step 4: Force push

```bash
git push --force
```

## Option 4: Squash and Start Fresh (Nuclear Option)

If the repository is new or the history isn't valuable:

### Step 1: Create orphan branch

```bash
# Checkout current code
git checkout --orphan clean-start

# Add all files
git add -A

# Commit
git commit -m "Initial commit - clean history"
```

### Step 2: Delete old branch and rename

```bash
# Delete old main branch
git branch -D main

# Rename new branch to main
git branch -m main

# Force push
git push -f origin main
```

## Post-Cleanup Steps

### For Repository Owner

1. ✅ Verify history is clean:
   ```bash
   gitleaks detect --verbose
   ```

2. ✅ Enable branch protection (if not already):
   - Go to Settings → Branches → Branch protection rules
   - Require pull request reviews
   - Require status checks to pass (secret scanning)

3. ✅ Notify all contributors:
   ```markdown
   IMPORTANT: Git history has been rewritten to remove exposed secrets.
   
   All contributors must:
   1. Save any local uncommitted work
   2. Delete their local repository
   3. Clone the repository fresh
   4. Run ./setup-hooks.sh to enable secret scanning
   ```

### For Contributors

After the repository owner completes the cleanup:

1. **Backup any local uncommitted work**:
   ```bash
   git stash
   # Or commit to a temporary branch
   ```

2. **Delete your local repository**:
   ```bash
   cd ..
   rm -rf ytdl-go
   ```

3. **Clone fresh**:
   ```bash
   git clone https://github.com/lvcoi/ytdl-go.git
   cd ytdl-go
   ```

4. **Enable pre-commit hooks**:
   ```bash
   ./setup-hooks.sh
   ```

5. **Restore any uncommitted work** (if stashed)

## Verification Checklist

After cleanup, verify:

- [ ] No secrets found by gitleaks: `gitleaks detect --verbose`
- [ ] Manual search returns nothing: `git log --all -p | grep -i "secret_pattern"`
- [ ] Pre-commit hooks are working: `./setup-hooks.sh`
- [ ] GitHub Actions secret scanning workflow is enabled
- [ ] All contributors have been notified
- [ ] Branch protection rules are configured

## Troubleshooting

### "non-fast-forward" error during force push

This is expected when rewriting history. Use `--force`:

```bash
git push --force origin main
```

### Contributors can't pull after force push

They must reset their local repository:

```bash
git fetch origin
git reset --hard origin/main
```

Or delete and re-clone (safer).

### Secret still detected after cleanup

1. Verify you're searching all branches and tags:
   ```bash
   git log --all -p | grep "secret_pattern"
   ```

2. Check if it's in a tag:
   ```bash
   git tag | xargs -I {} git show {} | grep "secret_pattern"
   ```

3. Delete tags if needed:
   ```bash
   git tag -d old-tag
   git push origin :refs/tags/old-tag
   ```

## Resources

- [GitHub: Removing sensitive data](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/removing-sensitive-data-from-a-repository)
- [BFG Repo-Cleaner](https://rtyley.github.io/bfg-repo-cleaner/)
- [git-filter-repo](https://github.com/newren/git-filter-repo)
- [Gitleaks](https://github.com/gitleaks/gitleaks)

## Need Help?

If you encounter issues during cleanup:

1. Restore from backup: `git clone ytdl-go-backup.git ytdl-go-restored`
2. Contact the repository maintainer
3. Review the [SECRET_MANAGEMENT.md](SECRET_MANAGEMENT.md) guide
