// Package output provides unified output adapters for MOPS plugins.
//
// This package enables plugin developers to write business logic once and support
// multiple output modes automatically:
//
//   - CLI: Command-line interface with colors and formatting
//   - UI Static: Buffered output for UI action results
//   - UI Streaming: Real-time output for interactive functions
//
// Basic Usage:
//
//	// For CLI commands
//	func (p *MyPlugin) CLICommand(args []string) (string, error) {
//		adapter := output.NewCLIAdapter(true) // with colors
//		err := p.runBusinessLogic(adapter, parseArgs(args))
//		return adapter.GetOutput(), err
//	}
//
//	// For UI streaming
//	func (p *MyPlugin) StreamingFunction(ctx context.Context, outputChan chan<- string, ...) error {
//		adapter := output.NewStreamingAdapter(ctx, outputChan)
//		return p.runBusinessLogic(adapter, params)
//	}
//
//	// Business logic (shared across all modes)
//	func (p *MyPlugin) runBusinessLogic(out output.Adapter, params map[string]interface{}) error {
//		out.WriteHeader("🎯 My Plugin Demo")
//		out.WriteSection("Processing items...")
//
//		for i, item := range items {
//			out.WriteProgress(i+1, len(items), fmt.Sprintf("Processing %s", item))
//			// ... business logic
//			out.WriteBulletPoint(fmt.Sprintf("✅ Completed %s", item))
//		}
//
//		return out.Flush()
//	}
//
// Advanced Features:
//
//   - Delayed adapters for animated effects
//   - Custom styling and theming
//   - Helper utilities for common patterns
//   - Progress sequences and statistics tables
//
package output

import (
	"context"
	"fmt"
)

// Factory Functions - Convenience constructors for common use cases

// NewForCLI creates an adapter optimized for command-line interfaces
func NewForCLI(enableColors bool) StaticAdapter {
	return NewCLIAdapter(enableColors)
}

// NewForUI creates an adapter optimized for UI static displays
func NewForUI() StaticAdapter {
	return NewStaticAdapterWithStyle(DefaultOutputStyle())
}

// NewForUIMinimal creates an adapter with minimal styling for UI
func NewForUIMinimal() StaticAdapter {
	return NewStaticAdapterWithStyle(MinimalOutputStyle())
}

// NewForStreaming creates an adapter for real-time streaming output
func NewForStreaming(ctx context.Context, outputChan chan<- string) StreamingAdapter {
	return NewStreamingAdapter(ctx, outputChan)
}

// NewForStreamingAnimated creates a streaming adapter with animation delays
func NewForStreamingAnimated(ctx context.Context, outputChan chan<- string) Adapter {
	streaming := NewStreamingAdapter(ctx, outputChan)
	return NewDelayedAdapterWithDefaults(streaming, false) // enable delays
}

// Plugin Helper Functions - Common patterns for plugin development

// ExecuteWithProgress runs a function with progress feedback.
// This is a common pattern for plugins that process multiple items.
func ExecuteWithProgress(adapter Adapter, title string, items []string, processor func(string) error) error {
	progressItems := make([]ProgressItem, len(items))
	for i, item := range items {
		item := item // capture loop variable
		progressItems[i] = ProgressItem{
			Name:      item,
			Processor: func() error { return processor(item) },
		}
	}
	
	return WriteProgressSequence(adapter, title, progressItems)
}

// ShowResults displays results in a consistent format across all adapters
func ShowResults(adapter Adapter, title string, results map[string]interface{}, success bool) error {
	if err := adapter.WriteLine(""); err != nil {
		return err
	}
	
	// Status indicator
	var statusLine string
	if success {
		statusLine = "✅ " + title + " completed successfully!"
	} else {
		statusLine = "❌ " + title + " failed"
	}
	
	if err := adapter.WriteLine(statusLine); err != nil {
		return err
	}
	
	// Results
	if len(results) > 0 {
		if err := adapter.WriteLine(""); err != nil {
			return err
		}
		if err := adapter.WriteStats(results); err != nil {
			return err
		}
	}
	
	return adapter.Flush()
}

// Demo Framework - Helper for creating consistent demos across plugins

// DemoStep represents a single step in a demonstration
type DemoStep struct {
	Name        string
	Description string
	Action      func(Adapter) error
}

// RunDemo executes a sequence of demo steps with consistent formatting
func RunDemo(adapter Adapter, title string, steps []DemoStep) error {
	// Header
	if err := adapter.WriteHeader(title); err != nil {
		return err
	}
	if err := adapter.WriteLine(""); err != nil {
		return err
	}
	
	// Execute steps
	for i, step := range steps {
		// Show progress
		if err := adapter.WriteProgress(i+1, len(steps), step.Name); err != nil {
			return err
		}
		
		// Execute step
		if step.Action != nil {
			if err := step.Action(adapter); err != nil {
				if err := adapter.WriteError(fmt.Errorf("step %s failed: %w", step.Name, err)); err != nil {
					return err
				}
				continue
			}
		}
		
		// Show completion
		completionMsg := fmt.Sprintf("✅ %s", step.Name)
		if step.Description != "" {
			completionMsg += fmt.Sprintf(" - %s", step.Description)
		}
		if err := adapter.WriteBulletPoint(completionMsg); err != nil {
			return err
		}
	}
	
	if err := adapter.WriteLine(""); err != nil {
		return err
	}
	if err := adapter.WriteLine("🎉 Demo completed successfully!"); err != nil {
		return err
	}
	
	return adapter.Flush()
}
