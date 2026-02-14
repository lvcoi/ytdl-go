---
title: "Pull Request Process"
weight: 50
---

# Pull Request Process

This guide explains the workflow and best practices for submitting pull requests to ytdl-go.

## Table of Contents

- [Before You Start](#before-you-start)
- [Creating a Pull Request](#creating-a-pull-request)
- [PR Guidelines](#pr-guidelines)
- [Review Process](#review-process)
- [After Your PR is Merged](#after-your-pr-is-merged)

## Before You Start

### 1. Check for Existing Issues

Before starting work:

1. **Search existing issues:** https://github.com/lvcoi/ytdl-go/issues
2. **Check open PRs:** Someone may already be working on it
3. **Create an issue if needed:** Discuss approach before implementing

### 2. Fork and Clone

```bash
# Fork on GitHub, then clone
git clone https://github.com/YOUR_USERNAME/ytdl-go.git
cd ytdl-go

# Add upstream remote
git remote add upstream https://github.com/lvcoi/ytdl-go.git
```

### 3. Create a Branch

Use descriptive branch names:

```bash
# Feature branches
git checkout -b feature/add-subtitle-download
git checkout -b feature/improve-error-messages

# Bug fix branches
git checkout -b fix/handle-private-videos
git checkout -b fix/progress-bar-flicker

# Documentation branches
git checkout -b docs/update-contributing-guide
```

**Branch naming:**
- `feature/` - New features
- `fix/` - Bug fixes
- `docs/` - Documentation changes
- `refactor/` - Code refactoring
- `test/` - Adding tests

## Creating a Pull Request

### 1. Make Your Changes

Write code following the [code style guide](code-style):

```bash
# Edit files
vim internal/downloader/youtube.go

# Test locally
go test ./...
go build .
./ytdl-go -info https://www.youtube.com/watch?v=dQw4w9WgXcQ
```

### 2. Commit Your Changes

Write clear, descriptive commit messages:

```bash
# Good commit messages
git commit -m "Add subtitle download support"
git commit -m "Fix progress bar flickering on macOS"
git commit -m "Update API documentation for SSE events"

# Avoid vague messages
git commit -m "Fix bug"
git commit -m "Update code"
git commit -m "WIP"
```

**Commit message guidelines:**
- Use present tense ("Add feature" not "Added feature")
- Be concise but descriptive
- Reference issue numbers if applicable: "Fix #123: Handle private videos"
- First line < 72 characters

**Multiple commits:**
- Logical, focused commits are good
- Each commit should be a coherent change
- Don't worry about too many commits (can be squashed later)

### 3. Push to Your Fork

```bash
git push origin feature/your-feature-name
```

### 4. Create the PR

On GitHub:

1. Go to your fork: `https://github.com/YOUR_USERNAME/ytdl-go`
2. Click "Compare & pull request"
3. Fill out the PR template (if present)

## PR Guidelines

### PR Title

Use a clear, descriptive title:

**Good:**
- "Add subtitle download support"
- "Fix progress bar flickering on macOS"
- "Update API documentation for SSE events"

**Avoid:**
- "Update"
- "Fix bug"
- "Changes"

### PR Description

Include these sections:

#### 1. What

Describe what changed:

```markdown
This PR adds support for downloading video subtitles in multiple formats.
```

#### 2. Why

Explain motivation:

```markdown
Users have requested subtitle downloads for accessibility and language learning.
Closes #123.
```

#### 3. How

Describe the approach:

```markdown
- Added `SubtitleDownloader` in `internal/downloader/subtitles.go`
- Updated `DownloadOptions` to include subtitle preferences
- Added `-subtitles` flag to CLI
```

#### 4. Testing

Describe how you tested:

```markdown
Tested with:
- Videos with multiple subtitle languages
- Videos with auto-generated captions
- Videos without subtitles (graceful fallback)
```

#### 5. Screenshots (if applicable)

For UI changes, include before/after screenshots.

#### 6. Breaking Changes

Clearly mark breaking changes:

```markdown
## Breaking Changes

- The `-format` flag now requires explicit codec (e.g., `-format mp4:h264`)
- Old behavior: `-format mp4`
- New behavior: `-format mp4:h264`
```

### Example PR Description

```markdown
# Add subtitle download support

## What
Adds the ability to download video subtitles in SRT and VTT formats.

## Why
Closes #123. Many users have requested subtitle downloads for accessibility and language learning purposes.

## How
- Created `SubtitleDownloader` in `internal/downloader/subtitles.go`
- Added subtitle extraction logic using YouTube API
- Updated `DownloadOptions` struct with subtitle fields
- Added `-subtitles` flag (default: disabled)
- Added `-subtitle-lang` flag (default: `en`)
- Added `-subtitle-format` flag (default: `srt`)

## Testing
Tested with:
- ✅ Videos with multiple subtitle languages
- ✅ Videos with auto-generated captions
- ✅ Videos without subtitles (logs warning, continues)
- ✅ Different formats (SRT, VTT)
- ✅ Multiple languages simultaneously

## Breaking Changes
None.

## Additional Notes
Subtitles are saved as `{video_name}.{lang}.{format}` (e.g., `video.en.srt`).
```

### Checklist

Include a checklist in your PR:

```markdown
## Checklist
- [x] Code follows style guidelines
- [x] Self-reviewed code
- [x] Commented complex logic
- [x] Updated documentation
- [x] No new warnings from linter
- [x] Tests pass locally
- [x] Tested with real videos
```

## Review Process

### What Reviewers Look For

Reviewers will check:

1. **Correctness:** Does it work as intended?
2. **Code Quality:** Is it clean, readable, maintainable?
3. **Performance:** Does it maintain/improve performance?
4. **Tests:** Are there tests (or plans to add them)?
5. **Documentation:** Are docs updated if needed?
6. **Breaking Changes:** Are they necessary and documented?
7. **Security:** Are there security implications?

### Responding to Feedback

When reviewers provide feedback:

1. **Be respectful:** Feedback is about code, not you
2. **Ask questions:** If unclear, ask for clarification
3. **Discuss alternatives:** Suggest better approaches if you disagree
4. **Make changes:** Address feedback with new commits
5. **Mark resolved:** Mark conversations resolved when addressed

### Making Changes After Review

```bash
# Make requested changes
vim internal/downloader/youtube.go

# Commit changes
git add .
git commit -m "Address review feedback: improve error handling"

# Push to same branch
git push origin feature/your-feature-name
```

The PR will automatically update with new commits.

### Commit Squashing

Don't squash commits yourself unless asked. Maintainers can squash when merging.

If maintainers request squashing:

```bash
# Interactive rebase (squash last 3 commits)
git rebase -i HEAD~3

# Mark commits as 'squash' or 's' in editor

# Force push
git push -f origin feature/your-feature-name
```

## After Your PR is Merged

### Update Your Local Repository

```bash
# Switch to main branch
git checkout main

# Pull latest changes
git pull upstream main

# Delete feature branch (optional)
git branch -d feature/your-feature-name
git push origin -delete feature/your-feature-name
```

### Future Contributions

For your next contribution:

```bash
# Ensure main is up-to-date
git checkout main
git pull upstream main

# Create new branch
git checkout -b feature/next-feature
```

## Common Scenarios

### Scenario: Merge Conflicts

If your PR has merge conflicts:

```bash
# Update your branch with latest main
git checkout main
git pull upstream main
git checkout feature/your-feature
git rebase main

# Resolve conflicts in editor
vim conflicted-file.go

# Mark as resolved
git add conflicted-file.go
git rebase -continue

# Force push
git push -f origin feature/your-feature
```

### Scenario: Need to Update from Main

If main has moved ahead while you were working:

```bash
# Fetch latest
git fetch upstream

# Rebase onto main
git checkout feature/your-feature
git rebase upstream/main

# Force push if needed
git push -f origin feature/your-feature
```

### Scenario: Want to Add More Changes

Simply commit and push:

```bash
# Make changes
vim file.go

# Commit
git add file.go
git commit -m "Additional improvements"

# Push
git push origin feature/your-feature
```

The PR automatically updates.

### Scenario: Accidentally Committed to Main

If you committed to `main` instead of a feature branch:

```bash
# Create feature branch from current state
git checkout -b feature/accidental-commits

# Reset main to upstream
git checkout main
git reset --hard upstream/main

# Push feature branch
git push origin feature/accidental-commits
```

## Best Practices

### 1. Keep PRs Focused

- **One PR per feature/fix**
- Avoid mixing multiple unrelated changes
- If fixing multiple bugs, create separate PRs

### 2. Keep PRs Small

- Easier to review
- Faster to merge
- Less likely to have conflicts

**Guideline:** Aim for < 500 lines changed per PR

### 3. Update Documentation

If your change affects user-facing behavior:

- Update README.md
- Update docs/ files
- Update CLI help text
- Update code comments

### 4. Add Tests (When Possible)

While ytdl-go currently has minimal tests:

- Add tests for new functions
- Add tests for bug fixes
- Help improve test coverage

### 5. Be Patient

- Reviews take time
- Maintainers are volunteers
- Don't ping reviewers repeatedly

### 6. Stay Engaged

- Respond to feedback promptly
- Keep the conversation going
- Don't abandon your PR

## What Makes a Great PR

A great PR:

- ✅ Solves a real problem
- ✅ Has clear title and description
- ✅ Includes tests (or explains why not)
- ✅ Updates documentation
- ✅ Follows code style
- ✅ Has logical commit messages
- ✅ Is reasonably sized
- ✅ Author is responsive to feedback

## Code of Conduct

When contributing:

- **Be respectful:** Treat others with respect
- **Be constructive:** Provide helpful feedback
- **Be patient:** Everyone is learning
- **Be inclusive:** Welcome all contributors
- **Be professional:** Focus on the code, not the person

## Getting Help

If you're stuck:

- **Ask in the PR:** Tag maintainers for guidance
- **Open a Discussion:** For general questions
- **Check Documentation:** Review architecture docs

## Related Documentation

- [Getting Started](getting-started) - Development setup
- [Backend Development](backend) - Go-specific practices
- [Frontend Development](frontend) - JavaScript-specific practices
- [Code Style](code-style) - Coding standards
