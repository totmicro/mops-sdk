package sdk

import (
	"fmt"
	_ "embed"
	"gopkg.in/yaml.v3"
)

// LoadPluginMetadataFromBytes loads metadata from YAML bytes (for embedded data)
func LoadPluginMetadataFromBytes(data []byte) (*PluginMetadata, error) {
	var metadata PluginMetadata
	if err := yaml.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse embedded plugin.yaml: %w", err)
	}

	// Validate required fields
	if metadata.Name == "" {
		return nil, fmt.Errorf("plugin name is required in metadata")
	}
	if metadata.Version == "" {
		return nil, fmt.Errorf("plugin version is required in metadata")
	}
	if metadata.Description == "" {
		return nil, fmt.Errorf("plugin description is required in metadata")
	}

	return &metadata, nil
}

// NewPluginBuilderFromEmbeddedYAML creates a PluginBuilder from embedded YAML data
func NewPluginBuilderFromEmbeddedYAML(embeddedData []byte) (*PluginBuilder, error) {
	metadata, err := LoadPluginMetadataFromBytes(embeddedData)
	if err != nil {
		return nil, fmt.Errorf("failed to load embedded metadata: %w", err)
	}

	return NewPluginBuilderFromMetadata(metadata), nil
}

// NewPluginBuilderFromMetadata creates a PluginBuilder using provided metadata
func NewPluginBuilderFromMetadata(metadata *PluginMetadata) *PluginBuilder {
	builder := NewPluginBuilder(metadata.Name, metadata.Version, metadata.Description)

	// Set optional fields if available
	if metadata.DisplayName != "" {
		builder.SetDisplayName(metadata.DisplayName)
	}
	if metadata.Author != "" {
		builder.SetAuthor(metadata.Author)
	}
	if metadata.License != "" {
		builder.SetLicense(metadata.License)
	}
	if metadata.Homepage != "" {
		builder.SetHomepage(metadata.Homepage)
	}

	// Add tags
	for _, tag := range metadata.Tags {
		builder.AddTag(tag)
	}

	// Handle hierarchical config presets with global base
	var globalConfig map[string]interface{}
	if globalPreset, exists := metadata.ConfigPresets["global"]; exists {
		globalConfig = globalPreset.Config
	}
	
	// Process each preset (excluding global)
	for presetName, preset := range metadata.ConfigPresets {
		if presetName == "global" {
			continue // Skip global preset as it's not a selectable preset
		}
		
		// Start with global config as base
		mergedConfig := make(map[string]interface{})
		for key, value := range globalConfig {
			mergedConfig[key] = value
		}
		
		// Merge preset-specific config (overwrites global values)
		for key, value := range preset.Config {
			mergedConfig[key] = value
		}
		
		// Register the merged preset
		builder.WithConfigPreset(presetName, preset.Name, preset.Description, mergedConfig)
		
		// If this is the default preset, also set it as the plugin's default config
		if presetName == "default" {
			builder.SetDefaultConfig(mergedConfig)
		}
	}

	// Set default config if available and no default preset was found
	hasDefaultPreset := false
	hasOtherPresets := false
	
	// Check if default preset exists and if there are other presets besides global
	for presetName := range metadata.ConfigPresets {
		if presetName == "default" {
			hasDefaultPreset = true
		} else if presetName != "global" {
			hasOtherPresets = true
		}
	}
	
	// Set default config if available and no default preset was found
	if metadata.DefaultConfig != nil && !hasDefaultPreset {
		builder.SetDefaultConfig(metadata.DefaultConfig)
	}
	
	// If no DefaultConfig and no default preset, use global preset as fallback
	// only if there are other presets (not just global alone)
	if metadata.DefaultConfig == nil && !hasDefaultPreset && hasOtherPresets {
		if globalPreset, exists := metadata.ConfigPresets["global"]; exists {
			// Use the global preset config as default
			globalConfig := make(map[string]any)
			for key, value := range globalPreset.Config {
				globalConfig[key] = value
			}
			builder.SetDefaultConfig(globalConfig)
		}
	}

	return builder
}

// MustNewPluginBuilderFromEmbeddedYAML creates a PluginBuilder from embedded YAML and panics on error
func MustNewPluginBuilderFromEmbeddedYAML(embeddedData []byte) *PluginBuilder {
	builder, err := NewPluginBuilderFromEmbeddedYAML(embeddedData)
	if err != nil {
		panic(fmt.Sprintf("Failed to create plugin builder from embedded YAML: %v", err))
	}
	return builder
}
