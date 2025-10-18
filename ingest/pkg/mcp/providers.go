package mcp

import (
	"fmt"
	"sort"
	"strings"
)

// ProviderPreset describes built-in configuration defaults for known MCP providers.
type ProviderPreset struct {
	Name            string
	Description     string
	DefaultEndpoint string
}

var providerPresets = map[string]ProviderPreset{
	"linear": {
		Name:            "linear",
		Description:     "Linear issue tracker official MCP server",
		DefaultEndpoint: "https://mcp.linear.app/sse",
	},
	"github": {
		Name:            "github",
		Description:     "GitHub hosted MCP server (issues, pull requests, repos)",
		DefaultEndpoint: "https://api.githubcopilot.com/mcp/",
	},
}

// ProviderPresets returns a sorted list of supported provider presets.
func ProviderPresets() []ProviderPreset {
	out := make([]ProviderPreset, 0, len(providerPresets))
	for _, preset := range providerPresets {
		out = append(out, preset)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Name < out[j].Name
	})
	return out
}

// ProviderHelp renders a human-friendly description of available presets.
func ProviderHelp() string {
	presets := ProviderPresets()
	if len(presets) == 0 {
		return "  (no built-in providers)"
	}

	var b strings.Builder
	for _, preset := range presets {
		fmt.Fprintf(&b, "  - %s: %s (default endpoint %s)\n", preset.Name, preset.Description, preset.DefaultEndpoint)
	}
	return strings.TrimRight(b.String(), "\n")
}

func providerDefaultEndpoint(name string) (string, bool) {
	preset, ok := providerPresets[name]
	if !ok {
		return "", false
	}
	return preset.DefaultEndpoint, true
}
