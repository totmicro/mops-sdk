package sdk

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
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
		Params: map[string]interface{}{
			"command": command, // Pass the shell command as a parameter
		},
		Hidden:         false,
	}
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
