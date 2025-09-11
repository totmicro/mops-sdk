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

	// Set default config if available
	if metadata.DefaultConfig != nil {
		builder.SetDefaultConfig(metadata.DefaultConfig)
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
