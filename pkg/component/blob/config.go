package blob

// FilesystemConfig represents configuration for a single filesystem.
type FilesystemConfig struct {
	URL string `config:"url"`
}

// FilesystemsConfig represents the configuration for the filesystems module.
type BlobConfig struct {
	Filesystems map[string]FilesystemConfig `config:"filesystems"`
}
