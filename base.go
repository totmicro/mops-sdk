package sdk

// PluginBase provides a base implementation for mops plugins
type PluginBase struct {
	info                 PluginInfo
	config               map[string]any
	providers            []DynamicProvider
	executors            []ActionExecutor
	interactiveFunctions map[string]InteractiveGoFunction
	cliCommands          map[string]CLICommandHandler
	menuEntries          map[string][]MenuEntry
}

// NewPluginBase creates a new plugin base
func NewPluginBase(info PluginInfo) *PluginBase {
	return &PluginBase{
		info:                 info,
		providers:            []DynamicProvider{},
		executors:            []ActionExecutor{},
		interactiveFunctions: make(map[string]InteractiveGoFunction),
		cliCommands:          make(map[string]CLICommandHandler),
		menuEntries:          make(map[string][]MenuEntry),
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

// RegisterExecutors registers action executors with mops
func (p *PluginBase) RegisterExecutors() []ActionExecutor {
	return p.executors
}

// RegisterInteractiveFunctions registers interactive functions with mops
func (p *PluginBase) RegisterInteractiveFunctions() map[string]InteractiveGoFunction {
	return p.interactiveFunctions
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

// AddExecutor adds an action executor to the plugin
func (p *PluginBase) AddExecutor(executor ActionExecutor) {
	p.executors = append(p.executors, executor)
}

// AddInteractiveFunction adds an interactive function to the plugin
func (p *PluginBase) AddInteractiveFunction(name string, fn InteractiveGoFunction) {
	p.interactiveFunctions[name] = fn
}

// AddCLICommand adds a CLI command to the plugin
func (p *PluginBase) AddCLICommand(name string, handler CLICommandHandler) {
	p.cliCommands[name] = handler
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

// GetEntries returns the menu entries
func (p *SimpleProvider) GetEntries(param string) ([]MenuEntry, error) {
	return p.fn(param)
}

// SimpleExecutor provides a convenient way to create basic executors
type SimpleExecutor struct {
	actionType string
	fn         func(entry MenuEntry, input string) ActionResult
}

// NewSimpleExecutor creates a new simple executor
func NewSimpleExecutor(actionType string, fn func(entry MenuEntry, input string) ActionResult) *SimpleExecutor {
	return &SimpleExecutor{
		actionType: actionType,
		fn:         fn,
	}
}

// GetActionType returns the action type
func (e *SimpleExecutor) GetActionType() string {
	return e.actionType
}

// Execute executes the action
func (e *SimpleExecutor) Execute(entry MenuEntry, input string) ActionResult {
	return e.fn(entry, input)
}

// WithSimpleProvider adds a simple provider to the plugin base
func (p *PluginBase) WithSimpleProvider(name string, fn func(param string) ([]MenuEntry, error)) *PluginBase {
	p.AddProvider(NewSimpleProvider(name, fn))
	return p
}

// WithSimpleExecutor adds a simple executor to the plugin base
func (p *PluginBase) WithSimpleExecutor(actionType string, fn func(entry MenuEntry, input string) ActionResult) *PluginBase {
	p.AddExecutor(NewSimpleExecutor(actionType, fn))
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

// WithMenuEntry adds a menu entry to the plugin base
func (p *PluginBase) WithMenuEntry(menuID string, entry MenuEntry) *PluginBase {
	p.AddMenuEntry(menuID, entry)
	return p
}
