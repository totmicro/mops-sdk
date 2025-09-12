package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/rpc"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/go-plugin"
)

const (
	PluginName      = "mops-plugin"
	ProtocolVersion = 1
)

// PluginMap is the map of plugins we can dispense.
var PluginMap = map[string]plugin.Plugin{
	PluginName: &MopsPlugin{},
}

// MopsPlugin is the implementation of plugin.Plugin so we can serve/consume this
type MopsPlugin struct {
	// Impl Injection
	Impl Plugin
}

func (p *MopsPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &PluginRPCServer{Impl: p.Impl}, nil
}

func (p *MopsPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &PluginRPCClient{client: c}, nil
}

// PluginRPCClient is an implementation of Plugin that talks over RPC.
type PluginRPCClient struct {
	client *rpc.Client
}

func (m *PluginRPCClient) GetInfo() (*PluginInfo, error) {
	var resp PluginInfoRPC
	err := m.client.Call("Plugin.GetInfo", new(interface{}), &resp)
	if err != nil {
		return nil, err
	}
	
	info, err := resp.ToPluginInfo()
	if err != nil {
		return nil, err
	}
	
	return &info, nil
}

func (g *PluginRPCClient) Initialize(config map[string]any) error {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	var resp error
	err = g.client.Call("Plugin.Initialize", string(configJSON), &resp)
	if err != nil {
		return err
	}
	return resp
}

func (g *PluginRPCClient) ReloadConfig(config map[string]any) error {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	var resp error
	err = g.client.Call("Plugin.ReloadConfig", string(configJSON), &resp)
	if err != nil {
		return err
	}
	return resp
}

func (g *PluginRPCClient) RegisterProviders() []DynamicProvider {
	var resp []DynamicProviderRPC
	err := g.client.Call("Plugin.RegisterProviders", new(interface{}), &resp)
	if err != nil {
		return nil
	}

	providers := make([]DynamicProvider, len(resp))
	for i, p := range resp {
		providers[i] = &DynamicProviderRPCClient{
			client:      g.client,
			name:        p.Name,
			description: p.Description,
		}
	}
	return providers
}

func (g *PluginRPCClient) RegisterInteractiveFunctions() map[string]InteractiveGoFunction {
	var resp map[string]string
	err := g.client.Call("Plugin.RegisterInteractiveFunctions", new(interface{}), &resp)
	if err != nil {
		return nil
	}

	functions := make(map[string]InteractiveGoFunction)
	for name := range resp {
		funcName := name
		functions[name] = func(ctx context.Context, outputChan chan<- string, inputChan <-chan string, params map[string]interface{}) error {
			return g.callInteractiveFunction(ctx, funcName, outputChan, inputChan, params)
		}
	}
	return functions
}

func (g *PluginRPCClient) callInteractiveFunction(ctx context.Context, name string, outputChan chan<- string, inputChan <-chan string, params map[string]interface{}) error {
	// Use the new streaming RPC implementation
	return g.callInteractiveFunctionStreaming(ctx, name, outputChan, inputChan, params)
}

// callInteractiveFunctionStreaming implements proper bidirectional streaming over RPC
func (g *PluginRPCClient) callInteractiveFunctionStreaming(ctx context.Context, name string, outputChan chan<- string, inputChan <-chan string, params map[string]interface{}) error {
	// Step 1: Initialize the interactive session
	initArgs := &InteractiveStreamInitArgs{
		FunctionName: name,
		Params:       params,
	}
	
	var initResp InteractiveStreamInitResponse
	err := g.client.Call("Plugin.InitInteractiveStream", initArgs, &initResp)
	if err != nil {
		return fmt.Errorf("failed to initialize interactive stream: %w", err)
	}
	
	if initResp.Error != nil {
		return initResp.Error
	}
	
	sessionID := initResp.SessionID
	defer func() {
		// Clean up the session
		cleanupArgs := &InteractiveStreamCleanupArgs{SessionID: sessionID}
		var cleanupResp InteractiveStreamCleanupResponse
		g.client.Call("Plugin.CleanupInteractiveStream", cleanupArgs, &cleanupResp)
	}()
	
	// Step 2: Start concurrent goroutines for input and output handling
	errChan := make(chan error, 2)
	done := make(chan bool, 1)
	
	// Output polling goroutine
	go func() {
		defer func() { done <- true }()
		
		for {
			select {
			case <-ctx.Done():
				errChan <- ctx.Err()
				return
			default:
				// Poll for output
				outputArgs := &InteractiveStreamOutputArgs{SessionID: sessionID}
				var outputResp InteractiveStreamOutputResponse
				
				err := g.client.Call("Plugin.GetInteractiveStreamOutput", outputArgs, &outputResp)
				if err != nil {
					errChan <- fmt.Errorf("failed to get stream output: %w", err)
					return
				}
				
				// Send any new output
				for _, line := range outputResp.NewOutput {
					select {
					case outputChan <- line:
					case <-ctx.Done():
						errChan <- ctx.Err()
						return
					}
				}
				
				// Check if function completed
				if outputResp.Completed {
					if outputResp.Error != nil {
						errChan <- outputResp.Error
					} else {
						errChan <- nil
					}
					return
				}
				
				// Adaptive delay: if we got output, poll immediately for more
				// If no output, use a small delay to avoid busy polling
				var delay time.Duration
				if len(outputResp.NewOutput) > 0 {
					delay = 1 * time.Millisecond // Very fast when there's active output
				} else {
					delay = 10 * time.Millisecond // Slower when idle
				}
				
				select {
				case <-ctx.Done():
					errChan <- ctx.Err()
					return
				case <-time.After(delay):
					// Continue polling
				}
			}
		}
	}()
	
	// Input forwarding goroutine
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case input, ok := <-inputChan:
				if !ok {
					return
				}
				
				// Send input to the plugin
				inputArgs := &InteractiveStreamInputArgs{
					SessionID: sessionID,
					Input:     input,
				}
				var inputResp InteractiveStreamInputResponse
				
				err := g.client.Call("Plugin.SendInteractiveStreamInput", inputArgs, &inputResp)
				if err != nil {
					// Log error but don't fail the entire function
					continue
				}
			}
		}
	}()
	
	// Wait for completion
	select {
	case <-done:
		return <-errChan
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (g *PluginRPCClient) GetCLICommands() (map[string]CLICommandHandler, error) {
	var resp map[string]CLICommandInfo
	err := g.client.Call("Plugin.GetCLICommands", new(interface{}), &resp)
	if err != nil {
		return nil, err
	}

	commands := make(map[string]CLICommandHandler)
	for name := range resp {
		cmdName := name
		commands[name] = func(args []string) error {
			var execResp CLICommandExecuteResponse
			execArgs := &CLICommandExecuteArgs{
				Name: cmdName,
				Args: args,
			}
			err := g.client.Call("Plugin.ExecuteCLICommand", execArgs, &execResp)
			if err != nil {
				return err
			}
			if execResp.Output != "" {
				// Output is handled by MOPS host, don't print to avoid UI interference
				// fmt.Print(execResp.Output)
			}
			return execResp.Error
		}
	}
	return commands, nil
}

func (g *PluginRPCClient) GetStreamingCLICommands() (map[string]StreamingCLICommandHandler, error) {
	var resp map[string]CLICommandInfo
	err := g.client.Call("Plugin.GetStreamingCLICommands", new(interface{}), &resp)
	if err != nil {
		return nil, err
	}

	commands := make(map[string]StreamingCLICommandHandler)
	for name := range resp {
		cmdName := name
		commands[name] = &StreamingCommandWrapper{
			name:   cmdName,
			client: g.client,
		}
	}
	return commands, nil
}

// StreamingCommandWrapper wraps RPC calls for streaming CLI commands
type StreamingCommandWrapper struct {
	name   string
	client *rpc.Client
}

// Execute provides fallback execution for streaming commands
func (w *StreamingCommandWrapper) Execute(ctx context.Context, args []string) error {
	var execResp CLICommandExecuteResponse
	execArgs := &CLICommandExecuteArgs{
		Name: w.name,
		Args: args,
	}
	err := w.client.Call("Plugin.ExecuteCLICommand", execArgs, &execResp)
	if err != nil {
		return err
	}
	if execResp.Output != "" {
		// Output is handled by MOPS host, don't print to avoid UI interference
		// fmt.Print(execResp.Output)
	}
	return execResp.Error
}

// GetHelp returns help for the streaming command
func (w *StreamingCommandWrapper) GetHelp() string {
	return "Streaming CLI command"
}

// ExecuteStreaming provides real-time streaming execution
func (w *StreamingCommandWrapper) ExecuteStreaming(ctx context.Context, args []string, outputChan chan<- string) error {
	var execResp CLICommandExecuteResponse
	execArgs := &CLICommandExecuteArgs{
		Name: w.name,
		Args: args,
	}
	
	// Call the streaming execution RPC method
	err := w.client.Call("Plugin.ExecuteStreamingCLICommand", execArgs, &execResp)
	if err != nil {
		return err
	}
	
	// For now, output the result to the channel
	if execResp.Output != "" {
		outputChan <- execResp.Output
	}
	
	return execResp.Error
}

// SupportsStreaming indicates this command supports streaming
func (w *StreamingCommandWrapper) SupportsStreaming() bool {
	return true
}

func (g *PluginRPCClient) GetMenuEntries() (map[string][]MenuEntry, error) {
	var resp map[string][]MenuEntry
	err := g.client.Call("Plugin.GetMenuEntries", new(interface{}), &resp)
	return resp, err
}

func (g *PluginRPCClient) Cleanup() error {
	var resp error
	err := g.client.Call("Plugin.Cleanup", new(interface{}), &resp)
	if err != nil {
		return err
	}
	return resp
}

func (g *PluginRPCClient) ValidateConfig(config map[string]any) error {
	var resp error
	err := g.client.Call("Plugin.ValidateConfig", config, &resp)
	if err != nil {
		return err
	}
	return resp
}

// PluginRPCServer is the RPC server that PluginRPCClient talks to
type PluginRPCServer struct {
	Impl Plugin
}

func (s *PluginRPCServer) GetInfo(args interface{}, resp *PluginInfoRPC) error {
	info := s.Impl.GetInfo()
	rpcInfo, err := NewPluginInfoRPC(info)
	if err != nil {
		return err
	}
	*resp = rpcInfo
	return nil
}

func (s *PluginRPCServer) Initialize(configJSON string, resp *error) error {
	var config map[string]any
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		*resp = fmt.Errorf("failed to unmarshal config: %w", err)
		return nil
	}

	*resp = s.Impl.Initialize(config)
	return nil
}

func (s *PluginRPCServer) ReloadConfig(configJSON string, resp *error) error {
	var config map[string]any
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		*resp = fmt.Errorf("failed to unmarshal config: %w", err)
		return nil
	}

	*resp = s.Impl.ReloadConfig(config)
	return nil
}

func (s *PluginRPCServer) RegisterProviders(args interface{}, resp *[]DynamicProviderRPC) error {
	providers := s.Impl.RegisterProviders()
	result := make([]DynamicProviderRPC, len(providers))
	for i, p := range providers {
		result[i] = DynamicProviderRPC{
			Name:        p.GetName(),
			Description: p.GetDescription(),
		}
	}
	*resp = result
	return nil
}

func (s *PluginRPCServer) RegisterInteractiveFunctions(args interface{}, resp *map[string]string) error {
	functions := s.Impl.RegisterInteractiveFunctions()
	result := make(map[string]string)
	for name := range functions {
		result[name] = name
	}
	*resp = result
	return nil
}

func (s *PluginRPCServer) CallInteractiveFunction(args *InteractiveFunctionCallArgs, resp *InteractiveFunctionCallResponse) error {
	functions := s.Impl.RegisterInteractiveFunctions()

	fn, exists := functions[args.Name]
	if !exists {
		return fmt.Errorf("function %s not found", args.Name)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	outputChan := make(chan string, 1000)
	inputChan := make(chan string, 100)
	defer close(inputChan)

	errChan := make(chan error, 1)
	go func() {
		defer close(outputChan)
		defer close(errChan)

		if err := fn(ctx, outputChan, inputChan, args.Params); err != nil {
			errChan <- err
		}
	}()

	var output []string
	var funcError error

	for {
		select {
		case line, ok := <-outputChan:
			if !ok {
				select {
				case err := <-errChan:
					funcError = err
				default:
				}
				goto done
			}
			output = append(output, line)
		case err := <-errChan:
			funcError = err
		case <-ctx.Done():
			funcError = fmt.Errorf("function timed out")
			goto done
		}
	}

done:
	resp.Output = output
	resp.Error = funcError
	return nil
}

func (s *PluginRPCServer) GetCLICommands(args interface{}, resp *map[string]CLICommandInfo) error {
	commands, err := s.Impl.GetCLICommands()
	if err != nil {
		return err
	}

	result := make(map[string]CLICommandInfo)
	info := s.Impl.GetInfo()

	for name := range commands {
		// Find command info from plugin metadata
		for _, cmdInfo := range info.CLICommands {
			if cmdInfo.Name == name {
				result[name] = cmdInfo
				break
			}
		}
	}

	*resp = result
	return nil
}

func (s *PluginRPCServer) ExecuteCLICommand(args *CLICommandExecuteArgs, resp *CLICommandExecuteResponse) error {
	commands, err := s.Impl.GetCLICommands()
	if err != nil {
		resp.Error = err
		return nil
	}

	handler, exists := commands[args.Name]
	if !exists {
		resp.Error = fmt.Errorf("command %s not found", args.Name)
		return nil
	}

	// Capture stdout during command execution
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	execErr := handler(args.Args)

	w.Close()
	os.Stdout = oldStdout

	output, _ := io.ReadAll(r)
	resp.Output = string(output)
	resp.Error = execErr

	return nil
}

func (s *PluginRPCServer) GetStreamingCLICommands(args interface{}, resp *map[string]CLICommandInfo) error {
	commands, err := s.Impl.GetStreamingCLICommands()
	if err != nil {
		return err
	}

	result := make(map[string]CLICommandInfo)
	info := s.Impl.GetInfo()

	for name := range commands {
		// Find command info from plugin metadata
		for _, cmdInfo := range info.CLICommands {
			if cmdInfo.Name == name {
				result[name] = cmdInfo
				break
			}
		}
	}

	*resp = result
	return nil
}

func (s *PluginRPCServer) ExecuteStreamingCLICommand(args *CLICommandExecuteArgs, resp *CLICommandExecuteResponse) error {
	commands, err := s.Impl.GetStreamingCLICommands()
	if err != nil {
		resp.Error = err
		return nil
	}

	handler, exists := commands[args.Name]
	if !exists {
		resp.Error = fmt.Errorf("streaming command %s not found", args.Name)
		return nil
	}

	// Create output channel for streaming
	outputChan := make(chan string, 100)
	var outputLines []string
	
	// Collect streaming output
	done := make(chan error, 1)
	go func() {
		defer close(outputChan)
		ctx := context.Background()
		done <- handler.ExecuteStreaming(ctx, args.Args, outputChan)
	}()
	
	// Collect all output lines
	for {
		select {
		case line, ok := <-outputChan:
			if !ok {
				// Channel closed, wait for completion
				resp.Error = <-done
				resp.Output = strings.Join(outputLines, "")
				return nil
			}
			outputLines = append(outputLines, line)
		case execErr := <-done:
			resp.Error = execErr
			resp.Output = strings.Join(outputLines, "")
			return nil
		}
	}
}

func (s *PluginRPCServer) GetMenuEntries(args interface{}, resp *map[string][]MenuEntry) error {
	entries, err := s.Impl.GetMenuEntries()
	if err != nil {
		return err
	}
	*resp = entries
	return nil
}

func (s *PluginRPCServer) Cleanup(args interface{}, resp *error) error {
	*resp = s.Impl.Cleanup()
	return nil
}

func (s *PluginRPCServer) ValidateConfig(args map[string]any, resp *error) error {
	*resp = s.Impl.ValidateConfig(args)
	return nil
}

// Streaming Interactive Function Server Methods
var (
	streamingSessions = make(map[string]*InteractiveStreamSession)
	sessionMutex      sync.RWMutex
)

// setSessionInputState updates the input state for a session
func setSessionInputState(sessionID string, waitingForInput bool, prompt string) {
	sessionMutex.RLock()
	session, exists := streamingSessions[sessionID]
	sessionMutex.RUnlock()
	
	if !exists {
		return // Session not found, ignore
	}
	
	session.mu.Lock()
	session.WaitingForInput = waitingForInput
	session.InputPrompt = prompt
	session.mu.Unlock()
}

type InteractiveStreamSession struct {
	FunctionName   string
	Params         map[string]interface{}
	OutputChan     chan string
	InputChan      chan string
	ErrorChan      chan error
	Context        context.Context
	Cancel         context.CancelFunc
	OutputBuffer   []string
	LastSentIndex  int  // Track what has been sent to avoid duplicates
	Completed      bool
	FinalError     error
	WaitingForInput bool   // Track if function is waiting for input
	InputPrompt     string // Current input prompt
	mu             sync.RWMutex
}

func (s *PluginRPCServer) InitInteractiveStream(args *InteractiveStreamInitArgs, resp *InteractiveStreamInitResponse) error {
	sessionID := fmt.Sprintf("session_%d_%s", time.Now().UnixNano(), args.FunctionName)
	
	// Get regular interactive functions
	functions := s.Impl.RegisterInteractiveFunctions()
	fn, exists := functions[args.FunctionName]
	if !exists {
		resp.Error = fmt.Errorf("function %s not found", args.FunctionName)
		return nil
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	session := &InteractiveStreamSession{
		FunctionName:    args.FunctionName,
		Params:          args.Params,
		OutputChan:      make(chan string, 1000),
		InputChan:       make(chan string, 100),
		ErrorChan:       make(chan error, 1),
		Context:         ctx,
		Cancel:          cancel,
		OutputBuffer:    make([]string, 0),
		LastSentIndex:   0,
		Completed:       false,
		WaitingForInput: false,
		InputPrompt:     "",
	}
	
	sessionMutex.Lock()
	streamingSessions[sessionID] = session
	sessionMutex.Unlock()
	
	// Start the function in a goroutine
	go func() {
		defer func() {
			session.mu.Lock()
			session.Completed = true
			close(session.OutputChan)
			close(session.InputChan)
			session.mu.Unlock()
		}()
		
		err := fn(session.Context, session.OutputChan, session.InputChan, session.Params)
		
		session.mu.Lock()
		session.FinalError = err
		session.mu.Unlock()
		
		session.ErrorChan <- err
	}()
	
	// Start output collection goroutine
	go func() {
		for {
			select {
			case line, ok := <-session.OutputChan:
				if !ok {
					return
				}
				session.mu.Lock()
				session.OutputBuffer = append(session.OutputBuffer, line)
				session.mu.Unlock()
				
			case <-session.Context.Done():
				return
			}
		}
	}()
	
	resp.SessionID = sessionID
	resp.Error = nil
	return nil
}

func (s *PluginRPCServer) SendInteractiveStreamInput(args *InteractiveStreamInputArgs, resp *InteractiveStreamInputResponse) error {
	sessionMutex.RLock()
	session, exists := streamingSessions[args.SessionID]
	sessionMutex.RUnlock()
	
	if !exists {
		resp.Error = fmt.Errorf("session %s not found", args.SessionID)
		return nil
	}
	
	session.mu.RLock()
	completed := session.Completed
	session.mu.RUnlock()
	
	if completed {
		resp.Success = false
		resp.Error = fmt.Errorf("session %s already completed", args.SessionID)
		return nil
	}
	
	select {
	case session.InputChan <- args.Input:
		resp.Success = true
		resp.Error = nil
	case <-time.After(1 * time.Second):
		resp.Success = false
		resp.Error = fmt.Errorf("timeout sending input to session %s", args.SessionID)
	}
	
	return nil
}

func (s *PluginRPCServer) GetInteractiveStreamOutput(args *InteractiveStreamOutputArgs, resp *InteractiveStreamOutputResponse) error {
	sessionMutex.RLock()
	session, exists := streamingSessions[args.SessionID]
	sessionMutex.RUnlock()
	
	if !exists {
		resp.Error = fmt.Errorf("session %s not found", args.SessionID)
		return nil
	}
	
	session.mu.Lock()
	defer session.mu.Unlock()
	
	// Get only new output since last call using the LastSentIndex
	var newOutput []string
	if session.LastSentIndex < len(session.OutputBuffer) {
		newOutput = make([]string, len(session.OutputBuffer)-session.LastSentIndex)
		copy(newOutput, session.OutputBuffer[session.LastSentIndex:])
		session.LastSentIndex = len(session.OutputBuffer)
	} else {
		newOutput = []string{} // No new output
	}
	
	completed := session.Completed
	finalError := session.FinalError
	
	resp.NewOutput = newOutput
	resp.Completed = completed
	resp.Error = finalError
	resp.NeedsInput = session.WaitingForInput
	resp.InputPrompt = session.InputPrompt
	
	return nil
}

func (s *PluginRPCServer) CleanupInteractiveStream(args *InteractiveStreamCleanupArgs, resp *InteractiveStreamCleanupResponse) error {
	sessionMutex.Lock()
	session, exists := streamingSessions[args.SessionID]
	if exists {
		session.Cancel()
		delete(streamingSessions, args.SessionID)
	}
	sessionMutex.Unlock()
	
	resp.Success = exists
	if !exists {
		resp.Error = fmt.Errorf("session %s not found", args.SessionID)
	}
	
	return nil
}

func (s *PluginRPCServer) GenerateProviderEntries(args *ProviderGenerateEntriesArgs, resp *[]MenuEntry) error {
	providers := s.Impl.RegisterProviders()
	for _, provider := range providers {
		if provider.GetName() == args.Name {
			entries, err := provider.GenerateEntries(args.Param)
			if err != nil {
				return err
			}
			*resp = entries
			return nil
		}
	}
	return fmt.Errorf("provider %s not found", args.Name)
}

// RPC client implementations
type DynamicProviderRPCClient struct {
	client      *rpc.Client
	name        string
	description string
}

func (d *DynamicProviderRPCClient) GetName() string {
	return d.name
}

func (d *DynamicProviderRPCClient) GetDescription() string {
	return d.description
}

func (d *DynamicProviderRPCClient) GenerateEntries(param string) ([]MenuEntry, error) {
	args := &ProviderGenerateEntriesArgs{
		Name:  d.name,
		Param: param,
	}
	var resp []MenuEntry
	err := d.client.Call("Plugin.GenerateProviderEntries", args, &resp)
	return resp, err
}

func (d *DynamicProviderRPCClient) SupportsRefresh() bool {
	return false // Default implementation for RPC clients
}

// RPC argument types
type DynamicProviderRPC struct {
	Name        string
	Description string
}

type InteractiveFunctionCallArgs struct {
	Name   string
	Params map[string]interface{}
}

type InteractiveFunctionCallResponse struct {
	Output []string
	Error  error
}

type CLICommandExecuteArgs struct {
	Name string
	Args []string
}

type CLICommandExecuteResponse struct {
	Output string
	Error  error
}

type ProviderGenerateEntriesArgs struct {
	Name  string
	Param string
}

type ExecutorExecuteArgs struct {
	ActionType string
	Entry      MenuEntry
	Input      string
}

// Streaming Interactive Function RPC Types
type InteractiveStreamInitArgs struct {
	FunctionName string
	Params       map[string]interface{}
}

type InteractiveStreamInitResponse struct {
	SessionID string
	Error     error
}

type InteractiveStreamInputArgs struct {
	SessionID string
	Input     string
}

type InteractiveStreamInputResponse struct {
	Success bool
	Error   error
}

type InteractiveStreamOutputArgs struct {
	SessionID string
}

type InteractiveStreamOutputResponse struct {
	NewOutput     []string
	Completed     bool
	Error         error
	NeedsInput    bool   // Indicates if the function is waiting for user input
	InputPrompt   string // Optional prompt message for the input
}

type InteractiveStreamCleanupArgs struct {
	SessionID string
}

type InteractiveStreamCleanupResponse struct {
	Success bool
	Error   error
}
