package storage

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/gosimple/slug"
	"github.com/talav/talav/pkg/component/media"
	"gocloud.dev/blob"
)

// MediaStorage provides file storage operations for media files
// This is a pure infrastructure service that wraps blob.Bucket.
type MediaStorage struct {
	bucket *blob.Bucket
	config media.MediaConfig
}

// NewMediaStorage creates a new media storage service.
func NewMediaStorage(bucket *blob.Bucket, config media.MediaConfig) *MediaStorage {
	return &MediaStorage{
		bucket: bucket,
		config: config,
	}
}

// Store stores a file and returns the relative storage path
// The path is generated based on preset, media ID, and slugified filename.
func (s *MediaStorage) Store(ctx context.Context, preset, mediaID, originalFileName string, reader io.Reader) (string, error) {
	// Generate storage path
	storagePath := s.GenerateStoragePath(preset, mediaID, originalFileName)

	// Write to blob storage
	writer, err := s.bucket.NewWriter(ctx, storagePath, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create writer: %w", err)
	}

	_, err = io.Copy(writer, reader)
	if err != nil {
		_ = writer.Close()

		return "", fmt.Errorf("failed to write file: %w", err)
	}

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("failed to close writer: %w", err)
	}

	return storagePath, nil
}

// Retrieve retrieves a file by its storage path.
func (s *MediaStorage) Retrieve(ctx context.Context, storagePath string) (io.ReadCloser, error) {
	reader, err := s.bucket.NewReader(ctx, storagePath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return reader, nil
}

// Delete deletes a file by its storage path.
func (s *MediaStorage) Delete(ctx context.Context, storagePath string) error {
	if err := s.bucket.Delete(ctx, storagePath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// GenerateStoragePath generates a storage path without storing the file
// Useful for pre-generating paths before file upload.
func (s *MediaStorage) GenerateStoragePath(preset, mediaID, originalFileName string) string {
	timestamp := time.Now().Format("20060102150405")
	ext := filepath.Ext(originalFileName)
	baseName := strings.TrimSuffix(originalFileName, ext)
	slugified := slug.Make(baseName)

	return fmt.Sprintf("%s/%s/%s_%s%s", preset, mediaID, slugified, timestamp, ext)
}
