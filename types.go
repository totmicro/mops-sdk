// Package sdk provides the MOPS SDK for building plugins
package sdk

import (
	"context"
)

// PlatformInfo contains platform-specific information
type PlatformInfo struct {
	OS         string // "linux", "darwin", "windows"
	Arch       string // "amd64", "arm64"
	Platform   string // "linux-amd64", "darwin-arm64", etc.
	BinaryExt  string // ".exe" on Windows, "" on Unix
	ArchiveExt string // Expected archive extension: ".tar.gz", ".zip"
	InstallDir string // Default installation directory
	InPath     bool   // Whether install directory is in PATH
}

// ActionResult represents the result of executing an action
type ActionResult struct {
	Success     bool
	Output      string
	Error       error
	ShowOutput  bool
	Title       string
	IsStreaming bool
	RefreshMenu bool // Indicates if the parent menu should refresh its dynamic content
}

// InteractiveGoFunction represents a Go function that can be executed with real-time interaction
type InteractiveGoFunction func(ctx context.Context, outputChan chan<- string, inputChan <-chan string, params map[string]interface{}) error

// InputRequester provides methods for functions to explicitly request user input
type InputRequester interface {
	// RequestInput sends a prompt and waits for user input
	RequestInput(prompt string) (string, error)
	// RequestInputWithDefault sends a prompt with a default value
	RequestInputWithDefault(prompt string, defaultValue string) (string, error)
	// RequestPassword sends a prompt and waits for hidden password input
	RequestPassword(prompt string) (string, error)
}

// InteractiveGoFunction represents a Go function that can handle bidirectional communication

// DynamicProvider is an interface that provides dynamic menu entries
type DynamicProvider interface {
	GetName() string
	GetDescription() string
	GenerateEntries(param string) ([]MenuEntry, error)
	SupportsRefresh() bool
}

// AppState represents the application state
type AppState int

const (
	MainMenu AppState = iota
	DoneScreen
)

// NavMsg represents navigation messages
type NavMsg struct {
	Next          AppState
	NextID        string
	Data          any
	IsBack        bool
	Params        map[string]interface{}
	TitleOverride string
}

// MenuConfig represents the menu configuration
type MenuConfig struct {
	Menus []Menu `yaml:"menus"`
}

// Menu represents a menu definition
type Menu struct {
	ID            string          `yaml:"id"`
	Title         string          `yaml:"title"`
	Input         bool            `yaml:"input,omitempty"`
	Entries       []MenuEntry     `yaml:"entries,omitempty"`
	ConfirmAction *MenuAction     `yaml:"confirm_action,omitempty"`
	CancelKey     string          `yaml:"cancel_key,omitempty"`
	CancelNextID  string          `yaml:"cancel_next_id,omitempty"`
	Provider      string          `yaml:"provider,omitempty"`
	ProviderParam string          `yaml:"provider_param,omitempty"`
	Map           map[string]Menu `yaml:"-"`
	// Checkbox functionality
	IsCheckboxMenu bool        `yaml:"is_checkbox_menu,omitempty"` // Marks this menu as supporting checkboxes
	ExecuteAction  *MenuAction `yaml:"execute_action,omitempty"`   // Action to run with selected checkboxes
	ExecuteKey     string      `yaml:"execute_key,omitempty"`      // Key to trigger execute action (defaults to 'e')
}

// MenuEntry represents a menu entry
type MenuEntry struct {
	Key           string                 `yaml:"key"`
	Label         string                 `yaml:"label"`
	ID            string                 `yaml:"id,omitempty"`          // Unique identifier for CLI auto-execution
	Action        string                 `yaml:"action"`
	Target        string                 `yaml:"target,omitempty"`
	Message       string                 `yaml:"message,omitempty"`
	Command       string                 `yaml:"command,omitempty"`
	NextID        string                 `yaml:"next_id,omitempty"`
	FilePath      string                 `yaml:"file_path,omitempty"`
	Params        map[string]interface{} `yaml:"params,omitempty"`
	IsDynamic     bool                   `yaml:"-"`
	KeyBind       string                 `yaml:"key_bind,omitempty"`
	IsInteractive bool                   `yaml:"is_interactive,omitempty"`
	// Checkbox functionality
	IsCheckbox    bool   `yaml:"is_checkbox,omitempty"`    // Marks this entry as a checkbox
	IsChecked     bool   `yaml:"-"`                        // Runtime state - not persisted in YAML
	CheckboxGroup string `yaml:"checkbox_group,omitempty"` // Group name for related checkboxes
}

// MenuAction represents a menu action
type MenuAction struct {
	Action  string                 `yaml:"action"`
	Target  string                 `yaml:"target,omitempty"`
	Command string                 `yaml:"command,omitempty"`
	NextID  string                 `yaml:"next_id,omitempty"`
	Params  map[string]interface{} `yaml:"params,omitempty"`
}

// ToolInstallerInterface provides methods for cross-platform tool installation with custom commands
type ToolInstallerInterface interface {
	// InstallTool performs cross-platform tool installation using custom commands
	InstallTool(config *ToolInstallConfig) (*ToolInstallResult, error)
	// CheckToolStatus checks if a tool is installed and gets version information
	CheckToolStatus(config *ToolInstallConfig) (*ToolInstallResult, error)
	// GetCurrentPlatform returns the current platform string in "os-arch" format
	GetCurrentPlatform() string
	// GetSupportedPlatforms returns a list of commonly supported platforms
	GetSupportedPlatforms() []string
}
