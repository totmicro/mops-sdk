# GitHub Release Helper with Authentication

The `GitHubReleaseHelper` now supports authentication for private repositories using the GitHub CLI (`gh`) tool.

## Features

### Public Repositories (No Auth Required)
```go
// Works for public repos like netbirdio/netbird
helper := sdk.NewGitHubReleaseHelper()
release, err := helper.GetLatestRelease("netbirdio", "netbird")
```

### Private Repositories (Authentication Required)

#### Option 1: Automatic Authentication (Recommended)
```go
// Automatically uses gh CLI authentication
helper, err := sdk.NewAuthenticatedGitHubReleaseHelper()
if err != nil {
    // Handle error - gh CLI not installed or not authenticated
    return fmt.Errorf("GitHub authentication failed: %w", err)
}

release, err := helper.GetLatestRelease("myorg", "private-repo")
```

#### Option 2: Manual Token
```go
// Use a specific token
token := "ghp_xxxxxxxxxxxxxxxxxxxx"
helper := sdk.NewGitHubReleaseHelperWithAuth(token)
release, err := helper.GetLatestRelease("myorg", "private-repo")
```

#### Option 3: Fallback Authentication
```go
// Try public first, fallback to authenticated if needed
helper := sdk.NewGitHubReleaseHelper()
release, err := helper.GetLatestRelease("myorg", "maybe-private-repo")
if err != nil && strings.Contains(err.Error(), "not found") {
    // Try with authentication
    authHelper, authErr := sdk.NewAuthenticatedGitHubReleaseHelper()
    if authErr == nil {
        release, err = authHelper.GetLatestRelease("myorg", "maybe-private-repo")
    }
}
```

## Prerequisites for Private Repositories

1. **Install GitHub CLI**:
   ```bash
   # macOS
   brew install gh
   
   # Ubuntu/Debian
   curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | sudo dd of=/usr/share/keyrings/githubcli-archive-keyring.gpg
   echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | sudo tee /etc/apt/sources.list.d/github-cli.list > /dev/null
   sudo apt update && sudo apt install gh
   ```

2. **Authenticate with GitHub**:
   ```bash
   gh auth login
   ```

3. **Verify Authentication**:
   ```bash
   gh auth status
   ```

## Error Handling

The helper provides clear error messages for common authentication issues:

- `"repository not found or no releases available (may be private - ensure gh CLI is authenticated)"`
- `"access denied to repository (may be private - ensure gh CLI is authenticated with 'gh auth login')"`
- `"failed to get GitHub token from gh CLI: (make sure 'gh auth login' has been run)"`

## Rate Limits

- **Unauthenticated**: 60 requests per hour per IP
- **Authenticated**: 5,000 requests per hour per user
- **GitHub App**: 15,000 requests per hour per installation

Using authentication significantly increases the rate limit and allows access to private repositories.