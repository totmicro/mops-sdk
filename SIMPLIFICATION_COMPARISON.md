# Simple Example Plugin Simplification

## Before vs After Comparison

### Original Implementation (lines count: ~250+ lines)

The original plugin had:
- Complex `SmartCLICommandConfig` setup
- Manual menu provider with custom selection handling  
- Repetitive action mappings
- Multiple `WithInteractiveFunction` calls
- Complex parameter extraction logic
- Boilerplate for handling both CLI args and menu selection

### Simplified Implementation (lines count: ~180 lines)

The simplified version using SDK action helpers:

#### Key Improvements:

1. **Single Configuration Line for CLI + Menu**:
   ```go
   // Before: ~30 lines of SmartCLICommandConfig + menu provider
   WithArgumentBasedAction(sdk.ArgumentBasedActionConfig{
       CommandName:     "login",
       Description:     "Login to a profile", 
       MenuID:          "profile-selector",
       MenuTitle:       "🔐 Profile Login",
       Items:           profiles,
       DirectFunction:  profileLoginDirect,
       ExecuteFunction: profileLoginFromMenu,
   })
   ```

2. **Unified Profile Data Structure**:
   ```go
   // Clean, structured profile definitions
   profiles := []sdk.SelectionItem{
       {
           Name:        "production",
           Description: "Production environment", 
           Icon:        "🌐",
           Params: map[string]interface{}{
               "profile_name":        "production",
               "profile_description": "Production environment",
               "profile_server":      "prod.example.com", 
               "profile_command":     "ssh user@prod.example.com",
           },
       },
       // ... other profiles
   }
   ```

3. **Automatic Parameter Handling**:
   - CLI args automatically parsed and passed to `directFunction`
   - Menu selection parameters automatically passed to `executeFunction`
   - No manual parameter extraction needed

4. **Simplified Action Mappings**:
   ```go
   // Before: Multiple WithInteractiveFunction calls
   WithActionMappings().
   AddInteractive(sdk.StreamingAction("basic", "Basic demo", "simple-example_basic-demo", basicDemo)).
   AddInteractive(sdk.StreamingAction("config", "Show configuration", "simple-example_config-demo", configDemo)).
   // ...
   ```

5. **Clean Menu Definition**:
   ```go
   WithBasicMenuProvider([]sdk.MenuEntry{
       {Key: "1", Label: "🎯 Basic Demo", Action: "core_interactive-go", Command: "simple-example_basic-demo"},
       {Key: "7", Label: "🔐 Profile Login", Action: "goto", Target: "profile-selector"},
       // ...
   })
   ```

### Boilerplate Reduction:

| Aspect | Before | After | Reduction |
|--------|--------|-------|-----------|
| Lines of code | ~250+ | ~180 | ~30% |
| CLI + Menu setup | ~50 lines | ~10 lines | ~80% |
| Parameter handling | Manual extraction | Automatic | ~100% |
| Function registration | Multiple calls | Single config | ~70% |
| Menu definitions | Complex provider | Simple array | ~60% |

### Functionality Maintained:

✅ **CLI Direct Execution**: `mops simple-example login production`  
✅ **Interactive Menu**: Navigate to profile selector and choose  
✅ **Parameter Validation**: Both CLI args and menu params validated  
✅ **Error Handling**: Proper error messages and validation  
✅ **Real-time Output**: Streaming output maintained  
✅ **All Demo Functions**: All existing functionality preserved  

### Benefits of New SDK Pattern:

1. **Generic & Reusable**: The `ArgumentBasedActionConfig` can be used by ANY plugin that needs CLI-with-args + menu selection
2. **Less Boilerplate**: Automatic parameter handling eliminates manual extraction code
3. **Type Safety**: Structured configuration reduces runtime errors
4. **Maintainable**: Clear separation of concerns between data definition and execution logic
5. **Consistent UX**: Unified behavior across all plugins using this pattern

### Usage Examples:

```bash
# CLI direct usage
mops simple-example login production
mops simple-example login staging
mops simple-example login development

# Interactive menu usage  
mops simple-example
# Select option 7 for Profile Login
# Choose from available profiles
```

## SDK Pattern Benefits

This demonstrates the power of the new SDK action helpers:

- **Any CLI tool** with arguments can now use `WithArgumentBasedAction`
- **Zero boilerplate** for parameter handling between CLI and menu modes
- **Consistent UX** across all plugins
- **Easy maintenance** with centralized configuration
- **Type safety** through structured configuration

The `ArgumentBasedActionConfig` pattern is truly **agnostic and reusable** for any future CLI-with-args actions.
