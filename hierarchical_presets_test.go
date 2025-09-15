package sdk

import (
	"testing"
	"reflect"
)

func TestHierarchicalPresets(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]ConfigPreset
		expectedDefault map[string]interface{}
	}{
		{
			name: "Global with default preset",
			input: `
name: test-plugin
version: 1.0.0
description: Test plugin
config_presets:
  global:
    config:
      url: "https://example.com/api"
      timeout: 30
  default:
    name: "Default Settings"
    description: "Default configuration"
    config:
      enabled: true
      max_items: 20
`,
			expected: map[string]ConfigPreset{
				"default": {
					Name:        "Default Settings",
					Description: "Default configuration",
					Config: map[string]interface{}{
						"url":       "https://example.com/api", // from global
						"timeout":   30,                        // from global
						"enabled":   true,                      // from default
						"max_items": 20,                        // from default
					},
				},
			},
			expectedDefault: map[string]interface{}{
				"url":       "https://example.com/api",
				"timeout":   30,
				"enabled":   true,
				"max_items": 20,
			},
		},
		{
			name: "Global with multiple presets",
			input: `
name: test-plugin
version: 1.0.0
description: Test plugin
config_presets:
  global:
    config:
      base_url: "https://api.example.com"
      retry_count: 3
  default:
    name: "Default Settings"
    description: "Default configuration"
    config:
      enabled: true
      timeout: 30
  performance:
    name: "Performance Mode"
    description: "High performance settings"
    config:
      enabled: true
      timeout: 10
      batch_size: 100
`,
			expected: map[string]ConfigPreset{
				"default": {
					Name:        "Default Settings",
					Description: "Default configuration",
					Config: map[string]interface{}{
						"base_url":    "https://api.example.com", // from global
						"retry_count": 3,                         // from global
						"enabled":     true,                      // from default
						"timeout":     30,                        // from default
					},
				},
				"performance": {
					Name:        "Performance Mode",
					Description: "High performance settings",
					Config: map[string]interface{}{
						"base_url":    "https://api.example.com", // from global
						"retry_count": 3,                         // from global
						"enabled":     true,                      // from performance
						"timeout":     10,                        // from performance (overrides global)
						"batch_size":  100,                       // from performance
					},
				},
			},
			expectedDefault: map[string]interface{}{
				"base_url":    "https://api.example.com",
				"retry_count": 3,
				"enabled":     true,
				"timeout":     30,
			},
		},
		{
			name: "No global preset",
			input: `
name: test-plugin
version: 1.0.0
description: Test plugin
config_presets:
  default:
    name: "Default Settings"
    description: "Default configuration"
    config:
      enabled: true
      timeout: 30
`,
			expected: map[string]ConfigPreset{
				"default": {
					Name:        "Default Settings",
					Description: "Default configuration",
					Config: map[string]interface{}{
						"enabled": true,
						"timeout": 30,
					},
				},
			},
			expectedDefault: map[string]interface{}{
				"enabled": true,
				"timeout": 30,
			},
		},
		{
			name: "Global only (no other presets)",
			input: `
name: test-plugin
version: 1.0.0
description: Test plugin
config_presets:
  global:
    config:
      base_url: "https://api.example.com"
      timeout: 30
`,
			expected:        map[string]ConfigPreset{}, // Global should not appear as a selectable preset
			expectedDefault: nil,                       // No default preset
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the metadata
			metadata, err := LoadPluginMetadataFromBytes([]byte(tt.input))
			if err != nil {
				t.Fatalf("Failed to load metadata: %v", err)
			}

			// Create plugin builder
			builder := NewPluginBuilderFromMetadata(metadata)
			plugin := builder.Build()
			
			// Get the plugin info
			info := plugin.GetInfo()

			// Check config presets
			if !reflect.DeepEqual(info.ConfigPresets, tt.expected) {
				t.Errorf("Config presets mismatch.\nExpected: %+v\nGot: %+v", tt.expected, info.ConfigPresets)
			}

			// Check default config
			if !reflect.DeepEqual(info.DefaultConfig, tt.expectedDefault) {
				t.Errorf("Default config mismatch.\nExpected: %+v\nGot: %+v", tt.expectedDefault, info.DefaultConfig)
			}

			// Verify global preset is not included as a selectable preset
			if _, exists := info.ConfigPresets["global"]; exists {
				t.Error("Global preset should not be included as a selectable preset")
			}
		})
	}
}

func TestGlobalPresetAsDefaultWhenNoDefaultPreset(t *testing.T) {
	// Test the specific case where there's no default preset and no DefaultConfig,
	// but there is a global preset that should be used as the default
	input := `
name: test-plugin
version: 1.0.0
description: Test plugin with only global preset
config_presets:
  global:
    config:
      url: "https://api.example.com"
      timeout: 30
      enabled: true
  production:
    name: "Production Settings"
    description: "Production configuration"
    config:
      timeout: 10
      max_connections: 100
`

	metadata, err := LoadPluginMetadataFromBytes([]byte(input))
	if err != nil {
		t.Fatalf("Failed to load metadata: %v", err)
	}

	builder := NewPluginBuilderFromMetadata(metadata)
	plugin := builder.Build()
	info := plugin.GetInfo()

	// The default config should be the global preset config
	expectedDefault := map[string]interface{}{
		"url":     "https://api.example.com",
		"timeout": 30,
		"enabled": true,
	}

	if !reflect.DeepEqual(info.DefaultConfig, expectedDefault) {
		t.Errorf("Default config should use global preset when no default preset exists.\nExpected: %+v\nGot: %+v", expectedDefault, info.DefaultConfig)
	}

	// Verify that production preset exists and has merged global config
	productionPreset, exists := info.ConfigPresets["production"]
	if !exists {
		t.Error("Production preset should exist")
	} else {
		expectedProduction := map[string]interface{}{
			"url":             "https://api.example.com", // from global
			"timeout":         10,                        // from production (overrides global)
			"enabled":         true,                      // from global
			"max_connections": 100,                       // from production
		}

		if !reflect.DeepEqual(productionPreset.Config, expectedProduction) {
			t.Errorf("Production preset config mismatch.\nExpected: %+v\nGot: %+v", expectedProduction, productionPreset.Config)
		}
	}

	// Verify global preset is not included as a selectable preset
	if _, exists := info.ConfigPresets["global"]; exists {
		t.Error("Global preset should not be included as a selectable preset")
	}
}

func TestHierarchicalPresetsWithSAML2AWSStructure(t *testing.T) {
	// Test with the actual SAML2AWS structure that uses "settings" instead of "config"
	input := `
name: saml2aws
version: 2.0.0
description: SAML2AWS authentication plugin
config_presets:
  global:
    config:
      url: "https://example.com/api"
      timeout: 30
  default:
    name: "Default Settings"
    description: "Okta SAML Provider Configuration"
    config:
      idp_provider: "Okta"
      cache_saml: true
      skip_prompt: true
  azure_ad:
    name: "Azure AD"
    description: "Azure Active Directory SAML Configuration"
    config:
      idp_provider: "AzureAD"
      cache_saml: true
      skip_prompt: false
`

	metadata, err := LoadPluginMetadataFromBytes([]byte(input))
	if err != nil {
		t.Fatalf("Failed to load metadata: %v", err)
	}

	builder := NewPluginBuilderFromMetadata(metadata)
	plugin := builder.Build()
	info := plugin.GetInfo()

	// Check that default preset has both global and default values
	defaultPreset, exists := info.ConfigPresets["default"]
	if !exists {
		t.Fatal("Default preset not found")
	}

	expectedDefault := map[string]interface{}{
		"url":          "https://example.com/api", // from global
		"timeout":      30,                        // from global
		"idp_provider": "Okta",                    // from default
		"cache_saml":   true,                      // from default
		"skip_prompt":  true,                      // from default
	}

	if !reflect.DeepEqual(defaultPreset.Config, expectedDefault) {
		t.Errorf("Default preset config mismatch.\nExpected: %+v\nGot: %+v", expectedDefault, defaultPreset.Config)
	}

	// Check that azure_ad preset has both global and azure_ad values
	azurePreset, exists := info.ConfigPresets["azure_ad"]
	if !exists {
		t.Fatal("Azure AD preset not found")
	}

	expectedAzure := map[string]interface{}{
		"url":          "https://example.com/api", // from global
		"timeout":      30,                        // from global
		"idp_provider": "AzureAD",                 // from azure_ad
		"cache_saml":   true,                      // from azure_ad
		"skip_prompt":  false,                     // from azure_ad (overrides global)
	}

	if !reflect.DeepEqual(azurePreset.Config, expectedAzure) {
		t.Errorf("Azure AD preset config mismatch.\nExpected: %+v\nGot: %+v", expectedAzure, azurePreset.Config)
	}

	// Check that default config is set to the default preset
	if !reflect.DeepEqual(info.DefaultConfig, expectedDefault) {
		t.Errorf("Plugin default config mismatch.\nExpected: %+v\nGot: %+v", expectedDefault, info.DefaultConfig)
	}
}

func TestDebugPresetProcessing(t *testing.T) {
	// Debug test to see what's actually happening
	input := `
name: debug-test
version: 1.0.0
description: Debug test
config_presets:
  global:
    config:
      global_value: "from_global"
  default:
    name: "Default"
    description: "Default preset"
    config:
      default_value: "from_default"
`

	metadata, err := LoadPluginMetadataFromBytes([]byte(input))
	if err != nil {
		t.Fatalf("Failed to load metadata: %v", err)
	}

	t.Logf("Raw metadata config presets: %+v", metadata.ConfigPresets)

	builder := NewPluginBuilderFromMetadata(metadata)
	plugin := builder.Build()
	info := plugin.GetInfo()

	t.Logf("Final plugin info config presets: %+v", info.ConfigPresets)
	t.Logf("Final plugin info default config: %+v", info.DefaultConfig)
}
