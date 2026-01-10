package handler

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"mime/multipart"
	"net/http"

	"github.com/talav/talav/pkg/component/media/app/command"
	"github.com/talav/talav/pkg/module/media/dto"
)

// CreateMediaHandler handles HTTP requests for media creation.
type CreateMediaHandler struct {
	createMediaHandler *command.CreateMediaHandler
	logger             *slog.Logger
}

// NewCreateMediaHandler creates a new CreateMediaHandler instance.
func NewCreateMediaHandler(createMediaHandler *command.CreateMediaHandler, logger *slog.Logger) *CreateMediaHandler {
	return &CreateMediaHandler{
		createMediaHandler: createMediaHandler,
		logger:             logger,
	}
}

// Handle handles HTTP POST requests to create a new media entry.
func (h *CreateMediaHandler) Handle(ctx context.Context, req *dto.CreateMediaRequest) (*dto.CreateMediaResponse, error) {
	// Create a temporary file header from the raw bytes
	fileHeader, err := createFileHeader(req.Body.File, "file")
	if err != nil {
		return nil, fmt.Errorf("failed to process uploaded file: %w", err)
	}

	// Create command
	cmd := &command.CreateMediaCommand{
		Preset:       req.Body.Preset,
		ProviderName: req.Body.Provider,
		File:         fileHeader,
		Description:  req.Body.Description,
	}

	// Execute command
	result, err := h.createMediaHandler.Handle(ctx, cmd)
	if err != nil {
		return nil, err
	}

	return dto.FromCreateResult(result), nil
}

// createFileHeader creates a multipart.FileHeader from raw bytes.
func createFileHeader(data []byte, filename string) (*multipart.FileHeader, error) {
	// Create a buffer to write our multipart form
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	// Create the form file field
	fw, err := w.CreateFormFile("file", filename)
	if err != nil {
		return nil, err
	}

	// Write the file data
	if _, err := fw.Write(data); err != nil {
		return nil, err
	}

	// Close the multipart writer to finalize
	if err := w.Close(); err != nil {
		return nil, err
	}

	// Parse the multipart form to get the FileHeader
	r := multipart.NewReader(&b, w.Boundary())
	form, err := r.ReadForm(int64(len(data)) + 1024)
	if err != nil {
		return nil, err
	}

	files := form.File["file"]
	if len(files) == 0 {
		return nil, fmt.Errorf("failed to create file header")
	}

	return files[0], nil
}

// DefaultStatusCode returns the default HTTP status code for successful creation.
func (h *CreateMediaHandler) DefaultStatusCode() int {
	return http.StatusCreated
}
