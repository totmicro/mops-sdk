package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/rpc"
	"os"
	"strings"
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
			client: g.client,
			name:   p.Name,
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
	args := &InteractiveFunctionCallArgs{
		Name:   name,
		Params: params,
	}

	var resp InteractiveFunctionCallResponse
	err := g.client.Call("Plugin.CallInteractiveFunction", args, &resp)
	if err != nil {
		return fmt.Errorf("RPC call failed: %w", err)
	}

	for _, line := range resp.Output {
		outputChan <- line
	}

	return resp.Error
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
			Name: p.GetName(),
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

func (s *PluginRPCServer) GenerateProviderEntries(args *ProviderGenerateEntriesArgs, resp *[]MenuEntry) error {
	providers := s.Impl.RegisterProviders()
	for _, provider := range providers {
		if provider.GetName() == args.Name {
			entries, err := provider.GetEntries(args.Param)
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
	client *rpc.Client
	name   string
}

func (d *DynamicProviderRPCClient) GetName() string {
	return d.name
}

func (d *DynamicProviderRPCClient) GetEntries(param string) ([]MenuEntry, error) {
	args := &ProviderGenerateEntriesArgs{
		Name:  d.name,
		Param: param,
	}
	var resp []MenuEntry
	err := d.client.Call("Plugin.GenerateProviderEntries", args, &resp)
	return resp, err
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
	Name string
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
