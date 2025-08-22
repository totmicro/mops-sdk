package sdk


// ActionExecutorFunc represents a simple function that can execute a UI action
// This wraps the interface-based ActionExecutor for easier use
type ActionExecutorFunc func(ctx context.Context, params map[string]interface{}) (interface{}, error)

// SimpleActionExecutor wraps a function to implement the ActionExecutor interface
type SimpleActionExecutor struct {
	handler ActionExecutorFunc
}

// WithBasicMenuProvider: Quickly define a main menu with minimal code
func (b *PluginBuilder) WithBasicMenuProvider(entries []MenuEntry) *PluginBuilder {
	return b.WithMenuProvider("main", "main", "Main menu", func(param string) ([]MenuEntry, error) {
		return entries, nil
	})
}

// WithMenuStartupShortcut: Register a CLI command to start UI at a specific menu
func (b *PluginBuilder) WithMenuStartupShortcut(command, menuID, description string) *PluginBuilder {
	return b.WithCLICommand(command, description, func(args []string) error {
		// This should trigger the UI to open at menuID (requires core support)
		fmt.Printf("[SDK] Would start UI at menu: %s\n", menuID)
		// TODO: Integrate with MOPS core to actually start UI at menuID
		return nil
	})
}

// WithAutoConfigPresets: Auto-load config presets from plugin.yaml
func (b *PluginBuilder) WithAutoConfigPresets() *PluginBuilder {
	for name, preset := range b.getInfo().ConfigPresets {
		b.WithConfigPreset(name, preset.Name, preset.Description, preset.Config)
	}
	return b
}
}

// ActionMapping defines the relationship between CLI commands and UI actions
type ActionMapping struct {
	CLIName        string             // Name of the CLI command (e.g., "basic")
	CLIDescription string             // Description for CLI help
	UIAction       string             // UI action type (e.g., "interactive_go", "action")
	UITarget       string             // Target menu/provider for goto actions
	UICommand      string             // Command name for interactive actions
	Executor       ActionExecutorFunc // Function to execute the action
}

// InteractiveMapping defines CLI to Interactive Function mappings
type InteractiveMapping struct {
	CLIName        string                // Name of the CLI command (e.g., "streaming")
	CLIDescription string                // Description for CLI help
	UICommand      string                // Interactive function name
	Function       InteractiveGoFunction // The interactive function implementation
}

// ActionMappingBuilder provides a fluent interface for creating action mappings
type ActionMappingBuilder struct {
	builder *PluginBuilder
}

// Exported helpers and types for plugin authors
type PluginActionSet = ActionSet
var WithStandardActions = (*PluginBuilder).WithStandardActions
var WithInteractiveFunction = (*PluginBuilder).WithInteractiveFunction

// WithActionMappings adds multiple action mappings with a fluent interface
func (b *PluginBuilder) WithActionMappings() *ActionMappingBuilder {
	return &ActionMappingBuilder{builder: b}
}

// AddAction adds a CLI command that maps to a UI action executor
func (amb *ActionMappingBuilder) AddAction(mapping ActionMapping) *ActionMappingBuilder {
	// Add the CLI command with UI mapping
	amb.builder.WithCLICommandAndUIMapping(
		mapping.CLIName,
		mapping.CLIDescription,
		func(args []string) error {
			// Convert CLI args to parameters
			params := make(map[string]interface{})
			for i, arg := range args {
				params[fmt.Sprintf("arg%d", i)] = arg
			}
			
			// Execute the action
			result, err := mapping.Executor(context.Background(), params)
			if err != nil {
				return err
			}
			
			// Print result if it's a string or convertible
			if result != nil {
				fmt.Println(result)
			}
			return nil
		},
		mapping.UIAction,
		mapping.UITarget,
		mapping.UICommand,
	)
	
	// Register the action executor for UI use
	if mapping.Executor != nil {
		amb.builder.WithStandardExecutor(mapping.CLIName, func(entry MenuEntry, input string) ActionResult {
			// Convert input and entry to params
			params := make(map[string]interface{})
			params["input"] = input
			params["entry"] = entry
			
			result, err := mapping.Executor(context.Background(), params)
			if err != nil {
				return ActionResult{
					Success: false,
					Output:  err.Error(),
					Error:   err,
				}
			}
			
			output := ""
			if result != nil {
				output = fmt.Sprintf("%v", result)
			}
			
			return ActionResult{
				Success:    true,
				Output:     output,
				ShowOutput: true,
			}
		})
	}
	
	return amb
}

// AddInteractive adds a CLI command that maps to an interactive function
func (amb *ActionMappingBuilder) AddInteractive(mapping InteractiveMapping) *ActionMappingBuilder {
	// Add the CLI command with interactive UI mapping
	amb.builder.WithCLICommandAndUIMapping(
		mapping.CLIName,
		mapping.CLIDescription,
		func(args []string) error {
			fmt.Println("🚀 Launching streaming demo in UI mode...")
			fmt.Println("📺 This provides real-time output with rich formatting!")
			return nil
		},
		"interactive_go",     // Always use interactive_go for streaming
		"",                   // No target needed for interactive functions
		mapping.UICommand,    // The function name to call
	)
	
	// Register the interactive function
	amb.builder.WithInteractiveFunction(mapping.UICommand, mapping.Function)
	return amb
}

// Build finalizes the action mappings and returns the original builder
func (amb *ActionMappingBuilder) Build() *PluginBuilder {
	return amb.builder
}

// Common action mapping patterns

// SimpleAction creates a basic action mapping
func SimpleAction(name, description, uiAction string, executor ActionExecutorFunc) ActionMapping {
	return ActionMapping{
		CLIName:        name,
		CLIDescription: description,
		UIAction:       uiAction,
		UITarget:       "",
		UICommand:      "",
		Executor:       executor,
	}
}

// MenuGotoAction creates an action mapping that navigates to a menu
func MenuGotoAction(name, description, targetMenu string, executor ActionExecutorFunc) ActionMapping {
	return ActionMapping{
		CLIName:        name,
		CLIDescription: description,
		UIAction:       "goto",
		UITarget:       targetMenu,
		UICommand:      "",
		Executor:       executor,
	}
}

// StreamingAction creates an interactive function mapping for real-time output
func StreamingAction(name, description, functionName string, function InteractiveGoFunction) InteractiveMapping {
	return InteractiveMapping{
		CLIName:        name,
		CLIDescription: description,
		UICommand:      functionName,
		Function:       function,
	}
}

// PluginActionSet provides a complete set of common plugin actions
type PluginActionSet struct {
	Basic       ActionExecutorFunc    // Basic demo functionality
	Config      ActionExecutorFunc    // Configuration display
	Help        ActionExecutorFunc    // Help information
	Streaming   InteractiveGoFunction // Real-time streaming demo
}

// WithStandardActions adds a standard set of plugin actions
func (b *PluginBuilder) WithStandardActions(actions PluginActionSet) *PluginBuilder {
	amb := b.WithActionMappings()
	
	if actions.Basic != nil {
		amb.AddAction(SimpleAction("basic", "Execute basic demo", "action", actions.Basic))
	}
	
	if actions.Config != nil {
		amb.AddAction(SimpleAction("config", "Show configuration", "action", actions.Config))
	}
	
	if actions.Help != nil {
		amb.AddAction(SimpleAction("help", "Show help information", "action", actions.Help))
	}
	
	if actions.Streaming != nil {
		amb.AddInteractive(StreamingAction("streaming", "Execute streaming demo", "streaming-demo", actions.Streaming))
	}
	
	return amb.Build()
}
