package sdk

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Shared storage markers for interactive function communication
const (
	SharedStorageSetMarker    = "MOPS_SHARED_SET:"
	SharedStorageGetMarker    = "MOPS_SHARED_GET:"
	SharedStorageListMarker   = "MOPS_SHARED_LIST"
	SharedStorageResultMarker = "MOPS_SHARED_RESULT:"
	SharedStorageErrorMarker  = "MOPS_SHARED_ERROR:"
)

// InteractiveSharedStorage provides shared storage access during interactive function execution
type InteractiveSharedStorage struct {
	outputChan chan<- string
	inputChan  <-chan string
	ctx        context.Context
}

// NewInteractiveSharedStorage creates a new shared storage accessor for interactive functions
func NewInteractiveSharedStorage(ctx context.Context, outputChan chan<- string, inputChan <-chan string) *InteractiveSharedStorage {
	return &InteractiveSharedStorage{
		outputChan: outputChan,
		inputChan:  inputChan,
		ctx:        ctx,
	}
}

// SetVar sets a shared variable during interactive function execution
func (s *InteractiveSharedStorage) SetVar(key, value string) error {
	// Send set command via output channel
	command := fmt.Sprintf("%s%s:%s", SharedStorageSetMarker, key, value)
	
	select {
	case s.outputChan <- command:
		// Command sent successfully
	case <-s.ctx.Done():
		return fmt.Errorf("context cancelled while setting shared variable")
	}
	
	// Wait for response
	select {
	case response := <-s.inputChan:
		if strings.HasPrefix(response, SharedStorageResultMarker) {
			// Success - extract result if any
			return nil
		} else if strings.HasPrefix(response, SharedStorageErrorMarker) {
			// Error - extract error message
			errorMsg := strings.TrimPrefix(response, SharedStorageErrorMarker)
			return fmt.Errorf("shared storage error: %s", errorMsg)
		} else {
			// Unexpected response - put it back somehow or ignore
			GetPluginLogger().Debugf("Unexpected response to SetVar: %s", response)
			return fmt.Errorf("unexpected response from host")
		}
	case <-time.After(5 * time.Second):
		return fmt.Errorf("timeout waiting for shared storage response")
	case <-s.ctx.Done():
		return fmt.Errorf("context cancelled while waiting for shared storage response")
	}
}

// GetVar gets a shared variable during interactive function execution
func (s *InteractiveSharedStorage) GetVar(key string) (string, error) {
	// Send get command via output channel
	command := fmt.Sprintf("%s%s", SharedStorageGetMarker, key)
	
	select {
	case s.outputChan <- command:
		// Command sent successfully
	case <-s.ctx.Done():
		return "", fmt.Errorf("context cancelled while getting shared variable")
	}
	
	// Wait for response
	select {
	case response := <-s.inputChan:
		if strings.HasPrefix(response, SharedStorageResultMarker) {
			// Success - extract value
			value := strings.TrimPrefix(response, SharedStorageResultMarker)
			return value, nil
		} else if strings.HasPrefix(response, SharedStorageErrorMarker) {
			// Error - extract error message
			errorMsg := strings.TrimPrefix(response, SharedStorageErrorMarker)
			return "", fmt.Errorf("shared storage error: %s", errorMsg)
		} else {
			// Unexpected response
			GetPluginLogger().Debugf("Unexpected response to GetVar: %s", response)
			return "", fmt.Errorf("unexpected response from host")
		}
	case <-time.After(5 * time.Second):
		return "", fmt.Errorf("timeout waiting for shared storage response")
	case <-s.ctx.Done():
		return "", fmt.Errorf("context cancelled while waiting for shared storage response")
	}
}

// ListVars lists all shared variables during interactive function execution
func (s *InteractiveSharedStorage) ListVars() ([]string, error) {
	// Send list command via output channel
	command := SharedStorageListMarker
	
	select {
	case s.outputChan <- command:
		// Command sent successfully
	case <-s.ctx.Done():
		return nil, fmt.Errorf("context cancelled while listing shared variables")
	}
	
	// Wait for response
	select {
	case response := <-s.inputChan:
		if strings.HasPrefix(response, SharedStorageResultMarker) {
			// Success - extract keys (comma-separated)
			keysStr := strings.TrimPrefix(response, SharedStorageResultMarker)
			if keysStr == "" {
				return []string{}, nil
			}
			keys := strings.Split(keysStr, ",")
			return keys, nil
		} else if strings.HasPrefix(response, SharedStorageErrorMarker) {
			// Error - extract error message
			errorMsg := strings.TrimPrefix(response, SharedStorageErrorMarker)
			return nil, fmt.Errorf("shared storage error: %s", errorMsg)
		} else {
			// Unexpected response
			GetPluginLogger().Debugf("Unexpected response to ListVars: %s", response)
			return nil, fmt.Errorf("unexpected response from host")
		}
	case <-time.After(5 * time.Second):
		return nil, fmt.Errorf("timeout waiting for shared storage response")
	case <-s.ctx.Done():
		return nil, fmt.Errorf("context cancelled while waiting for shared storage response")
	}
}