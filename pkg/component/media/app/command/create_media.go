package command

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/gosimple/slug"
	"github.com/talav/talav/pkg/component/media/app/preset"
	"github.com/talav/talav/pkg/component/media/app/provider"
	"github.com/talav/talav/pkg/component/media/domain"
	"github.com/talav/talav/pkg/component/media/infra/repo"
)

// CreateMediaCommand represents the command to create a new media entry.
type CreateMediaCommand struct {
	Preset       string                // Image processing preset (e.g., "gallery", "product", "avatar")
	ProviderName string                // Provider name (e.g., "any_name_1", "any_name_2")
	File         *multipart.FileHeader // File header from multipart form (contains filename, size, MIME type)
	Description  string                // Optional description/alt text
}

// CreateMediaResult represents the result of creating a media entry.
type CreateMediaResult struct {
	domain.MediaView
}

// CreateMediaHandler handles media creation commands.
type CreateMediaHandler struct {
	mediaRepo        repo.MediaRepository
	providerRegistry provider.Registry
	presetRegistry   preset.Registry
	logger           *slog.Logger
}

// NewCreateMediaHandler creates a new CreateMediaHandler.
func NewCreateMediaHandler(mediaRepo repo.MediaRepository, providerRegistry provider.Registry, presetRegistry preset.Registry, logger *slog.Logger) *CreateMediaHandler {
	return &CreateMediaHandler{
		mediaRepo:        mediaRepo,
		providerRegistry: providerRegistry,
		presetRegistry:   presetRegistry,
		logger:           logger,
	}
}

// Handle processes the CreateMediaCommand and returns the created media entry.
func (h *CreateMediaHandler) Handle(ctx context.Context, cmd *CreateMediaCommand) (*CreateMediaResult, error) {
	// Validate and get provider
	provider, err := h.validateAndGetProvider(cmd.Preset, cmd.ProviderName)
	if err != nil {
		return nil, err
	}

	// Extract file metadata and create media entity
	fileMeta := h.extractFileMetadata(cmd.File)
	media, err := h.createMediaEntity(cmd, fileMeta)
	if err != nil {
		return nil, err
	}

	// Validate provider can process this media
	if !provider.CanProcess(media) {
		h.logger.Error("Provider cannot process file", "provider", cmd.ProviderName, "filename", fileMeta.originalFileName)

		return nil, fmt.Errorf("provider %q cannot process file with extension %q", cmd.ProviderName, fileMeta.ext)
	}

	// Process file (read, extract dimensions, store, generate thumbnails)
	if err := h.processFile(ctx, provider, media, cmd.File, cmd.Preset); err != nil {
		return nil, err
	}

	// Save to repository
	err = h.mediaRepo.Create(ctx, media)
	if err != nil {
		h.logger.Error("Failed to save media to repository", "error", err, "mediaID", media.ID)
		// Attempt to clean up stored file if repository save fails
		_ = provider.Delete(ctx, media)

		return nil, fmt.Errorf("failed to save media: %w", err)
	}

	return &CreateMediaResult{
		MediaView: domain.MediaView{
			Media:     media,
			PublicURL: provider.GetPublicURL(media),
		},
	}, nil
}

// fileMetadata contains extracted file information.
type fileMetadata struct {
	originalFileName string
	fileSize         int64
	mimeType         string
	ext              string
	slugified        string
	mediaType        domain.MediaType
}

// extractFileMetadata extracts metadata from the multipart file header.
func (h *CreateMediaHandler) extractFileMetadata(fileHeader *multipart.FileHeader) fileMetadata {
	originalFileName := fileHeader.Filename
	fileSize := fileHeader.Size
	mimeType := fileHeader.Header.Get("Content-Type")
	ext := filepath.Ext(originalFileName)

	// If Content-Type header is missing, try to detect from extension
	if mimeType == "" {
		mimeType = mime.TypeByExtension(ext)
		if mimeType == "" {
			mimeType = "application/octet-stream" // Default fallback
		}
	}

	// Determine media type from MIME type
	mediaType := determineMediaType(mimeType, originalFileName)

	// Generate slug from original filename without extension
	originalBaseName := strings.TrimSuffix(originalFileName, ext)
	slugified := slug.Make(originalBaseName)

	return fileMetadata{
		originalFileName: originalFileName,
		fileSize:         fileSize,
		mimeType:         mimeType,
		ext:              ext,
		slugified:        slugified,
		mediaType:        mediaType,
	}
}

// createMediaEntity creates a new Media domain entity from command and metadata.
func (h *CreateMediaHandler) createMediaEntity(cmd *CreateMediaCommand, meta fileMetadata) (*domain.Media, error) {
	media, err := domain.NewMedia(
		cmd.Preset,
		cmd.ProviderName,
		meta.mediaType,
		meta.originalFileName,
		"", // fileName parameter is ignored, generated from slug and extension
		"", // URL will be set after storing
		meta.fileSize,
		meta.mimeType,
		0,               // width - not extracted yet
		0,               // height - not extracted yet
		meta.ext,        // extension
		meta.slugified,  // slug
		cmd.Description, // Description (can be empty)
	)
	if err != nil {
		h.logger.Error("Failed to create media domain entity", "error", err)

		return nil, fmt.Errorf("failed to create media: %w", err)
	}

	return media, nil
}

// processFile reads the file, extracts dimensions, stores it, and generates thumbnails.
func (h *CreateMediaHandler) processFile(ctx context.Context, provider provider.Provider, media *domain.Media, fileHeader *multipart.FileHeader, preset string) error {
	// Read file into memory (needed for storage and thumbnails)
	fileBytes, err := h.readMultipartFile(fileHeader)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Extract image dimensions using provider
	h.extractDimensions(provider, media, fileBytes)

	// Use provider to create/store the original file
	url, err := provider.Create(ctx, media, fileBytes)
	if err != nil {
		h.logger.Error("Failed to create file via provider", "error", err, "mediaID", media.ID)

		return fmt.Errorf("failed to create file: %w", err)
	}
	media.URL = url

	// Generate thumbnails if preset has formats configured
	h.generateThumbnails(ctx, provider, media, fileBytes, preset)

	return nil
}

// extractDimensions extracts image dimensions and updates media metadata.
func (h *CreateMediaHandler) extractDimensions(provider provider.Provider, media *domain.Media, fileBytes []byte) {
	width, height, err := provider.GetDimensions(fileBytes)
	if err != nil {
		h.logger.Warn("Failed to extract image dimensions", "error", err, "mediaID", media.ID)

		return
	}

	// Update media metadata with dimensions
	media.Metadata.Width = width
	media.Metadata.Height = height
}

// generateThumbnails generates thumbnails if preset has formats configured.
func (h *CreateMediaHandler) generateThumbnails(ctx context.Context, provider provider.Provider, media *domain.Media, fileBytes []byte, preset string) {
	formats, err := h.presetRegistry.GetPresetFormats(preset)
	if err != nil {
		return
	}

	thumbnails, err := provider.GenerateThumbnails(ctx, media, fileBytes, formats)
	if err != nil {
		h.logger.Warn("Failed to generate thumbnails", "error", err, "mediaID", media.ID)

		return
	}

	if thumbnails != nil {
		media.Thumbnails = thumbnails
	}
}

// validateAndGetProvider validates that the provider is allowed for the preset and retrieves it.
func (h *CreateMediaHandler) validateAndGetProvider(preset, providerName string) (provider.Provider, error) {
	// Validate provider is allowed for preset
	allowed, err := h.presetRegistry.IsProviderAllowed(preset, providerName)
	if err != nil {
		h.logger.Error("Failed to check provider for preset", "error", err, "preset", preset, "provider", providerName)

		return nil, fmt.Errorf("failed to validate preset: %w", err)
	}
	if !allowed {
		h.logger.Error("Provider not allowed for preset", "preset", preset, "provider", providerName)

		return nil, fmt.Errorf("provider %q is not allowed for preset %q", providerName, preset)
	}

	// Get provider from registry
	provider, err := h.providerRegistry.GetProvider(providerName)
	if err != nil {
		h.logger.Error("Failed to get provider", "error", err, "provider", providerName)

		return nil, fmt.Errorf("failed to get provider: %w", err)
	}

	return provider, nil
}

// determineMediaType determines the MediaType from MIME type and filename.
func determineMediaType(mimeType, fileName string) domain.MediaType {
	if mediaType := mediaTypeFromMIME(mimeType); mediaType != domain.MediaTypeFile {
		return mediaType
	}
	if detectedMimeType := mime.TypeByExtension(filepath.Ext(fileName)); detectedMimeType != "" {
		if mediaType := mediaTypeFromMIME(detectedMimeType); mediaType != domain.MediaTypeFile {
			return mediaType
		}
	}

	return domain.MediaTypeFile
}

// mediaTypeFromMIME extracts MediaType from MIME type string.
func mediaTypeFromMIME(mimeType string) domain.MediaType {
	switch {
	case strings.HasPrefix(mimeType, "image/"):
		return domain.MediaTypeImage
	case strings.HasPrefix(mimeType, "video/"):
		return domain.MediaTypeVideo
	default:
		return domain.MediaTypeFile
	}
}

// readMultipartFile reads a multipart file into memory
// multipart.File is typically a one-time stream, so we buffer it once for multiple uses.
func (h *CreateMediaHandler) readMultipartFile(fileHeader *multipart.FileHeader) ([]byte, error) {
	file, err := fileHeader.Open()
	if err != nil {
		h.logger.Error("Failed to open uploaded file", "error", err)

		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			h.logger.Warn("Failed to close uploaded file", "error", closeErr)
		}
	}()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		h.logger.Error("Failed to read file", "error", err)

		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return fileBytes, nil
}
