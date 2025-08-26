package sdk

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"gopkg.in/yaml.v2"
)

// SimplePluginConfig provides a configuration structure for quick plugin setup
type SimplePluginConfig struct {
	Author              string                            // Plugin author
	DisplayName         string                            // Plugin display name
	AutoLoadPresets     bool                              // Whether to auto-load presets from plugin.yaml
	InteractiveFunctions map[string]InteractiveGoFunction // Interactive functions
	MenuEntries         []MenuEntry                       // Menu entries for UI
	CLIShortcuts        []CLIShortcut                     // CLI shortcuts for menu access
}

// CLIShortcut represents a CLI shortcut configuration
type CLIShortcut struct {
	Command     string // CLI command name
	MenuID      string // Target menu ID
	Description string // Description for help
}

// FormField represents a form field configuration for input menus
type FormField struct {
	Key    string // Field identifier
	Label  string // Field display label
	Type   string // Field type (text, number, etc.)
}

// ParameterType represents the data type of a parameter
type ParameterType string

const (
	StringParam  ParameterType = "string"
	IntParam     ParameterType = "int"
	BoolParam    ParameterType = "bool"
	PortParam    ParameterType = "port"
	UrlParam     ParameterType = "url"
)

// ParameterDefinition defines validation rules and metadata for a parameter
type ParameterDefinition struct {
	Name        string        // Parameter name (used for flag parsing, e.g., "profile")
	Description string        // Human-readable description
	Type        ParameterType // Parameter data type
	Required    bool          // Whether parameter is required
	Default     string        // Default value if not provided
	Pattern     string        // Regex pattern for validation (optional)
	MinValue    *int          // Minimum value for numeric types (optional)
	MaxValue    *int          // Maximum value for numeric types (optional)
	Choices     []string      // Valid choices for the parameter (optional)
}

// ValidationError represents a parameter validation error
type ValidationError struct {
	Parameter string
	Message   string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("parameter '%s': %s", e.Parameter, e.Message)
}

// ValidateValue validates a parameter value against its definition
func (p *ParameterDefinition) ValidateValue(value string) error {
	// Required check
	if p.Required && value == "" {
		return ValidationError{Parameter: p.Name, Message: "is required"}
	}
	
	// If empty and not required, use default or skip validation
	if value == "" {
		return nil
	}
	
	// Type validation
	switch p.Type {
	case IntParam:
		intVal, err := strconv.Atoi(value)
		if err != nil {
			return ValidationError{Parameter: p.Name, Message: "must be a valid integer"}
		}
		if p.MinValue != nil && intVal < *p.MinValue {
			return ValidationError{Parameter: p.Name, Message: fmt.Sprintf("must be at least %d", *p.MinValue)}
		}
		if p.MaxValue != nil && intVal > *p.MaxValue {
			return ValidationError{Parameter: p.Name, Message: fmt.Sprintf("must be at most %d", *p.MaxValue)}
		}
		
	case PortParam:
		port, err := strconv.Atoi(value)
		if err != nil {
			return ValidationError{Parameter: p.Name, Message: "must be a valid port number"}
		}
		if port < 1 || port > 65535 {
			return ValidationError{Parameter: p.Name, Message: "must be between 1 and 65535"}
		}
		
	case BoolParam:
		if value != "true" && value != "false" && value != "1" && value != "0" {
			return ValidationError{Parameter: p.Name, Message: "must be true, false, 1, or 0"}
		}
		
	case UrlParam:
		if !strings.HasPrefix(value, "http://") && !strings.HasPrefix(value, "https://") {
			return ValidationError{Parameter: p.Name, Message: "must be a valid URL (http:// or https://)"}
		}
	}
	
	// Pattern validation
	if p.Pattern != "" {
		matched, err := regexp.MatchString(p.Pattern, value)
		if err != nil {
			return ValidationError{Parameter: p.Name, Message: "invalid pattern validation"}
		}
		if !matched {
			return ValidationError{Parameter: p.Name, Message: fmt.Sprintf("does not match required pattern: %s", p.Pattern)}
		}
	}
	
	// Choices validation
	if len(p.Choices) > 0 {
		valid := false
		for _, choice := range p.Choices {
			if value == choice {
				valid = true
				break
			}
		}
		if !valid {
			return ValidationError{Parameter: p.Name, Message: fmt.Sprintf("must be one of: %s", strings.Join(p.Choices, ", "))}
		}
	}
	
	return nil
}

// ParseAndValidate parses CLI arguments against parameter definitions and returns validated values
func ParseAndValidateArgs(args []string, definitions []ParameterDefinition) (map[string]string, error) {
	result := make(map[string]string)
	
	// Initialize with defaults
	for _, def := range definitions {
		if def.Default != "" {
			result[def.Name] = def.Default
		}
	}
	
	// Parse arguments
	for i := 0; i < len(args); i++ {
		arg := args[i]
		
		if strings.HasPrefix(arg, "--") {
			var key, value string
			
			if strings.Contains(arg, "=") {
				// Format: --key=value
				parts := strings.SplitN(arg[2:], "=", 2)
				key = parts[0]
				value = parts[1]
			} else {
				// Format: --key value
				key = arg[2:]
				if i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
					i++
					value = args[i]
				} else {
					// Boolean flag without value
					value = "true"
				}
			}
			
			// Find parameter definition
			var paramDef *ParameterDefinition
			for _, def := range definitions {
				if def.Name == key {
					paramDef = &def
					break
				}
			}
			
			if paramDef == nil {
				return nil, fmt.Errorf("unknown parameter: --%s", key)
			}
			
			// Validate and store
			if err := paramDef.ValidateValue(value); err != nil {
				return nil, err
			}
			
			result[key] = value
		}
	}
	
	// Check required parameters
	for _, def := range definitions {
		if def.Required {
			if _, exists := result[def.Name]; !exists || result[def.Name] == "" {
				return nil, ValidationError{Parameter: def.Name, Message: "is required"}
			}
		}
	}
	
	return result, nil
}

// Helper functions for creating parameter definitions

// RequiredStringParam creates a required string parameter
func RequiredStringParam(name, description string) ParameterDefinition {
	return ParameterDefinition{
		Name:        name,
		Description: description,
		Type:        StringParam,
		Required:    true,
	}
}

// OptionalStringParam creates an optional string parameter with default value
func OptionalStringParam(name, description, defaultValue string) ParameterDefinition {
	return ParameterDefinition{
		Name:        name,
		Description: description,
		Type:        StringParam,
		Required:    false,
		Default:     defaultValue,
	}
}

// RequiredIntParam creates a required integer parameter
func RequiredIntParam(name, description string) ParameterDefinition {
	return ParameterDefinition{
		Name:        name,
		Description: description,
		Type:        IntParam,
		Required:    true,
	}
}

// OptionalIntParam creates an optional integer parameter with default value
func OptionalIntParam(name, description string, defaultValue int) ParameterDefinition {
	return ParameterDefinition{
		Name:        name,
		Description: description,
		Type:        IntParam,
		Required:    false,
		Default:     fmt.Sprintf("%d", defaultValue),
	}
}

// PortParameter creates a port parameter (required by default)
func PortParameter(name, description string, required bool) ParameterDefinition {
	return ParameterDefinition{
		Name:        name,
		Description: description,
		Type:        PortParam,
		Required:    required,
	}
}

// ChoiceParam creates a parameter with predefined choices
func ChoiceParam(name, description string, choices []string, required bool) ParameterDefinition {
	return ParameterDefinition{
		Name:        name,
		Description: description,
		Type:        StringParam,
		Required:    required,
		Choices:     choices,
	}
}

// BoolParameter creates a boolean parameter
func BoolParameter(name, description string, defaultValue bool) ParameterDefinition {
	return ParameterDefinition{
		Name:        name,
		Description: description,
		Type:        BoolParam,
		Required:    false,
		Default:     fmt.Sprintf("%t", defaultValue),
	}
}

// WithBasicMenuProvider: Quickly define a main menu with minimal code
func (b *PluginBuilder) WithBasicMenuProvider(entries []MenuEntry) *PluginBuilder {
	return b.WithMenuProvider("main", "main", "Main menu", func(param string) ([]MenuEntry, error) {
		return entries, nil
	})
}

// WithMenuStartupShortcut: Register a CLI command to start UI at a specific menu
func (b *PluginBuilder) WithMenuStartupShortcut(command, menuID, description string) *PluginBuilder {
	return b.WithHiddenCLICommand(command, description, func(args []string) error {
		// This should trigger the UI to open at menuID (requires core support)
		fmt.Printf("[SDK] Would start UI at menu: %s\n", menuID)
		// TODO: Integrate with MOPS core to actually start UI at menuID
		return nil
	})
}

// WithAutoConfigPresets: Auto-load config presets from plugin.yaml
func (b *PluginBuilder) WithAutoConfigPresets() *PluginBuilder {
	// Get the plugin name for path construction
	pluginName := b.getInfo().Name
	
	// List of possible paths to search for plugin.yaml
	searchPaths := []string{
		"plugin.yaml",
		filepath.Join("..", "mops-plugins", "plugins", pluginName, "plugin.yaml"),
		filepath.Join("plugins", pluginName, "plugin.yaml"),
	}
	
	var yamlData []byte
	var err error
	
	// Try each search path
	for _, path := range searchPaths {
		yamlData, err = ioutil.ReadFile(path)
		if err == nil {
			break
		}
	}
	
	if err != nil {
		// If plugin.yaml doesn't exist in any location, just return without error
		return b
	}
	
	// Parse the YAML file
	var metadata struct {
		ConfigPresets map[string]struct {
			Name        string                 `yaml:"name"`
			Description string                 `yaml:"description"`
			Config      map[string]interface{} `yaml:"config"`
		} `yaml:"config_presets"`
	}
	
	if err := yaml.Unmarshal(yamlData, &metadata); err != nil {
		// If parsing fails, just return without error
		return b
	}
	
	// Register each config preset
	for presetName, preset := range metadata.ConfigPresets {
		b.WithConfigPreset(presetName, preset.Name, preset.Description, preset.Config)
	}
	
	return b
}

// WithQuickMenu: Create a menu with minimal configuration using menu provider
func (b *PluginBuilder) WithQuickMenu(menuID, title string, entries []MenuEntry) *PluginBuilder {
	return b.WithMenuProvider(menuID, menuID, title, func(param string) ([]MenuEntry, error) {
		return entries, nil
	})
}

// WithAutoStartMenu: Set a menu to start automatically when UI opens
func (b *PluginBuilder) WithAutoStartMenu(menuID string) *PluginBuilder {
	return b.WithMenuStartupShortcut("auto", menuID, "Auto-start menu")
}

// WithFormMenu: Create a form-based menu for input collection
func (b *PluginBuilder) WithFormMenu(menuID, title string, fields []FormField, submitAction string) *PluginBuilder {
	entries := []MenuEntry{}
	for _, field := range fields {
		entries = append(entries, MenuEntry{
			Key:    field.Key,
			Label:  field.Label,
			Action: "input",
		})
	}
	
	// Add submit button
	entries = append(entries, MenuEntry{
		Key:    "submit",
		Label:  "✅ Submit",
		Action: "action",
		Command: submitAction,
	})
	
	return b.WithMenuProvider(menuID, menuID, title, func(param string) ([]MenuEntry, error) {
		return entries, nil
	})
}

// WithCommandShortcuts: Add multiple CLI shortcuts at once
func (b *PluginBuilder) WithCommandShortcuts(shortcuts map[string]string) *PluginBuilder {
	for command, target := range shortcuts {
		b.WithMenuStartupShortcut(command, target, fmt.Sprintf("Shortcut to %s", target))
	}
	return b
}

// WithTags: Set multiple tags at once
func (b *PluginBuilder) WithTags(tags []string) *PluginBuilder {
	for _, tag := range tags {
		b.AddTag(tag)
	}
	return b
}

// WithSimplePlugin: Creates a complete simple plugin with minimal configuration
func (b *PluginBuilder) WithSimplePlugin(config SimplePluginConfig) *PluginBuilder {
	// Set basic info
	if config.Author != "" {
		b.SetAuthor(config.Author)
	}
	if config.DisplayName != "" {
		b.SetDisplayName(config.DisplayName)
	}
	
	// Auto-load presets if enabled
	if config.AutoLoadPresets {
		b.WithAutoConfigPresets()
	}
	
	// Add interactive functions
	for name, fn := range config.InteractiveFunctions {
		b.WithInteractiveFunction(name, fn)
	}
	
	// Add menu provider if provided
	if len(config.MenuEntries) > 0 {
		b.WithBasicMenuProvider(config.MenuEntries)
	}
	
	// Add CLI shortcuts
	for _, shortcut := range config.CLIShortcuts {
		b.WithMenuStartupShortcut(shortcut.Command, shortcut.MenuID, shortcut.Description)
	}
	
	return b
}

// WithPluginFromYAML: Auto-configure plugin from plugin.yaml metadata
func (b *PluginBuilder) WithPluginFromYAML() *PluginBuilder {
	// Try to load plugin.yaml and extract metadata
	searchPaths := []string{
		"plugin.yaml",
		"../plugin.yaml",
		"../../mops-plugins/plugins/example-plugin/plugin.yaml",
	}
	
	var yamlData []byte
	var err error
	
	for _, path := range searchPaths {
		yamlData, err = ioutil.ReadFile(path)
		if err == nil {
			break
		}
	}
	
	if err != nil {
		return b
	}
	
	// Parse the full YAML metadata
	var metadata struct {
		Name        string `yaml:"name"`
		DisplayName string `yaml:"display_name"`
		Author      string `yaml:"author"`
		Homepage    string `yaml:"homepage"`
		License     string `yaml:"license"`
		Tags        []string `yaml:"tags"`
	}
	
	if err := yaml.Unmarshal(yamlData, &metadata); err == nil {
		// Apply metadata to plugin
		if metadata.DisplayName != "" {
			b.SetDisplayName(metadata.DisplayName)
		}
		if metadata.Author != "" {
			b.SetAuthor(metadata.Author)
		}
		if metadata.Homepage != "" {
			b.SetHomepage(metadata.Homepage)
		}
		if metadata.License != "" {
			b.SetLicense(metadata.License)
		}
		for _, tag := range metadata.Tags {
			b.AddTag(tag)
		}
	}
	
	// Also load presets
	b.WithAutoConfigPresets()
	
	return b
}

// WithQuickSetup: One-liner to create a fully configured plugin (simplified to interactive functions only)
func (b *PluginBuilder) WithQuickSetup(menuEntries []MenuEntry) *PluginBuilder {
	return b.
		WithPluginFromYAML().                    // Auto-load from YAML
		WithBasicMenuProvider(menuEntries).      // Add menu
		WithMenuStartupShortcut("ui", "main", "Open UI") // Add UI shortcut
}

// InteractiveMapping defines CLI to Interactive Function mappings
type InteractiveMapping struct {
	CLIName        string                       // Name of the CLI command (e.g., "streaming")
	CLIDescription string                       // Description for CLI help
	UICommand      string                       // Interactive function name
	Function       InteractiveGoFunction        // The interactive function implementation (nil to use built-in)
	Params         map[string]interface{}       // Parameters to pass to the interactive function
	Hidden         bool                         // Hide this command from CLI help and prevent CLI execution
	MenuTarget     string                       // Target menu for navigation when no CLI args provided
}

// ActionMappingBuilder provides a fluent interface for creating action mappings
type ActionMappingBuilder struct {
	builder *PluginBuilder
}

// Exported helpers and types for plugin authors
var WithInteractiveFunction = (*PluginBuilder).WithInteractiveFunction

// WithActionMappings adds multiple action mappings with a fluent interface
func (b *PluginBuilder) WithActionMappings() *ActionMappingBuilder {
	return &ActionMappingBuilder{builder: b}
}

// AddInteractive adds a CLI command that maps to an interactive function
func (amb *ActionMappingBuilder) AddInteractive(mapping InteractiveMapping) *ActionMappingBuilder {
	// Create the CLI command handler
	handler := func(args []string) error {
		fmt.Println("🚀 Launching streaming demo in UI mode...")
		fmt.Println("📺 This provides real-time output with rich formatting!")
		return nil
	}
	
	// Add CLI command (hidden or visible based on mapping.Hidden)
	if mapping.Hidden {
		amb.builder.WithHiddenCLICommand(mapping.CLIName, mapping.CLIDescription, handler)
	} else {
		amb.builder.WithCLICommandAndUIMapping(
			mapping.CLIName,
			mapping.CLIDescription,
			handler,
			"core_interactive-go",     // Always use core_interactive-go for streaming
			"",                   // No target needed for interactive functions
			mapping.UICommand,    // The function name to call
		)
	}
	
	// Register the interactive function only if it's not nil (don't override built-in functions)
	if mapping.Function != nil {
		amb.builder.WithInteractiveFunction(mapping.UICommand, mapping.Function)
	}
	return amb
}

// Build finalizes the action mappings and returns the original builder
func (amb *ActionMappingBuilder) Build() *PluginBuilder {
	return amb.builder
}

// Common action mapping patterns

// StreamingAction creates an interactive function mapping for real-time output
func StreamingAction(name, description, functionName string, function InteractiveGoFunction) InteractiveMapping {
	return InteractiveMapping{
		CLIName:        name,
		CLIDescription: description,
		UICommand:      functionName,
		Function:       function,
		Hidden:         false,
	}
}

// CLIHiddenStreamingAction creates an interactive mapping that's hidden from CLI help but available in UI
func CLIHiddenStreamingAction(name, description, functionName string, function InteractiveGoFunction) InteractiveMapping {
	return InteractiveMapping{
		CLIName:        name,
		CLIDescription: description,
		UICommand:      functionName,
		Function:       function,
		Hidden:         true,
	}
}

// ShellAction creates an interactive mapping for shell command execution using the built-in shell_command function
func ShellAction(name, description, command string) InteractiveMapping {
	return InteractiveMapping{
		CLIName:        name,
		CLIDescription: description,
		UICommand:      "shell_command", // Use the built-in shell command function
		Function:       nil,             // We don't need a custom function, use the built-in one
		Hidden:         false,
		Params: map[string]interface{}{
			"command": command,
		},
	}
}

// SelectionAction creates a CLI command that shows a selection menu when no args provided,
// or executes directly when args are provided. Generic pattern for any CLI command with arguments.
func SelectionAction(name, description, menuTarget string, directFunction InteractiveGoFunction) InteractiveMapping {
	return InteractiveMapping{
		CLIName:        name,
		CLIDescription: description,
		UICommand:      fmt.Sprintf("%s_direct", name), // Function name for direct execution
		Function:       directFunction,
		Hidden:         false,
		MenuTarget:     menuTarget, // Target menu for selection when no args provided
	}
}

// MenuProviderAction creates a standardized menu provider with auto-numbered entries
// that work with CLI arguments and UI interaction
func (b *PluginBuilder) WithSelectionMenuProvider(menuID, title string, items []SelectionItem, executeFunction InteractiveGoFunction) *PluginBuilder {
	return b.WithMenuProvider(menuID, menuID, title, func(param string) ([]MenuEntry, error) {
		var entries []MenuEntry
		
		// Add selection entries directly without useless header
		for i, item := range items {
			// Use custom title if provided, otherwise default format
			var message string
			if item.Title != "" {
				message = item.Title
			} else {
				message = fmt.Sprintf("🚀 %s Login", item.Description)
			}
			
			entries = append(entries, MenuEntry{
				Key:     fmt.Sprintf("%d", i+1),
				Label:   fmt.Sprintf("%s %s - %s", item.Icon, item.Name, item.Description),
				Action:  "core_interactive-go",
				Command: fmt.Sprintf("%s_execute", menuID),
				Message: message,
				Params: item.Params,
			})
		}
		
		return entries, nil
	}).WithInteractiveFunction(fmt.Sprintf("%s_execute", menuID), executeFunction)
}

// SelectionItem represents a selectable item in a menu
type SelectionItem struct {
	Name        string                 // Item name/identifier
	Description string                 // Item description
	Icon        string                 // Item icon/emoji
	Title       string                 // Custom title for Bubble Tea UI (optional)
	Params      map[string]interface{} // Parameters to pass to the execute function
}

// WithArgumentBasedAction creates a complete CLI-with-args pattern:
// - CLI command with arguments executes directly 
// - CLI command without arguments navigates to selection menu
// - UI menu provides interactive selection
func (b *PluginBuilder) WithArgumentBasedAction(config ArgumentBasedActionConfig) *PluginBuilder {
	// Create the full menu ID with plugin prefix
	fullMenuID := b.getStandardName(config.MenuID)
	
	// Determine the CLI title to use
	cliTitle := config.CLITitle
	if cliTitle == "" {
		cliTitle = config.MenuTitle // Fall back to menu title if no CLI title specified
	}
	
	// Create the smart CLI command
	smartConfig := SmartCLICommandConfig{
		Command:             config.CommandName,
		Description:         config.Description,
		Usage:               fmt.Sprintf("%s [args...]", config.CommandName),
		SmartFunctionName:   fmt.Sprintf("%s_smart_%s", config.CommandName, strings.ReplaceAll(strings.ToLower(cliTitle), " ", "_")),
		UITarget:            fullMenuID, // Use the full prefixed menu ID
		UITitle:             cliTitle,   // Pass CLI title for Bubble Tea display
		DirectHandler:       nil, // We'll use DirectExecutor instead
		DirectExecutor:      convertToDirectExecutor(config.DirectFunction, cliTitle, config.Parameters),
	}
	
	// Add the smart CLI command
	b = b.WithSmartCLICommand(smartConfig)
	
	// Add the selection menu provider
	return b.WithSelectionMenuProvider(config.MenuID, config.MenuTitle, config.Items, config.ExecuteFunction)
}

// ArgumentBasedActionConfig configures a complete CLI-with-args action pattern
type ArgumentBasedActionConfig struct {
	CommandName     string                    // CLI command name
	Description     string                    // CLI command description
	MenuID          string                    // Menu ID for selection
	MenuTitle       string                    // Menu title for interactive selection
	CLITitle        string                    // Title for CLI execution (optional, defaults to MenuTitle)
	Items           []SelectionItem           // Selectable items
	Parameters      []ParameterDefinition     // Parameter definitions for validation
	DirectFunction  InteractiveGoFunction     // Function for direct CLI execution with args
	ExecuteFunction InteractiveGoFunction     // Function for menu-based execution
}

// convertToDirectExecutor converts an InteractiveGoFunction to work with DirectExecutor signature
func convertToDirectExecutor(fn InteractiveGoFunction, cliTitle string, parameters []ParameterDefinition) func(ctx context.Context, outputChan chan<- string, inputChan <-chan string, args []string) error {
	return func(ctx context.Context, outputChan chan<- string, inputChan <-chan string, args []string) error {
		var params map[string]interface{}
		
		if len(parameters) > 0 {
			// Use parameter validation if definitions are provided
			validatedParams, err := ParseAndValidateArgs(args, parameters)
			if err != nil {
				// Instead of returning an error, send user-friendly output
				outputChan <- ""
				outputChan <- "❌ Parameter Validation Error"
				outputChan <- "============================"
				outputChan <- ""
				outputChan <- fmt.Sprintf("Error: %v", err)
				outputChan <- ""
				outputChan <- "📖 Usage Help:"
				outputChan <- generateUsageHelp(cliTitle, parameters)
				outputChan <- ""
				outputChan <- "💡 Examples:"
				outputChan <- generateExamples(cliTitle, parameters)
				return nil // Return nil to avoid "unexpected EOF" errors
			}
			
			// Convert to interface{} map
			params = make(map[string]interface{})
			for key, value := range validatedParams {
				params[key] = value
			}
		} else {
			// Fallback to legacy behavior for backward compatibility
			params = make(map[string]interface{})
			for i, arg := range args {
				params[fmt.Sprintf("arg%d", i)] = arg
			}
			params["cliArgs"] = strings.Join(args, " ")
		}
		
		// Add CLI title to provide better context
		if cliTitle != "" {
			params["cliTitle"] = cliTitle
		}
		
		return fn(ctx, outputChan, inputChan, params)
	}
}

// generateUsageHelp creates a usage help string from parameter definitions
func generateUsageHelp(commandTitle string, parameters []ParameterDefinition) string {
	var help strings.Builder
	
	help.WriteString(fmt.Sprintf("Command: %s\n", commandTitle))
	help.WriteString("\nParameters:\n")
	
	for _, param := range parameters {
		required := ""
		if param.Required {
			required = " (required)"
		}
		
		defaultVal := ""
		if param.Default != "" {
			defaultVal = fmt.Sprintf(" [default: %s]", param.Default)
		}
		
		choices := ""
		if len(param.Choices) > 0 {
			choices = fmt.Sprintf(" [choices: %s]", strings.Join(param.Choices, ", "))
		}
		
		help.WriteString(fmt.Sprintf("  --%s: %s%s%s%s\n", 
			param.Name, 
			param.Description, 
			required, 
			defaultVal,
			choices))
	}
	
	return help.String()
}

// generateExamples creates usage examples from parameter definitions
func generateExamples(commandTitle string, parameters []ParameterDefinition) string {
	var examples strings.Builder
	
	// Extract command name from title (remove emojis and extra text)
	commandName := strings.ToLower(strings.Fields(commandTitle)[len(strings.Fields(commandTitle))-1])
	// Keep the command name as extracted, without hardcoded plugin references
	
	// Find required parameters for minimal example
	var requiredParams []ParameterDefinition
	var optionalParams []ParameterDefinition
	
	for _, param := range parameters {
		if param.Required {
			requiredParams = append(requiredParams, param)
		} else {
			optionalParams = append(optionalParams, param)
		}
	}
	
	// Generate minimal example with required parameters only
	if len(requiredParams) > 0 {
		example := fmt.Sprintf("  mops %s", commandName)
		for _, param := range requiredParams {
			exampleValue := "value"
			if len(param.Choices) > 0 {
				exampleValue = param.Choices[0]
			} else if param.Name == "profile" {
				exampleValue = "production"
			} else if param.Name == "port" {
				exampleValue = "22"
			}
			example += fmt.Sprintf(" --%s %s", param.Name, exampleValue)
		}
		examples.WriteString(example + "\n")
	}
	
	// Generate full example with optional parameters
	if len(parameters) > len(requiredParams) {
		example := fmt.Sprintf("  mops %s", commandName)
		for _, param := range parameters {
			exampleValue := "value"
			if len(param.Choices) > 0 {
				exampleValue = param.Choices[0]
			} else if param.Name == "profile" {
				exampleValue = "staging"
			} else if param.Name == "server" {
				exampleValue = "custom.example.com"
			} else if param.Name == "port" {
				exampleValue = "2222"
			} else if param.Name == "user" {
				exampleValue = "myuser"
			}
			example += fmt.Sprintf(" --%s %s", param.Name, exampleValue)
		}
		examples.WriteString(example + "\n")
	}
	
	return examples.String()
}

// WithShellCommands adds multiple shell command menu entries
func (b *PluginBuilder) WithShellCommands(commands map[string]ShellCommandEntry) *PluginBuilder {
	var entries []MenuEntry
	
	for key, shellCmd := range commands {
		entries = append(entries, MenuEntry{
			Key:     key,
			Label:   shellCmd.Label,
			Action:  "core_interactive-go",
			Command: "shell_command",
			Message: shellCmd.Description,
			Params: map[string]interface{}{
				"command": shellCmd.Command,
			},
		})
	}
	
	// Add these entries to the existing menu provider
	return b.WithBasicMenuProvider(entries)
}

// WithShellCommand adds a single shell command menu entry
func (b *PluginBuilder) WithShellCommand(key, label, description, command string) *PluginBuilder {
	return b.WithShellCommands(map[string]ShellCommandEntry{
		key: {
			Label:       label,
			Description: description,
			Command:     command,
		},
	})
}

// ShellCommandEntry represents a shell command configuration
type ShellCommandEntry struct {
	Label       string // Display label for the menu entry
	Description string // Description/message for the entry
	Command     string // Shell command to execute
}
