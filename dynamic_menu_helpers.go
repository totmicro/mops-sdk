package sdk

import (
	"fmt"
	"strconv"
)

// MenuEntryConfig represents configuration for generating a single menu entry
type MenuEntryConfig struct {
	Key     string
	Label   string
	Action  string
	Command string
	Target  string
	Message string
	Params  map[string]interface{}
}

// DynamicMenuConfig represents configuration for generating dynamic menus
type DynamicMenuConfig struct {
	MaxEntries       int
	EnableExitButton bool
	ExitKey          string
	ExitLabel        string
	EmptyMessage     string
	EmptyKey         string
}

// DefaultDynamicMenuConfig returns a default configuration for dynamic menus
func DefaultDynamicMenuConfig() DynamicMenuConfig {
	return DynamicMenuConfig{
		MaxEntries:       9,
		EnableExitButton: true,
		ExitKey:          "x",
		ExitLabel:        "❌ Exit",
		EmptyMessage:     "No entries found",
		EmptyKey:         "e",
	}
}

// CLIDynamicMenuConfig returns a CLI-optimized configuration for dynamic menus
func CLIDynamicMenuConfig() DynamicMenuConfig {
	config := DefaultDynamicMenuConfig()
	config.ExitKey = "q"
	config.ExitLabel = "🚪 Exit"
	return config
}

// GenerateAutoNumberedEntries creates auto-numbered menu entries from a list of entry configs
func GenerateAutoNumberedEntries(entries []MenuEntryConfig, config DynamicMenuConfig) []MenuEntry {
	var menuEntries []MenuEntry

	// Add auto-numbered entries (1-9)
	for i, entryConfig := range entries {
		if i >= config.MaxEntries {
			break
		}

		// Use auto-generated key if not specified
		key := entryConfig.Key
		if key == "" {
			key = strconv.Itoa(i + 1)
		}

		menuEntries = append(menuEntries, MenuEntry{
			Key:     key,
			Label:   entryConfig.Label,
			Action:  entryConfig.Action,
			Command: entryConfig.Command,
			Target:  entryConfig.Target,
			Message: entryConfig.Message,
			Params:  entryConfig.Params,
		})
	}

	// Add empty message if no entries found
	if len(entries) == 0 {
		menuEntries = append(menuEntries, MenuEntry{
			Key:     config.EmptyKey,
			Label:   fmt.Sprintf("📭 %s", config.EmptyMessage),
			Action:  "noop",
			Message: config.EmptyMessage,
		})
	}

	// Add exit button if enabled
	if config.EnableExitButton {
		menuEntries = append(menuEntries, MenuEntry{
			Key:     config.ExitKey,
			Label:   config.ExitLabel,
			Action:  "exit",
			Message: "Exit the current menu",
		})
	}

	return menuEntries
}