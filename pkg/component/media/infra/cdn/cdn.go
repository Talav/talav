package cdn

// CDN defines the interface for CDN operations.
type CDN interface {
	// GetPath returns the full URL for a given relative path
	GetPath(relativePath string) string
}
