# Hello World Plugin Example

A simple example plugin demonstrating the MOPS SDK capabilities.

## Features

- Basic action execution
- Interactive chat function
- Configuration management
- CLI command integration

## Build

```bash
make build
```

## Install

```bash
make install
```

## Usage

After installation, the plugin provides:

- **Hello Action**: Execute via MOPS menu or CLI
- **Interactive Chat**: Real-time chat interface
- **Configuration**: Customizable greeting message

### CLI Commands
```bash
# Direct plugin execution
./hello-world

# Via MOPS
mops hello-world hello
mops hello-world chat
```

## Configuration

Edit `config.yaml` to customize the greeting message:

```yaml
greeting_message: "Hello from my custom plugin!"
```
