package sdk

import (
	"fmt"
	
	"github.com/hashicorp/go-plugin"
)

// PluginBuilder provides a fluent interface for creating plugins
type PluginBuilder struct {
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
		ConfigPresets:  make(map[string]ConfigPreset),
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
		base: base,
	}
}

// getInfo returns a reference to the plugin info from the base
func (b *PluginBuilder) getInfo() *PluginInfo {
	return &b.base.info
}

// SetAuthor sets the plugin author
func (b *PluginBuilder) SetAuthor(author string) *PluginBuilder {
	b.getInfo().Author = author
	return b
}

// SetDisplayName sets the plugin display name for UI
func (b *PluginBuilder) SetDisplayName(displayName string) *PluginBuilder {
	b.getInfo().DisplayName = displayName
	return b
}

// SetLicense sets the plugin license
func (b *PluginBuilder) SetLicense(license string) *PluginBuilder {
	b.getInfo().License = license
	return b
}

// SetHomepage sets the plugin homepage
func (b *PluginBuilder) SetHomepage(homepage string) *PluginBuilder {
	b.getInfo().Homepage = homepage
	return b
}

// SetMopsVersions sets the MOPS version constraints
func (b *PluginBuilder) SetMopsVersions(minVersion, maxVersion string) *PluginBuilder {
	info := b.getInfo()
	info.MopsMinVersion = minVersion
	info.MopsMaxVersion = maxVersion
	return b
}

// AddTag adds a tag to the plugin
func (b *PluginBuilder) AddTag(tag string) *PluginBuilder {
	b.getInfo().Tags = append(b.getInfo().Tags, tag)
	return b
}

// AddDependency adds a dependency to the plugin
func (b *PluginBuilder) AddDependency(dependency string) *PluginBuilder {
	b.getInfo().Dependencies = append(b.getInfo().Dependencies, dependency)
	return b
}

// SetDefaultConfig sets the default configuration
func (b *PluginBuilder) SetDefaultConfig(config map[string]any) *PluginBuilder {
	b.getInfo().DefaultConfig = config
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
	b.getInfo().CLICommands = append(b.getInfo().CLICommands, cmdInfo)
	b.base.WithCLICommand(name, handler)
	return b
}

// WithCLICommandAndUIMapping adds a CLI command with UI action mapping
func (b *PluginBuilder) WithCLICommandAndUIMapping(name, description string, handler CLICommandHandler, uiAction, uiTarget, uiCommand string) *PluginBuilder {
	cmdInfo := CLICommandInfo{
		Name:        name,
		Description: description,
		UIAction:    uiAction,
		UITarget:    uiTarget,
		UICommand:   uiCommand,
	}
	b.getInfo().CLICommands = append(b.getInfo().CLICommands, cmdInfo)
	b.base.WithCLICommand(name, handler)
	return b
}

// WithStreamingCLICommand adds a streaming CLI command that supports real-time output
func (b *PluginBuilder) WithStreamingCLICommand(name, description string, handler StreamingCLICommandHandler) *PluginBuilder {
	cmdInfo := CLICommandInfo{
		Name:        name,
		Description: description + " (streaming)",
	}
	b.getInfo().CLICommands = append(b.getInfo().CLICommands, cmdInfo)
	b.base.WithStreamingCLICommand(name, handler)
	return b
}

// WithMenuEntry adds a menu entry
func (b *PluginBuilder) WithMenuEntry(menuID string, entry MenuEntry) *PluginBuilder {
	b.base.WithMenuEntry(menuID, entry)
	return b
}

// WithMenuIntegration configures automatic menu integration
func (b *PluginBuilder) WithMenuIntegration(autoRegister bool, key, label, icon string, priority int) *PluginBuilder {
	b.getInfo().MenuIntegration = &PluginMenuIntegration{
		AutoRegister: autoRegister,
		MenuID:       "main",
		Key:          key,
		Label:        label,
		Icon:         icon,
		Priority:     priority,
		Group:        "plugins",
	}
	return b
}

// WithMenuIntegrationAdvanced configures automatic menu integration with all options
func (b *PluginBuilder) WithMenuIntegrationAdvanced(integration PluginMenuIntegration) *PluginBuilder {
	b.getInfo().MenuIntegration = &integration
	return b
}

// WithConfigPreset adds a configuration preset
func (b *PluginBuilder) WithConfigPreset(name, displayName, description string, config map[string]interface{}) *PluginBuilder {
	preset := ConfigPreset{
		Name:        displayName,
		Description: description,
		Config:      config,
	}
	b.getInfo().ConfigPresets[name] = preset
	return b
}

// Unified naming helper methods for standardized component registration

// getStandardName creates a standardized name with plugin prefix
func (b *PluginBuilder) getStandardName(componentName string) string {
	return fmt.Sprintf("%s-%s", b.getInfo().Name, componentName)
}

// WithStandardProvider adds a provider with standardized naming and type
func (b *PluginBuilder) WithStandardProvider(componentName, providerType, param, description string, providerFunc func(string) ([]MenuEntry, error)) *PluginBuilder {
	standardName := b.getStandardName(componentName)
	// Create a provider that includes type and param metadata
	provider := NewStandardProvider(standardName, providerType, param, description, providerFunc)
	b.base.providers = append(b.base.providers, provider)
	return b
}

// WithMenuProvider adds a menu provider with standardized naming (type: "menu")
func (b *PluginBuilder) WithMenuProvider(componentName, param, description string, providerFunc func(string) ([]MenuEntry, error)) *PluginBuilder {
	return b.WithStandardProvider(componentName, "menu", param, description, providerFunc)
}

// WithMainMenuProvider adds the main menu provider (type: "menu", param: "main")
func (b *PluginBuilder) WithMainMenuProvider(description string, providerFunc func(string) ([]MenuEntry, error)) *PluginBuilder {
	return b.WithMenuProvider("menu", "main", description, providerFunc)
}

// WithDataProvider adds a data provider with standardized naming (type: "data")
func (b *PluginBuilder) WithDataProvider(componentName, param, description string, providerFunc func(string) ([]MenuEntry, error)) *PluginBuilder {
	return b.WithStandardProvider(componentName, "data", param, description, providerFunc)
}

// WithConfigProvider adds a config provider with standardized naming (type: "config")
func (b *PluginBuilder) WithConfigProvider(componentName, param, description string, providerFunc func(string) ([]MenuEntry, error)) *PluginBuilder {
	return b.WithStandardProvider(componentName, "config", param, description, providerFunc)
}

// WithStandardExecutor adds an executor with standardized naming
func (b *PluginBuilder) WithStandardExecutor(actionName string, executorFunc func(MenuEntry, string) ActionResult) *PluginBuilder {
	standardName := b.getStandardName(actionName)
	executor := NewSimpleExecutor(standardName, executorFunc)
	b.base.executors = append(b.base.executors, executor)
	return b
}

// WithStandardInteractiveFunction adds an interactive function with standardized naming
func (b *PluginBuilder) WithStandardInteractiveFunction(functionName string, function InteractiveGoFunction) *PluginBuilder {
	standardName := b.getStandardName(functionName)
	b.base.interactiveFunctions[standardName] = function
	return b
}

// WithUnifiedProvider adds a single unified provider that handles multiple contexts
func (b *PluginBuilder) WithUnifiedProvider(description string) *UnifiedProviderBuilder {
	pluginName := b.getInfo().Name
	provider := NewUnifiedProvider(pluginName, description)
	
	return &UnifiedProviderBuilder{
		pluginBuilder: b,
		provider:      provider,
	}
}

// UnifiedProviderBuilder provides a fluent interface for building unified providers
type UnifiedProviderBuilder struct {
	pluginBuilder *PluginBuilder
	provider      *UnifiedProvider
}

// WithMainMenu adds the main menu handler (auto-registered)
func (upb *UnifiedProviderBuilder) WithMainMenu(handler func(param string) ([]MenuEntry, error)) *UnifiedProviderBuilder {
	upb.provider.WithMenuHandler("main", handler)
	return upb
}

// WithMenu adds a named menu handler
func (upb *UnifiedProviderBuilder) WithMenu(title string, handler func(param string) ([]MenuEntry, error)) *UnifiedProviderBuilder {
	upb.provider.WithMenuHandler(title, handler)
	return upb
}

// WithData adds a data handler
func (upb *UnifiedProviderBuilder) WithData(title string, handler func(param string) ([]MenuEntry, error)) *UnifiedProviderBuilder {
	upb.provider.WithDataHandler(title, handler)
	return upb
}

// WithConfig adds a config handler
func (upb *UnifiedProviderBuilder) WithConfig(title string, handler func(param string) ([]MenuEntry, error)) *UnifiedProviderBuilder {
	upb.provider.WithConfigHandler(title, handler)
	return upb
}

// Done completes the unified provider and returns to the main plugin builder
func (upb *UnifiedProviderBuilder) Done() *PluginBuilder {
	upb.pluginBuilder.base.providers = append(upb.pluginBuilder.base.providers, upb.provider)
	return upb.pluginBuilder
}

// Convenience methods for common patterns

// WithMainMenuAndExecutor adds both a main menu provider and corresponding executors
func (b *PluginBuilder) WithMainMenuAndExecutor(description string, providerFunc func(string) ([]MenuEntry, error), executors map[string]func(MenuEntry, string) ActionResult) *PluginBuilder {
	// Add main menu provider
	b.WithMainMenuProvider(description, providerFunc)
	
	// Add executors with standard naming
	for actionName, executorFunc := range executors {
		b.WithStandardExecutor(actionName, executorFunc)
	}
	
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
