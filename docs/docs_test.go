package docs_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestDocsConfigExists verifies zensical.toml is present and contains required keys.
func TestDocsConfigExists(t *testing.T) {
	root := findRepoRoot(t)
	path := filepath.Join(root, "zensical.toml")

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("zensical.toml not found: %v", err)
	}

	content := string(data)
	requiredKeys := []string{"[project]", "site_name =", "nav = [", "[project.theme]", "features = ["}
	for _, key := range requiredKeys {
		if !strings.Contains(content, key) {
			t.Errorf("zensical.toml missing required key: %s", key)
		}
	}
}

// TestRequiredDocsExist verifies all .md pages referenced in zensical.toml nav exist on disk.
func TestRequiredDocsExist(t *testing.T) {
	root := findRepoRoot(t)
	pages := extractNavPagesFromFile(t, filepath.Join(root, "zensical.toml"))

	if len(pages) == 0 {
		t.Fatal("no pages found in zensical.toml nav")
	}

	docsDir := filepath.Join(root, "docs")
	for _, page := range pages {
		pagePath := filepath.Join(docsDir, page)
		if _, err := os.Stat(pagePath); os.IsNotExist(err) {
			t.Errorf("page referenced in nav but missing on disk: docs/%s", page)
		}
	}
}

// TestDocsWorkflowExists verifies the GitHub Actions docs workflow is present and correct.
func TestDocsWorkflowExists(t *testing.T) {
	root := findRepoRoot(t)
	path := filepath.Join(root, ".github", "workflows", "Publish.yml")

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf(".github/workflows/Publish.yml not found: %v", err)
	}

	content := string(data)
	requiredStrings := []string{
		"Documentation",
		"actions/configure-pages",
		"actions/checkout",
		"actions/setup-python",
		"pip install zensical",
		"zensical build --clean",
		"actions/deploy-pages",
	}
	for _, s := range requiredStrings {
		if !strings.Contains(content, s) {
			t.Errorf("docs workflow missing expected content: %q", s)
		}
	}
}

// TestNavStructureHasRequiredSections verifies the nav has all required top-level sections.
func TestNavStructureHasRequiredSections(t *testing.T) {
	root := findRepoRoot(t)
	data, err := os.ReadFile(filepath.Join(root, "zensical.toml"))
	if err != nil {
		t.Fatalf("zensical.toml not found: %v", err)
	}

	content := string(data)
	requiredSections := []string{"{\"Home\" = \"index.md\"}", "{\"User Guide\" = [", "{\"Developer Guide\" = [", "{\"Reference\" = ["}
	for _, section := range requiredSections {
		if !strings.Contains(content, section) {
			t.Errorf("zensical.toml nav missing required section: %s", section)
		}
	}
}

// TestDocsNoOrphanedFiles checks that markdown files in wiki subdirectories are referenced in nav.
func TestDocsNoOrphanedFiles(t *testing.T) {
	root := findRepoRoot(t)
	navPages := make(map[string]bool)
	for _, page := range extractNavPagesFromFile(t, filepath.Join(root, "zensical.toml")) {
		navPages[page] = true
	}

	docsDir := filepath.Join(root, "docs")
	wikiDirs := []string{"user-guide", "developer-guide", "reference"}

	for _, dir := range wikiDirs {
		dirPath := filepath.Join(docsDir, dir)
		entries, err := os.ReadDir(dirPath)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
				continue
			}
			relPath := filepath.Join(dir, entry.Name())
			if !navPages[relPath] {
				t.Errorf("docs/%s exists but is not referenced in zensical.toml nav", relPath)
			}
		}
	}
}

// TestIndexPageHasLinks verifies the home page links to key sections.
func TestIndexPageHasLinks(t *testing.T) {
	root := findRepoRoot(t)
	data, err := os.ReadFile(filepath.Join(root, "docs", "index.md"))
	if err != nil {
		t.Fatalf("docs/index.md not found: %v", err)
	}

	content := string(data)
	requiredLinks := []string{
		"user-guide/installation.md",
		"user-guide/quick-start.md",
		"developer-guide/architecture.md",
		"reference/cli-options.md",
	}
	for _, link := range requiredLinks {
		if !strings.Contains(content, link) {
			t.Errorf("docs/index.md missing link to: %s", link)
		}
	}
}

// TestDocsThemeConfig verifies theme configuration matches project requirements.
func TestDocsThemeConfig(t *testing.T) {
	root := findRepoRoot(t)
	data, err := os.ReadFile(filepath.Join(root, "zensical.toml"))
	if err != nil {
		t.Fatalf("zensical.toml not found: %v", err)
	}

	content := string(data)

	checks := map[string]string{
		"theme section":    "[project.theme]",
		"default palette":  "scheme = \"default\"",
		"dark palette":     "scheme = \"slate\"",
		"navigation tabs":  "\"navigation.tabs\"",
		"search highlight": "\"search.highlight\"",
		"code copy":        "\"content.code.copy\"",
		"repo URL":         "link = \"https://github.com/lvcoi/ytdl-go\"",
	}

	for desc, expected := range checks {
		if !strings.Contains(content, expected) {
			t.Errorf("zensical.toml missing %s (expected %q)", desc, expected)
		}
	}
}

// extractNavPagesFromFile extracts all .md file paths from zensical.toml nav using regex.
func extractNavPagesFromFile(t *testing.T, path string) []string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("cannot read %s: %v", path, err)
	}

	// Match TOML nav entries like {"Title" = "path/to/file.md"}
	re := regexp.MustCompile(`=\s*"([a-zA-Z0-9_\-/]+\.md)"`)
	var pages []string
	for _, line := range strings.Split(string(data), "\n") {
		matches := re.FindStringSubmatch(line)
		if len(matches) > 1 {
			pages = append(pages, matches[1])
		}
	}
	return pages
}

// findRepoRoot walks up from the test file to find the repository root (contains go.mod).
func findRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("cannot get working directory: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find repository root (no go.mod found)")
		}
		dir = parent
	}
}
