package sdk

import (
	"context"
	"fmt"
)

// inputRequesterImpl implements the InputRequester interface
type inputRequesterImpl struct {
	ctx         context.Context
	outputChan  chan<- string
	inputChan   <-chan string
}

// NewInputRequester creates a new InputRequester instance
func NewInputRequester(ctx context.Context, outputChan chan<- string, inputChan <-chan string) InputRequester {
	return &inputRequesterImpl{
		ctx:        ctx,
		outputChan: outputChan,
		inputChan:  inputChan,
	}
}

// NewInputRequesterWithSession creates a new InputRequester instance with session ID
func NewInputRequesterWithSession(ctx context.Context, outputChan chan<- string, inputChan <-chan string, sessionID string) InputRequester {
	// For now, ignore the session ID as we're using the marker approach
	return &inputRequesterImpl{
		ctx:        ctx,
		outputChan: outputChan,
		inputChan:  inputChan,
	}
}

// RequestInput sends a prompt and waits for user input
func (r *inputRequesterImpl) RequestInput(prompt string) (string, error) {
	return r.RequestInputWithDefault(prompt, "")
}

// RequestInputWithDefault sends a prompt with a default value
func (r *inputRequesterImpl) RequestInputWithDefault(prompt string, defaultValue string) (string, error) {
	// Send the prompt as regular output (no special marker needed for now)
	fullPrompt := prompt
	if defaultValue != "" {
		fullPrompt = fmt.Sprintf("%s (default: %s)", prompt, defaultValue)
	}
	
	select {
	case r.outputChan <- fullPrompt:
		// Prompt sent successfully, now wait for user input
	case <-r.ctx.Done():
		return "", r.ctx.Err()
	}
	
	// Wait for user input
	select {
	case input := <-r.inputChan:
		// If input is empty and we have a default, use the default
		if input == "" && defaultValue != "" {
			return defaultValue, nil
		}
		return input, nil
	case <-r.ctx.Done():
		return "", r.ctx.Err()
	}
}
