package sdk

import (
	"testing"
)

func TestToolChecker(t *testing.T) {
	checker := NewToolChecker()

	// Test basic tool detection
	t.Run("IsInstalled", func(t *testing.T) {
		// Test with a tool that should be available on most systems
		if !checker.IsInstalled("go") {
			t.Skip("Go not installed, skipping test")
		}
		
		// This should pass if go is installed
		if !checker.IsInstalled("go") {
			t.Error("Expected go to be installed")
		}
		
		// This should fail for a non-existent tool
		if checker.IsInstalled("definitely-not-a-real-tool-12345") {
			t.Error("Expected non-existent tool to not be found")
		}
	})

	t.Run("FindTool", func(t *testing.T) {
		info, err := checker.FindTool("go")
		if err != nil {
			t.Skip("Go not installed, skipping test")
		}
		
		if !info.Installed {
			t.Error("Expected tool to be marked as installed")
		}
		
		if info.Path == "" {
			t.Error("Expected tool path to be populated")
		}
		
		if info.Name != "go" {
			t.Errorf("Expected tool name to be 'go', got '%s'", info.Name)
		}
	})

	t.Run("CheckMultipleTools", func(t *testing.T) {
		tools := []string{"go", "definitely-not-a-real-tool-12345"}
		results := checker.CheckMultipleTools(tools)
		
		if len(results) != 2 {
			t.Errorf("Expected 2 results, got %d", len(results))
		}
		
		// Check that non-existent tool is marked as not installed
		if nonExistent, exists := results["definitely-not-a-real-tool-12345"]; exists {
			if nonExistent.Installed {
				t.Error("Expected non-existent tool to be marked as not installed")
			}
		}
	})

	t.Run("RequireTool", func(t *testing.T) {
		// This should fail
		err := checker.RequireTool("definitely-not-a-real-tool-12345")
		if err == nil {
			t.Error("Expected error for non-existent tool")
		}
	})

	t.Run("RequireTools", func(t *testing.T) {
		// Test with mix of existing and non-existing tools
		tools := []string{"go", "definitely-not-a-real-tool-12345"}
		err := checker.RequireTools(tools)
		if err == nil {
			// Only fail if go is actually installed
			if checker.IsInstalled("go") {
				t.Error("Expected error for mixed tool list with non-existent tool")
			}
		}
	})
}

func TestToolCheckerGeneric(t *testing.T) {
	checker := NewToolChecker()
	
	// Test with commonly available tools
	commonTools := []string{"go", "git"}
	results := checker.CheckMultipleTools(commonTools)
	
	// Just verify the function runs without error
	// Results will vary based on what's installed on the system
	if results == nil {
		t.Error("Expected results map to not be nil")
	}
	
	// Log what tools were found for debugging
	t.Logf("Checked %d tools", len(results))
	for toolName, info := range results {
		if info.Installed {
			t.Logf("✓ %s: %s (version: %s)", toolName, info.Path, info.Version)
		} else {
			t.Logf("✗ %s: not found", toolName)
		}
	}
}