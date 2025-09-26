# GitHub Release Helper Usage Guide

The `GitHubReleaseHelper` in mops-sdk provides a generic way for plugins to interact with GitHub releases API to detect latest versions, compare versions, and check for updates.

## Basic Usage

### 1. Import and Create Helper

```go
import sdk "github.com/totmicro/mops-sdk"

githubHelper := sdk.NewGitHubReleaseHelper()
```

### 2. Get Latest Version

```go
// Get just the version string
latestVersion, err := githubHelper.GetLatestVersion("owner", "repo")
if err != nil {
    // Handle error
}
fmt.Printf("Latest version: %s\n", latestVersion)
```

### 3. Get Full Release Information

```go
// Get complete release information
release, err := githubHelper.GetLatestRelease("owner", "repo")
if err != nil {
    // Handle error
}

fmt.Printf("Version: %s\n", release.TagName)
fmt.Printf("Published: %s\n", release.PublishedAt)
fmt.Printf("URL: %s\n", release.HTMLURL)
fmt.Printf("Is prerelease: %t\n", release.Prerelease)
```

### 4. Compare Versions

```go
currentVersion := "1.2.0"
latestVersion := "1.3.0"

comparison := githubHelper.CompareVersions(currentVersion, latestVersion)
switch comparison {
case -1:
    fmt.Println("Update available")
case 0:
    fmt.Println("Up to date")
case 1:
    fmt.Println("Newer than latest")
}
```

### 5. Check for Updates (Convenience Method)

```go
versionCheck := githubHelper.CheckVersions("owner", "repo", currentVersion)
if versionCheck.Error != nil {
    // Handle error
}

fmt.Printf("Current: %s\n", versionCheck.CurrentVersion)
fmt.Printf("Latest: %s\n", versionCheck.LatestVersion)
fmt.Printf("Update available: %t\n", versionCheck.UpdateAvailable)
```

## Example Plugin Implementation

Here's how a plugin can integrate GitHub release detection:

```go
func installTool(ctx context.Context, outputChan chan<- string, inputChan <-chan string, params map[string]interface{}) error {
    // Check current installation
    currentVersion := getCurrentToolVersion() // Your implementation
    
    // Check GitHub for latest version
    outputChan <- "🌐 Checking for latest release..."
    githubHelper := sdk.NewGitHubReleaseHelper()
    
    versionCheck := githubHelper.CheckVersions("owner", "repo", currentVersion)
    if versionCheck.Error != nil {
        outputChan <- fmt.Sprintf("⚠️  Could not check latest version: %v", versionCheck.Error)
        // Continue with installation using fallback method
    } else {
        outputChan <- fmt.Sprintf("🆕 Latest version: v%s", versionCheck.LatestVersion)
        
        if versionCheck.UpdateAvailable {
            outputChan <- "📈 Update available! Proceeding with upgrade..."
        } else if versionCheck.Comparison == 0 {
            outputChan <- "✅ Already up to date"
            return nil
        }
    }
    
    // Proceed with installation using the detected version
    return installWithVersion(versionCheck.LatestVersion)
}
```

## Features

- **Generic**: Works with any GitHub repository
- **Error Handling**: Graceful fallback when GitHub API is unavailable
- **Version Comparison**: Semantic version comparison with support for various formats
- **Rate Limiting Aware**: Includes proper headers to avoid GitHub API limits
- **Multiple Release Support**: Can fetch multiple releases, not just the latest

## Error Handling

The helper handles common scenarios:
- Network connectivity issues
- GitHub API rate limiting
- Invalid repository names
- Missing releases

Always check for errors and provide fallback mechanisms in your plugins.

## Real-World Example: NetBird Plugin

The NetBird plugin demonstrates full integration:

1. **Check Current Version**: Uses ToolInstaller to detect installed version
2. **Check Latest Version**: Uses GitHubReleaseHelper to get latest from GitHub
3. **Smart Updates**: Only installs/updates when necessary
4. **User Feedback**: Provides clear status and version comparison info
5. **Standalone Version Check**: Separate action just for checking versions

See `/mops-plugins/plugins/netbird/installation.go` for the complete implementation.