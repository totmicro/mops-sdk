package sdk

import (
	"github.com/hashicorp/go-plugin"
)

// PluginBuilder provides a fluent interface for creating plugins
type PluginBuilder struct {
	info PluginInfo
	base *PluginBase
}

// NewPluginBuilder creates a new plugin builder
func NewPluginBuilder(name, version, description string) *PluginBuilder {
	info := PluginInfo{
		Name:           name,
		Version:        version,
		Description:    description,
		License:        "MIT",
		Tags:           []string{},
		Dependencies:   []string{},
		CLICommands:    []CLICommandInfo{},
		DefaultConfig:  make(map[string]any),
		Platform:       GetCurrentPlatform(),
		SupportedPlatforms: []PlatformInfo{
			{OS: "linux", Arch: "amd64"},
			{OS: "linux", Arch: "arm64"},
			{OS: "macos", Arch: "amd64"},
			{OS: "macos", Arch: "arm64"},
			{OS: "windows", Arch: "amd64"},
		},
	}
	
	base := NewPluginBase(info)
	
	return &PluginBuilder{
		info: info,
		base: base,
	}
}

// SetAuthor sets the plugin author
func (b *PluginBuilder) SetAuthor(author string) *PluginBuilder {
	b.info.Author = author
	b.base.info.Author = author
	return b
}

// SetLicense sets the plugin license
func (b *PluginBuilder) SetLicense(license string) *PluginBuilder {
	b.info.License = license
	b.base.info.License = license
	return b
}

// SetHomepage sets the plugin homepage
func (b *PluginBuilder) SetHomepage(homepage string) *PluginBuilder {
	b.info.Homepage = homepage
	b.base.info.Homepage = homepage
	return b
}

// SetMopsVersions sets the MOPS version constraints
func (b *PluginBuilder) SetMopsVersions(minVersion, maxVersion string) *PluginBuilder {
	b.info.MopsMinVersion = minVersion
	b.info.MopsMaxVersion = maxVersion
	b.base.info.MopsMinVersion = minVersion
	b.base.info.MopsMaxVersion = maxVersion
	return b
}

// AddTag adds a tag to the plugin
func (b *PluginBuilder) AddTag(tag string) *PluginBuilder {
	b.info.Tags = append(b.info.Tags, tag)
	b.base.info.Tags = append(b.base.info.Tags, tag)
	return b
}

// AddDependency adds a dependency to the plugin
func (b *PluginBuilder) AddDependency(dependency string) *PluginBuilder {
	b.info.Dependencies = append(b.info.Dependencies, dependency)
	b.base.info.Dependencies = append(b.base.info.Dependencies, dependency)
	return b
}

// SetDefaultConfig sets the default configuration
func (b *PluginBuilder) SetDefaultConfig(config map[string]any) *PluginBuilder {
	b.info.DefaultConfig = config
	b.base.info.DefaultConfig = config
	return b
}

// WithSimpleProvider adds a simple provider
func (b *PluginBuilder) WithSimpleProvider(name, description string, fn func(param string) ([]MenuEntry, error)) *PluginBuilder {
	b.base.WithSimpleProvider(name, fn)
	return b
}

// WithSimpleExecutor adds a simple executor
func (b *PluginBuilder) WithSimpleExecutor(actionType string, fn func(entry MenuEntry, input string) ActionResult) *PluginBuilder {
	b.base.WithSimpleExecutor(actionType, fn)
	return b
}

// WithInteractiveFunction adds an interactive function
func (b *PluginBuilder) WithInteractiveFunction(name string, fn InteractiveGoFunction) *PluginBuilder {
	b.base.WithInteractiveFunction(name, fn)
	return b
}

// WithCLICommand adds a CLI command
func (b *PluginBuilder) WithCLICommand(name, description string, handler CLICommandHandler) *PluginBuilder {
	cmdInfo := CLICommandInfo{
		Name:        name,
		Description: description,
	}
	b.info.CLICommands = append(b.info.CLICommands, cmdInfo)
	b.base.info.CLICommands = append(b.base.info.CLICommands, cmdInfo)
	b.base.WithCLICommand(name, handler)
	return b
}

// WithMenuEntry adds a menu entry
func (b *PluginBuilder) WithMenuEntry(menuID string, entry MenuEntry) *PluginBuilder {
	b.base.WithMenuEntry(menuID, entry)
	return b
}

// Build creates the plugin
func (b *PluginBuilder) Build() Plugin {
	return b.base
}

// Main is a convenience function to serve a plugin
func Main(pluginImpl Plugin) {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: plugin.HandshakeConfig{
			ProtocolVersion:  ProtocolVersion,
			MagicCookieKey:   "MOPS_PLUGIN",
			MagicCookieValue: "mops-plugin-interface",
		},
		Plugins: map[string]plugin.Plugin{
			PluginName: &MopsPlugin{Impl: pluginImpl},
		},
	})
}
