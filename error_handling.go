package sdk

import (
	"context"
	"fmt"
)

// ErrorType represents different categories of errors in the plugin system
type ErrorType int

const (
	// ErrTypeCommandFailure represents a command/operation that failed but doesn't affect plugin health
	ErrTypeCommandFailure ErrorType = iota
	// ErrTypePluginFailure represents a critical plugin error that should make the plugin unavailable
	ErrTypePluginFailure
	// ErrTypeSystemFailure represents a system-level error
	ErrTypeSystemFailure
)

// PluginError represents a classified error that can be handled appropriately
type PluginError struct {
	Type    ErrorType
	Message string
	Cause   error
}

func (pe *PluginError) Error() string {
	if pe.Cause != nil {
		return fmt.Sprintf("%s: %v", pe.Message, pe.Cause)
	}
	return pe.Message
}

// IsPluginFailure checks if this error should cause plugin unavailability
func (pe *PluginError) IsPluginFailure() bool {
	return pe.Type == ErrTypePluginFailure || pe.Type == ErrTypeSystemFailure
}

// NewCommandError creates an error for command failures that shouldn't affect plugin availability
func NewCommandError(message string, cause error) *PluginError {
	return &PluginError{
		Type:    ErrTypeCommandFailure,
		Message: message,
		Cause:   cause,
	}
}

// NewPluginError creates an error that indicates plugin failure
func NewPluginError(message string, cause error) *PluginError {
	return &PluginError{
		Type:    ErrTypePluginFailure,
		Message: message,
		Cause:   cause,
	}
}

// NewSystemError creates an error that indicates system failure
func NewSystemError(message string, cause error) *PluginError {
	return &PluginError{
		Type:    ErrTypeSystemFailure,
		Message: message,
		Cause:   cause,
	}
}

// ResilientFunction wraps an InteractiveGoFunction to handle command errors gracefully
// 
// DEPRECATED: This wrapper is now redundant since the system-level safety net
// automatically handles all plugin function errors. The system-level protection
// is more comprehensive and handles panics as well. This wrapper is kept for
// backward compatibility only.
type ResilientFunction struct {
	fn InteractiveGoFunction
}

// NewResilientFunction creates a wrapper that handles command errors gracefully
//
// DEPRECATED: System-level safety net now handles all errors automatically.
// You can now return errors directly from functions - they will be converted
// to user messages automatically without crashing the plugin.
func NewResilientFunction(fn InteractiveGoFunction) *ResilientFunction {
	return &ResilientFunction{fn: fn}
}

// Execute runs the function and handles errors according to their type
// 
// DEPRECATED: The system-level safety net now provides superior error handling
// that includes panic recovery and automatic error conversion.
func (rf *ResilientFunction) Execute(ctx context.Context, outputChan chan<- string, inputChan <-chan string, params map[string]interface{}) error {
	// With system-level safety net, we can just call the function directly
	// The system will handle any errors or panics automatically
	return rf.fn(ctx, outputChan, inputChan, params)
}

// GetFunction returns the wrapped InteractiveGoFunction for compatibility
func (rf *ResilientFunction) GetFunction() InteractiveGoFunction {
	return rf.Execute
}