# Hello World Standalone Plugin

This is a standalone example plugin built using the MOPS SDK. It demonstrates how to create a plugin independently from the main MOPS repository.

## Features

- **Simple Action Executor**: Executes a "hello" action with customizable greetings
- **Dynamic Provider**: Provides greeting options in multiple languages
- **Interactive Function**: Real-time chat interface
- **CLI Commands**: Command-line interface for direct plugin interaction

## Building

### Quick Build

```bash
make build
```

### Multi-platform Build

```bash
make build-all
```

This will create binaries for:
- Linux AMD64/ARM64
- macOS AMD64/ARM64 (Intel/Apple Silicon)
- Windows AMD64

### Development

```bash
# Install dependencies
make deps

# Format code
make format

# Run tests
make test

# Run the plugin directly
make run
```

## Installation

### Local Installation

```bash
make install
```

This installs the plugin to `~/.mops/plugins/` where MOPS can find it.

### Manual Installation

```bash
# Build the plugin
go build -o hello-world-standalone .

# Copy to MOPS plugins directory
mkdir -p ~/.mops/plugins
cp hello-world-standalone ~/.mops/plugins/
```

## Usage

Once installed, the plugin will be automatically loaded by MOPS and its functionality will be available in the menu system.

### CLI Commands

The plugin also provides CLI commands that can be executed directly:

```bash
# Print a hello message
mops plugin hello-world-standalone hello [name]

# Greet in different languages
mops plugin hello-world-standalone greet en World
mops plugin hello-world-standalone greet es Mundo

# Show plugin status
mops plugin hello-world-standalone status
mops plugin hello-world-standalone status --verbose
```

## Plugin Features

### Action Executor

The plugin registers a "hello" action that can be triggered from menu entries. It accepts parameters to customize the greeting.

### Dynamic Provider

The "greetings" provider generates menu entries for different language greetings:

- English: Hello 👋
- Spanish: Hola 🇪🇸
- French: Bonjour 🇫🇷
- German: Guten Tag 🇩🇪
- Italian: Ciao 🇮🇹

### Interactive Function

The "chat" interactive function provides a real-time chat interface where users can interact with the plugin.

## Configuration

The plugin supports configuration through the MOPS configuration system:

```yaml
plugins:
  hello-world-standalone:
    greeting: "Hello"
    enable_colors: true
    timeout: 30
```

## Development

This plugin serves as a template for creating standalone MOPS plugins. Key files:

- `main.go` - Plugin implementation using the SDK
- `plugin.yaml` - Plugin metadata and configuration
- `go.mod` - Go module with SDK dependency
- `Makefile` - Build and development commands

## License

MIT License
