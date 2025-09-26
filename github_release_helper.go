package sdk

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

// GitHubRelease represents a GitHub release
type GitHubRelease struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	Body        string    `json:"body"`
	Draft       bool      `json:"draft"`
	Prerelease  bool      `json:"prerelease"`
	PublishedAt time.Time `json:"published_at"`
	HTMLURL     string    `json:"html_url"`
	Assets      []GitHubReleaseAsset `json:"assets"`
}

// GitHubReleaseAsset represents a release asset
type GitHubReleaseAsset struct {
	Name               string `json:"name"`
	Label              string `json:"label"`
	ContentType        string `json:"content_type"`
	Size               int64  `json:"size"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// GitHubReleaseHelper provides utilities for working with GitHub releases
type GitHubReleaseHelper struct {
	client *http.Client
}

// NewGitHubReleaseHelper creates a new GitHub release helper
func NewGitHubReleaseHelper() *GitHubReleaseHelper {
	return &GitHubReleaseHelper{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetLatestRelease fetches the latest release for a GitHub repository
func (g *GitHubReleaseHelper) GetLatestRelease(owner, repo string) (*GitHubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	// Set headers to avoid rate limiting and get proper response
	req.Header.Set("User-Agent", "mops-plugin/1.0")
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	
	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}
	
	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	return &release, nil
}

// GetLatestVersion returns just the version string from the latest release
func (g *GitHubReleaseHelper) GetLatestVersion(owner, repo string) (string, error) {
	release, err := g.GetLatestRelease(owner, repo)
	if err != nil {
		return "", err
	}
	
	// Clean up version string (remove 'v' prefix if present)
	version := strings.TrimPrefix(release.TagName, "v")
	return version, nil
}

// CompareVersions compares two semantic version strings
// Returns: -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2
func (g *GitHubReleaseHelper) CompareVersions(v1, v2 string) int {
	// Clean up versions
	v1 = strings.TrimPrefix(strings.TrimSpace(v1), "v")
	v2 = strings.TrimPrefix(strings.TrimSpace(v2), "v")
	
	if v1 == v2 {
		return 0
	}
	
	// Handle empty versions
	if v1 == "" && v2 != "" {
		return -1
	}
	if v2 == "" && v1 != "" {
		return 1
	}
	if v1 == "" && v2 == "" {
		return 0
	}
	
	// Split versions into main version and pre-release parts
	v1Main, v1Pre := g.splitVersionPrerelease(v1)
	v2Main, v2Pre := g.splitVersionPrerelease(v2)
	
	// Split main versions into parts
	parts1 := strings.Split(v1Main, ".")
	parts2 := strings.Split(v2Main, ".")
	
	// Ensure both have at least 3 parts (major.minor.patch)
	for len(parts1) < 3 {
		parts1 = append(parts1, "0")
	}
	for len(parts2) < 3 {
		parts2 = append(parts2, "0")
	}
	
	// Compare each part of main version
	maxParts := len(parts1)
	if len(parts2) > maxParts {
		maxParts = len(parts2)
	}
	
	for i := 0; i < maxParts; i++ {
		part1 := "0"
		part2 := "0"
		
		if i < len(parts1) {
			part1 = parts1[i]
		}
		if i < len(parts2) {
			part2 = parts2[i]
		}
		
		// Convert to integers for numeric comparison
		num1, err1 := strconv.Atoi(part1)
		num2, err2 := strconv.Atoi(part2)
		
		if err1 != nil || err2 != nil {
			// If can't convert to int, do string comparison
			if part1 < part2 {
				return -1
			} else if part1 > part2 {
				return 1
			}
		} else {
			// Numeric comparison
			if num1 < num2 {
				return -1
			} else if num1 > num2 {
				return 1
			}
		}
	}
	
	// Main versions are equal, compare pre-release
	// According to semver: pre-release versions have lower precedence than normal versions
	if v1Pre == "" && v2Pre != "" {
		return 1 // v1 (release) > v2 (pre-release)
	}
	if v1Pre != "" && v2Pre == "" {
		return -1 // v1 (pre-release) < v2 (release)
	}
	if v1Pre != "" && v2Pre != "" {
		// Both are pre-releases, compare pre-release identifiers
		if v1Pre < v2Pre {
			return -1
		} else if v1Pre > v2Pre {
			return 1
		}
	}
	
	return 0
}

// splitVersionPrerelease splits a version string into main version and pre-release parts
func (g *GitHubReleaseHelper) splitVersionPrerelease(version string) (main, prerelease string) {
	parts := strings.Split(version, "-")
	main = parts[0]
	if len(parts) > 1 {
		prerelease = strings.Join(parts[1:], "-")
	}
	return main, prerelease
}

// IsUpdateAvailable checks if an update is available for the given current version
func (g *GitHubReleaseHelper) IsUpdateAvailable(owner, repo, currentVersion string) (bool, string, error) {
	latestVersion, err := g.GetLatestVersion(owner, repo)
	if err != nil {
		return false, "", err
	}
	
	comparison := g.CompareVersions(currentVersion, latestVersion)
	return comparison < 0, latestVersion, nil
}

// VersionCheckResult represents the result of a version check
type VersionCheckResult struct {
	CurrentVersion  string
	LatestVersion   string
	UpdateAvailable bool
	Comparison      int  // -1: current < latest, 0: equal, 1: current > latest
	Error           error
}

// CheckVersions is a convenience method that checks both current and latest versions
func (g *GitHubReleaseHelper) CheckVersions(owner, repo, currentVersion string) *VersionCheckResult {
	result := &VersionCheckResult{
		CurrentVersion: currentVersion,
	}
	
	latestVersion, err := g.GetLatestVersion(owner, repo)
	if err != nil {
		result.Error = err
		return result
	}
	
	result.LatestVersion = latestVersion
	result.Comparison = g.CompareVersions(currentVersion, latestVersion)
	result.UpdateAvailable = result.Comparison < 0
	
	return result
}

// GetReleases fetches all releases for a GitHub repository (up to 100)
func (g *GitHubReleaseHelper) GetReleases(owner, repo string, limit int) ([]GitHubRelease, error) {
	if limit <= 0 || limit > 100 {
		limit = 30 // Default reasonable limit
	}
	
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases?per_page=%d", owner, repo, limit)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("User-Agent", "mops-plugin/1.0")
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	
	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch releases: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}
	
	var releases []GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	// Sort releases by version (newest first)
	sort.Slice(releases, func(i, j int) bool {
		return g.CompareVersions(releases[i].TagName, releases[j].TagName) > 0
	})
	
	return releases, nil
}