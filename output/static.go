package output

import (
	"fmt"
	"strings"
)

// staticAdapter implements StaticAdapter for buffered output
type staticAdapter struct {
	buffer *strings.Builder
	style  OutputStyle
}

// NewStaticAdapter creates a new static adapter for buffered output.
// This is typically used for CLI commands and UI static actions.
func NewStaticAdapter() StaticAdapter {
	return NewStaticAdapterWithStyle(DefaultOutputStyle())
}

// NewStaticAdapterWithStyle creates a static adapter with custom styling
func NewStaticAdapterWithStyle(style OutputStyle) StaticAdapter {
	return &staticAdapter{
		buffer: &strings.Builder{},
		style:  style,
	}
}

func (s *staticAdapter) WriteLine(line string) error {
	s.buffer.WriteString(line)
	s.buffer.WriteString("\n")
	return nil
}

func (s *staticAdapter) WriteError(err error) error {
	var errorMsg string
	if s.style.EnableEmojis {
		errorMsg = fmt.Sprintf("❌ Error: %v", err)
	} else {
		errorMsg = fmt.Sprintf("Error: %v", err)
	}
	return s.WriteLine(errorMsg)
}

func (s *staticAdapter) WriteHeader(title string) error {
	if err := s.WriteLine(title); err != nil {
		return err
	}
	separator := strings.Repeat("=", len(title))
	return s.WriteLine(separator)
}

func (s *staticAdapter) WriteSection(section string) error {
	var sectionText string
	if s.style.EnableEmojis {
		sectionText = fmt.Sprintf("📋 %s", section)
	} else {
		sectionText = section
	}
	return s.WriteLine(sectionText)
}

func (s *staticAdapter) WriteBulletPoint(point string) error {
	return s.WriteLine(fmt.Sprintf("  • %s", point))
}

func (s *staticAdapter) WriteProgress(current, total int, message string) error {
	var progress string
	if s.style.EnableEmojis {
		progress = fmt.Sprintf("📈 Progress [%d/%d]: %s", current, total, message)
	} else {
		progress = fmt.Sprintf("Progress [%d/%d]: %s", current, total, message)
	}
	return s.WriteLine(progress)
}

func (s *staticAdapter) WriteStats(stats map[string]interface{}) error {
	var statsHeader string
	if s.style.EnableEmojis {
		statsHeader = "📊 Statistics:"
	} else {
		statsHeader = "Statistics:"
	}
	
	if err := s.WriteLine(statsHeader); err != nil {
		return err
	}
	
	for key, value := range stats {
		if err := s.WriteLine(fmt.Sprintf("  • %s: %v", key, value)); err != nil {
			return err
		}
	}
	return nil
}

func (s *staticAdapter) Flush() error {
	return nil // Nothing to flush for string builder
}

func (s *staticAdapter) SupportsRealTime() bool {
	return false
}

// StaticAdapter specific methods
func (s *staticAdapter) GetOutput() string {
	return s.buffer.String()
}

func (s *staticAdapter) Clear() {
	s.buffer.Reset()
}

// cliAdapter extends staticAdapter with CLI-specific features
type cliAdapter struct {
	*staticAdapter
	enableColors bool
}

// NewCLIAdapter creates a new CLI adapter with color support.
// This is optimized for command-line interfaces.
func NewCLIAdapter(enableColors bool) StaticAdapter {
	style := DefaultOutputStyle()
	style.EnableColors = enableColors
	style.EnableEmojis = true // CLI typically supports emojis
	
	return &cliAdapter{
		staticAdapter: NewStaticAdapterWithStyle(style).(*staticAdapter),
		enableColors:  enableColors,
	}
}

// Add color support methods for CLI
func (c *cliAdapter) WriteHeaderWithColor(title string) error {
	if c.enableColors {
		coloredTitle := fmt.Sprintf("\033[1;34m%s\033[0m", title) // Blue bold
		return c.WriteHeader(coloredTitle)
	}
	return c.WriteHeader(title)
}

func (c *cliAdapter) WriteErrorWithColor(err error) error {
	if c.enableColors {
		coloredError := fmt.Sprintf("\033[1;31m❌ Error: %v\033[0m", err) // Red bold
		return c.WriteLine(coloredError)
	}
	return c.WriteError(err)
}

func (c *cliAdapter) WriteSectionWithColor(section string) error {
	if c.enableColors {
		coloredSection := fmt.Sprintf("\033[1;32m📋 %s\033[0m", section) // Green bold
		return c.WriteLine(coloredSection)
	}
	return c.WriteSection(section)
}
