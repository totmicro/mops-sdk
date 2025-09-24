package sdk

import (
	"context"
	"testing"
)

func TestNewPluginBuilder(t *testing.T) {
	builder := NewPluginBuilder("test-plugin", "1.0.0", "A test plugin")

	if builder == nil {
		t.Fatal("NewPluginBuilder returned nil")
	}

	info := builder.base.GetInfo()
	if info.Name != "test-plugin" {
		t.Errorf("Expected name 'test-plugin', got '%s'", info.Name)
	}
	if info.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", info.Version)
	}
	if info.Description != "A test plugin" {
		t.Errorf("Expected description 'A test plugin', got '%s'", info.Description)
	}
	if info.License != "MIT" {
		t.Errorf("Expected default license 'MIT', got '%s'", info.License)
	}

	// Check default supported platforms
	if len(info.SupportedPlatforms) != 5 {
		t.Errorf("Expected 5 supported platforms, got %d", len(info.SupportedPlatforms))
	}

	// Verify base is created
	if builder.base == nil {
		t.Error("Expected base to be created")
	}
}

func TestBuilderFluentInterface(t *testing.T) {
	builder := NewPluginBuilder("test", "1.0.0", "test description").
		SetAuthor("Test Author").
		SetLicense("Apache-2.0").
		SetHomepage("https://example.com").
		SetMopsVersions("1.0.0", "2.0.0").
		AddTag("utility").
		AddTag("cli").
		AddDependency("dep1").
		SetDefaultConfig(map[string]any{"key": "value"})

	info := builder.base.GetInfo()

	// Test author
	if info.Author != "Test Author" {
		t.Errorf("Expected author 'Test Author', got '%s'", info.Author)
	}

	// Test license
	if info.License != "Apache-2.0" {
		t.Errorf("Expected license 'Apache-2.0', got '%s'", info.License)
	}

	// Test homepage
	if info.Homepage != "https://example.com" {
		t.Errorf("Expected homepage 'https://example.com', got '%s'", info.Homepage)
	}

	// Test MOPS versions
	if info.MopsMinVersion != "1.0.0" {
		t.Errorf("Expected min version '1.0.0', got '%s'", info.MopsMinVersion)
	}
	if info.MopsMaxVersion != "2.0.0" {
		t.Errorf("Expected max version '2.0.0', got '%s'", info.MopsMaxVersion)
	}

	// Test tags
	if len(info.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(info.Tags))
	}
	if info.Tags[0] != "utility" || info.Tags[1] != "cli" {
		t.Errorf("Expected tags [utility, cli], got %v", info.Tags)
	}

	// Test dependencies
	if len(info.Dependencies) != 1 {
		t.Errorf("Expected 1 dependency, got %d", len(info.Dependencies))
	}
	if info.Dependencies[0] != "dep1" {
		t.Errorf("Expected dependency 'dep1', got '%s'", info.Dependencies[0])
	}

	// Test default config
	if len(info.DefaultConfig) != 1 {
		t.Errorf("Expected 1 config item, got %d", len(info.DefaultConfig))
	}
	if info.DefaultConfig["key"] != "value" {
		t.Errorf("Expected config value 'value', got '%v'", info.DefaultConfig["key"])
	}
}

func TestBuilderWithSimpleProvider(t *testing.T) {
	providerFunc := func(param string) ([]MenuEntry, error) {
		return []MenuEntry{
			{Key: "1", Label: "Test Entry", Action: "test"},
		}, nil
	}

	builder := NewPluginBuilder("test", "1.0.0", "test").
		WithSimpleProvider("test-provider", "A test provider", providerFunc)

	if len(builder.base.providers) != 1 {
		t.Errorf("Expected 1 provider, got %d", len(builder.base.providers))
	}

	provider := builder.base.providers[0]
	if provider.GetName() != "test-provider" {
		t.Errorf("Expected provider name 'test-provider', got '%s'", provider.GetName())
	}

	entries, err := provider.GenerateEntries("test")
	if err != nil {
		t.Errorf("Provider should not return error: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(entries))
	}
	if entries[0].Key != "1" {
		t.Errorf("Expected entry key '1', got '%s'", entries[0].Key)
	}
}



func TestBuilderWithInteractiveFunction(t *testing.T) {
	interactiveFunc := func(ctx context.Context, outputChan chan<- string, inputChan <-chan string, params map[string]interface{}) error {
		return nil
	}

	builder := NewPluginBuilder("test", "1.0.0", "test").
		WithInteractiveFunction("test-func", interactiveFunc)

	if len(builder.base.interactiveFunctions) != 1 {
		t.Errorf("Expected 1 interactive function, got %d", len(builder.base.interactiveFunctions))
	}

	_, exists := builder.base.interactiveFunctions["test-func"]
	if !exists {
		t.Error("Expected interactive function 'test-func' to exist")
	}
}

func TestBuilderWithCLICommand(t *testing.T) {
	commandHandler := func(args []string) error {
		return nil
	}

	builder := NewPluginBuilder("test", "1.0.0", "test").
		WithCLICommand("test-cmd", "A test command", commandHandler)

	info := builder.base.GetInfo()
	if len(info.CLICommands) != 1 {
		t.Errorf("Expected 1 CLI command info, got %d", len(info.CLICommands))
	}
	if len(builder.base.cliCommands) != 1 {
		t.Errorf("Expected 1 CLI command handler, got %d", len(builder.base.cliCommands))
	}

	cmdInfo := info.CLICommands[0]
	if cmdInfo.Name != "test-cmd" {
		t.Errorf("Expected command name 'test-cmd', got '%s'", cmdInfo.Name)
	}
	if cmdInfo.Description != "A test command" {
		t.Errorf("Expected command description 'A test command', got '%s'", cmdInfo.Description)
	}

	_, exists := builder.base.cliCommands["test-cmd"]
	if !exists {
		t.Error("Expected CLI command handler 'test-cmd' to exist")
	}
}

func TestBuilderWithMenuIntegration(t *testing.T) {
	builder := NewPluginBuilder("test", "1.0.0", "test").
		WithMenuIntegration(true, "p", "Test Plugin", "🧪", 10)

	info := builder.base.GetInfo()
	if info.MenuIntegration == nil {
		t.Fatal("Expected MenuIntegration to be set")
	}

	integration := info.MenuIntegration
	if !integration.AutoRegister {
		t.Error("Expected AutoRegister to be true")
	}
	if integration.MenuID != "main" {
		t.Errorf("Expected MenuID 'main', got '%s'", integration.MenuID)
	}
	if integration.Key != "p" {
		t.Errorf("Expected Key 'p', got '%s'", integration.Key)
	}
	if integration.Label != "Test Plugin" {
		t.Errorf("Expected Label 'Test Plugin', got '%s'", integration.Label)
	}
	if integration.Icon != "🧪" {
		t.Errorf("Expected Icon '🧪', got '%s'", integration.Icon)
	}
	if integration.Priority != 10 {
		t.Errorf("Expected Priority 10, got %d", integration.Priority)
	}
	if integration.Group != "plugins" {
		t.Errorf("Expected Group 'plugins', got '%s'", integration.Group)
	}
}

func TestBuilderWithMenuIntegrationAdvanced(t *testing.T) {
	customIntegration := PluginMenuIntegration{
		AutoRegister: false,
		MenuID:       "custom-menu",
		Key:          "c",
		Label:        "Custom Plugin",
		Icon:         "🔧",
		Priority:     5,
		Group:        "custom",
	}

	builder := NewPluginBuilder("test", "1.0.0", "test").
		WithMenuIntegrationAdvanced(customIntegration)

	info := builder.base.GetInfo()
	if info.MenuIntegration == nil {
		t.Fatal("Expected MenuIntegration to be set")
	}

	integration := info.MenuIntegration
	if integration.AutoRegister {
		t.Error("Expected AutoRegister to be false")
	}
	if integration.MenuID != "custom-menu" {
		t.Errorf("Expected MenuID 'custom-menu', got '%s'", integration.MenuID)
	}
	if integration.Group != "custom" {
		t.Errorf("Expected Group 'custom', got '%s'", integration.Group)
	}
}

func TestGetStandardName(t *testing.T) {
	builder := NewPluginBuilder("my-plugin", "1.0.0", "test")
	
	standardName := builder.getStandardName("component")
	expected := "my-plugin-component"
	
	if standardName != expected {
		t.Errorf("Expected standard name '%s', got '%s'", expected, standardName)
	}
}

func TestBuilderBuild(t *testing.T) {
	builder := NewPluginBuilder("test", "1.0.0", "test description")
	plugin := builder.Build()

	if plugin == nil {
		t.Fatal("Build() returned nil")
	}

	// Verify it returns the base plugin
	if plugin != builder.base {
		t.Error("Build() should return the base plugin")
	}

	// Verify plugin implements the Plugin interface
	info := plugin.GetInfo()
	if info.Name != "test" {
		t.Errorf("Expected plugin name 'test', got '%s'", info.Name)
	}
}

func TestBuilderInfoAndBaseSync(t *testing.T) {
	// This test verifies that the builder maintains single source of truth
	// All changes should be reflected through the base
	builder := NewPluginBuilder("sync-test", "1.0.0", "test sync").
		SetAuthor("Sync Author").
		SetLicense("BSD-3-Clause").
		AddTag("sync-test").
		SetMopsVersions("1.5.0", "2.5.0")

	info := builder.base.GetInfo()

	// Check all fields are set correctly
	if info.Name != "sync-test" {
		t.Errorf("Expected name 'sync-test', got '%s'", info.Name)
	}
	if info.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", info.Version)
	}
	if info.Author != "Sync Author" {
		t.Errorf("Expected author 'Sync Author', got '%s'", info.Author)
	}
	if info.License != "BSD-3-Clause" {
		t.Errorf("Expected license 'BSD-3-Clause', got '%s'", info.License)
	}
	if info.MopsMinVersion != "1.5.0" {
		t.Errorf("Expected min version '1.5.0', got '%s'", info.MopsMinVersion)
	}
	if info.MopsMaxVersion != "2.5.0" {
		t.Errorf("Expected max version '2.5.0', got '%s'", info.MopsMaxVersion)
	}
	if len(info.Tags) != 1 || info.Tags[0] != "sync-test" {
		t.Errorf("Expected tags [sync-test], got %v", info.Tags)
	}
}
