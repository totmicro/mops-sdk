package sdk

import (
	"context"
	"testing"
)

func TestMopsPlugin(t *testing.T) {
	// Create a mock plugin implementation
	mockPlugin := &MockPluginImpl{
		info: PluginInfo{
			Name:        "test-rpc-plugin",
			Version:     "1.0.0",
			Description: "A test plugin for RPC testing",
		},
	}

	// Create the MOPS plugin wrapper
	mopsPlugin := &MopsPlugin{Impl: mockPlugin}

	if mopsPlugin.Impl != mockPlugin {
		t.Error("Expected plugin implementation to be set correctly")
	}
}

func TestPluginRPCServer_GetInfo(t *testing.T) {
	mockPlugin := &MockPluginImpl{
		info: PluginInfo{
			Name:        "test-info-plugin",
			Version:     "2.0.0",
			Description: "Test plugin for info retrieval",
			Author:      "Test Author",
			License:     "MIT",
		},
	}

	server := &PluginRPCServer{Impl: mockPlugin}

	var resp PluginInfoRPC
	err := server.GetInfo(nil, &resp)

	if err != nil {
		t.Errorf("GetInfo should not return error: %v", err)
	}

	if resp.Name != "test-info-plugin" {
		t.Errorf("Expected name 'test-info-plugin', got '%s'", resp.Name)
	}
	if resp.Version != "2.0.0" {
		t.Errorf("Expected version '2.0.0', got '%s'", resp.Version)
	}
	if resp.Author != "Test Author" {
		t.Errorf("Expected author 'Test Author', got '%s'", resp.Author)
	}
}

func TestPluginRPCServer_RegisterProviders(t *testing.T) {
	mockProvider := &MockProvider{
		name:        "test-provider",
		description: "A test provider",
	}

	mockPlugin := &MockPluginImpl{
		providers: []DynamicProvider{mockProvider},
	}

	server := &PluginRPCServer{Impl: mockPlugin}

	var resp []DynamicProviderRPC
	err := server.RegisterProviders(nil, &resp)

	if err != nil {
		t.Errorf("RegisterProviders should not return error: %v", err)
	}
	if len(resp) != 1 {
		t.Errorf("Expected 1 provider, got %d", len(resp))
	}
	if resp[0].Name != "test-provider" {
		t.Errorf("Expected provider name 'test-provider', got '%s'", resp[0].Name)
	}
}

func TestPluginRPCServer_RegisterExecutors(t *testing.T) {
	mockExecutor := &MockExecutor{
		actionType: "test-action",
	}

	mockPlugin := &MockPluginImpl{
		executors: []ActionExecutor{mockExecutor},
	}

	server := &PluginRPCServer{Impl: mockPlugin}

	var resp []ActionExecutorRPC
	err := server.RegisterExecutors(nil, &resp)

	if err != nil {
		t.Errorf("RegisterExecutors should not return error: %v", err)
	}
	if len(resp) != 1 {
		t.Errorf("Expected 1 executor, got %d", len(resp))
	}
	if resp[0].ActionType != "test-action" {
		t.Errorf("Expected action type 'test-action', got '%s'", resp[0].ActionType)
	}
}

func TestPluginRPCServer_RegisterInteractiveFunctions(t *testing.T) {
	testFunc := func(ctx context.Context, outputChan chan<- string, inputChan <-chan string, params map[string]interface{}) error {
		outputChan <- "Test output"
		return nil
	}

	mockPlugin := &MockPluginImpl{
		interactiveFunctions: map[string]InteractiveGoFunction{
			"test-func": testFunc,
		},
	}

	server := &PluginRPCServer{Impl: mockPlugin}

	var resp map[string]string
	err := server.RegisterInteractiveFunctions(nil, &resp)

	if err != nil {
		t.Errorf("RegisterInteractiveFunctions should not return error: %v", err)
	}
	if len(resp) != 1 {
		t.Errorf("Expected 1 interactive function, got %d", len(resp))
	}
	if _, exists := resp["test-func"]; !exists {
		t.Error("Expected 'test-func' to exist in response")
	}
}

func TestPluginRPCServer_GetCLICommands(t *testing.T) {
	testHandler := func(args []string) error { return nil }

	mockPlugin := &MockPluginImpl{
		info: PluginInfo{
			CLICommands: []CLICommandInfo{
				{
					Name:        "test-cmd",
					Description: "A test command",
					Usage:       "test-cmd [options]",
				},
			},
		},
		cliCommands: map[string]CLICommandHandler{
			"test-cmd": testHandler,
		},
	}

	server := &PluginRPCServer{Impl: mockPlugin}

	var resp map[string]CLICommandInfo
	err := server.GetCLICommands(nil, &resp)

	if err != nil {
		t.Errorf("GetCLICommands should not return error: %v", err)
	}
	if len(resp) != 1 {
		t.Errorf("Expected 1 CLI command, got %d", len(resp))
	}
	if cmdInfo, exists := resp["test-cmd"]; !exists {
		t.Error("Expected 'test-cmd' to exist in response")
	} else {
		if cmdInfo.Name != "test-cmd" {
			t.Errorf("Expected command name 'test-cmd', got '%s'", cmdInfo.Name)
		}
	}
}

func TestPluginRPCServer_GetMenuEntries(t *testing.T) {
	mockPlugin := &MockPluginImpl{
		menuEntries: map[string][]MenuEntry{
			"main": {
				{Key: "1", Label: "Test Entry", Action: "test-action"},
			},
		},
	}

	server := &PluginRPCServer{Impl: mockPlugin}

	var resp map[string][]MenuEntry
	err := server.GetMenuEntries(nil, &resp)

	if err != nil {
		t.Errorf("GetMenuEntries should not return error: %v", err)
	}
	if len(resp) != 1 {
		t.Errorf("Expected 1 menu, got %d", len(resp))
	}
	if entries, exists := resp["main"]; !exists {
		t.Error("Expected 'main' menu to exist")
	} else if len(entries) != 1 {
		t.Errorf("Expected 1 entry in main menu, got %d", len(entries))
	} else if entries[0].Key != "1" {
		t.Errorf("Expected entry key '1', got '%s'", entries[0].Key)
	}
}

func TestPluginRPCServer_Cleanup(t *testing.T) {
	mockPlugin := &MockPluginImpl{}
	server := &PluginRPCServer{Impl: mockPlugin}

	var resp error
	err := server.Cleanup(nil, &resp)

	if err != nil {
		t.Errorf("Cleanup should not return error: %v", err)
	}
	if resp != nil {
		t.Errorf("Cleanup response should be nil, got: %v", resp)
	}
}

// Mock implementation for testing
type MockPluginImpl struct {
	info                 PluginInfo
	providers            []DynamicProvider
	executors            []ActionExecutor
	interactiveFunctions map[string]InteractiveGoFunction
	cliCommands          map[string]CLICommandHandler
	menuEntries          map[string][]MenuEntry
	initError            error
	cleanupError         error
}

func (m *MockPluginImpl) GetInfo() PluginInfo {
	return m.info
}

func (m *MockPluginImpl) Initialize(config map[string]any) error {
	return m.initError
}

func (m *MockPluginImpl) RegisterProviders() []DynamicProvider {
	if m.providers == nil {
		return []DynamicProvider{}
	}
	return m.providers
}

func (m *MockPluginImpl) RegisterExecutors() []ActionExecutor {
	if m.executors == nil {
		return []ActionExecutor{}
	}
	return m.executors
}

func (m *MockPluginImpl) RegisterInteractiveFunctions() map[string]InteractiveGoFunction {
	if m.interactiveFunctions == nil {
		return make(map[string]InteractiveGoFunction)
	}
	return m.interactiveFunctions
}

func (m *MockPluginImpl) GetCLICommands() (map[string]CLICommandHandler, error) {
	if m.cliCommands == nil {
		return make(map[string]CLICommandHandler), nil
	}
	return m.cliCommands, nil
}

func (m *MockPluginImpl) GetMenuEntries() (map[string][]MenuEntry, error) {
	if m.menuEntries == nil {
		return make(map[string][]MenuEntry), nil
	}
	return m.menuEntries, nil
}

func (m *MockPluginImpl) Cleanup() error {
	return m.cleanupError
}

func (m *MockPluginImpl) ValidateConfig(config map[string]any) error {
	return nil
}

// Test RPC protocol version and magic cookie
func TestMopsPluginConstants(t *testing.T) {
	// These constants should be defined correctly for the protocol to work
	if ProtocolVersion == 0 {
		t.Error("ProtocolVersion should be defined")
	}
	if PluginName == "" {
		t.Error("PluginName should be defined")
	}
}

// Test provider and executor RPC conversion
func TestProviderRPCConversion(t *testing.T) {
	provider := &MockProvider{
		name:        "test-provider",
		description: "A test provider for RPC",
	}

	// Test converting to RPC format
	rpcProvider := DynamicProviderRPC{
		Name:        provider.GetName(),
		Description: provider.GetDescription(),
	}

	if rpcProvider.Name != "test-provider" {
		t.Errorf("Expected RPC provider name 'test-provider', got '%s'", rpcProvider.Name)
	}
	if rpcProvider.Description != "A test provider for RPC" {
		t.Errorf("Expected RPC provider description to match")
	}
}

func TestExecutorRPCConversion(t *testing.T) {
	executor := &MockExecutor{
		actionType: "test-action",
	}

	// Test converting to RPC format
	rpcExecutor := ActionExecutorRPC{
		ActionType: executor.GetActionType(),
	}

	if rpcExecutor.ActionType != "test-action" {
		t.Errorf("Expected RPC executor action type 'test-action', got '%s'", rpcExecutor.ActionType)
	}
}
