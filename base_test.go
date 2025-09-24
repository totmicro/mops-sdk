package sdk

import (
	"context"
	"testing"
)

func TestNewPluginBase(t *testing.T) {
	info := PluginInfo{
		Name:        "test-base-plugin",
		Version:     "1.0.0",
		Description: "A test base plugin",
		Author:      "Test Author",
		License:     "MIT",
	}

	base := NewPluginBase(info)

	if base == nil {
		t.Fatal("NewPluginBase returned nil")
	}

	// Check info is set correctly
	baseInfo := base.GetInfo()
	if baseInfo.Name != info.Name {
		t.Errorf("Expected name '%s', got '%s'", info.Name, baseInfo.Name)
	}
	if baseInfo.Version != info.Version {
		t.Errorf("Expected version '%s', got '%s'", info.Version, baseInfo.Version)
	}

	// Check collections are initialized
	if base.providers == nil {
		t.Error("Expected providers to be initialized")
	}
	if base.interactiveFunctions == nil {
		t.Error("Expected interactiveFunctions to be initialized")
	}
	if base.cliCommands == nil {
		t.Error("Expected cliCommands to be initialized")
	}
	if base.menuEntries == nil {
		t.Error("Expected menuEntries to be initialized")
	}

	// Check initial state
	if len(base.providers) != 0 {
		t.Errorf("Expected 0 providers initially, got %d", len(base.providers))
	}

	if len(base.interactiveFunctions) != 0 {
		t.Errorf("Expected 0 interactive functions initially, got %d", len(base.interactiveFunctions))
	}
	if len(base.cliCommands) != 0 {
		t.Errorf("Expected 0 CLI commands initially, got %d", len(base.cliCommands))
	}
	if len(base.menuEntries) != 0 {
		t.Errorf("Expected 0 menu entries initially, got %d", len(base.menuEntries))
	}
}

func TestPluginBaseInitialize(t *testing.T) {
	info := PluginInfo{Name: "test", Version: "1.0.0"}
	base := NewPluginBase(info)

	config := map[string]any{
		"key1": "value1",
		"key2": 42,
		"key3": true,
	}

	err := base.Initialize(config)
	if err != nil {
		t.Errorf("Initialize should not return error: %v", err)
	}

	// Check config is stored
	if len(base.config) != 3 {
		t.Errorf("Expected 3 config items, got %d", len(base.config))
	}
	if base.config["key1"] != "value1" {
		t.Errorf("Expected config key1='value1', got '%v'", base.config["key1"])
	}
	if base.config["key2"] != 42 {
		t.Errorf("Expected config key2=42, got '%v'", base.config["key2"])
	}
	if base.config["key3"] != true {
		t.Errorf("Expected config key3=true, got '%v'", base.config["key3"])
	}
}

func TestPluginBaseRegisterProviders(t *testing.T) {
	info := PluginInfo{Name: "test", Version: "1.0.0"}
	base := NewPluginBase(info)

	// Initially empty
	providers := base.RegisterProviders()
	if len(providers) != 0 {
		t.Errorf("Expected 0 providers initially, got %d", len(providers))
	}

	// Add a provider
	mockProvider := &MockProvider{name: "test-provider"}
	base.AddProvider(mockProvider)

	providers = base.RegisterProviders()
	if len(providers) != 1 {
		t.Errorf("Expected 1 provider after adding, got %d", len(providers))
	}
	if providers[0].GetName() != "test-provider" {
		t.Errorf("Expected provider name 'test-provider', got '%s'", providers[0].GetName())
	}
}



func TestPluginBaseRegisterInteractiveFunctions(t *testing.T) {
	info := PluginInfo{Name: "test", Version: "1.0.0"}
	base := NewPluginBase(info)

	// Initially empty
	functions := base.RegisterInteractiveFunctions()
	if len(functions) != 0 {
		t.Errorf("Expected 0 interactive functions initially, got %d", len(functions))
	}

	// Add an interactive function
	testFunc := func(ctx context.Context, outputChan chan<- string, inputChan <-chan string, params map[string]interface{}) error {
		return nil
	}
	base.AddInteractiveFunction("test-func", testFunc)

	functions = base.RegisterInteractiveFunctions()
	if len(functions) != 1 {
		t.Errorf("Expected 1 interactive function after adding, got %d", len(functions))
	}
	if _, exists := functions["test-func"]; !exists {
		t.Error("Expected 'test-func' to exist in interactive functions")
	}
}

func TestPluginBaseGetCLICommands(t *testing.T) {
	info := PluginInfo{Name: "test", Version: "1.0.0"}
	base := NewPluginBase(info)

	// Initially empty
	commands, err := base.GetCLICommands()
	if err != nil {
		t.Errorf("GetCLICommands should not return error: %v", err)
	}
	if len(commands) != 0 {
		t.Errorf("Expected 0 CLI commands initially, got %d", len(commands))
	}

	// Add a CLI command
	testHandler := func(args []string) error { return nil }
	base.AddCLICommand("test-cmd", testHandler)

	commands, err = base.GetCLICommands()
	if err != nil {
		t.Errorf("GetCLICommands should not return error: %v", err)
	}
	if len(commands) != 1 {
		t.Errorf("Expected 1 CLI command after adding, got %d", len(commands))
	}
	if _, exists := commands["test-cmd"]; !exists {
		t.Error("Expected 'test-cmd' to exist in CLI commands")
	}
}

func TestPluginBaseGetMenuEntries(t *testing.T) {
	info := PluginInfo{Name: "test", Version: "1.0.0"}
	base := NewPluginBase(info)

	// Initially empty
	entries, err := base.GetMenuEntries()
	if err != nil {
		t.Errorf("GetMenuEntries should not return error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("Expected 0 menu entries initially, got %d", len(entries))
	}

	// Add a menu entry
	entry := MenuEntry{
		Key:    "1",
		Label:  "Test Entry",
		Action: "test-action",
	}
	base.AddMenuEntry("test-menu", entry)

	entries, err = base.GetMenuEntries()
	if err != nil {
		t.Errorf("GetMenuEntries should not return error: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("Expected 1 menu after adding entry, got %d", len(entries))
	}
	if menuEntries, exists := entries["test-menu"]; !exists {
		t.Error("Expected 'test-menu' to exist")
	} else if len(menuEntries) != 1 {
		t.Errorf("Expected 1 entry in 'test-menu', got %d", len(menuEntries))
	} else if menuEntries[0].Key != "1" {
		t.Errorf("Expected entry key '1', got '%s'", menuEntries[0].Key)
	}
}

func TestPluginBaseCleanup(t *testing.T) {
	info := PluginInfo{Name: "test", Version: "1.0.0"}
	base := NewPluginBase(info)

	err := base.Cleanup()
	if err != nil {
		t.Errorf("Cleanup should not return error: %v", err)
	}
}

func TestPluginBaseValidateConfig(t *testing.T) {
	info := PluginInfo{Name: "test", Version: "1.0.0"}
	base := NewPluginBase(info)

	config := map[string]any{"key": "value"}
	err := base.ValidateConfig(config)
	if err != nil {
		t.Errorf("ValidateConfig should not return error: %v", err)
	}
}

func TestPluginBaseAddMultipleMenuEntries(t *testing.T) {
	info := PluginInfo{Name: "test", Version: "1.0.0"}
	base := NewPluginBase(info)

	// Add multiple entries to the same menu
	entry1 := MenuEntry{Key: "1", Label: "Entry 1"}
	entry2 := MenuEntry{Key: "2", Label: "Entry 2"}
	
	base.AddMenuEntry("main", entry1)
	base.AddMenuEntry("main", entry2)

	entries, _ := base.GetMenuEntries()
	if len(entries) != 1 {
		t.Errorf("Expected 1 menu, got %d", len(entries))
	}
	
	mainEntries := entries["main"]
	if len(mainEntries) != 2 {
		t.Errorf("Expected 2 entries in main menu, got %d", len(mainEntries))
	}

	// Add entry to different menu
	entry3 := MenuEntry{Key: "3", Label: "Entry 3"}
	base.AddMenuEntry("settings", entry3)

	entries, _ = base.GetMenuEntries()
	if len(entries) != 2 {
		t.Errorf("Expected 2 menus, got %d", len(entries))
	}
	
	settingsEntries := entries["settings"]
	if len(settingsEntries) != 1 {
		t.Errorf("Expected 1 entry in settings menu, got %d", len(settingsEntries))
	}
}

// Mock implementations for testing

type MockProvider struct {
	name        string
	description string
}

func (m *MockProvider) GetName() string {
	return m.name
}

func (m *MockProvider) GetDescription() string {
	return m.description
}

func (m *MockProvider) GenerateEntries(param string) ([]MenuEntry, error) {
	return []MenuEntry{
		{Key: "1", Label: "Mock Entry", Action: "mock-action"},
	}, nil
}

func (m *MockProvider) SupportsRefresh() bool {
	return true
}

type MockExecutor struct {
	actionType string
}

func (m *MockExecutor) GetActionType() string {
	return m.actionType
}

func (m *MockExecutor) Execute(entry MenuEntry, input string) ActionResult {
	return ActionResult{
		Success: true,
		Output:  "Mock execution result",
	}
}
