package provider

// Config represents configuration for a provider.
type ProviderConfig struct {
	Type       string   `config:"type"`       // file or image
	Extensions []string `config:"extensions"` // list of supported extensions
	Filesystem string   `config:"filesystem"` // filesystem name for this provider
	CDN        string   `config:"cdn"`        // CDN name for this provider (defaults to "default")
}

// FormatConfig represents configuration for a thumbnail format.
type FormatConfig struct {
	Width   int            // thumbnail width
	Height  int            // thumbnail height
	Resizer string         // resizer name to use
	Format  string         // optional: thumbnail format (e.g., "jpg", "png", "gif"). If empty, uses original format
	Options map[string]any // optional format-specific options
}
