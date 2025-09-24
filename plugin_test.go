package sdk

import (
	"context"
	"testing"
	"time"
)

func TestPluginInfo(t *testing.T) {
	info := PluginInfo{
		Name:        "test-plugin",
		Version:     "1.0.0",
		Description: "A test plugin",
		Author:      "Test Author",
		License:     "MIT",
	}

	if info.Name != "test-plugin" {
		t.Errorf("Expected name 'test-plugin', got '%s'", info.Name)
	}
	if info.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", info.Version)
	}
}

func TestPlatformInfo(t *testing.T) {
	platform := PlatformInfo{
		OS:   "linux",
		Arch: "amd64",
	}

	if platform.OS != "linux" {
		t.Errorf("Expected OS 'linux', got '%s'", platform.OS)
	}
	if platform.Arch != "amd64" {
		t.Errorf("Expected Arch 'amd64', got '%s'", platform.Arch)
	}
}

func TestCLICommandInfo(t *testing.T) {
	cmd := CLICommandInfo{
		Name:        "test-cmd",
		Description: "A test command",
		Usage:       "test-cmd [args]",
		Examples:    []string{"test-cmd hello"},
	}

	if cmd.Name != "test-cmd" {
		t.Errorf("Expected name 'test-cmd', got '%s'", cmd.Name)
	}
	if len(cmd.Examples) != 1 {
		t.Errorf("Expected 1 example, got %d", len(cmd.Examples))
	}
}

func TestPluginMenuIntegration(t *testing.T) {
	integration := PluginMenuIntegration{
		AutoRegister: true,
		MenuID:       "main",
		Key:          "p",
		Label:        "Test Plugin",
		Icon:         "🧪",
		Priority:     10,
		Group:        "plugins",
	}

	if !integration.AutoRegister {
		t.Error("Expected AutoRegister to be true")
	}
	if integration.MenuID != "main" {
		t.Errorf("Expected MenuID 'main', got '%s'", integration.MenuID)
	}
}

func TestInteractiveGoFunction(t *testing.T) {
	// Test that InteractiveGoFunction type is properly defined
	var fn InteractiveGoFunction = func(ctx context.Context, outputChan chan<- string, inputChan <-chan string, params map[string]interface{}) error {
		outputChan <- "test output"
		return nil
	}

	if fn == nil {
		t.Error("InteractiveGoFunction should not be nil")
	}

	// Test function execution
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	outputChan := make(chan string, 1)
	inputChan := make(chan string)
	params := make(map[string]interface{})

	go func() {
		defer close(outputChan)
		err := fn(ctx, outputChan, inputChan, params)
		if err != nil {
			t.Errorf("Function execution failed: %v", err)
		}
	}()

	select {
	case output := <-outputChan:
		if output != "test output" {
			t.Errorf("Expected 'test output', got '%s'", output)
		}
	case <-ctx.Done():
		t.Error("Function execution timed out")
	}
}

func TestActionResult(t *testing.T) {
	result := ActionResult{
		Success:     true,
		Output:      "Command executed successfully",
		Error:       nil,
		ShowOutput:  true,
		Title:       "Success",
		IsStreaming: false,
		RefreshMenu: true,
	}

	if !result.Success {
		t.Error("Expected Success to be true")
	}
	if result.Output != "Command executed successfully" {
		t.Errorf("Expected output 'Command executed successfully', got '%s'", result.Output)
	}
	if !result.RefreshMenu {
		t.Error("Expected RefreshMenu to be true")
	}
}

func TestMenuEntry(t *testing.T) {
	entry := MenuEntry{
		Key:           "1",
		Label:         "Test Entry",
		Action:        "test-action",
		Target:        "test-target",
		Message:       "Test message",
		Command:       "test-command",
		NextID:        "next-menu",
		FilePath:      "/path/to/file",
		Params:        map[string]interface{}{"key": "value"},
		IsDynamic:     true,
		KeyBind:       "Ctrl+T",
		IsInteractive: false,
	}

	if entry.Key != "1" {
		t.Errorf("Expected Key '1', got '%s'", entry.Key)
	}
	if entry.Label != "Test Entry" {
		t.Errorf("Expected Label 'Test Entry', got '%s'", entry.Label)
	}
	if !entry.IsDynamic {
		t.Error("Expected IsDynamic to be true")
	}
	if len(entry.Params) != 1 {
		t.Errorf("Expected 1 param, got %d", len(entry.Params))
	}
	if entry.Params["key"] != "value" {
		t.Errorf("Expected param value 'value', got '%v'", entry.Params["key"])
	}
}

func TestRepositoryPlugin(t *testing.T) {
	plugin := RepositoryPlugin{
		Name:        "test-repo-plugin",
		Version:     "2.0.0",
		Description: "A repository plugin",
		Author:      "Repo Author",
		License:     "Apache-2.0",
		Homepage:    "https://example.com",
		Repository:  "https://github.com/test/repo",
		Tags:        []string{"utility", "cli"},
		MopsVersion: MopsVersionConstraint{
			MinVersion: "1.0.0",
			MaxVersion: "3.0.0",
		},
		Platforms: map[string]RepositoryPluginAsset{
			"linux-amd64": {
				Platform: struct {
					OS   string `json:"os"`
					Arch string `json:"arch"`
				}{
					OS:   "linux",
					Arch: "amd64",
				},
				DownloadURL: "https://example.com/download",
				Size:        1024,
				Checksum:    "sha256:abc123",
				Filename:    "plugin.tar.gz",
			},
		},
		Checksums:   map[string]string{"linux-amd64": "sha256:abc123"},
		LastUpdated: time.Now(),
	}

	if plugin.Name != "test-repo-plugin" {
		t.Errorf("Expected name 'test-repo-plugin', got '%s'", plugin.Name)
	}
	if len(plugin.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(plugin.Tags))
	}
	if len(plugin.Platforms) != 1 {
		t.Errorf("Expected 1 platform, got %d", len(plugin.Platforms))
	}

	platform, exists := plugin.Platforms["linux-amd64"]
	if !exists {
		t.Error("Expected linux-amd64 platform to exist")
	}
	if platform.Size != 1024 {
		t.Errorf("Expected size 1024, got %d", platform.Size)
	}
}
