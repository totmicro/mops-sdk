package output

import (
	"fmt"
	"time"
)

// WriteProgressSequence executes a sequence of items with progress indicators.
// This is a common pattern for processing multiple items with visual feedback.
func WriteProgressSequence(adapter Adapter, title string, items []ProgressItem) error {
	if title != "" {
		if err := adapter.WriteSection(title); err != nil {
			return err
		}
		if err := adapter.WriteLine(""); err != nil {
			return err
		}
	}
	
	for i, item := range items {
		// Show progress
		if err := adapter.WriteProgress(i+1, len(items), fmt.Sprintf("Processing %s", item.Name)); err != nil {
			return err
		}
		
		// Process the item
		if item.Processor != nil {
			if err := item.Processor(); err != nil {
				if err := adapter.WriteError(fmt.Errorf("failed to process %s: %w", item.Name, err)); err != nil {
					return err
				}
				continue
			}
		}
		
		// Show completion
		completionMsg := fmt.Sprintf("✅ Completed %s", item.Name)
		if item.Description != "" {
			completionMsg += fmt.Sprintf(" - %s", item.Description)
		}
		
		if err := adapter.WriteBulletPoint(completionMsg); err != nil {
			return err
		}
	}
	
	return nil
}

// WriteStatsTable writes statistics in a formatted table-like structure
func WriteStatsTable(adapter Adapter, title string, stats []StatsEntry) error {
	if title != "" {
		if err := adapter.WriteSection(title); err != nil {
			return err
		}
	}
	
	for _, stat := range stats {
		var line string
		if stat.Unit != "" {
			line = fmt.Sprintf("  • %s: %v %s", stat.Key, stat.Value, stat.Unit)
		} else {
			line = fmt.Sprintf("  • %s: %v", stat.Key, stat.Value)
		}
		
		if err := adapter.WriteBulletPoint(line); err != nil {
			return err
		}
	}
	
	return nil
}

// WriteTimedSequence writes a sequence of lines with specific delays between them.
// Useful for creating dramatic effects in streaming contexts.
func WriteTimedSequence(adapter StreamingAdapter, lines []string, delay time.Duration) error {
	for _, line := range lines {
		if err := adapter.WriteWithDelay(line, delay); err != nil {
			return err
		}
	}
	return nil
}

// WriteSeparator writes a visual separator line
func WriteSeparator(adapter Adapter, length int, char string) error {
	if char == "" {
		char = "-"
	}
	if length <= 0 {
		length = 50
	}
	
	separator := ""
	for i := 0; i < length; i++ {
		separator += char
	}
	
	return adapter.WriteLine(separator)
}

// WriteBoxedMessage writes a message surrounded by a box
func WriteBoxedMessage(adapter Adapter, message string, title string) error {
	maxLen := len(message)
	if len(title) > maxLen {
		maxLen = len(title)
	}
	
	// Top border
	topBorder := "┌" + repeatString("─", maxLen+2) + "┐"
	if err := adapter.WriteLine(topBorder); err != nil {
		return err
	}
	
	// Title if provided
	if title != "" {
		titleLine := "│ " + padString(title, maxLen) + " │"
		if err := adapter.WriteLine(titleLine); err != nil {
			return err
		}
		
		// Separator
		separator := "├" + repeatString("─", maxLen+2) + "┤"
		if err := adapter.WriteLine(separator); err != nil {
			return err
		}
	}
	
	// Message
	messageLine := "│ " + padString(message, maxLen) + " │"
	if err := adapter.WriteLine(messageLine); err != nil {
		return err
	}
	
	// Bottom border
	bottomBorder := "└" + repeatString("─", maxLen+2) + "┘"
	return adapter.WriteLine(bottomBorder)
}

// WriteMultilineBoxedMessage writes a multi-line message in a box
func WriteMultilineBoxedMessage(adapter Adapter, lines []string, title string) error {
	if len(lines) == 0 {
		return nil
	}
	
	// Find max length
	maxLen := 0
	for _, line := range lines {
		if len(line) > maxLen {
			maxLen = len(line)
		}
	}
	if title != "" && len(title) > maxLen {
		maxLen = len(title)
	}
	
	// Top border
	topBorder := "┌" + repeatString("─", maxLen+2) + "┐"
	if err := adapter.WriteLine(topBorder); err != nil {
		return err
	}
	
	// Title if provided
	if title != "" {
		titleLine := "│ " + padString(title, maxLen) + " │"
		if err := adapter.WriteLine(titleLine); err != nil {
			return err
		}
		
		// Separator
		separator := "├" + repeatString("─", maxLen+2) + "┤"
		if err := adapter.WriteLine(separator); err != nil {
			return err
		}
	}
	
	// Content lines
	for _, line := range lines {
		contentLine := "│ " + padString(line, maxLen) + " │"
		if err := adapter.WriteLine(contentLine); err != nil {
			return err
		}
	}
	
	// Bottom border
	bottomBorder := "└" + repeatString("─", maxLen+2) + "┘"
	return adapter.WriteLine(bottomBorder)
}

// Helper functions
func repeatString(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}

func padString(s string, length int) string {
	for len(s) < length {
		s += " "
	}
	return s
}
