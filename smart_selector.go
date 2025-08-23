package sdk

import (
	"context"
	"fmt"
	"strings"
)

// SmartSelectorConfig defines the configuration for a smart selector function
type SmartSelectorConfig struct {
	// Direct execution function - called when CLI arguments are provided
	DirectExecutor func(ctx context.Context, outputChan chan<- string, inputChan <-chan string, args []string) error
	
	// Interactive selector function - called when no CLI arguments are provided
	InteractiveSelector func(ctx context.Context, outputChan chan<- string, inputChan <-chan string, params map[string]interface{}) error
	
	// Name of the command for error messages
	CommandName string
	
	// Usage string for help
	Usage string
}

// CreateSmartSelector creates a smart selector function that handles both direct execution and interactive selection
func CreateSmartSelector(config SmartSelectorConfig) InteractiveGoFunction {
	return func(ctx context.Context, outputChan chan<- string, inputChan <-chan string, params map[string]interface{}) error {
		// Extract CLI arguments from various parameter sources
		args := extractCLIArguments(params)
		
		if len(args) > 0 {
			// CLI arguments provided - use direct executor
			return config.DirectExecutor(ctx, outputChan, inputChan, args)
		} else {
			// No CLI arguments - use interactive selector
			return config.InteractiveSelector(ctx, outputChan, inputChan, params)
		}
	}
}

// extractCLIArguments extracts CLI arguments from the params map
func extractCLIArguments(params map[string]interface{}) []string {
	var args []string
	
	// Check for CLI arguments passed via auto-execution (space-separated string)
	if cliArgs, exists := params["cliArgs"].(string); exists && cliArgs != "" {
		return strings.Fields(cliArgs)
	}
	
	// Check for individual CLI argument parameters (arg0, arg1, arg2, etc.)
	for i := 0; i < 10; i++ { // Support up to 10 arguments
		argKey := fmt.Sprintf("arg%d", i)
		if arg, exists := params[argKey].(string); exists && arg != "" {
			args = append(args, arg)
		} else {
			break // Stop at the first missing argument
		}
	}
	
	return args
}

// SmartCLICommand creates a CLI command configuration with UI mapping that uses the smart selector pattern
func SmartCLICommand(command, description string, directExecutor func(args []string) error, interactiveFunctionName string) CLICommandConfig {
	return CLICommandConfig{
		Command:     command,
		Description: description,
		Handler:     directExecutor,
		UIAction:    "core_interactive-go",
		UITarget:    "",
		UICommand:   interactiveFunctionName,
	}
}

// CLICommandConfig represents the configuration for a CLI command with UI mapping
type CLICommandConfig struct {
	Command     string
	Description string
	Handler     func(args []string) error
	UIAction    string
	UITarget    string
	UICommand   string
}

// SmartCLICommandConfig defines the configuration for a smart CLI command
type SmartCLICommandConfig struct {
	// Command name
	Command string
	
	// Command description
	Description string
	
	// Usage string for help
	Usage string
	
	// Name of the smart selector function to register
	SmartFunctionName string
	
	// UI target menu to navigate to when no CLI arguments are provided (optional)
	UITarget string
	
	// Direct CLI handler - called when the command is executed directly from CLI
	DirectHandler func(args []string) error
	
	// Direct executor for UI mode - called when CLI arguments are provided to the UI function
	DirectExecutor func(ctx context.Context, outputChan chan<- string, inputChan <-chan string, args []string) error
	
	// Interactive selector function - called when no CLI arguments are provided in UI mode
	InteractiveSelector func(ctx context.Context, outputChan chan<- string, inputChan <-chan string, params map[string]interface{}) error
}
