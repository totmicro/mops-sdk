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

func (g *PluginRPCClient) GetInfo() PluginInfo {
	var resp PluginInfoRPC
	err := g.client.Call("Plugin.GetInfo", new(interface{}), &resp)
	if err != nil {
		return PluginInfo{}
	}
	info, _ := resp.ToPluginInfo()
	return info
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

func (g *PluginRPCClient) RegisterExecutors() []ActionExecutor {
	var resp []ActionExecutorRPC
	err := g.client.Call("Plugin.RegisterExecutors", new(interface{}), &resp)
	if err != nil {
		return nil
	}

	executors := make([]ActionExecutor, len(resp))
	for i, e := range resp {
		executors[i] = &ActionExecutorRPCClient{
			client:     g.client,
			actionType: e.ActionType,
		}
	}
	return executors
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
				
				// Small delay to avoid busy polling
				select {
				case <-ctx.Done():
					errChan <- ctx.Err()
					return
				case <-time.After(10 * time.Millisecond):
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
				fmt.Print(execResp.Output)
			}
			return execResp.Error
		}
	}
	return commands, nil
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

func (s *PluginRPCServer) RegisterExecutors(args interface{}, resp *[]ActionExecutorRPC) error {
	executors := s.Impl.RegisterExecutors()
	result := make([]ActionExecutorRPC, len(executors))
	for i, e := range executors {
		result[i] = ActionExecutorRPC{
			ActionType: e.GetActionType(),
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

	s.populateInputChannel(inputChan, args.Name)

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
			if !s.isDebugMessage(line) {
				output = append(output, line)
			}
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

func (s *PluginRPCServer) populateInputChannel(inputChan chan<- string, functionName string) {
	defer close(inputChan)

	switch functionName {
	case "chat":
		inputChan <- "Hello from test!"
		inputChan <- "How are you?"
		inputChan <- "quit"
	default:
		inputChan <- "test input"
		inputChan <- "quit"
	}
}

func (s *PluginRPCServer) isDebugMessage(line string) bool {
	debugPatterns := []string{
		"[SERVER DEBUG]", "[CLIENT DEBUG]", "🔧 [SERVER DEBUG]",
		"✅ [SERVER DEBUG]", "❌ [SERVER DEBUG]", "⏰ [SERVER DEBUG]",
	}
	for _, pattern := range debugPatterns {
		if strings.Contains(line, pattern) {
			return true
		}
	}
	return false
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

type InteractiveStreamSession struct {
	FunctionName string
	Params       map[string]interface{}
	OutputChan   chan string
	InputChan    chan string
	ErrorChan    chan error
	Context      context.Context
	Cancel       context.CancelFunc
	OutputBuffer []string
	Completed    bool
	FinalError   error
	mu           sync.RWMutex
}

func (s *PluginRPCServer) InitInteractiveStream(args *InteractiveStreamInitArgs, resp *InteractiveStreamInitResponse) error {
	sessionID := fmt.Sprintf("session_%d_%s", time.Now().UnixNano(), args.FunctionName)
	
	functions := s.Impl.RegisterInteractiveFunctions()
	fn, exists := functions[args.FunctionName]
	if !exists {
		resp.Error = fmt.Errorf("function %s not found", args.FunctionName)
		return nil
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	session := &InteractiveStreamSession{
		FunctionName: args.FunctionName,
		Params:       args.Params,
		OutputChan:   make(chan string, 1000),
		InputChan:    make(chan string, 100),
		ErrorChan:    make(chan error, 1),
		Context:      ctx,
		Cancel:       cancel,
		OutputBuffer: make([]string, 0),
		Completed:    false,
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
	// Get new output since last call (this is simplified - in production we'd track the last sent index)
	newOutput := make([]string, len(session.OutputBuffer))
	copy(newOutput, session.OutputBuffer)
	session.OutputBuffer = session.OutputBuffer[:0] // Clear buffer after reading
	
	completed := session.Completed
	finalError := session.FinalError
	session.mu.Unlock()
	
	resp.NewOutput = newOutput
	resp.Completed = completed
	resp.Error = finalError
	
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

func (s *PluginRPCServer) ExecuteAction(args *ExecutorExecuteArgs, resp *ActionResult) error {
	executors := s.Impl.RegisterExecutors()

	for _, executor := range executors {
		if executor.GetActionType() == args.ActionType {
			*resp = executor.Execute(args.Entry, args.Input)
			return nil
		}
	}

	*resp = ActionResult{
		Success: false,
		Error:   fmt.Errorf("executor not found for action type: %s", args.ActionType),
	}
	return nil
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

type ActionExecutorRPCClient struct {
	client     *rpc.Client
	actionType string
}

func (a *ActionExecutorRPCClient) GetActionType() string {
	return a.actionType
}

func (a *ActionExecutorRPCClient) Execute(entry MenuEntry, input string) ActionResult {
	args := &ExecutorExecuteArgs{
		ActionType: a.actionType,
		Entry:      entry,
		Input:      input,
	}
	var resp ActionResult
	err := a.client.Call("Plugin.ExecuteAction", args, &resp)
	if err != nil {
		return ActionResult{
			Success: false,
			Error:   err,
		}
	}
	return resp
}

// RPC argument types
type DynamicProviderRPC struct {
	Name        string
	Description string
}

type ActionExecutorRPC struct {
	ActionType string
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
	NewOutput []string
	Completed bool
	Error     error
}

type InteractiveStreamCleanupArgs struct {
	SessionID string
}

type InteractiveStreamCleanupResponse struct {
	Success bool
	Error   error
}
