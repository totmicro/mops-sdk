package output_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/totmicro/mops-sdk/output"
)

func TestStaticAdapter(t *testing.T) {
	adapter := output.NewForUI()
	
	err := adapter.WriteHeader("Test Header")
	if err != nil {
		t.Fatalf("WriteHeader failed: %v", err)
	}
	
	err = adapter.WriteSection("Test Section")
	if err != nil {
		t.Fatalf("WriteSection failed: %v", err)
	}
	
	err = adapter.WriteBulletPoint("Test bullet point")
	if err != nil {
		t.Fatalf("WriteBulletPoint failed: %v", err)
	}
	
	output := adapter.GetOutput()
	
	// Verify content
	if !strings.Contains(output, "Test Header") {
		t.Error("Output should contain header")
	}
	
	if !strings.Contains(output, "📋 Test Section") {
		t.Error("Output should contain section with emoji")
	}
	
	if !strings.Contains(output, "• Test bullet point") {
		t.Error("Output should contain bullet point")
	}
}

func TestCLIAdapter(t *testing.T) {
	adapter := output.NewForCLI(false) // no colors for testing
	
	err := adapter.WriteProgress(1, 3, "Processing item 1")
	if err != nil {
		t.Fatalf("WriteProgress failed: %v", err)
	}
	
	stats := map[string]interface{}{
		"Items processed": 3,
		"Duration":        "1.5s",
		"Success":         true,
	}
	
	err = adapter.WriteStats(stats)
	if err != nil {
		t.Fatalf("WriteStats failed: %v", err)
	}
	
	output := adapter.GetOutput()
	
	if !strings.Contains(output, "Progress [1/3]") {
		t.Error("Output should contain progress indicator")
	}
	
	if !strings.Contains(output, "Statistics:") {
		t.Error("Output should contain statistics header")
	}
	
	if !strings.Contains(output, "Items processed: 3") {
		t.Error("Output should contain stats data")
	}
}

func TestStreamingAdapter(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	outputChan := make(chan string, 10)
	adapter := output.NewForStreaming(ctx, outputChan)
	
	// Test basic streaming
	go func() {
		defer close(outputChan)
		
		_ = adapter.WriteHeader("Streaming Test")
		_ = adapter.WriteLine("Line 1")
		_ = adapter.WriteLine("Line 2")
		_ = adapter.WriteProgress(1, 2, "Processing")
	}()
	
	// Collect output
	var lines []string
	for line := range outputChan {
		lines = append(lines, line)
	}
	
	if len(lines) < 4 {
		t.Errorf("Expected at least 4 lines, got %d", len(lines))
	}
	
	// Verify streaming worked
	found := false
	for _, line := range lines {
		if strings.Contains(line, "Streaming Test") {
			found = true
			break
		}
	}
	
	if !found {
		t.Error("Header should be streamed")
	}
}

func TestProgressSequence(t *testing.T) {
	adapter := output.NewForUI()
	
	items := []output.ProgressItem{
		{
			Name:        "Item 1",
			Description: "First item",
			Processor: func() error {
				return nil // Success
			},
		},
		{
			Name:        "Item 2", 
			Description: "Second item",
			Processor: func() error {
				return nil // Success
			},
		},
	}
	
	err := output.WriteProgressSequence(adapter, "Processing Items", items)
	if err != nil {
		t.Fatalf("WriteProgressSequence failed: %v", err)
	}
	
	result := adapter.GetOutput()
	
	if !strings.Contains(result, "Processing Items") {
		t.Error("Output should contain title")
	}
	
	if !strings.Contains(result, "Progress [1/2]") {
		t.Error("Output should contain progress for first item")
	}
	
	if !strings.Contains(result, "✅ Completed Item 1") {
		t.Error("Output should contain completion message")
	}
}

func TestDemoFramework(t *testing.T) {
	adapter := output.NewForUI()
	
	steps := []output.DemoStep{
		{
			Name:        "Initialize",
			Description: "Setting up demo",
			Action: func(a output.Adapter) error {
				return a.WriteLine("Demo initialized")
			},
		},
		{
			Name:        "Execute",
			Description: "Running main logic",
			Action: func(a output.Adapter) error {
				return a.WriteBulletPoint("Main logic executed")
			},
		},
	}
	
	err := output.RunDemo(adapter, "Test Demo", steps)
	if err != nil {
		t.Fatalf("RunDemo failed: %v", err)
	}
	
	result := adapter.GetOutput()
	
	if !strings.Contains(result, "Test Demo") {
		t.Error("Output should contain demo title")
	}
	
	if !strings.Contains(result, "✅ Initialize") {
		t.Error("Output should contain step completion")
	}
	
	if !strings.Contains(result, "Demo completed successfully") {
		t.Error("Output should contain success message")
	}
}
