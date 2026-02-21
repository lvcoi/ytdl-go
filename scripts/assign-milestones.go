package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"
)

// MilestoneRule defines a rule for assigning milestones based on close date
type MilestoneRule struct {
	Name      string
	StartDate time.Time
	EndDate   time.Time
}

func main() {
	var (
		token    = flag.String("token", os.Getenv("GITHUB_TOKEN"), "GitHub personal access token (or set GITHUB_TOKEN env var)")
		owner    = flag.String("owner", "lvcoi", "GitHub repository owner")
		repo     = flag.String("repo", "ytdl-go", "GitHub repository name")
		dryRun   = flag.Bool("dry-run", true, "Dry run mode - don't actually update issues")
		verbose  = flag.Bool("verbose", false, "Verbose output")
		issueNum = flag.Int("issue", 0, "Process only this specific issue number (0 = all)")
	)
	flag.Parse()

	if *token == "" {
		fmt.Fprintln(os.Stderr, "GitHub token is required. Use -token flag or set GITHUB_TOKEN environment variable")
		os.Exit(1)
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: *token})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// Get all milestones
	milestones, err := getAllMilestones(ctx, client, *owner, *repo)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to fetch milestones: %v\n", err)
		os.Exit(1)
	}

	if *verbose {
		fmt.Printf("Found %d milestones:\n", len(milestones))
		for _, m := range milestones {
			fmt.Printf("  - %s (ID: %d, State: %s)\n", m.GetTitle(), m.GetNumber(), m.GetState())
		}
		fmt.Println()
	}

	// Get closed issues
	var issues []*github.Issue
	if *issueNum > 0 {
		issue, _, err := client.Issues.Get(ctx, *owner, *repo, *issueNum)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to fetch issue #%d: %v\n", *issueNum, err)
			os.Exit(1)
		}
		if issue.GetState() != "closed" {
			fmt.Fprintf(os.Stderr, "Issue #%d is not closed\n", *issueNum)
			os.Exit(1)
		}
		issues = []*github.Issue{issue}
	} else {
		issues, err = getAllClosedIssues(ctx, client, *owner, *repo)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to fetch closed issues: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Printf("Found %d closed issue(s) to process\n", len(issues))

	// Process issues and assign milestones
	rules := buildMilestoneRules(milestones)
	if len(rules) == 0 {
		fmt.Println("Warning: No milestone rules created. Using fallback strategy.")
	}

	processIssues(ctx, client, *owner, *repo, issues, milestones, rules, *dryRun, *verbose)
}

// processIssues processes issues and assigns milestones based on rules
func processIssues(ctx context.Context, client *github.Client, owner, repo string, issues []*github.Issue, milestones []*github.Milestone, rules []MilestoneRule, dryRun bool, verbose bool) {
	updated := 0
	skipped := 0
	errors := 0

	for _, issue := range issues {
		// Skip pull requests (GitHub API includes PRs in issues endpoint)
		if issue.PullRequestLinks != nil {
			skipped++
			continue
		}

		// Skip issues that already have a milestone
		if issue.Milestone != nil {
			if verbose {
				fmt.Printf("Issue #%d already has milestone '%s', skipping\n", issue.GetNumber(), issue.Milestone.GetTitle())
			}
			skipped++
			continue
		}

		// Determine appropriate milestone
		milestone := determineMilestone(issue, rules, milestones)
		if milestone == nil {
			if verbose {
				fmt.Printf("Issue #%d: No appropriate milestone found, skipping\n", issue.GetNumber())
			}
			skipped++
			continue
		}

		fmt.Printf("Issue #%d: '%s' -> Milestone '%s'\n",
			issue.GetNumber(),
			truncate(issue.GetTitle(), 60),
			milestone.GetTitle())

		if !dryRun {
			// Update the issue with the milestone
			issueReq := &github.IssueRequest{
				Milestone: github.Int(milestone.GetNumber()),
			}
			_, _, err := client.Issues.Edit(ctx, owner, repo, issue.GetNumber(), issueReq)
			if err != nil {
				fmt.Fprintf(os.Stderr, "  ERROR: Failed to update issue #%d: %v\n", issue.GetNumber(), err)
				errors++
				continue
			}
			fmt.Printf("  ✓ Updated successfully\n")
		} else {
			fmt.Printf("  (dry-run: no changes made)\n")
		}
		updated++
	}

	fmt.Println("\n=== Summary ===")
	fmt.Printf("Issues to update: %d\n", updated)
	fmt.Printf("Skipped: %d\n", skipped)
	if errors > 0 {
		fmt.Printf("Errors: %d\n", errors)
	}
	if dryRun {
		fmt.Println("\nThis was a DRY RUN. Use -dry-run=false to actually update issues.")
	}
}

// getAllMilestones fetches all milestones (open and closed)
func getAllMilestones(ctx context.Context, client *github.Client, owner, repo string) ([]*github.Milestone, error) {
	var allMilestones []*github.Milestone

	// Get open milestones
	opts := &github.MilestoneListOptions{
		State: "open",
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}
	openMilestones, _, err := client.Issues.ListMilestones(ctx, owner, repo, opts)
	if err != nil {
		return nil, err
	}
	allMilestones = append(allMilestones, openMilestones...)

	// Get closed milestones
	opts.State = "closed"
	closedMilestones, _, err := client.Issues.ListMilestones(ctx, owner, repo, opts)
	if err != nil {
		return nil, err
	}
	allMilestones = append(allMilestones, closedMilestones...)

	return allMilestones, nil
}

// getAllClosedIssues fetches all closed issues without milestones
func getAllClosedIssues(ctx context.Context, client *github.Client, owner, repo string) ([]*github.Issue, error) {
	var allIssues []*github.Issue

	opts := &github.IssueListByRepoOptions{
		State: "closed",
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	for {
		issues, resp, err := client.Issues.ListByRepo(ctx, owner, repo, opts)
		if err != nil {
			return nil, err
		}
		allIssues = append(allIssues, issues...)

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allIssues, nil
}

// buildMilestoneRules creates rules based on milestone due dates
func buildMilestoneRules(milestones []*github.Milestone) []MilestoneRule {
	var rules []MilestoneRule

	for _, m := range milestones {
		if m.DueOn == nil {
			continue
		}

		rule := MilestoneRule{
			Name:    m.GetTitle(),
			EndDate: m.DueOn.Time,
			// Start date is 30 days before due date (arbitrary, can be adjusted)
			StartDate: m.DueOn.Time.Add(-30 * 24 * time.Hour),
		}
		rules = append(rules, rule)
	}

	return rules
}

// determineMilestone determines the appropriate milestone for an issue
func determineMilestone(issue *github.Issue, rules []MilestoneRule, milestones []*github.Milestone) *github.Milestone {
	closedAt := issue.GetClosedAt().Time
	if closedAt.IsZero() {
		return nil
	}

	// Strategy 1: Use rules based on milestone due dates
	for _, rule := range rules {
		if (closedAt.After(rule.StartDate) || closedAt.Equal(rule.StartDate)) && (closedAt.Before(rule.EndDate) || closedAt.Equal(rule.EndDate)) {
			for _, m := range milestones {
				if m.GetTitle() == rule.Name {
					return m
				}
			}
		}
	}

	// Strategy 2: Find the most recent milestone that was created before the issue was closed
	var bestMilestone *github.Milestone
	var bestDate time.Time

	for _, m := range milestones {
		createdAt := m.GetCreatedAt().Time
		if createdAt.Before(closedAt) && (bestMilestone == nil || createdAt.After(bestDate)) {
			bestMilestone = m
			bestDate = createdAt
		}
	}

	// Strategy 3: If no good match, use the earliest open milestone
	if bestMilestone == nil {
		for _, m := range milestones {
			if m.GetState() == "open" {
				if bestMilestone == nil || m.GetCreatedAt().Time.Before(bestMilestone.GetCreatedAt().Time) {
					bestMilestone = m
				}
			}
		}
	}

	return bestMilestone
}

// truncate truncates a string to maxLen characters
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
