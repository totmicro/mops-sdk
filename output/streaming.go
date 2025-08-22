package output

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// streamingAdapter implements StreamingAdapter for real-time output via channels
type streamingAdapter struct {
	outputChan chan<- string
	ctx        context.Context
	style      OutputStyle
}

// NewStreamingAdapter creates a new streaming adapter for real-time output.
// This is typically used for UI interactive functions.
func NewStreamingAdapter(ctx context.Context, outputChan chan<- string) StreamingAdapter {
	return NewStreamingAdapterWithStyle(ctx, outputChan, DefaultOutputStyle())
}

// NewStreamingAdapterWithStyle creates a streaming adapter with custom styling
func NewStreamingAdapterWithStyle(ctx context.Context, outputChan chan<- string, style OutputStyle) StreamingAdapter {
	return &streamingAdapter{
		outputChan: outputChan,
		ctx:        ctx,
		style:      style,
	}
}

func (s *streamingAdapter) WriteLine(line string) error {
	select {
	case s.outputChan <- line:
		return nil
	case <-s.ctx.Done():
		return s.ctx.Err()
	}
}

func (s *streamingAdapter) WriteError(err error) error {
	var errorMsg string
	if s.style.EnableEmojis {
		errorMsg = fmt.Sprintf("❌ Error: %v", err)
	} else {
		errorMsg = fmt.Sprintf("Error: %v", err)
	}
	return s.WriteLine(errorMsg)
}

func (s *streamingAdapter) WriteHeader(title string) error {
	if err := s.WriteLine(title); err != nil {
		return err
	}
	separator := strings.Repeat("=", len(title))
	return s.WriteLine(separator)
}

func (s *streamingAdapter) WriteSection(section string) error {
	var sectionText string
	if s.style.EnableEmojis {
		sectionText = fmt.Sprintf("📋 %s", section)
	} else {
		sectionText = section
	}
	return s.WriteLine(sectionText)
}

func (s *streamingAdapter) WriteBulletPoint(point string) error {
	return s.WriteLine(fmt.Sprintf("  • %s", point))
}

func (s *streamingAdapter) WriteProgress(current, total int, message string) error {
	var progress string
	if s.style.EnableEmojis {
		progress = fmt.Sprintf("📈 Progress [%d/%d]: %s", current, total, message)
	} else {
		progress = fmt.Sprintf("Progress [%d/%d]: %s", current, total, message)
	}
	return s.WriteLine(progress)
}

func (s *streamingAdapter) WriteStats(stats map[string]interface{}) error {
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

func (s *streamingAdapter) Flush() error {
	return nil // No buffering for streaming
}

func (s *streamingAdapter) SupportsRealTime() bool {
	return true
}

// StreamingAdapter specific methods
func (s *streamingAdapter) WriteRaw(content string) error {
	return s.WriteLine(content)
}

func (s *streamingAdapter) WriteWithDelay(line string, delay time.Duration) error {
	if err := s.WriteLine(line); err != nil {
		return err
	}
	
	select {
	case <-time.After(delay):
		return nil
	case <-s.ctx.Done():
		return s.ctx.Err()
	}
}
