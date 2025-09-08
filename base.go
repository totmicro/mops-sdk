package sdk

import (
	"fmt"
	"strings"
)

// PluginBase provides a base implementation for mops plugins
type PluginBase struct {
	info                        PluginInfo
	config                      map[string]any
	providers                   []DynamicProvider
	interactiveFunctions        map[string]InteractiveGoFunction
	enhancedInteractiveFunctions map[string]EnhancedInteractiveFunction
	cliCommands                 map[string]CLICommandHandler
	streamingCLICommands        map[string]StreamingCLICommandHandler
	menuEntries                 map[string][]MenuEntry
}

// NewPluginBase creates a new plugin base
func NewPluginBase(info PluginInfo) *PluginBase {
	return &PluginBase{
		info:                        info,
		providers:                   []DynamicProvider{},
		interactiveFunctions:        make(map[string]InteractiveGoFunction),
		enhancedInteractiveFunctions: make(map[string]EnhancedInteractiveFunction),
		cliCommands:                 make(map[string]CLICommandHandler),
		streamingCLICommands:        make(map[string]StreamingCLICommandHandler),
		menuEntries:                 make(map[string][]MenuEntry),
	}
}

// GetInfo returns metadata about the plugin
func (p *PluginBase) GetInfo() PluginInfo {
	return p.info
}

// Initialize sets up the plugin with the provided config
func (p *PluginBase) Initialize(config map[string]any) error {
	p.config = config
	return nil
}

// RegisterProviders registers dynamic providers with mops
func (p *PluginBase) RegisterProviders() []DynamicProvider {
	return p.providers
}

// RegisterInteractiveFunctions registers interactive functions with mops
func (p *PluginBase) RegisterInteractiveFunctions() map[string]InteractiveGoFunction {
	return p.interactiveFunctions
}

// RegisterEnhancedInteractiveFunctions registers enhanced interactive functions with mops
func (p *PluginBase) RegisterEnhancedInteractiveFunctions() map[string]EnhancedInteractiveFunction {
	return p.enhancedInteractiveFunctions
}

// AddEnhancedInteractiveFunction adds an enhanced interactive function with explicit input request capabilities
func (p *PluginBase) AddEnhancedInteractiveFunction(name string, fn EnhancedInteractiveFunction) {
	p.enhancedInteractiveFunctions[name] = fn
}

// GetCLICommands returns CLI command handlers
func (p *PluginBase) GetCLICommands() (map[string]CLICommandHandler, error) {
	return p.cliCommands, nil
}

// GetMenuEntries returns menu entries that should be added to menus
func (p *PluginBase) GetMenuEntries() (map[string][]MenuEntry, error) {
	return p.menuEntries, nil
}

// Cleanup performs cleanup when the plugin is being unloaded
func (p *PluginBase) Cleanup() error {
	return nil
}

// ValidateConfig validates the plugin configuration
func (p *PluginBase) ValidateConfig(config map[string]any) error {
	return nil
}

// Helper methods for plugin developers

// AddProvider adds a dynamic provider to the plugin
func (p *PluginBase) AddProvider(provider DynamicProvider) {
	p.providers = append(p.providers, provider)
}

// AddInteractiveFunction adds an interactive function to the plugin
func (p *PluginBase) AddInteractiveFunction(name string, fn InteractiveGoFunction) {
	p.interactiveFunctions[name] = fn
}

// AddCLICommand adds a CLI command to the plugin
func (p *PluginBase) AddCLICommand(name string, handler CLICommandHandler) {
	p.cliCommands[name] = handler
}

// AddStreamingCLICommand adds a streaming CLI command to the plugin
func (p *PluginBase) AddStreamingCLICommand(name string, handler StreamingCLICommandHandler) {
	p.streamingCLICommands[name] = handler
}

// GetStreamingCLICommands returns all streaming CLI commands
func (p *PluginBase) GetStreamingCLICommands() (map[string]StreamingCLICommandHandler, error) {
	return p.streamingCLICommands, nil
}

// AddMenuEntry adds a menu entry to a specific menu
func (p *PluginBase) AddMenuEntry(menuID string, entry MenuEntry) {
	if p.menuEntries[menuID] == nil {
		p.menuEntries[menuID] = []MenuEntry{}
	}
	p.menuEntries[menuID] = append(p.menuEntries[menuID], entry)
}

// SimpleProvider provides a convenient way to create basic providers
type SimpleProvider struct {
	name string
	fn   func(param string) ([]MenuEntry, error)
}

// NewSimpleProvider creates a new simple provider
func NewSimpleProvider(name string, fn func(param string) ([]MenuEntry, error)) *SimpleProvider {
	return &SimpleProvider{
		name: name,
		fn:   fn,
	}
}

// GetName returns the provider name
func (p *SimpleProvider) GetName() string {
	return p.name
}

// GetDescription returns the provider description  
func (p *SimpleProvider) GetDescription() string {
	return "Simple provider" // Default description
}

// GenerateEntries returns the menu entries
func (p *SimpleProvider) GenerateEntries(param string) ([]MenuEntry, error) {
	return p.fn(param)
}

// SupportsRefresh indicates if this provider supports real-time updates
func (p *SimpleProvider) SupportsRefresh() bool {
	return false // Default implementation
}
// StandardProvider provides a provider with standardized naming and metadata
type StandardProvider struct {
	name         string
	providerType string
	param        string
	description  string
	fn           func(param string) ([]MenuEntry, error)
}

// NewStandardProvider creates a new standard provider
func NewStandardProvider(name, providerType, param, description string, fn func(param string) ([]MenuEntry, error)) *StandardProvider {
	return &StandardProvider{
		name:         name,
		providerType: providerType,
		param:        param,
		description:  description,
		fn:           fn,
	}
}

// GetName returns the provider name
func (p *StandardProvider) GetName() string {
	return p.name
}

// GetType returns the provider type
func (p *StandardProvider) GetType() string {
	return p.providerType
}

// GetParam returns the provider param
func (p *StandardProvider) GetParam() string {
	return p.param
}

// GetDescription returns the provider description
func (p *StandardProvider) GetDescription() string {
	return p.description
}

// GenerateEntries returns the menu entries
func (p *StandardProvider) GenerateEntries(param string) ([]MenuEntry, error) {
	return p.fn(param)
}

// SupportsRefresh indicates if this provider supports real-time updates
func (p *StandardProvider) SupportsRefresh() bool {
	return false // Default implementation
}

// UnifiedProvider provides a single provider that handles multiple contexts based on parameters
type UnifiedProvider struct {
	name        string
	description string
	handlers    map[string]func(param string) ([]MenuEntry, error)
}

// NewUnifiedProvider creates a new unified provider
func NewUnifiedProvider(name, description string) *UnifiedProvider {
	return &UnifiedProvider{
		name:        name,
		description: description,
		handlers:    make(map[string]func(param string) ([]MenuEntry, error)),
	}
}

// WithMenuHandler adds a menu handler for the unified provider
func (p *UnifiedProvider) WithMenuHandler(title string, handler func(param string) ([]MenuEntry, error)) *UnifiedProvider {
	key := fmt.Sprintf("menu:%s", title)
	p.handlers[key] = handler
	return p
}

// WithDataHandler adds a data handler for the unified provider
func (p *UnifiedProvider) WithDataHandler(title string, handler func(param string) ([]MenuEntry, error)) *UnifiedProvider {
	key := fmt.Sprintf("data:%s", title)
	p.handlers[key] = handler
	return p
}

// WithConfigHandler adds a config handler for the unified provider
func (p *UnifiedProvider) WithConfigHandler(title string, handler func(param string) ([]MenuEntry, error)) *UnifiedProvider {
	key := fmt.Sprintf("config:%s", title)
	p.handlers[key] = handler
	return p
}

// GetName returns the provider name
func (p *UnifiedProvider) GetName() string {
	return p.name
}

// GetDescription returns the provider description
func (p *UnifiedProvider) GetDescription() string {
	return p.description
}

// GenerateEntries returns the menu entries based on the parameter
func (p *UnifiedProvider) GenerateEntries(param string) ([]MenuEntry, error) {
	// Parse the parameter to determine type and title
	// Expected format: "type:title" or just "title" (defaults to menu)
	var providerType, title string
	
	if strings.Contains(param, ":") {
		parts := strings.SplitN(param, ":", 2)
		providerType = parts[0]
		title = parts[1]
	} else {
		providerType = "menu"
		title = param
	}
	
	key := fmt.Sprintf("%s:%s", providerType, title)
	
	if handler, exists := p.handlers[key]; exists {
		return handler(param)
	}
	
	// Return empty if no handler found
	return []MenuEntry{}, nil
}

// SupportsRefresh indicates if this provider supports real-time updates
func (p *UnifiedProvider) SupportsRefresh() bool {
	return false // Default implementation
}

// HasMainMenu checks if this provider has a main menu handler
func (p *UnifiedProvider) HasMainMenu() bool {
	_, exists := p.handlers["menu:main"]
	return exists
}

// WithSimpleProvider adds a simple provider to the plugin base
func (p *PluginBase) WithSimpleProvider(name string, fn func(param string) ([]MenuEntry, error)) *PluginBase {
	p.AddProvider(NewSimpleProvider(name, fn))
	return p
}

// WithInteractiveFunction adds an interactive function to the plugin base
func (p *PluginBase) WithInteractiveFunction(name string, fn InteractiveGoFunction) *PluginBase {
	p.AddInteractiveFunction(name, fn)
	return p
}

// WithCLICommand adds a CLI command to the plugin base
func (p *PluginBase) WithCLICommand(name string, handler CLICommandHandler) *PluginBase {
	p.AddCLICommand(name, handler)
	return p
}

// WithStreamingCLICommand adds a streaming CLI command to the plugin base
func (p *PluginBase) WithStreamingCLICommand(name string, handler StreamingCLICommandHandler) *PluginBase {
	p.AddStreamingCLICommand(name, handler)
	return p
}

// WithMenuEntry adds a menu entry to the plugin base
func (p *PluginBase) WithMenuEntry(menuID string, entry MenuEntry) *PluginBase {
	p.AddMenuEntry(menuID, entry)
	return p
}
