// Package output provides unified output adapters for MOPS plugins.
// This package enables plugins to write business logic once and support
// multiple output modes: CLI, UI static display, and real-time streaming.
package output

import (
	"time"
)

// Adapter defines the unified interface for different output mechanisms.
// Implementations handle the specifics of outputting to CLI, UI, streaming channels, etc.
type Adapter interface {
	// Basic output methods
	WriteLine(line string) error
	WriteError(err error) error
	
	// Structured output methods
	WriteHeader(title string) error
	WriteSection(section string) error
	WriteBulletPoint(point string) error
	WriteProgress(current, total int, message string) error
	WriteStats(stats map[string]interface{}) error
	
	// Control methods
	Flush() error
	SupportsRealTime() bool
}

// StaticAdapter extends Adapter for adapters that buffer output.
// Used for CLI and UI static displays.
type StaticAdapter interface {
	Adapter
	GetOutput() string
	Clear()
}

// StreamingAdapter extends Adapter for real-time streaming adapters.
// Used for UI interactive functions and live updates.
type StreamingAdapter interface {
	Adapter
	WriteRaw(content string) error
	WriteWithDelay(line string, delay time.Duration) error
}

// DelayConfig configures timing behavior for delayed adapters
type DelayConfig struct {
	SimulateDelays bool
	LineDelay      time.Duration
	ProgressDelay  time.Duration
	SectionDelay   time.Duration
}

// DefaultDelayConfig returns sensible defaults for delay configuration
func DefaultDelayConfig() DelayConfig {
	return DelayConfig{
		SimulateDelays: true,
		LineDelay:      200 * time.Millisecond,
		ProgressDelay:  300 * time.Millisecond,
		SectionDelay:   500 * time.Millisecond,
	}
}

// OutputStyle configures visual styling for adapters
type OutputStyle struct {
	EnableColors bool
	EnableEmojis bool
	EnableBold   bool
	Theme        string // "default", "minimal", "rich"
}

// DefaultOutputStyle returns sensible defaults for output styling
func DefaultOutputStyle() OutputStyle {
	return OutputStyle{
		EnableColors: true,
		EnableEmojis: true,
		EnableBold:   true,
		Theme:        "default",
	}
}

// MinimalOutputStyle returns a minimal style configuration
func MinimalOutputStyle() OutputStyle {
	return OutputStyle{
		EnableColors: false,
		EnableEmojis: false,
		EnableBold:   false,
		Theme:        "minimal",
	}
}

// RichOutputStyle returns a rich style configuration with all features
func RichOutputStyle() OutputStyle {
	return OutputStyle{
		EnableColors: true,
		EnableEmojis: true,
		EnableBold:   true,
		Theme:        "rich",
	}
}

// ProgressItem represents a single item in a progress sequence
type ProgressItem struct {
	Name        string
	Description string
	Processor   func() error
}

// StatsEntry represents a single statistics entry
type StatsEntry struct {
	Key   string
	Value interface{}
	Unit  string // optional unit like "ms", "MB", etc.
}
