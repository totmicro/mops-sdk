// Package sdk provides the MOPS SDK for building plugins
package sdk

import (
	"context"
)

// ActionResult represents the result of executing an action
type ActionResult struct {
	Success     bool
	Output      string
	Error       error
	ShowOutput  bool
	Title       string
	IsStreaming bool
	RefreshMenu bool // Indicates if the parent menu should refresh its dynamic content
}

// InteractiveGoFunction represents a Go function that can be executed with real-time interaction
type InteractiveGoFunction func(ctx context.Context, outputChan chan<- string, inputChan <-chan string, params map[string]interface{}) error

// DynamicProvider is an interface that provides dynamic menu entries
type DynamicProvider interface {
	GetName() string
	GetDescription() string
	GenerateEntries(param string) ([]MenuEntry, error)
	SupportsRefresh() bool
}

// AppState represents the application state
type AppState int

const (
	MainMenu AppState = iota
	DoneScreen
)

// NavMsg represents navigation messages
type NavMsg struct {
	Next          AppState
	NextID        string
	Data          any
	IsBack        bool
	Params        map[string]interface{}
	TitleOverride string
}

// MenuConfig represents the menu configuration
type MenuConfig struct {
	Menus []Menu `yaml:"menus"`
}

// Menu represents a menu definition
type Menu struct {
	ID            string          `yaml:"id"`
	Title         string          `yaml:"title"`
	Input         bool            `yaml:"input,omitempty"`
	Entries       []MenuEntry     `yaml:"entries,omitempty"`
	ConfirmAction *MenuAction     `yaml:"confirm_action,omitempty"`
	CancelKey     string          `yaml:"cancel_key,omitempty"`
	CancelNextID  string          `yaml:"cancel_next_id,omitempty"`
	Provider      string          `yaml:"provider,omitempty"`
	ProviderParam string          `yaml:"provider_param,omitempty"`
	Map           map[string]Menu `yaml:"-"`
}

// MenuEntry represents a menu entry
type MenuEntry struct {
	Key           string                 `yaml:"key"`
	Label         string                 `yaml:"label"`
	ID            string                 `yaml:"id,omitempty"`          // Unique identifier for CLI auto-execution
	Action        string                 `yaml:"action"`
	Target        string                 `yaml:"target,omitempty"`
	Message       string                 `yaml:"message,omitempty"`
	Command       string                 `yaml:"command,omitempty"`
	NextID        string                 `yaml:"next_id,omitempty"`
	FilePath      string                 `yaml:"file_path,omitempty"`
	Params        map[string]interface{} `yaml:"params,omitempty"`
	IsDynamic     bool                   `yaml:"-"`
	KeyBind       string                 `yaml:"key_bind,omitempty"`
	IsInteractive bool                   `yaml:"is_interactive,omitempty"`
}

// MenuAction represents a menu action
type MenuAction struct {
	Action  string                 `yaml:"action"`
	Target  string                 `yaml:"target,omitempty"`
	Command string                 `yaml:"command,omitempty"`
	NextID  string                 `yaml:"next_id,omitempty"`
	Params  map[string]interface{} `yaml:"params,omitempty"`
}
