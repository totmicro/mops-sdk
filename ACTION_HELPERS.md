# Action Helpers - Unified CLI/UI Plugin Development

The MOPS SDK now provides powerful action helpers that make it extremely easy to create plugins with unified CLI and UI functionality.

## Key Benefits

✅ **Unified Actions**: Write once, works in both CLI and UI modes  
✅ **Automatic Mapping**: CLI commands automatically map to UI actions  
✅ **Streaming Support**: Built-in support for real-time output  
✅ **Reduced Boilerplate**: Minimal code needed for full functionality  
✅ **Type Safety**: Compile-time checks for action mappings  

## Quick Start

### 1. Basic Action Mapping

```go
// Create actions that work in both CLI and UI
plugin := sdk.NewPluginBuilder("myplugin", "1.0.0", "My awesome plugin").
    WithActionMappings().
        // Simple action that works everywhere
        AddAction(sdk.SimpleAction("demo", "Run demo", "action", demoFunction)).
        
        // Streaming action for real-time output
        AddInteractive(sdk.StreamingAction("stream", "Stream demo", "stream-demo", streamFunction)).
        
        Build().
    Build()
```

### 2. Standard Action Set

```go
// Use predefined common actions
plugin := sdk.NewPluginBuilder("myplugin", "1.0.0", "My plugin").
    WithStandardActions(sdk.PluginActionSet{
        Basic:     basicDemo,     // Basic functionality
        Config:    showConfig,    // Configuration display
        Help:      showHelp,      // Help information  
        Streaming: streamingDemo, // Real-time streaming
    }).
    Build()
```

## Action Function Signatures

### Regular Actions
```go
func myAction(ctx context.Context, params map[string]interface{}) (interface{}, error) {
    // Your action logic here
    return "Action result", nil
}
```

### Streaming Actions  
```go
func myStreamingAction(ctx context.Context, outputChan chan<- string, inputChan <-chan string, params map[string]interface{}) error {
    outputChan <- "Real-time output!"
    return nil
}
```

## Complete Example

```go
package main

import (
    "context"
    "fmt"
    sdk "github.com/totmicro/mops-sdk"
)

func main() {
    plugin := sdk.NewPluginBuilder("example", "1.0.0", "Example plugin").
        WithActionMappings().
            AddAction(sdk.SimpleAction("demo", "Run demo", "action", demo)).
            AddInteractive(sdk.StreamingAction("stream", "Stream", "stream", stream)).
            Build().
        WithMainMenuProvider("Main menu", provideMenu).
        Build()
    
    sdk.Main(plugin)
}

func demo(ctx context.Context, params map[string]interface{}) (interface{}, error) {
    return "🎯 Demo executed successfully!", nil
}

func stream(ctx context.Context, out chan<- string, in <-chan string, params map[string]interface{}) error {
    out <- "🚀 Streaming output...\n"
    return nil
}

func provideMenu(param string) ([]sdk.MenuEntry, error) {
    return []sdk.MenuEntry{
        {Key: "1", Label: "🎯 Demo", Action: "demo-action"},
        {Key: "2", Label: "⚡ Stream", Action: "interactive_go", Command: "stream"},
    }, nil
}
```

## What Happens Automatically

When you use action helpers, the SDK automatically:

1. **Registers CLI Commands**: `mops plugin example demo`
2. **Creates UI Actions**: Menu entries that execute the same logic
3. **Maps CLI to UI**: `demo` CLI command → `demo-action` UI action
4. **Handles Parameters**: Converts CLI args to action parameters
5. **Manages Output**: Formats results for CLI/UI appropriately
6. **Enables Streaming**: Real-time output in both modes

## Migration from Manual Mapping

### Before (Manual Setup)
```go
// Lots of boilerplate...
WithCLICommandAndUIMapping("demo", "Run demo", func(args []string) error {
    result, err := executeDemo(args)
    if err != nil { return err }
    fmt.Println(result)
    return nil
}, "action", "", "")

WithStandardExecutor("demo-action", func(entry MenuEntry, input string) ActionResult {
    result, err := executeDemo([]string{input})
    return ActionResult{Success: err == nil, Output: result}
})
```

### After (Action Helpers)
```go
// Clean and simple!
WithActionMappings().
    AddAction(sdk.SimpleAction("demo", "Run demo", "action", executeDemo)).
    Build()
```

## Supported Action Types

- **`action`**: Standard UI action execution
- **`goto`**: Navigation to another menu
- **`interactive_go`**: Real-time streaming functions

## Best Practices

1. **Use Standard Actions**: Start with `WithStandardActions()` for common patterns
2. **Leverage Streaming**: Use `AddInteractive()` for real-time output needs
3. **Consistent Naming**: Use descriptive action names that work in both contexts
4. **Error Handling**: Return meaningful errors that work in CLI and UI
5. **Parameter Design**: Design actions to work with both CLI args and UI inputs

The action helpers eliminate the complexity of manual CLI/UI mapping while providing a more powerful and maintainable approach to plugin development.
