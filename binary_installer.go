package sdk

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// BinaryInstaller provides cross-platform binary installation capabilities
type BinaryInstaller struct {
	ctx              context.Context
	outputChan       chan<- string
	inputChan        <-chan string
	inputRequester   InputRequester
	downloadTimeout  time.Duration
}

// BinaryInstallConfig contains configuration for binary installation
type BinaryInstallConfig struct {
	BinaryName       string            // e.g., "saml2aws"
	Version          string            // e.g., "2.36.19" or "latest"
	DownloadURLs     map[string]string // Platform-specific URLs: "linux-amd64" -> "https://..."
	InstallPath      string            // Optional custom install path
	RequiresSudo     bool              // Whether installation requires sudo (Unix only)
	VersionExtractor func(string) string // Optional: extract version from URL or other source
}

// PlatformInfo contains platform-specific information
type PlatformInfo struct {
	OS           string // "linux", "darwin", "windows"
	Arch         string // "amd64", "arm64"
	Platform     string // "linux-amd64", "darwin-arm64", etc.
	BinaryExt    string // ".exe" on Windows, "" on Unix
	ArchiveExt   string // Expected archive extension: ".tar.gz", ".zip"
	InstallDir   string // Default installation directory
	InPath       bool   // Whether install directory is in PATH
}

// NewBinaryInstaller creates a new binary installer instance
func NewBinaryInstaller(ctx context.Context, outputChan chan<- string, inputChan <-chan string) *BinaryInstaller {
	return &BinaryInstaller{
		ctx:             ctx,
		outputChan:      outputChan,
		inputChan:       inputChan,
		inputRequester:  NewInputRequester(ctx, outputChan, inputChan),
		downloadTimeout: 5 * time.Minute,
	}
}

// SetDownloadTimeout sets the timeout for download operations
func (bi *BinaryInstaller) SetDownloadTimeout(timeout time.Duration) {
	bi.downloadTimeout = timeout
}

// GetPlatformInfo returns information about the current platform
func (bi *BinaryInstaller) GetPlatformInfo() *PlatformInfo {
	osType := runtime.GOOS
	arch := runtime.GOARCH
	platform := fmt.Sprintf("%s-%s", osType, arch)
	
	info := &PlatformInfo{
		OS:       osType,
		Arch:     arch,
		Platform: platform,
	}
	
	// Set platform-specific defaults
	switch osType {
	case "windows":
		info.BinaryExt = ".exe"
		info.ArchiveExt = ".zip"
		info.InstallDir = "C:\\Program Files\\MOPS\\bin"
		info.InPath = false
	case "linux":
		info.BinaryExt = ""
		info.ArchiveExt = ".tar.gz"
		info.InstallDir = "/usr/local/bin"
		info.InPath = true
	case "darwin":
		info.BinaryExt = ""
		info.ArchiveExt = ".tar.gz"
		info.InstallDir = "/usr/local/bin"
		info.InPath = true
	}
	
	return info
}

// InstallBinary performs cross-platform binary installation
func (bi *BinaryInstaller) InstallBinary(config *BinaryInstallConfig) error {
	platform := bi.GetPlatformInfo()
	
	bi.outputChan <- fmt.Sprintf("📥 %s Installation", strings.Title(config.BinaryName))
	bi.outputChan <- strings.Repeat("=", len(config.BinaryName)+14)
	bi.outputChan <- ""
	
	// Handle version resolution
	version := config.Version
	if version == "latest" && config.VersionExtractor != nil {
		if downloadURL, exists := config.DownloadURLs[platform.Platform]; exists {
			version = config.VersionExtractor(downloadURL)
		}
		if version == "latest" {
			return fmt.Errorf("could not determine latest version")
		}
	}
	
	bi.outputChan <- "🔍 Installation Details:"
	bi.outputChan <- fmt.Sprintf("   📦 Binary: %s", config.BinaryName)
	bi.outputChan <- fmt.Sprintf("   🏷️  Version: %s", version)
	bi.outputChan <- fmt.Sprintf("   🖥️  OS: %s", platform.OS)
	bi.outputChan <- fmt.Sprintf("   🏗️  Architecture: %s", platform.Arch)
	bi.outputChan <- ""
	
	// Check if download URL exists for this platform
	downloadURL, exists := config.DownloadURLs[platform.Platform]
	if !exists {
		return fmt.Errorf("no download URL available for platform %s", platform.Platform)
	}
	
	// Handle Windows separately (manual installation)
	if platform.OS == "windows" {
		return bi.provideWindowsInstructions(config, downloadURL, platform)
	}
	
	// Automatic installation for Unix systems
	return bi.performAutomaticInstallation(config, downloadURL, version, platform)
}

// provideWindowsInstructions provides manual installation instructions for Windows
func (bi *BinaryInstaller) provideWindowsInstructions(config *BinaryInstallConfig, downloadURL string, platform *PlatformInfo) error {
	bi.outputChan <- "🪟 Windows Manual Installation Required"
	bi.outputChan <- "======================================"
	bi.outputChan <- ""
	bi.outputChan <- "📋 Manual Installation Steps:"
	bi.outputChan <- ""
	bi.outputChan <- "1. 📥 Download the Windows binary:"
	bi.outputChan <- fmt.Sprintf("   %s", downloadURL)
	bi.outputChan <- ""
	bi.outputChan <- "2. 📂 Extract the archive to a temporary location"
	bi.outputChan <- ""
	bi.outputChan <- fmt.Sprintf("3. 📁 Copy %s%s to one of these locations:", config.BinaryName, platform.BinaryExt)
	bi.outputChan <- fmt.Sprintf("   • %s\\%s%s", platform.InstallDir, config.BinaryName, platform.BinaryExt)
	bi.outputChan <- fmt.Sprintf("   • C:\\Windows\\System32\\%s%s", config.BinaryName, platform.BinaryExt)
	bi.outputChan <- "   • Any directory in your PATH environment variable"
	bi.outputChan <- ""
	
	if !platform.InPath {
		bi.outputChan <- "4. 🔧 Add to PATH (if not using System32):"
		bi.outputChan <- "   • Open System Properties → Advanced → Environment Variables"
		bi.outputChan <- fmt.Sprintf("   • Add %s to your PATH", platform.InstallDir)
		bi.outputChan <- ""
	}
	
	bi.outputChan <- "5. ✅ Verify installation:"
	bi.outputChan <- "   • Open Command Prompt or PowerShell"
	bi.outputChan <- fmt.Sprintf("   • Run: %s --version", config.BinaryName)
	bi.outputChan <- ""
	bi.outputChan <- fmt.Sprintf("🚀 %s is ready to use after manual installation!", strings.Title(config.BinaryName))
	
	return nil
}

// performAutomaticInstallation handles automatic installation on Unix systems
func (bi *BinaryInstaller) performAutomaticInstallation(config *BinaryInstallConfig, downloadURL, version string, platform *PlatformInfo) error {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", fmt.Sprintf("%s-install", config.BinaryName))
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Download
	bi.outputChan <- "📦 Downloading binary..."
	archivePath, err := bi.downloadBinary(downloadURL, tempDir, platform)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	
	// Extract
	bi.outputChan <- "📂 Extracting archive..."
	binaryPath, err := bi.extractArchive(archivePath, tempDir, config.BinaryName, platform)
	if err != nil {
		return fmt.Errorf("extraction failed: %w", err)
	}
	
	// Install
	bi.outputChan <- "🔧 Installing binary..."
	if err := bi.installBinary(binaryPath, config, platform); err != nil {
		return fmt.Errorf("installation failed: %w", err)
	}
	
	bi.outputChan <- ""
	bi.outputChan <- fmt.Sprintf("✅ %s installation completed successfully!", strings.Title(config.BinaryName))
	
	return nil
}

// downloadBinary downloads the binary archive
func (bi *BinaryInstaller) downloadBinary(downloadURL, tempDir string, platform *PlatformInfo) (string, error) {
	// Determine archive filename from URL
	urlParts := strings.Split(downloadURL, "/")
	filename := urlParts[len(urlParts)-1]
	archivePath := filepath.Join(tempDir, filename)
	
	bi.outputChan <- fmt.Sprintf("📥 Downloading from: %s", downloadURL)
	
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: bi.downloadTimeout,
	}
	
	// Download
	resp, err := client.Get(downloadURL)
	if err != nil {
		return "", fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}
	
	// Save to file
	file, err := os.Create(archivePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()
	
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to save file: %w", err)
	}
	
	return archivePath, nil
}

// extractArchive extracts the downloaded archive and finds the binary
func (bi *BinaryInstaller) extractArchive(archivePath, tempDir, binaryName string, platform *PlatformInfo) (string, error) {
	extractDir := filepath.Join(tempDir, "extracted")
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create extract directory: %w", err)
	}
	
	// Extract based on file extension
	if strings.HasSuffix(archivePath, ".tar.gz") {
		if err := bi.extractTarGz(archivePath, extractDir); err != nil {
			return "", fmt.Errorf("failed to extract tar.gz: %w", err)
		}
	} else if strings.HasSuffix(archivePath, ".zip") {
		if err := bi.extractZip(archivePath, extractDir); err != nil {
			return "", fmt.Errorf("failed to extract zip: %w", err)
		}
	} else {
		return "", fmt.Errorf("unsupported archive format: %s", filepath.Ext(archivePath))
	}
	
	// Find the binary
	binaryFilename := binaryName + platform.BinaryExt
	binaryPath, err := bi.findBinary(extractDir, binaryFilename)
	if err != nil {
		return "", fmt.Errorf("binary not found after extraction: %w", err)
	}
	
	return binaryPath, nil
}

// extractTarGz extracts a .tar.gz archive
func (bi *BinaryInstaller) extractTarGz(archivePath, extractDir string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()
	
	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzr.Close()
	
	tr := tar.NewReader(gzr)
	
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		
		target := filepath.Join(extractDir, header.Name)
		
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return err
			}
			f.Close()
		}
	}
	
	return nil
}

// extractZip extracts a .zip archive
func (bi *BinaryInstaller) extractZip(archivePath, extractDir string) error {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer r.Close()
	
	for _, f := range r.File {
		target := filepath.Join(extractDir, f.Name)
		
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, f.FileInfo().Mode()); err != nil {
				return err
			}
			continue
		}
		
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}
		
		rc, err := f.Open()
		if err != nil {
			return err
		}
		
		outFile, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.FileInfo().Mode())
		if err != nil {
			rc.Close()
			return err
		}
		
		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		
		if err != nil {
			return err
		}
	}
	
	return nil
}

// findBinary searches for the binary file in the extracted directory
func (bi *BinaryInstaller) findBinary(baseDir, binaryName string) (string, error) {
	var foundPath string
	
	err := filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if !info.IsDir() && info.Name() == binaryName {
			foundPath = path
			return fmt.Errorf("found") // Use error to break out of walk
		}
		
		return nil
	})
	
	if err != nil && err.Error() == "found" {
		return foundPath, nil
	}
	
	if foundPath == "" {
		return "", fmt.Errorf("binary %s not found in extracted files", binaryName)
	}
	
	return foundPath, nil
}

// installBinary installs the binary to the target location
func (bi *BinaryInstaller) installBinary(binaryPath string, config *BinaryInstallConfig, platform *PlatformInfo) error {
	// Determine target path
	targetPath := config.InstallPath
	if targetPath == "" {
		targetPath = filepath.Join(platform.InstallDir, config.BinaryName+platform.BinaryExt)
	}
	
	// Handle sudo requirement
	if config.RequiresSudo {
		// Use the new sudo checker to intelligently handle sudo
		sudoChecker := NewSudoChecker()
		sudoInfo, err := sudoChecker.CheckSudoPrivileges()
		if err != nil {
			return fmt.Errorf("failed to check sudo privileges: %w", err)
		}
		
		if !sudoInfo.HasSudo {
			return fmt.Errorf("sudo is required for installation but not available - please run with sudo")
		}
		
		bi.outputChan <- "🔒 Checking system privileges..."
		
		if sudoInfo.IsRoot {
			bi.outputChan <- "✅ Running as root - proceeding with installation"
			// Copy directly without sudo
			cmd := exec.CommandContext(bi.ctx, "cp", binaryPath, targetPath)
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to copy binary: %w", err)
			}
			
			// Make executable
			cmd = exec.CommandContext(bi.ctx, "chmod", "+x", targetPath)
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to make binary executable: %w", err)
			}
		} else if sudoInfo.CanSudoWithoutPwd || sudoInfo.SudoTimeLeft > 0 {
			if sudoInfo.CanSudoWithoutPwd {
				bi.outputChan <- "✅ Sudo available (passwordless) - proceeding with installation"
			} else {
				bi.outputChan <- fmt.Sprintf("✅ Sudo available (cached for %d minutes) - proceeding with installation", sudoInfo.SudoTimeLeft)
			}
			
			// Copy with sudo (no password needed)
			cmd := exec.CommandContext(bi.ctx, "sudo", "cp", binaryPath, targetPath)
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to copy binary with sudo: %w", err)
			}
			
			// Make executable with sudo
			cmd = exec.CommandContext(bi.ctx, "sudo", "chmod", "+x", targetPath)
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to make binary executable with sudo: %w", err)
			}
		} else {
			// Need to prompt for password
			bi.outputChan <- "⚠️  Sudo password required for installation"
			password, err := bi.inputRequester.RequestPassword("Enter sudo password: ")
			if err != nil {
				return fmt.Errorf("failed to get sudo password: %w", err)
			}
			
			// Copy with sudo and password
			cmd := exec.CommandContext(bi.ctx, "bash", "-c", 
				fmt.Sprintf("echo '%s' | sudo -S cp '%s' '%s'", password, binaryPath, targetPath))
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to copy binary with sudo: %w", err)
			}
			
			// Make executable with sudo
			cmd = exec.CommandContext(bi.ctx, "bash", "-c",
				fmt.Sprintf("echo '%s' | sudo -S chmod +x '%s'", password, targetPath))
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to make binary executable with sudo: %w", err)
			}
		}
	} else {
		// Direct copy (no sudo)
		if err := bi.copyFile(binaryPath, targetPath); err != nil {
			return fmt.Errorf("failed to copy binary: %w", err)
		}
		
		// Make executable
		if err := os.Chmod(targetPath, 0755); err != nil {
			return fmt.Errorf("failed to make binary executable: %w", err)
		}
	}
	
	return nil
}

// copyFile copies a file from src to dst
func (bi *BinaryInstaller) copyFile(src, dst string) error {
	// Ensure target directory exists
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()
	
	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()
	
	_, err = io.Copy(destFile, sourceFile)
	return err
}

// CheckBinaryInstalled checks if a binary is installed and returns version info
func (bi *BinaryInstaller) CheckBinaryInstalled(binaryName string) (bool, string, error) {
	cmd := exec.CommandContext(bi.ctx, binaryName, "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, "", nil // Binary not installed or not in PATH
	}
	
	version := strings.TrimSpace(string(output))
	return true, version, nil
}
