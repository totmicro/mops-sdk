package sdk

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// ToolInstallCommand represents a platform-specific installation command or script
type ToolInstallCommand struct {
	// Command to execute (can be a shell command, script path, or executable)
	Command string `yaml:"command"`
	// Arguments to pass to the command
	Args []string `yaml:"args,omitempty"`
	// Working directory for command execution
	WorkDir string `yaml:"work_dir,omitempty"`
	// Environment variables to set
	Env map[string]string `yaml:"env,omitempty"`
	// Whether this command requires sudo/elevated privileges
	RequiresSudo bool `yaml:"requires_sudo,omitempty"`
	// Whether to prime sudo cache before command execution (without running command as root)
	PrimeSudoCache bool `yaml:"prime_sudo_cache,omitempty"`
	// Shell to use for command execution (defaults to system shell)
	Shell string `yaml:"shell,omitempty"`
}

// ToolVersionCommand represents a command to check version or installation status
type ToolVersionCommand struct {
	// Command to check if tool is installed
	CheckCommand *ToolInstallCommand `yaml:"check_command,omitempty"`
	// Command to get current version
	VersionCommand *ToolInstallCommand `yaml:"version_command,omitempty"`
	// Whether version information is retrievable for this platform
	VersionRetrievable bool `yaml:"version_retrievable,omitempty"`
	// Custom version parser function (only available in Go code, not YAML)
	VersionParser func(string) string `yaml:"-"`
}

// ToolInstallConfig contains configuration for tool installation
type ToolInstallConfig struct {
	// Tool name for display purposes
	ToolName string `yaml:"tool_name"`
	// Default version to install (optional, can be used by commands)
	Version string `yaml:"version,omitempty"`
	// Force update even if tool is already installed
	ForceUpdate bool `yaml:"force_update,omitempty"`
	// Platform-specific installation commands
	// Keys should be in format: "os-arch" (e.g., "linux-amd64", "darwin-arm64", "windows-amd64")
	InstallCommands map[string]ToolInstallCommand `yaml:"install_commands"`
	// Platform-specific version checking commands
	VersionCommands map[string]ToolVersionCommand `yaml:"version_commands,omitempty"`
	// Post-install verification command (optional)
	VerifyCommand *ToolInstallCommand `yaml:"verify_command,omitempty"`
}

// ToolInstallResult represents the result of a tool installation or check
type ToolInstallResult struct {
	Success           bool
	AlreadyInstalled  bool
	Version           string
	VersionRetrievable bool
	Error             error
	Platform          string
}

// ToolInstaller provides cross-platform tool installation with custom scripts and commands
type ToolInstaller struct {
	ctx            context.Context
	outputChan     chan<- string
	inputChan      <-chan string
	sudoChecker    *SudoChecker
	inputRequester InputRequester
}

// NewToolInstaller creates a new tool installer instance
func NewToolInstaller(ctx context.Context, outputChan chan<- string, inputChan <-chan string) *ToolInstaller {
	return &ToolInstaller{
		ctx:            ctx,
		outputChan:     outputChan,
		inputChan:      inputChan,
		sudoChecker:    &SudoChecker{},
		inputRequester: NewInputRequester(ctx, outputChan, inputChan),
	}
}

// GetCurrentPlatform returns the current platform string in "os-arch" format
func (ti *ToolInstaller) GetCurrentPlatform() string {
	return fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)
}

// GetSupportedPlatforms returns a list of commonly supported platforms
func (ti *ToolInstaller) GetSupportedPlatforms() []string {
	return []string{
		"linux-amd64",
		"linux-arm64", 
		"darwin-amd64",
		"darwin-arm64",
		"windows-amd64",
		"windows-arm64",
	}
}

// CheckToolStatus checks if a tool is installed and gets version information
func (ti *ToolInstaller) CheckToolStatus(config *ToolInstallConfig) (*ToolInstallResult, error) {
	currentPlatform := ti.GetCurrentPlatform()
	
	result := &ToolInstallResult{
		Platform: currentPlatform,
	}

	// Check if version commands exist for current platform
	versionCmd, hasVersionCmd := config.VersionCommands[currentPlatform]
	if !hasVersionCmd {
		return nil, fmt.Errorf("no version commands available for platform %s", currentPlatform)
	}

	result.VersionRetrievable = versionCmd.VersionRetrievable

	// Check if tool is installed
	var installed bool
	if versionCmd.CheckCommand != nil {
		installed = ti.isToolInstalled(versionCmd.CheckCommand)
	} else if versionCmd.VersionCommand != nil {
		// If no CheckCommand, use the VersionCommand as a check
		installed = ti.isToolInstalled(versionCmd.VersionCommand)
	} else {
		// No way to check, assume not installed
		installed = false
	}
	
	if !installed {
		result.Success = true
		result.AlreadyInstalled = false
		return result, nil
	}
	
	result.AlreadyInstalled = true

	// Get version if retrievable and version command exists
	if result.VersionRetrievable && versionCmd.VersionCommand != nil {
		version, err := ti.getToolVersion(versionCmd.VersionCommand, versionCmd.VersionParser)
		if err != nil {
			// Don't fail if version retrieval fails, just log it
			ti.outputChan <- fmt.Sprintf("⚠️  Could not retrieve version: %v", err)
			result.Version = "unknown"
		} else {
			result.Version = version
		}
	} else if !result.VersionRetrievable {
		result.Version = "not retrievable"
	}

	result.Success = true
	return result, nil
}

// InstallTool performs cross-platform tool installation using custom commands
func (ti *ToolInstaller) InstallTool(config *ToolInstallConfig) (*ToolInstallResult, error) {
	currentPlatform := ti.GetCurrentPlatform()

	ti.outputChan <- fmt.Sprintf("🔧 %s Installation", strings.ToUpper(config.ToolName[:1])+config.ToolName[1:])
	ti.outputChan <- strings.Repeat("=", len(config.ToolName)+14)
	ti.outputChan <- ""

	ti.outputChan <- "📋 Installation Details:"
	ti.outputChan <- fmt.Sprintf("   🛠️  Tool: %s", config.ToolName)
	if config.Version != "" {
		ti.outputChan <- fmt.Sprintf("   🏷️  Version: %s", config.Version)
	}
	ti.outputChan <- fmt.Sprintf("   🖥️  Platform: %s", currentPlatform)
	ti.outputChan <- ""

	// Check if installation command exists for current platform
	installCmd, exists := config.InstallCommands[currentPlatform]
	if !exists {
		ti.outputChan <- fmt.Sprintf("❌ Platform not supported: %s", currentPlatform)
		ti.outputChan <- ""
		ti.outputChan <- "📝 Supported platforms:"
		for platform := range config.InstallCommands {
			ti.outputChan <- fmt.Sprintf("   • %s", platform)
		}
		return &ToolInstallResult{
			Success:  false,
			Platform: currentPlatform,
			Error:    fmt.Errorf("platform %s is not supported", currentPlatform),
		}, fmt.Errorf("platform %s is not supported", currentPlatform)
	}

	// Check current status
	ti.outputChan <- "🔍 Checking current installation status..."
	status, err := ti.CheckToolStatus(config)
	if err == nil && status.Success && status.AlreadyInstalled && !config.ForceUpdate {
		ti.outputChan <- fmt.Sprintf("✅ %s is already installed", config.ToolName)
		if status.VersionRetrievable && status.Version != "" {
			ti.outputChan <- fmt.Sprintf("📦 Current version: %s", status.Version)
		} else if !status.VersionRetrievable {
			ti.outputChan <- "📦 Version information is not retrievable for this platform"
		}
		
		// Run verification command if provided
		if config.VerifyCommand != nil {
			ti.outputChan <- "🔎 Verifying installation..."
			if err := ti.executeCommand(config.VerifyCommand, "Verification"); err != nil {
				ti.outputChan <- fmt.Sprintf("⚠️  Verification failed: %v", err)
			} else {
				ti.outputChan <- "✅ Installation verified successfully"
			}
		}
		return status, nil
	}

	if config.ForceUpdate && err == nil && status.Success && status.AlreadyInstalled {
		ti.outputChan <- fmt.Sprintf("🔄 %s is already installed, but forcing update as requested", config.ToolName)
		if status.VersionRetrievable && status.Version != "" {
			ti.outputChan <- fmt.Sprintf("📦 Current version: %s", status.Version)
		}
	} else if err == nil && (!status.Success || !status.AlreadyInstalled) {
		ti.outputChan <- fmt.Sprintf("📦 %s not found, proceeding with installation...", config.ToolName)
	} else {
		ti.outputChan <- fmt.Sprintf("⚠️  Could not check installation status: %v", err)
		ti.outputChan <- fmt.Sprintf("📦 Proceeding with %s installation...", config.ToolName)
	}
	ti.outputChan <- ""

	// Execute installation command
	ti.outputChan <- "🚀 Starting installation..."
	if err := ti.executeCommand(&installCmd, "Installation"); err != nil {
		result := &ToolInstallResult{
			Success:  false,
			Platform: currentPlatform,
			Error:    fmt.Errorf("installation failed: %w", err),
		}
		return result, fmt.Errorf("installation failed: %w", err)
	}

	// Check post-installation status
	postInstallStatus, err := ti.CheckToolStatus(config)
	if err == nil && postInstallStatus.Success && postInstallStatus.AlreadyInstalled {
		if postInstallStatus.VersionRetrievable && postInstallStatus.Version != "" {
			ti.outputChan <- fmt.Sprintf("📦 Installed version: %s", postInstallStatus.Version)
		}
	}

	// Run post-install verification if provided
	if config.VerifyCommand != nil {
		ti.outputChan <- ""
		ti.outputChan <- "🔎 Verifying installation..."
		if err := ti.executeCommand(config.VerifyCommand, "Verification"); err != nil {
			ti.outputChan <- fmt.Sprintf("⚠️  Installation completed but verification failed: %v", err)
		} else {
			ti.outputChan <- "✅ Installation verified successfully"
		}
	}

	ti.outputChan <- ""
	ti.outputChan <- fmt.Sprintf("🎉 %s installation completed successfully!", strings.ToUpper(config.ToolName[:1])+config.ToolName[1:])

	return &ToolInstallResult{
		Success:  true,
		Platform: currentPlatform,
	}, nil
}

// isToolInstalled checks if a tool is already installed using the check command
func (ti *ToolInstaller) isToolInstalled(checkCmd *ToolInstallCommand) bool {
	cmd := ti.buildCommand(checkCmd)
	
	// Run command but don't show output for check
	err := cmd.Run()
	return err == nil
}

// getToolVersion gets the current version of an installed tool
func (ti *ToolInstaller) getToolVersion(versionCmd *ToolInstallCommand, parser func(string) string) (string, error) {
	cmd := ti.buildCommand(versionCmd)
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("version command failed: %w", err)
	}

	versionStr := strings.TrimSpace(string(output))
	
	// Use custom parser if provided
	if parser != nil {
		versionStr = parser(versionStr)
	}

	return versionStr, nil
}

// executeCommand executes a tool install command with streaming output
func (ti *ToolInstaller) executeCommand(toolCmd *ToolInstallCommand, operation string) error {
	// Prime sudo cache if requested (but don't run the command as root)
	if toolCmd.PrimeSudoCache && runtime.GOOS != "windows" {
		if err := ti.primeSudoCacheOnly(); err != nil {
			return fmt.Errorf("failed to prime sudo cache: %w", err)
		}
	}

	cmd := ti.buildCommand(toolCmd)

	// Handle sudo password prompt if required (command will run as root)
	if toolCmd.RequiresSudo && runtime.GOOS != "windows" {
		return ti.executeSudoCommand(cmd, operation)
	}

	// Create pipes for real-time output streaming
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	// Stream output in real-time
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	// Stream stdout
	go ti.streamOutput(stdout, "")
	
	// Stream stderr  
	go ti.streamOutput(stderr, "")

	// Wait for command completion
	if err := <-done; err != nil {
		return fmt.Errorf("%s command failed: %w", operation, err)
	}

	return nil
}

// executeSudoCommand handles sudo authentication by priming the sudo cache, then runs the original command
func (ti *ToolInstaller) executeSudoCommand(cmd *exec.Cmd, operation string) error {
	// Check if we can run sudo without password prompt
	sudoInfo, err := ti.sudoChecker.CheckSudoPrivileges()
	if err == nil && sudoInfo.CanSudoWithoutPwd {
		// Execute directly without password prompt - sudo cache is already valid
		return ti.runCommandDirectly(cmd, operation)
	}

	// Get password from user using InputRequester
	password, err := ti.inputRequester.RequestPassword("Enter sudo password:")
	if err != nil {
		return fmt.Errorf("failed to get sudo password: %w", err)
	}

	// Prime the sudo cache using a dummy command with the provided password
	ti.outputChan <- "🔐 Authenticating with sudo..."
	primeCmd := exec.CommandContext(ti.ctx, "sudo", "-S", "echo", "sudo authentication successful")
	
	stdin, err := primeCmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe for sudo authentication: %w", err)
	}

	// Start the sudo prime command
	if err := primeCmd.Start(); err != nil {
		stdin.Close()
		return fmt.Errorf("failed to start sudo authentication: %w", err)
	}

	// Send password to authenticate
	go func() {
		defer stdin.Close()
		stdin.Write([]byte(password + "\n"))
	}()

	// Wait for authentication to complete
	if err := primeCmd.Wait(); err != nil {
		return fmt.Errorf("sudo authentication failed - please check your password: %w", err)
	}

	ti.outputChan <- "✅ Sudo authentication successful"

	// Now run the original command - sudo cache will be used for internal sudo calls
	return ti.runCommandDirectly(cmd, operation)
}

// runCommandDirectly executes a command directly with output streaming
func (ti *ToolInstaller) runCommandDirectly(cmd *exec.Cmd, operation string) error {
	// Set up output pipes for real-time streaming
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	// Stream outputs
	go ti.streamOutput(stdout, "")
	go ti.streamOutput(stderr, "")

	// Wait for command completion
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("%s command failed: %w", operation, err)
	}

	return nil
}

// primeSudoCacheOnly primes the sudo cache without running any subsequent commands as root
func (ti *ToolInstaller) primeSudoCacheOnly() error {
	// Check if we can run sudo without password prompt
	sudoInfo, err := ti.sudoChecker.CheckSudoPrivileges()
	if err == nil && sudoInfo.CanSudoWithoutPwd {
		// Sudo cache is already valid or no password required
		ti.outputChan <- "🔐 Sudo privileges available"
		return nil
	}

	// Get password from user using InputRequester
	password, err := ti.inputRequester.RequestPassword("Enter sudo password:")
	if err != nil {
		return fmt.Errorf("failed to get sudo password: %w", err)
	}

	// Prime the sudo cache using a dummy command with the provided password
	ti.outputChan <- "🔐 Authenticating with sudo..."
	primeCmd := exec.CommandContext(ti.ctx, "sudo", "-S", "echo", "Sudo access granted")
	
	stdin, err := primeCmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe for sudo authentication: %w", err)
	}

	// Start the sudo prime command
	if err := primeCmd.Start(); err != nil {
		stdin.Close()
		return fmt.Errorf("failed to start sudo authentication: %w", err)
	}

	// Send password to authenticate
	go func() {
		defer stdin.Close()
		stdin.Write([]byte(password + "\n"))
	}()

	// Wait for authentication to complete
	if err := primeCmd.Wait(); err != nil {
		return fmt.Errorf("sudo authentication failed - please check your password: %w", err)
	}

	ti.outputChan <- "✅ Sudo authentication successful"
	return nil
}

// buildCommand builds an exec.Cmd from a ToolInstallCommand
func (ti *ToolInstaller) buildCommand(toolCmd *ToolInstallCommand) *exec.Cmd {
	var cmd *exec.Cmd

	// Determine shell and command execution method
	shell := toolCmd.Shell
	if shell == "" {
		// Use system default shell
		switch runtime.GOOS {
		case "windows":
			shell = "cmd"
		default:
			shell = "bash"
		}
	}

	// Build command based on operating system and requirements
	if toolCmd.RequiresSudo && runtime.GOOS != "windows" {
		// Handle sudo on Unix systems
		if shell == "bash" || shell == "sh" {
			if len(toolCmd.Args) > 0 {
				// For shell execution with sudo, pass the first argument as the script content
				args := []string{shell, "-c", toolCmd.Args[0]}
				if len(toolCmd.Args) > 1 {
					// Additional arguments become script arguments ($1, $2, etc.)
					args = append(args, toolCmd.Args[1:]...)
				}
				cmd = exec.CommandContext(ti.ctx, "sudo", args...)
			} else {
				cmd = exec.CommandContext(ti.ctx, "sudo", shell, "-c", toolCmd.Command)
			}
		} else {
			// Direct sudo execution
			args := []string{toolCmd.Command}
			args = append(args, toolCmd.Args...)
			cmd = exec.CommandContext(ti.ctx, "sudo", args...)
		}
	} else if runtime.GOOS == "windows" {
		// Windows command execution
		if shell == "cmd" {
			args := []string{"/C", toolCmd.Command}
			args = append(args, toolCmd.Args...)
			cmd = exec.CommandContext(ti.ctx, "cmd", args...)
		} else if shell == "powershell" {
			fullCommand := toolCmd.Command
			if len(toolCmd.Args) > 0 {
				fullCommand += " " + strings.Join(toolCmd.Args, " ")
			}
			cmd = exec.CommandContext(ti.ctx, "powershell", "-Command", fullCommand)
		} else {
			// Direct execution
			cmd = exec.CommandContext(ti.ctx, toolCmd.Command, toolCmd.Args...)
		}
	} else {
		// Unix systems without sudo
		if shell == "bash" || shell == "sh" {
			if len(toolCmd.Args) > 0 {
				// For shell execution, pass the first argument as the script content
				// This handles multiline scripts properly with -c
				args := []string{"-c", toolCmd.Args[0]}
				if len(toolCmd.Args) > 1 {
					// Additional arguments become script arguments ($1, $2, etc.)
					args = append(args, toolCmd.Args[1:]...)
				}
				cmd = exec.CommandContext(ti.ctx, shell, args...)
			} else {
				cmd = exec.CommandContext(ti.ctx, shell, "-c", toolCmd.Command)
			}
		} else {
			// Direct execution
			cmd = exec.CommandContext(ti.ctx, toolCmd.Command, toolCmd.Args...)
		}
	}

	// Set working directory
	if toolCmd.WorkDir != "" {
		cmd.Dir = toolCmd.WorkDir
	}

	// Set environment variables
	if len(toolCmd.Env) > 0 {
		env := os.Environ()
		for key, value := range toolCmd.Env {
			env = append(env, fmt.Sprintf("%s=%s", key, value))
		}
		cmd.Env = env
	}

	return cmd
}

// streamOutput reads from a pipe and streams it to the output channel
func (ti *ToolInstaller) streamOutput(pipe io.ReadCloser, prefix string) {
	defer pipe.Close()
	
	buf := make([]byte, 1024)
	for {
		select {
		case <-ti.ctx.Done():
			return
		default:
			n, err := pipe.Read(buf)
			if n > 0 {
				output := strings.TrimSpace(string(buf[:n]))
				if output != "" {
					lines := strings.Split(output, "\n")
					for _, line := range lines {
						if strings.TrimSpace(line) != "" {
							ti.outputChan <- prefix + line
						}
					}
				}
			}
			if err != nil {
				if err != io.EOF {
					ti.outputChan <- fmt.Sprintf("Error reading output: %v", err)
				}
				return
			}
		}
	}
}