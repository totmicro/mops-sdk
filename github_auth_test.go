package sdk

import (
	"testing"
)

func TestGitHubReleaseHelperAuth(t *testing.T) {
	// Test basic helper (no auth)
	helper := NewGitHubReleaseHelper()
	if helper == nil {
		t.Fatal("NewGitHubReleaseHelper returned nil")
	}

	// Test auth availability check
	isAvailable := helper.isGHCLIAvailable()
	t.Logf("GitHub CLI available: %v", isAvailable)

	// Test getting token from gh CLI (will only work if gh is installed and authenticated)
	if isAvailable {
		token, err := helper.getGitHubTokenFromGHCLI()
		if err != nil {
			t.Logf("Could not get GitHub token from gh CLI: %v", err)
		} else {
			t.Logf("Successfully got GitHub token (length: %d characters)", len(token))
			
			// Test with auth
			authHelper := NewGitHubReleaseHelperWithAuth(token)
			if authHelper == nil {
				t.Fatal("NewGitHubReleaseHelperWithAuth returned nil")
			}
		}
	}

	// Test automatic auth helper
	authHelper, err := NewAuthenticatedGitHubReleaseHelper()
	if err != nil {
		t.Logf("Could not create authenticated helper: %v", err)
	} else {
		t.Log("Successfully created authenticated GitHub release helper")
		
		// Test with a public repository
		release, err := authHelper.GetLatestRelease("netbirdio", "netbird")
		if err != nil {
			t.Errorf("Failed to get latest release: %v", err)
		} else {
			t.Logf("Got latest release: %s (published: %s)", release.TagName, release.PublishedAt.Format("2006-01-02"))
		}
	}
}