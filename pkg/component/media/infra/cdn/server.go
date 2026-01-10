package cdn

import (
	"strings"
)

// Server implements a simple server CDN that prepends a base path to relative paths.
type Server struct {
	path string
}

// NewServer creates a new Server CDN instance.
func NewServer(path string) *Server {
	return &Server{
		path: strings.TrimRight(path, "/"),
	}
}

// GetPath returns the full URL by combining the base path with the relative path.
func (s *Server) GetPath(relativePath string) string {
	trimmed := strings.TrimLeft(relativePath, "/")

	return s.path + "/" + trimmed
}
