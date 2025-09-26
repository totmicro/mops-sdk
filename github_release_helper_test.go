package sdk

import (
	"testing"
)

func TestCompareVersions(t *testing.T) {
	helper := NewGitHubReleaseHelper()
	
	tests := []struct {
		v1       string
		v2       string
		expected int
		desc     string
	}{
		{"1.0.0", "1.0.0", 0, "equal versions"},
		{"1.0.0", "1.0.1", -1, "patch version upgrade available"},
		{"1.0.1", "1.0.0", 1, "current is newer patch"},
		{"1.0.0", "1.1.0", -1, "minor version upgrade available"},
		{"1.1.0", "1.0.0", 1, "current is newer minor"},
		{"1.0.0", "2.0.0", -1, "major version upgrade available"},
		{"2.0.0", "1.0.0", 1, "current is newer major"},
		{"v1.0.0", "1.0.0", 0, "v prefix handling"},
		{"1.0.0", "v1.0.0", 0, "v prefix handling reverse"},
		{"v1.0.0", "v1.0.0", 0, "both with v prefix"},
		{"1.2.3", "1.2.10", -1, "numeric comparison within patch"},
		{"1.10.0", "1.2.0", 1, "numeric comparison within minor"},
		{"10.0.0", "2.0.0", 1, "numeric comparison within major"},
		{"1.0", "1.0.0", 0, "missing patch version"},
		{"1", "1.0.0", 0, "missing minor and patch versions"},
		{"1.0.0-beta", "1.0.0", -1, "pre-release handling"},
		{"", "1.0.0", -1, "empty version"},
		{"1.0.0", "", 1, "empty version reverse"},
	}
	
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			result := helper.CompareVersions(test.v1, test.v2)
			if result != test.expected {
				t.Errorf("CompareVersions(%q, %q) = %d, expected %d", 
					test.v1, test.v2, result, test.expected)
			}
		})
	}
}

func TestVersionCheckResult(t *testing.T) {
	helper := NewGitHubReleaseHelper()
	
	// Test the convenience method logic
	tests := []struct {
		current         string
		latest          string
		expectedUpdate  bool
		expectedComp    int
		desc           string
	}{
		{"1.0.0", "1.0.1", true, -1, "update available"},
		{"1.0.1", "1.0.0", false, 1, "current newer"},
		{"1.0.0", "1.0.0", false, 0, "up to date"},
		{"v1.0.0", "1.0.1", true, -1, "update available with v prefix"},
	}
	
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			// Simulate the logic that would be in CheckVersions
			result := &VersionCheckResult{
				CurrentVersion: test.current,
				LatestVersion:  test.latest,
			}
			result.Comparison = helper.CompareVersions(test.current, test.latest)
			result.UpdateAvailable = result.Comparison < 0
			
			if result.UpdateAvailable != test.expectedUpdate {
				t.Errorf("UpdateAvailable = %t, expected %t", result.UpdateAvailable, test.expectedUpdate)
			}
			if result.Comparison != test.expectedComp {
				t.Errorf("Comparison = %d, expected %d", result.Comparison, test.expectedComp)
			}
		})
	}
}