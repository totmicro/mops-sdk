package sdk

import (
	"context"
	"time"
)

// PlatformInfo represents platform information for plugin distribution
type PlatformInfo struct {
	OS   string `json:"os"`
	Arch string `json:"arch"`
}

// ConfigPreset represents a plugin configuration preset
type ConfigPreset struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Config      map[string]interface{} `json:"config"`
}

// PluginInfo contains metadata about a plugin
type PluginInfo struct {
	Name               string                  `json:"name"`
	Version            string                  `json:"version"`
	Description        string                  `json:"description"`
	DisplayName        string                  `json:"display_name,omitempty"` // Optional display name for UI
	Author             string                  `json:"author"`
	License            string                  `json:"license"`
	Homepage           string                  `json:"homepage"`
	MopsMinVersion     string                  `json:"mops_min_version"`
	MopsMaxVersion     string                  `json:"mops_max_version"`
	Dependencies       []string                `json:"dependencies"`
	Tags               []string                `json:"tags"`
	CLICommands        []CLICommandInfo        `json:"cli_commands"`
	DefaultConfig      map[string]any          `json:"default_config"`
	ConfigPresets      map[string]ConfigPreset `json:"config_presets,omitempty"`
	Platform           PlatformInfo            `json:"platform"`
	SupportedPlatforms []PlatformInfo          `json:"supported_platforms"`
	MenuIntegration    *PluginMenuIntegration  `json:"menu_integration,omitempty"`
}

// CLICommandInfo describes a CLI command provided by the plugin
type CLICommandInfo struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Usage       string   `json:"usage"`
	Examples    []string `json:"examples"`
	Hidden      bool     `json:"hidden,omitempty"` // Hide this command from CLI help and prevent execution with args

	// UI Action mapping - allows CLI commands to specify their UI equivalent
	UIAction  string `json:"ui_action,omitempty"`  // The UI action type (e.g., "goto", "core_interactive-go", "action")
	UITarget  string `json:"ui_target,omitempty"`  // For goto actions - the target menu/provider
	UICommand string `json:"ui_command,omitempty"` // For core_interactive-go actions - the command to execute
	UITitle   string `json:"ui_title,omitempty"`   // Custom title for Bubble Tea UI when executed via CLI
}

// CLICommandHandler handles CLI command execution
type CLICommandHandler func(args []string) error

// StreamingCLICommandHandler extends CLICommandHandler to support real-time output streaming
type StreamingCLICommandHandler interface {
	// Execute runs the CLI command with given arguments (fallback method)
	Execute(ctx context.Context, args []string) error

	// GetHelp returns help information for the command
	GetHelp() string

	// ExecuteStreaming runs the CLI command with real-time output streaming
	// outputChan receives output lines as they are generated
	// The channel is closed when the command completes
	ExecuteStreaming(ctx context.Context, args []string, outputChan chan<- string) error

	// SupportsStreaming indicates if this command supports real-time streaming
	SupportsStreaming() bool
}

// Plugin is the interface that all plugins must implement
type Plugin interface {
	GetInfo() PluginInfo
	Initialize(config map[string]any) error
	RegisterProviders() []DynamicProvider
	RegisterInteractiveFunctions() map[string]InteractiveGoFunction
	GetCLICommands() (map[string]CLICommandHandler, error)
	GetStreamingCLICommands() (map[string]StreamingCLICommandHandler, error)
	GetMenuEntries() (map[string][]MenuEntry, error)
	Cleanup() error
	ValidateConfig(config map[string]any) error
}

// Registry interface for managing providers
type Registry interface {
	RegisterProvider(provider DynamicProvider)
	GetProvider(name string) (DynamicProvider, bool)
	ListProviders() []string
}

// PluginMenuIntegration defines how a plugin integrates with MOPS menus
type PluginMenuIntegration struct {
	AutoRegister bool   `yaml:"auto_register"` // Whether to automatically add to main menu
	MenuID       string `yaml:"menu_id"`       // Which menu to integrate with (default: "main")
	Key          string `yaml:"key"`           // Shortcut key for the menu entry
	Label        string `yaml:"label"`         // Display label for the menu entry
	Icon         string `yaml:"icon"`          // Optional icon/emoji for the menu entry
	Priority     int    `yaml:"priority"`      // Priority for ordering (lower = higher priority)
	Group        string `yaml:"group"`         // Optional grouping for organizing menu items
}

// PluginMetadata represents the plugin.yaml file structure
type PluginMetadata struct {
	Name               string                         `yaml:"name"`
	Version            string                         `yaml:"version"`
	Description        string                         `yaml:"description"`
	DisplayName        string                         `yaml:"display_name,omitempty"`
	Author             string                         `yaml:"author"`
	License            string                         `yaml:"license"`
	Homepage           string                         `yaml:"homepage"`
	Repository         string                         `yaml:"repository"`
	Category           string                         `yaml:"category"`
	Tags               []string                       `yaml:"tags"`
	MinimumMopsVersion string                         `yaml:"minimum_mops_version"`
	MopsVersion        MopsVersionConstraint          `yaml:"mops_version"`
	BuildTargets       []BuildTarget                  `yaml:"build_targets"`
	DefaultConfig      map[string]any                 `yaml:"default_config"`
	ConfigPresets      map[string]ConfigPreset        `yaml:"config_presets,omitempty"`
	CLICommands        []CLICommandInfo               `yaml:"cli_commands"`
	MenuIntegration    *PluginMenuIntegration         `yaml:"menu_integration,omitempty"`
}

// MopsVersionConstraint represents MOPS version compatibility
type MopsVersionConstraint struct {
	MinVersion string `yaml:"min_version"`
	MaxVersion string `yaml:"max_version"`
}

// BuildTarget represents a build target configuration
type BuildTarget struct {
	OS     string `yaml:"os"`
	Arch   string `yaml:"arch"`
	Output string `yaml:"output"`
}

// RepositoryRegistry represents a registry of all plugins in a repository
type RepositoryRegistry struct {
	Repository struct {
		Name        string    `json:"name"`
		Description string    `json:"description"`
		URL         string    `json:"url"`
		LastUpdated time.Time `json:"last_updated"`
		Version     string    `json:"version"`
	} `json:"repository"`
	Plugins map[string]RepositoryPlugin `json:"plugins"`
}

// RepositoryPlugin represents a plugin entry in a repository registry
type RepositoryPlugin struct {
	Name        string                           `json:"name"`
	Version     string                           `json:"version"`
	Description string                           `json:"description"`
	Author      string                           `json:"author"`
	License     string                           `json:"license"`
	Homepage    string                           `json:"homepage"`
	Repository  string                           `json:"repository"`
	Tags        []string                         `json:"tags"`
	MopsVersion MopsVersionConstraint            `json:"mops_version"`
	Platforms   map[string]RepositoryPluginAsset `json:"platforms"`
	Checksums   map[string]string                `json:"checksums"`
	LastUpdated time.Time                        `json:"last_updated"`
}

// RepositoryPluginAsset represents a platform-specific plugin asset
type RepositoryPluginAsset struct {
	Platform    PlatformInfo `json:"platform"`
	DownloadURL string       `json:"download_url"`
	Size        int64        `json:"size"`
	Checksum    string       `json:"checksum"`
	Filename    string       `json:"filename"`
}
