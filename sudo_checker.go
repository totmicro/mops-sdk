package sdk

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// SudoChecker provides methods to check sudo privileges and capabilities
type SudoChecker struct{}

// NewSudoChecker creates a new SudoChecker instance
func NewSudoChecker() *SudoChecker {
	return &SudoChecker{}
}

// SudoInfo contains information about sudo privileges and status
type SudoInfo struct {
	HasSudo           bool   // Whether sudo is available
	IsRoot            bool   // Whether running as root user
	CanSudoWithoutPwd bool   // Whether can sudo without password
	SudoPath          string // Path to sudo binary
	CurrentUser       string // Current username
	SudoTimeLeft      int    // Minutes left in sudo cache (if any)
	ErrorMessage      string // Error message if sudo check failed
}

// CheckSudoPrivileges performs a comprehensive check of sudo privileges
func (sc *SudoChecker) CheckSudoPrivileges() (*SudoInfo, error) {
	info := &SudoInfo{
		CurrentUser: sc.getCurrentUser(),
	}

	// Only supported on Unix-like systems
	if runtime.GOOS == "windows" {
		info.ErrorMessage = "sudo privilege checking is not supported on Windows"
		return info, nil
	}

	// Check if running as root
	info.IsRoot = sc.isRunningAsRoot()

	// Find sudo binary
	sudoPath, hasSudo := sc.findSudo()
	info.HasSudo = hasSudo
	info.SudoPath = sudoPath

	if !hasSudo {
		info.ErrorMessage = "sudo binary not found in system PATH"
		return info, nil
	}

	// Check if can sudo without password
	info.CanSudoWithoutPwd = sc.canSudoWithoutPassword()

	// Check sudo cache time remaining
	info.SudoTimeLeft = sc.getSudoCacheTimeLeft()

	return info, nil
}

// HasSudo returns true if sudo is available and user can use it
func (sc *SudoChecker) HasSudo() bool {
	info, err := sc.CheckSudoPrivileges()
	if err != nil {
		return false
	}
	return info.HasSudo && (info.IsRoot || info.CanSudoWithoutPwd || info.SudoTimeLeft > 0)
}

// RequireSudo checks if sudo privileges are available and returns an error if not
func (sc *SudoChecker) RequireSudo() error {
	if runtime.GOOS == "windows" {
		return NewError("sudo privileges are not supported on Windows")
	}

	info, err := sc.CheckSudoPrivileges()
	if err != nil {
		return NewError("failed to check sudo privileges: %v", err)
	}

	if !info.HasSudo {
		return NewError("sudo is not installed or not available in PATH")
	}

	if info.IsRoot {
		return nil // Already running as root
	}

	if info.CanSudoWithoutPwd {
		return nil // Can sudo without password
	}

	if info.SudoTimeLeft > 0 {
		return nil // Sudo cache is still valid
	}

	return NewError("sudo privileges are required but not available (try running 'sudo -v' first)")
}

// PromptForSudo attempts to prompt for sudo privileges
// WARNING: This method is INTERACTIVE and will prompt the user for a password.
// It should only be used when the plugin explicitly needs to escalate privileges
// and the user has explicitly requested a privileged operation.
// For checking sudo availability, use CheckSudoPrivileges() or HasSudo() instead.
func (sc *SudoChecker) PromptForSudo() error {
	if runtime.GOOS == "windows" {
		return NewError("sudo is not supported on Windows")
	}

	if sc.isRunningAsRoot() {
		return nil // Already root
	}

	// Try to validate sudo credentials
	cmd := exec.Command("sudo", "-v")
	err := cmd.Run()
	if err != nil {
		return NewError("failed to obtain sudo privileges: %v", err)
	}

	return nil
}

// isRunningAsRoot checks if the current process is running as root
func (sc *SudoChecker) isRunningAsRoot() bool {
	return os.Geteuid() == 0
}

// getCurrentUser returns the current username
func (sc *SudoChecker) getCurrentUser() string {
	if user := os.Getenv("USER"); user != "" {
		return user
	}
	if user := os.Getenv("USERNAME"); user != "" {
		return user
	}
	return "unknown"
}

// findSudo locates the sudo binary
func (sc *SudoChecker) findSudo() (string, bool) {
	path, err := exec.LookPath("sudo")
	return path, err == nil
}

// canSudoWithoutPassword checks if user can sudo without entering a password
func (sc *SudoChecker) canSudoWithoutPassword() bool {
	// Try a non-interactive sudo check
	cmd := exec.Command("sudo", "-n", "true")
	err := cmd.Run()
	return err == nil
}

// getSudoCacheTimeLeft returns minutes left in sudo cache, or 0 if no cache
func (sc *SudoChecker) getSudoCacheTimeLeft() int {
	// Check if sudo cache is valid by running a quick non-interactive command
	cmd := exec.Command("sudo", "-n", "true")
	err := cmd.Run()
	if err != nil {
		return 0
	}

	// Try to get more specific cache information using sudo -V
	cmd = exec.Command("sudo", "-V")
	output, err := cmd.Output()
	if err != nil {
		return 1 // Assume some time left since previous check passed
	}

	// Parse output for cache information
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "timestamp timeout") {
			// This would contain timeout information, but format varies
			// For simplicity, assume cache is valid if we got here
			return 15 // Default sudo timeout is usually 15 minutes
		}
	}

	return 1 // Conservative estimate
}

// TestSudoExecution tests if a command can be executed with sudo
func (sc *SudoChecker) TestSudoExecution(command string, args ...string) error {
	if runtime.GOOS == "windows" {
		return NewError("sudo is not supported on Windows")
	}

	// Prepare sudo command
	sudoArgs := append([]string{"-n", command}, args...)
	cmd := exec.Command("sudo", sudoArgs...)
	
	// Set a timeout to avoid hanging
	done := make(chan error, 1)
	go func() {
		done <- cmd.Run()
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(5 * time.Second):
		cmd.Process.Kill()
		return NewError("sudo command timed out")
	}
}

// GetSudoStatus returns a human-readable status string
func (sc *SudoChecker) GetSudoStatus() string {
	info, err := sc.CheckSudoPrivileges()
	if err != nil {
		return "Error checking sudo status"
	}

	if runtime.GOOS == "windows" {
		return "Windows (sudo not applicable)"
	}

	if info.IsRoot {
		return "Running as root user"
	}

	if !info.HasSudo {
		return "sudo not available"
	}

	if info.CanSudoWithoutPwd {
		return "sudo available (passwordless)"
	}

	if info.SudoTimeLeft > 0 {
		return "sudo available (cached credentials)"
	}

	return "sudo available (password required)"
}

// NewError creates a formatted error - helper function
func NewError(format string, args ...interface{}) error {
	if len(args) == 0 {
		return fmt.Errorf("%s", format)
	}
	return fmt.Errorf(format, args...)
}