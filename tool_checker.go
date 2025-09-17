package sdk

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// ToolChecker provides methods to check if tools are installed on the system
type ToolChecker struct{}

// NewToolChecker creates a new ToolChecker instance
func NewToolChecker() *ToolChecker {
	return &ToolChecker{}
}

// ToolInfo contains information about a detected tool
type ToolInfo struct {
	Name      string // Tool name
	Path      string // Full path to the executable
	Version   string // Version string (if detectable)
	Installed bool   // Whether the tool is installed
}

// IsInstalled checks if a tool is installed and accessible in the system PATH
func (tc *ToolChecker) IsInstalled(toolName string) bool {
	_, err := tc.FindTool(toolName)
	return err == nil
}

// FindTool locates a tool in the system and returns detailed information
func (tc *ToolChecker) FindTool(toolName string) (*ToolInfo, error) {
	info := &ToolInfo{
		Name:      toolName,
		Installed: false,
	}

	// Find the executable path
	execPath, err := tc.findExecutable(toolName)
	if err != nil {
		return info, fmt.Errorf("tool '%s' not found: %w", toolName, err)
	}

	info.Path = execPath
	info.Installed = true

	// Try to get version information
	if version := tc.getToolVersion(toolName, execPath); version != "" {
		info.Version = version
	}

	return info, nil
}

// CheckMultipleTools checks multiple tools at once and returns a map of results
func (tc *ToolChecker) CheckMultipleTools(toolNames []string) map[string]*ToolInfo {
	results := make(map[string]*ToolInfo)
	
	for _, toolName := range toolNames {
		if info, err := tc.FindTool(toolName); err == nil {
			results[toolName] = info
		} else {
			results[toolName] = &ToolInfo{
				Name:      toolName,
				Installed: false,
			}
		}
	}
	
	return results
}

// findExecutable locates an executable in the system PATH with OS-specific logic
func (tc *ToolChecker) findExecutable(toolName string) (string, error) {
	switch runtime.GOOS {
	case "windows":
		return tc.findExecutableWindows(toolName)
	case "darwin", "linux":
		return tc.findExecutableUnix(toolName)
	default:
		return tc.findExecutableUnix(toolName) // Default to Unix-like behavior
	}
}

// findExecutableWindows finds executables on Windows systems
func (tc *ToolChecker) findExecutableWindows(toolName string) (string, error) {
	// Windows executable extensions to try
	extensions := []string{"", ".exe", ".cmd", ".bat", ".ps1"}
	
	// First try with exec.LookPath which handles PATH and PATHEXT
	for _, ext := range extensions {
		nameWithExt := toolName + ext
		if path, err := exec.LookPath(nameWithExt); err == nil {
			return path, nil
		}
	}
	
	// If not found in PATH, try common installation directories
	commonDirs := []string{
		`C:\ProgramData\chocolatey\bin\`,
		`C:\tools\`,
		filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Local", "Programs"),
		filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming", "npm"),
		`C:\Program Files\`,
		`C:\Program Files (x86)\`,
	}
	
	for _, dir := range commonDirs {
		for _, ext := range extensions {
			fullPath := filepath.Join(dir, toolName+ext)
			if tc.fileExists(fullPath) && tc.isExecutable(fullPath) {
				return fullPath, nil
			}
		}
	}
	
	return "", fmt.Errorf("executable not found")
}

// findExecutableUnix finds executables on Unix-like systems (Linux, macOS)
func (tc *ToolChecker) findExecutableUnix(toolName string) (string, error) {
	// First try with exec.LookPath
	if path, err := exec.LookPath(toolName); err == nil {
		return path, nil
	}
	
	// If not found in PATH, try common installation directories
	commonDirs := []string{
		"/usr/local/bin",
		"/usr/bin",
		"/bin",
		"/opt/homebrew/bin", // macOS Homebrew on Apple Silicon
		"/usr/local/opt",    // macOS Homebrew on Intel
		filepath.Join(os.Getenv("HOME"), "bin"),
		filepath.Join(os.Getenv("HOME"), ".local", "bin"),
		"/snap/bin", // Ubuntu Snap packages
	}
	
	for _, dir := range commonDirs {
		fullPath := filepath.Join(dir, toolName)
		if tc.fileExists(fullPath) && tc.isExecutable(fullPath) {
			return fullPath, nil
		}
	}
	
	return "", fmt.Errorf("executable not found")
}

// getToolVersion attempts to get version information for a tool
func (tc *ToolChecker) getToolVersion(toolName, execPath string) string {
	// Common version flags to try
	versionFlags := []string{"--version", "-v", "-V", "version"}
	
	for _, flag := range versionFlags {
		if version := tc.tryGetVersion(execPath, flag); version != "" {
			return version
		}
	}
	
	return ""
}

// tryGetVersion executes a command with a version flag and parses the output
func (tc *ToolChecker) tryGetVersion(execPath, flag string) string {
	cmd := exec.Command(execPath, flag)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return ""
	}
	
	// Parse version from output (first line, remove extra whitespace)
	lines := strings.Split(string(output), "\n")
	if len(lines) > 0 {
		version := strings.TrimSpace(lines[0])
		// Limit version string length for readability
		if len(version) > 100 {
			version = version[:100] + "..."
		}
		return version
	}
	
	return ""
}

// fileExists checks if a file exists
func (tc *ToolChecker) fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// isExecutable checks if a file is executable
func (tc *ToolChecker) isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	
	// On Windows, check file extension
	if runtime.GOOS == "windows" {
		ext := strings.ToLower(filepath.Ext(path))
		execExts := []string{".exe", ".cmd", ".bat", ".ps1"}
		for _, execExt := range execExts {
			if ext == execExt {
				return true
			}
		}
		return false
	}
	
	// On Unix-like systems, check execute permission
	return info.Mode()&0111 != 0
}

// RequireTool checks if a tool is installed and returns an error if not
func (tc *ToolChecker) RequireTool(toolName string) error {
	if !tc.IsInstalled(toolName) {
		return fmt.Errorf("required tool '%s' is not installed or not accessible in PATH", toolName)
	}
	return nil
}

// RequireTools checks multiple tools and returns an error listing any missing tools
func (tc *ToolChecker) RequireTools(toolNames []string) error {
	var missingTools []string
	
	for _, toolName := range toolNames {
		if !tc.IsInstalled(toolName) {
			missingTools = append(missingTools, toolName)
		}
	}
	
	if len(missingTools) > 0 {
		return fmt.Errorf("required tools not found: %s", strings.Join(missingTools, ", "))
	}
	
	return nil
}