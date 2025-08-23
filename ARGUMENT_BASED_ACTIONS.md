# Argument-Based Actions - Simplified CLI with Args Pattern

The MOPS SDK now provides a streamlined way to create CLI commands that support both direct argument execution and interactive menu selection.

## Key Benefits

✅ **One Function Call**: Create complete CLI-with-args patterns in a single call  
✅ **Automatic Menu Generation**: Auto-generated selection menus with numbered entries  
✅ **Smart CLI Behavior**: CLI with args executes directly, without args shows menu  
✅ **Consistent UX**: Unified experience across CLI and UI modes  
✅ **Reusable Pattern**: Generic pattern works for any selection-based command  

## Quick Start

### Basic Pattern

```go
// Define your items
items := []sdk.SelectionItem{
    {
        Name:        "production",
        Description: "Production environment",
        Icon:        "🌐",
        Params: map[string]interface{}{
            "env":    "production",
            "server": "prod.example.com",
        },
    },
    {
        Name:        "staging", 
        Description: "Staging environment",
        Icon:        "🧪",
        Params: map[string]interface{}{
            "env":    "staging",
            "server": "stage.example.com",
        },
    },
}

// Create the complete action pattern
plugin := sdk.NewPluginBuilder("myplugin", "1.0.0", "My plugin").
    WithArgumentBasedAction(sdk.ArgumentBasedActionConfig{
        CommandName:     "deploy",
        Description:     "Deploy to environment",
        MenuID:          "deploy-selector",
        MenuTitle:       "🚀 Deployment Environments",
        Items:           items,
        DirectFunction:  deployDirect,     // Handles: myplugin deploy production
        ExecuteFunction: deployFromMenu,   // Handles: menu selection
    }).
    Build()
```

### Function Signatures

```go
// Direct execution function (called when CLI args provided)
func deployDirect(ctx context.Context, outputChan chan<- string, inputChan <-chan string, params map[string]interface{}) error {
    // params["arg0"] contains first CLI argument
    // params["cliArgs"] contains full argument string
    env := params["arg0"].(string)
    outputChan <- fmt.Sprintf("🚀 Deploying to %s...", env)
    return nil
}

// Menu execution function (called when menu item selected)
func deployFromMenu(ctx context.Context, outputChan chan<- string, inputChan <-chan string, params map[string]interface{}) error {
    // params contain the item's Params map
    env := params["env"].(string)
    server := params["server"].(string)
    outputChan <- fmt.Sprintf("🚀 Deploying to %s (%s)...", env, server)
    return nil
}
```

## What Happens Automatically

When you use `WithArgumentBasedAction`, the SDK automatically:

1. **Registers CLI Command**: `mops myplugin deploy [args...]`
2. **Creates Selection Menu**: Interactive menu with your items
3. **Handles Smart Routing**: 
   - `mops myplugin deploy production` → calls `deployDirect`
   - `mops myplugin deploy` → shows selection menu
4. **Generates Menu Entries**: Auto-numbered menu with icons and descriptions
5. **Parameter Mapping**: Converts CLI args and menu selections to function parameters

## Advanced Usage

### Custom Icons and Styling

```go
items := []sdk.SelectionItem{
    {
        Name:        "database",
        Description: "Database operations",
        Icon:        "🗄️",
        Params: map[string]interface{}{
            "type": "database",
            "operations": []string{"backup", "restore", "migrate"},
        },
    },
    {
        Name:        "api",
        Description: "API operations", 
        Icon:        "🔌",
        Params: map[string]interface{}{
            "type": "api",
            "endpoints": []string{"/health", "/metrics"},
        },
    },
}
```

### Multiple Argument Patterns

The same pattern works for various CLI argument styles:

- `myplugin connect server1` - Direct connection
- `myplugin connect` - Show server selection menu
- `myplugin backup daily` - Direct backup
- `myplugin backup` - Show backup type menu

## Migration from Manual Setup

### Before (Manual Setup)
```go
// Lots of boilerplate...
WithSmartCLICommand(sdk.SmartCLICommandConfig{
    Command: "deploy",
    Description: "Deploy to environment", 
    UITarget: "deploy-menu",
    DirectExecutor: deployHandler,
}).
WithMenuProvider("deploy-menu", "deploy-menu", "Environments", func(param string) ([]MenuEntry, error) {
    return []MenuEntry{
        {Key: "1", Label: "Production", Action: "core_interactive-go", Command: "deploy_prod"},
        {Key: "2", Label: "Staging", Action: "core_interactive-go", Command: "deploy_stage"},
    }, nil
}).
WithInteractiveFunction("deploy_prod", prodFunction).
WithInteractiveFunction("deploy_stage", stageFunction)
```

### After (Argument-Based Actions)
```go
// Clean and simple!
WithArgumentBasedAction(sdk.ArgumentBasedActionConfig{
    CommandName:     "deploy",
    Description:     "Deploy to environment",
    MenuID:          "deploy-selector", 
    MenuTitle:       "🚀 Deployment Environments",
    Items:           deployItems,
    DirectFunction:  deployDirect,
    ExecuteFunction: deployFromMenu,
})
```

## Best Practices

1. **Consistent Naming**: Use descriptive command names that work in both CLI and UI
2. **Rich Parameters**: Include all necessary data in SelectionItem.Params
3. **Error Handling**: Handle both CLI argument validation and menu parameter validation
4. **Icons & Descriptions**: Use clear icons and descriptions for better UX
5. **Parameter Design**: Design functions to work with both CLI args and menu parameters

The argument-based action pattern eliminates boilerplate while providing a powerful and consistent CLI-with-args experience.
