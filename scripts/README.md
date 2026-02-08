# GitHub Issue Milestone Assignment Script

This script automatically assigns closed issues to appropriate milestones based on when they were closed.

## Features

- ✅ Assigns closed issues to milestones based on smart matching logic
- ✅ Supports dry-run mode to preview changes before applying
- ✅ Can process all issues or a specific issue number
- ✅ Skips issues that already have milestones
- ✅ Filters out pull requests
- ✅ Multiple fallback strategies for milestone matching

## Prerequisites

- Go 1.24 or higher
- GitHub Personal Access Token with `repo` scope

## Installation

1. Navigate to the scripts directory:
   ```bash
   cd scripts
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

## Usage

### Basic Usage (Dry Run)

Preview what changes would be made without actually updating issues:

```bash
export GITHUB_TOKEN="your_github_token_here"
go run assign-milestones.go
```

### Actually Update Issues

To apply the changes:

```bash
export GITHUB_TOKEN="your_github_token_here"
go run assign-milestones.go -dry-run=false
```

### Process a Specific Issue

To update only one specific issue:

```bash
go run assign-milestones.go -issue=42 -dry-run=false
```

### Verbose Output

To see detailed information about what's happening:

```bash
go run assign-milestones.go -verbose
```

## Command Line Options

| Flag | Default | Description |
|------|---------|-------------|
| `-token` | `$GITHUB_TOKEN` | GitHub personal access token |
| `-dry-run` | `true` | Preview changes without applying them |
| `-verbose` | `false` | Show detailed output |
| `-issue` | `0` | Process only this specific issue (0 = all) |

## Milestone Matching Strategy

The script uses a three-tier strategy to assign milestones:

1. **Date-based matching**: If milestones have due dates, issues closed within 30 days before the due date are assigned to that milestone.

2. **Creation-based matching**: Issues are assigned to the most recent milestone that was created before the issue was closed.

3. **Fallback**: If no good match is found, issues are assigned to the earliest open milestone.

## Creating a GitHub Token

1. Go to GitHub Settings → Developer settings → Personal access tokens → Tokens (classic)
2. Click "Generate new token (classic)"
3. Give it a descriptive name (e.g., "Milestone Assignment Script")
4. Select the `repo` scope (full control of private repositories)
5. Click "Generate token"
6. Copy the token and save it securely

## Examples

### Example 1: Preview changes for all closed issues

```bash
export GITHUB_TOKEN="ghp_..."
go run assign-milestones.go -verbose
```

Output:
```
Found 3 milestones:
  - v1.0.0 (ID: 1, State: closed)
  - v1.1.0 (ID: 2, State: open)
  - v2.0.0 (ID: 3, State: open)

Found 15 closed issue(s) to process
Issue #42: 'Fix download speed regression' -> Milestone 'v1.0.0'
  (dry-run: no changes made)
Issue #43: 'Add progress bar' -> Milestone 'v1.1.0'
  (dry-run: no changes made)
...

=== Summary ===
Issues to update: 12
Skipped: 3

This was a DRY RUN. Use -dry-run=false to actually update issues.
```

### Example 2: Actually update a specific issue

```bash
export GITHUB_TOKEN="ghp_..."
go run assign-milestones.go -issue=42 -dry-run=false
```

Output:
```
Found 1 closed issue(s) to process
Issue #42: 'Fix download speed regression' -> Milestone 'v1.0.0'
  ✓ Updated successfully

=== Summary ===
Issues to update: 1
Skipped: 0
```

### Example 3: Build and run as a standalone binary

```bash
cd scripts
go build -o assign-milestones assign-milestones.go
export GITHUB_TOKEN="ghp_..."
./assign-milestones -dry-run=false
```

## Security Notes

- **Never commit your GitHub token** to the repository
- The token grants full access to your repositories, so keep it secure
- Use environment variables or secure credential storage
- Consider using a token with minimal necessary permissions
- Revoke tokens when they're no longer needed

## Troubleshooting

### "GitHub token is required"

Make sure you've set the `GITHUB_TOKEN` environment variable:
```bash
export GITHUB_TOKEN="your_token_here"
```

Or pass it directly:
```bash
go run assign-milestones.go -token="your_token_here"
```

### "Failed to fetch milestones: 401"

Your token may be invalid or expired. Generate a new token with the `repo` scope.

### "Failed to update issue: 403"

Your token doesn't have sufficient permissions. Ensure it has the `repo` scope.

## Contributing

If you find bugs or have suggestions for improvements:

1. Check the milestone matching logic in `determineMilestone()`
2. Adjust the date range in `buildMilestoneRules()` if needed
3. Test changes with `-dry-run` first
4. Submit a pull request with your improvements

## License

Same as the main ytdl-go project (MIT License).
