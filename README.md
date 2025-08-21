# MOPS SDK

The MOPS SDK provides a simple and powerful way to create plugins for the MOPS (Modular Operations Platform System).

## Overview

The SDK enables developers to create plugins that integrate seamlessly with MOPS, providing:

- Dynamic menu entries
- Custom action execution
- Interactive streaming functions
- CLI command extensions
- Configuration management

## Installation

```bash
go get github.com/totmicro/mops-sdk
```

## Quick Start

Create a simple plugin:

```go
package main

import (
    "context"
    "github.com/totmicro/mops-sdk"
)

func main() {
    plugin := sdk.NewPluginBuilder("my-plugin", "1.0.0", "My first MOPS plugin").
        AddExecutor("hello", func(ctx context.Context, params map[string]interface{}) (string, error) {
            return "Hello from my plugin!", nil
        }).
        Build()

    plugin.Start()
}
```

## Features

- **Plugin Builder**: Fluent API for plugin construction
- **Action Executors**: Execute custom business logic
- **Menu Providers**: Dynamic menu generation
- **Interactive Functions**: Real-time user interaction
- **CLI Commands**: Extend MOPS command-line interface
- **Configuration**: Plugin-specific configuration management

## Examples

See the `examples/` directory for complete plugin examples:
- `hello-world/` - Basic plugin with actions and interactive functions

## Development

### Building a Plugin
```bash
go build -o my-plugin .
```

### Testing with MOPS
```bash
# Copy plugin to MOPS plugins directory
cp my-plugin ~/.mops/plugins/

# Run MOPS
mops
```

## Documentation

- Plugin interface definitions in `plugin.go`
- RPC implementation in `rpc.go`
- Base plugin utilities in `base.go`
- Builder pattern in `builder.go`

## License

MIT License - See LICENSE file for details.
)

func main() {
    plugin := sdk.NewPluginBuilder("env-manager", "1.0.0", "A simple environment manager plugin").
        SetAuthor("Your Name").
        SetLicense("MIT").
        WithSimpleExecutor("hello", func(entry sdk.MenuEntry, input string) sdk.ActionResult {
            name := input
            if name == "" {
                name = "World"
            }
            return sdk.ActionResult{
                Success:    true,
                Output:     fmt.Sprintf("Hello, %s!", name),
                ShowOutput: true,
            }
        }).
        Build()

    sdk.Main(plugin)
}
```

## Plugin Builder API

### Creating a Plugin

```go
plugin := sdk.NewPluginBuilder("plugin-name", "1.0.0", "Plugin description")
```

### Setting Metadata

```go
plugin.SetAuthor("Author Name").
       SetLicense("MIT").
       SetHomepage("https://github.com/user/plugin").
       SetMopsVersions("1.0.0", "2.0.0").
       AddTag("utility").
       AddDependency("some-other-plugin")
```

### Adding Functionality

#### Simple Action Executor

```go
plugin.WithSimpleExecutor("my-action", func(entry sdk.MenuEntry, input string) sdk.ActionResult {
    return sdk.ActionResult{
        Success:    true,
        Output:     "Action executed successfully",
        ShowOutput: true,
    }
})
```

#### Dynamic Provider

```go
plugin.WithSimpleProvider("my-provider", "Provides dynamic entries", func(param string) ([]sdk.MenuEntry, error) {
    return []sdk.MenuEntry{
        {
            Key:    "1",
            Label:  "Option 1",
            Action: "my-action",
        },
        {
            Key:    "2", 
            Label:  "Option 2",
            Action: "my-action",
        },
    }, nil
})
```

#### Interactive Function

```go
plugin.WithInteractiveFunction("my-interactive", func(ctx context.Context, outputChan chan<- string, inputChan <-chan string, params map[string]interface{}) error {
    outputChan <- "Enter your name:"
    
    select {
    case name := <-inputChan:
        outputChan <- fmt.Sprintf("Hello, %s!", name)
    case <-ctx.Done():
        return ctx.Err()
    }
    
    return nil
})
```

#### CLI Command

```go
plugin.WithCLICommand("hello", "Print a greeting", func(args []string) error {
    name := "World"
    if len(args) > 0 {
        name = args[0]
    }
    fmt.Printf("Hello, %s!\n", name)
    return nil
})
```

## Plugin Structure

A typical plugin project structure:

```
my-plugin/
├── main.go           # Plugin entry point
├── go.mod           # Go module file
├── plugin.yaml      # Plugin metadata (optional)
├── Makefile         # Build configuration
├── README.md        # Plugin documentation
└── internal/        # Internal plugin code
    ├── core/        # Core plugin logic
    ├── cli/         # CLI command handlers
    └── ui/          # UI-related code
```

## Building and Distribution

### Local Development

```bash
go build -o my-plugin .
mkdir -p ~/.mops/plugins
cp my-plugin ~/.mops/plugins/
```

### Multi-platform Build

```bash
# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o my-plugin-linux-amd64 .

# macOS AMD64  
GOOS=darwin GOARCH=amd64 go build -o my-plugin-darwin-amd64 .

# Windows AMD64
GOOS=windows GOARCH=amd64 go build -o my-plugin-windows-amd64.exe .
```

## Plugin Metadata (plugin.yaml)

```yaml
name: my-plugin
version: 1.0.0
description: "My awesome plugin"
author: "Your Name"
license: "MIT"
homepage: "https://github.com/user/my-plugin"
tags:
  - "utility"
  - "example"

# MOPS version compatibility
mops_version:
  min_version: "1.0.0"
  max_version: "2.0.0"

# Build targets
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

# Default configuration
default_config:
  enabled: true
  timeout: 30

# CLI commands
cli_commands:
  - name: hello
    description: "Print a greeting"
    usage: "mops plugin my-plugin hello [name]"
    examples:
      - "mops plugin my-plugin hello"
      - "mops plugin my-plugin hello World"
```

## Examples

Check the `examples/` directory for complete plugin examples:

- `env-manager/` - Basic plugin with action executor
- `file-manager/` - Plugin with dynamic providers
- `interactive-demo/` - Plugin with interactive functions
- `cli-tools/` - Plugin with CLI commands

## API Reference

### Core Types

- `Plugin` - Main plugin interface
- `PluginInfo` - Plugin metadata
- `ActionResult` - Result of action execution
- `MenuEntry` - Menu entry definition
- `DynamicProvider` - Interface for dynamic menu providers
- `ActionExecutor` - Interface for action executors
- `InteractiveGoFunction` - Interactive function type

### Plugin Builder Methods

- `NewPluginBuilder(name, version, description)` - Create new builder
- `SetAuthor(author)` - Set plugin author
- `SetLicense(license)` - Set plugin license
- `SetHomepage(url)` - Set plugin homepage
- `AddTag(tag)` - Add plugin tag
- `WithSimpleExecutor(type, handler)` - Add action executor
- `WithSimpleProvider(name, description, handler)` - Add dynamic provider
- `WithInteractiveFunction(name, handler)` - Add interactive function
- `WithCLICommand(name, description, handler)` - Add CLI command
- `Build()` - Build the plugin

## License

MIT License - see LICENSE file for details.
