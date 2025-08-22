package output

import (
	"time"
)

// delayedAdapter wraps another adapter and adds configurable delays.
// This is useful for creating animated effects in streaming contexts.
type delayedAdapter struct {
	Adapter
	config     DelayConfig
	skipDelays bool
}

// NewDelayedAdapter creates a new delayed adapter wrapper.
// Set skipDelays to true for CLI/static contexts where delays are not desired.
func NewDelayedAdapter(adapter Adapter, config DelayConfig, skipDelays bool) Adapter {
	return &delayedAdapter{
		Adapter:    adapter,
		config:     config,
		skipDelays: skipDelays,
	}
}

// NewDelayedAdapterWithDefaults creates a delayed adapter with default timing
func NewDelayedAdapterWithDefaults(adapter Adapter, skipDelays bool) Adapter {
	return NewDelayedAdapter(adapter, DefaultDelayConfig(), skipDelays)
}

func (d *delayedAdapter) WriteLine(line string) error {
	if err := d.Adapter.WriteLine(line); err != nil {
		return err
	}
	
	// Add delay only for real-time adapters and if delays are enabled
	if d.Adapter.SupportsRealTime() && d.config.SimulateDelays && !d.skipDelays {
		time.Sleep(d.config.LineDelay)
	}
	
	return nil
}

func (d *delayedAdapter) WriteProgress(current, total int, message string) error {
	if err := d.Adapter.WriteProgress(current, total, message); err != nil {
		return err
	}
	
	if d.Adapter.SupportsRealTime() && d.config.SimulateDelays && !d.skipDelays {
		time.Sleep(d.config.ProgressDelay)
	}
	
	return nil
}

func (d *delayedAdapter) WriteSection(section string) error {
	if err := d.Adapter.WriteSection(section); err != nil {
		return err
	}
	
	if d.Adapter.SupportsRealTime() && d.config.SimulateDelays && !d.skipDelays {
		time.Sleep(d.config.SectionDelay)
	}
	
	return nil
}

func (d *delayedAdapter) WriteHeader(title string) error {
	if err := d.Adapter.WriteHeader(title); err != nil {
		return err
	}
	
	if d.Adapter.SupportsRealTime() && d.config.SimulateDelays && !d.skipDelays {
		time.Sleep(d.config.SectionDelay)
	}
	
	return nil
}
