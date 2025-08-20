# MOPS Plugin Development Guide

This guide explains how to create standalone MOPS plugins using the MOPS SDK in isolated repositories.

## Overview

The MOPS SDK allows you to develop plugins independently from the main MOPS repository. You can:

1. Create plugins in separate repositories
2. Use the SDK as a dependency
3. Build and distribute plugins independently
4. Publish plugins to plugin registries

## Project Setup

### 1. Initialize a New Plugin Project

```bash
# Create a new directory for your plugin
mkdir my-awesome-plugin
cd my-awesome-plugin

# Initialize Go module
go mod init my-awesome-plugin

# Add the MOPS SDK dependency
go get github.com/totmicro/mops-sdk@latest
```

### 2. Basic Project Structure

```
my-awesome-plugin/
├── main.go           # Plugin entry point
├── go.mod           # Go module with SDK dependency
├── plugin.yaml      # Plugin metadata (optional)
├── Makefile         # Build configuration
├── README.md        # Plugin documentation
├── internal/        # Internal plugin code (optional)
│   ├── core/        # Core plugin logic
│   ├── cli/         # CLI command handlers
│   └── ui/          # UI-related code
└── examples/        # Usage examples (optional)
```

### 3. Minimal Plugin Implementation

Create `main.go`:

```go
package main

import (
    "fmt"
    "github.com/totmicro/mops-sdk"
)

func main() {
    plugin := sdk.NewPluginBuilder("my-plugin", "1.0.0", "My awesome plugin").
        SetAuthor("Your Name").
        SetLicense("MIT").
        WithSimpleExecutor("my-action", func(entry sdk.MenuEntry, input string) sdk.ActionResult {
            return sdk.ActionResult{
                Success:    true,
                Output:     fmt.Sprintf("Hello from my plugin! Input: %s", input),
                ShowOutput: true,
            }
        }).
        Build()

    sdk.Main(plugin)
}
```

## Plugin Features

### Action Executors

Action executors handle specific action types triggered from menu entries:

```go
plugin.WithSimpleExecutor("my-action", func(entry sdk.MenuEntry, input string) sdk.ActionResult {
    // Process the action
    return sdk.ActionResult{
        Success:    true,
        Output:     "Action completed successfully",
        ShowOutput: true,
        Title:      "My Action",
    }
})
```

### Dynamic Providers

Dynamic providers generate menu entries at runtime:

```go
plugin.WithSimpleProvider("my-provider", "Provides dynamic entries", func(param string) ([]sdk.MenuEntry, error) {
    return []sdk.MenuEntry{
        {
            Key:    "1",
            Label:  "Dynamic Option 1",
            Action: "my-action",
            Params: map[string]interface{}{
                "option": "1",
            },
        },
        {
            Key:    "2",
            Label:  "Dynamic Option 2", 
            Action: "my-action",
            Params: map[string]interface{}{
                "option": "2",
            },
        },
    }, nil
})
```

### Interactive Functions

Interactive functions provide real-time streaming interaction:

```go
plugin.WithInteractiveFunction("my-interactive", func(ctx context.Context, outputChan chan<- string, inputChan <-chan string, params map[string]interface{}) error {
    outputChan <- "Welcome to my interactive function!"
    outputChan <- "Type something:"
    
    for {
        select {
        case input := <-inputChan:
            if input == "quit" {
                outputChan <- "Goodbye!"
                return nil
            }
            outputChan <- fmt.Sprintf("You typed: %s", input)
            
        case <-ctx.Done():
            return ctx.Err()
        }
    }
})
```

### CLI Commands

CLI commands extend the MOPS command-line interface:

```go
plugin.WithCLICommand("hello", "Print a greeting", func(args []string) error {
    name := "World"
    if len(args) > 0 {
        name = strings.Join(args, " ")
    }
    fmt.Printf("Hello, %s!\n", name)
    return nil
})
```

## Building and Distribution

### Local Development

```bash
# Build for current platform
go build -o my-plugin .

# Install locally
mkdir -p ~/.mops/plugins
cp my-plugin ~/.mops/plugins/
```

### Multi-platform Build

Create a `Makefile`:

```makefile
.PHONY: build-all clean

PLUGIN_NAME := my-plugin

build-all:
	# Linux AMD64
	GOOS=linux GOARCH=amd64 go build -o $(PLUGIN_NAME)-linux-amd64 .
	
	# macOS AMD64 (Intel)
	GOOS=darwin GOARCH=amd64 go build -o $(PLUGIN_NAME)-darwin-amd64 .
	
	# macOS ARM64 (Apple Silicon)
	GOOS=darwin GOARCH=arm64 go build -o $(PLUGIN_NAME)-darwin-arm64 .
	
	# Windows AMD64
	GOOS=windows GOARCH=amd64 go build -o $(PLUGIN_NAME)-windows-amd64.exe .

clean:
	rm -f $(PLUGIN_NAME)*
```

### Plugin Metadata

Create `plugin.yaml` for metadata:

```yaml
name: my-plugin
version: 1.0.0
description: "My awesome plugin"
author: "Your Name"
license: "MIT"
homepage: "https://github.com/yourname/my-plugin"
tags:
  - "utility"
  - "example"

mops_version:
  min_version: "1.0.0"
  max_version: "2.0.0"

build_targets:
  - os: linux
    arch: amd64
    output: my-plugin-linux-amd64
  - os: darwin
    arch: amd64
    output: my-plugin-darwin-amd64
  - os: windows
    arch: amd64
    output: my-plugin-windows-amd64.exe

default_config:
  enabled: true
  timeout: 30

cli_commands:
  - name: hello
    description: "Print a greeting"
    usage: "mops plugin my-plugin hello [name]"
```

## Publishing

### GitHub Releases

1. Tag your release:
```bash
git tag v1.0.0
git push origin v1.0.0
```

2. Create a GitHub release with built binaries
3. Users can install via:
```bash
mops plugin install github.com/yourname/my-plugin@v1.0.0
```

### Plugin Registry

Submit your plugin to the MOPS plugin registry:

1. Fork the registry repository
2. Add your plugin metadata
3. Submit a pull request

## Best Practices

### Code Organization

- Use internal packages for complex logic
- Separate CLI handlers from core functionality  
- Provide comprehensive error handling
- Include unit tests

### Configuration

- Use the default configuration system
- Validate configuration on initialization
- Provide sensible defaults

### Documentation

- Include clear README with usage examples
- Document all CLI commands
- Provide plugin.yaml metadata

### Testing

```go
func TestMyAction(t *testing.T) {
    plugin := createTestPlugin()
    
    result := plugin.Execute(sdk.MenuEntry{
        Action: "my-action",
    }, "test input")
    
    assert.True(t, result.Success)
    assert.Contains(t, result.Output, "test input")
}
```

## Examples

Check the SDK repository for complete examples:

- `examples/hello-world/` - Basic plugin with all features
- `examples/file-manager/` - File management plugin
- `examples/api-client/` - API integration plugin
- `examples/monitoring/` - System monitoring plugin

## Support

- GitHub Issues: https://github.com/totmicro/mops-sdk/issues
- Documentation: https://docs.mops.dev/plugins
- Community: https://discord.gg/mops
