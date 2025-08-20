package sdk

import (
	"time"
)

// PlatformInfo represents platform information for plugin distribution
type PlatformInfo struct {
	OS   string `json:"os"`
	Arch string `json:"arch"`
}

// PluginInfo contains metadata about a plugin
type PluginInfo struct {
	Name              string            `json:"name"`
	Version           string            `json:"version"`
	Description       string            `json:"description"`
	Author            string            `json:"author"`
	License           string            `json:"license"`
	Homepage          string            `json:"homepage"`
	MopsMinVersion    string            `json:"mops_min_version"`
	MopsMaxVersion    string            `json:"mops_max_version"`
	Dependencies      []string          `json:"dependencies"`
	Tags              []string          `json:"tags"`
	CLICommands       []CLICommandInfo  `json:"cli_commands"`
	DefaultConfig     map[string]any    `json:"default_config"`
	Platform          PlatformInfo      `json:"platform"`
	SupportedPlatforms []PlatformInfo   `json:"supported_platforms"`
}

// CLICommandInfo describes a CLI command provided by the plugin
type CLICommandInfo struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Usage       string   `json:"usage"`
	Examples    []string `json:"examples"`
}

// CLICommandHandler handles CLI command execution
type CLICommandHandler func(args []string) error

// Plugin is the interface that all plugins must implement
type Plugin interface {
	GetInfo() PluginInfo
	Initialize(config map[string]any) error
	RegisterProviders() []DynamicProvider
	RegisterExecutors() []ActionExecutor
	RegisterInteractiveFunctions() map[string]InteractiveGoFunction
	GetCLICommands() (map[string]CLICommandHandler, error)
	GetMenuEntries() (map[string][]MenuEntry, error)
	Cleanup() error
	ValidateConfig(config map[string]any) error
}

// Registry interface for managing providers and executors
type Registry interface {
	RegisterProvider(provider DynamicProvider)
	RegisterExecutor(executor ActionExecutor)
	GetProvider(name string) (DynamicProvider, bool)
	GetExecutor(actionType string) (ActionExecutor, bool)
	ListProviders() []string
	ListExecutors() []string
}

// PluginMetadata represents the plugin.yaml file structure
type PluginMetadata struct {
	Name         string                `yaml:"name"`
	Version      string                `yaml:"version"`
	Description  string                `yaml:"description"`
	Author       string                `yaml:"author"`
	License      string                `yaml:"license"`
	Homepage     string                `yaml:"homepage"`
	Repository   string                `yaml:"repository"`
	Tags         []string              `yaml:"tags"`
	MopsVersion  MopsVersionConstraint `yaml:"mops_version"`
	BuildTargets []BuildTarget         `yaml:"build_targets"`
	DefaultConfig map[string]any       `yaml:"default_config"`
	CLICommands  []CLICommandInfo      `yaml:"cli_commands"`
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
	Name         string                            `json:"name"`
	Version      string                            `json:"version"`
	Description  string                            `json:"description"`
	Author       string                            `json:"author"`
	License      string                            `json:"license"`
	Homepage     string                            `json:"homepage"`
	Repository   string                            `json:"repository"`
	Tags         []string                          `json:"tags"`
	MopsVersion  MopsVersionConstraint             `json:"mops_version"`
	Platforms    map[string]RepositoryPluginAsset  `json:"platforms"`
	Checksums    map[string]string                 `json:"checksums"`
	LastUpdated  time.Time                         `json:"last_updated"`
}

// RepositoryPluginAsset represents a platform-specific plugin asset
type RepositoryPluginAsset struct {
	Platform    PlatformInfo `json:"platform"`
	DownloadURL string       `json:"download_url"`
	Size        int64        `json:"size"`
	Checksum    string       `json:"checksum"`
	Filename    string       `json:"filename"`
}
