package sdk

import (
	"encoding/json"
	"fmt"
)

// PluginInfoRPC is a GOB-safe version of PluginInfo for RPC transport
type PluginInfoRPC struct {
	Name              string            `json:"name"`
	Version           string            `json:"version"`
	Description       string            `json:"description"`
	Author            string            `json:"author"`
	License           string            `json:"license"`
	Homepage          string            `json:"homepage"`
	MopsMinVersion    string            `json:"mops_min_version"`
	MopsMaxVersion    string            `json:"mops_max_version"`
	Dependencies      []string          `json:"dependencies"`
	Tags              []string          `json:"tags"`
	CLICommands       []CLICommandInfo  `json:"cli_commands"`
	DefaultConfigJSON string            `json:"default_config_json"` // JSON-encoded config to avoid GOB issues
	Platform          PlatformInfo      `json:"platform"`
	SupportedPlatforms []PlatformInfo   `json:"supported_platforms"`
	MenuIntegration   *PluginMenuIntegration `json:"menu_integration,omitempty"` // Menu integration config
}

// ToPluginInfo converts PluginInfoRPC to PluginInfo
func (rpc *PluginInfoRPC) ToPluginInfo() (PluginInfo, error) {
	info := PluginInfo{
		Name:               rpc.Name,
		Version:            rpc.Version,
		Description:        rpc.Description,
		Author:             rpc.Author,
		License:            rpc.License,
		Homepage:           rpc.Homepage,
		MopsMinVersion:     rpc.MopsMinVersion,
		MopsMaxVersion:     rpc.MopsMaxVersion,
		Dependencies:       rpc.Dependencies,
		Tags:               rpc.Tags,
		CLICommands:        rpc.CLICommands,
		Platform:           rpc.Platform,
		SupportedPlatforms: rpc.SupportedPlatforms,
		MenuIntegration:    rpc.MenuIntegration,
	}
	
	// Decode JSON config
	if rpc.DefaultConfigJSON != "" {
		if err := json.Unmarshal([]byte(rpc.DefaultConfigJSON), &info.DefaultConfig); err != nil {
			return info, fmt.Errorf("failed to decode default config: %w", err)
		}
	} else {
		info.DefaultConfig = make(map[string]any)
	}
	
	return info, nil
}

// NewPluginInfoRPC converts PluginInfo to PluginInfoRPC
func NewPluginInfoRPC(info PluginInfo) (PluginInfoRPC, error) {
	rpc := PluginInfoRPC{
		Name:               info.Name,
		Version:            info.Version,
		Description:        info.Description,
		Author:             info.Author,
		License:            info.License,
		Homepage:           info.Homepage,
		MopsMinVersion:     info.MopsMinVersion,
		MopsMaxVersion:     info.MopsMaxVersion,
		Dependencies:       info.Dependencies,
		Tags:               info.Tags,
		CLICommands:        info.CLICommands,
		Platform:           info.Platform,
		SupportedPlatforms: info.SupportedPlatforms,
		MenuIntegration:    info.MenuIntegration,
	}
	
	// Encode config as JSON
	if info.DefaultConfig != nil {
		configJSON, err := json.Marshal(info.DefaultConfig)
		if err != nil {
			return rpc, fmt.Errorf("failed to encode default config: %w", err)
		}
		rpc.DefaultConfigJSON = string(configJSON)
	}
	
	return rpc, nil
}
