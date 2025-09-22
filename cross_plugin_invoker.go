package sdk

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// CrossPluginInvoker provides functionality to invoke functions from other plugins
type CrossPluginInvoker struct {
	outputChan chan<- string
	inputChan  <-chan string
	ctx        context.Context
}

// NewCrossPluginInvoker creates a new cross-plugin invoker for interactive functions
func NewCrossPluginInvoker(ctx context.Context, outputChan chan<- string, inputChan <-chan string) *CrossPluginInvoker {
	return &CrossPluginInvoker{
		outputChan: outputChan,
		inputChan:  inputChan,
		ctx:        ctx,
	}
}

// CrossPluginInvocationResult represents the result of a cross-plugin function call
type CrossPluginInvocationResult struct {
	Success bool
	Error   error
	Output  []string // Collected output lines
}

// InvokeFunction executes a function from another plugin
func (c *CrossPluginInvoker) InvokeFunction(functionName string, params map[string]interface{}) (*CrossPluginInvocationResult, error) {
	// Send a cross-plugin invocation command
	command := fmt.Sprintf("MOPS_CROSS_PLUGIN_INVOKE:%s", functionName)
	if len(params) > 0 {
		// Serialize parameters as a simple key=value format for now
		var paramPairs []string
		for key, value := range params {
			paramPairs = append(paramPairs, fmt.Sprintf("%s=%v", key, value))
		}
		command = fmt.Sprintf("%s?%s", command, strings.Join(paramPairs, "&"))
	}
	
	select {
	case c.outputChan <- command:
		// Command sent successfully
	case <-c.ctx.Done():
		return nil, fmt.Errorf("context cancelled while invoking cross-plugin function")
	}
	
	// Wait for response - the host will execute the function and return results
	result := &CrossPluginInvocationResult{
		Output: make([]string, 0),
	}
	
	for {
		select {
		case response := <-c.inputChan:
			if strings.HasPrefix(response, "MOPS_CROSS_PLUGIN_RESULT:") {
				// Success result
				result.Success = true
				return result, nil
			} else if strings.HasPrefix(response, "MOPS_CROSS_PLUGIN_ERROR:") {
				// Error result
				errorMsg := strings.TrimPrefix(response, "MOPS_CROSS_PLUGIN_ERROR:")
				result.Success = false
				result.Error = fmt.Errorf("cross-plugin invocation error: %s", errorMsg)
				return result, result.Error
			} else if strings.HasPrefix(response, "MOPS_CROSS_PLUGIN_OUTPUT:") {
				// Output line from the invoked function
				output := strings.TrimPrefix(response, "MOPS_CROSS_PLUGIN_OUTPUT:")
				result.Output = append(result.Output, output)
				// Also forward the output to our own output channel for real-time display
				c.outputChan <- output
			} else {
				// Unexpected response - put it back somehow or ignore
				GetPluginLogger().Debugf("Unexpected response during cross-plugin invocation: %s", response)
			}
		case <-time.After(30 * time.Second):
			return nil, fmt.Errorf("timeout waiting for cross-plugin function response")
		case <-c.ctx.Done():
			return nil, fmt.Errorf("context cancelled while waiting for cross-plugin function response")
		}
	}
}

// ListAvailableFunctions returns a list of all available functions from all plugins
func (c *CrossPluginInvoker) ListAvailableFunctions() ([]string, error) {
	// Send a list functions command
	command := "MOPS_CROSS_PLUGIN_LIST"
	
	select {
	case c.outputChan <- command:
		// Command sent successfully
	case <-c.ctx.Done():
		return nil, fmt.Errorf("context cancelled while listing functions")
	}
	
	// Wait for response
	select {
	case response := <-c.inputChan:
		if strings.HasPrefix(response, "MOPS_CROSS_PLUGIN_FUNCTIONS:") {
			// Success - extract function list (comma-separated)
			functionsStr := strings.TrimPrefix(response, "MOPS_CROSS_PLUGIN_FUNCTIONS:")
			if functionsStr == "" {
				return []string{}, nil
			}
			functions := strings.Split(functionsStr, ",")
			return functions, nil
		} else if strings.HasPrefix(response, "MOPS_CROSS_PLUGIN_ERROR:") {
			// Error - extract error message
			errorMsg := strings.TrimPrefix(response, "MOPS_CROSS_PLUGIN_ERROR:")
			return nil, fmt.Errorf("cross-plugin list error: %s", errorMsg)
		} else {
			// Unexpected response
			GetPluginLogger().Debugf("Unexpected response to ListAvailableFunctions: %s", response)
			return nil, fmt.Errorf("unexpected response from host")
		}
	case <-time.After(10 * time.Second):
		return nil, fmt.Errorf("timeout waiting for function list response")
	case <-c.ctx.Done():
		return nil, fmt.Errorf("context cancelled while waiting for function list response")
	}
}

// InvokeFunctionWithInput executes a function that may require user input
// This version handles interactive functions that prompt for user input
func (c *CrossPluginInvoker) InvokeFunctionWithInput(functionName string, params map[string]interface{}, inputResponses []string) (*CrossPluginInvocationResult, error) {
	// Similar to InvokeFunction but with pre-defined input responses
	command := fmt.Sprintf("MOPS_CROSS_PLUGIN_INVOKE_INTERACTIVE:%s", functionName)
	if len(params) > 0 {
		var paramPairs []string
		for key, value := range params {
			paramPairs = append(paramPairs, fmt.Sprintf("%s=%v", key, value))
		}
		command = fmt.Sprintf("%s?%s", command, strings.Join(paramPairs, "&"))
	}
	
	// Add input responses to the command
	if len(inputResponses) > 0 {
		command = fmt.Sprintf("%s|%s", command, strings.Join(inputResponses, "|"))
	}
	
	select {
	case c.outputChan <- command:
		// Command sent successfully
	case <-c.ctx.Done():
		return nil, fmt.Errorf("context cancelled while invoking interactive cross-plugin function")
	}
	
	// Similar response handling as InvokeFunction
	result := &CrossPluginInvocationResult{
		Output: make([]string, 0),
	}
	
	for {
		select {
		case response := <-c.inputChan:
			if strings.HasPrefix(response, "MOPS_CROSS_PLUGIN_RESULT:") {
				result.Success = true
				return result, nil
			} else if strings.HasPrefix(response, "MOPS_CROSS_PLUGIN_ERROR:") {
				errorMsg := strings.TrimPrefix(response, "MOPS_CROSS_PLUGIN_ERROR:")
				result.Success = false
				result.Error = fmt.Errorf("cross-plugin invocation error: %s", errorMsg)
				return result, result.Error
			} else if strings.HasPrefix(response, "MOPS_CROSS_PLUGIN_OUTPUT:") {
				output := strings.TrimPrefix(response, "MOPS_CROSS_PLUGIN_OUTPUT:")
				result.Output = append(result.Output, output)
				c.outputChan <- output
			}
		case <-time.After(60 * time.Second): // Longer timeout for interactive functions
			return nil, fmt.Errorf("timeout waiting for interactive cross-plugin function response")
		case <-c.ctx.Done():
			return nil, fmt.Errorf("context cancelled while waiting for interactive cross-plugin function response")
		}
	}
}