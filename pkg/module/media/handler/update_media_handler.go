package handler

import (
	"context"
	"errors"

	"github.com/talav/talav/pkg/component/media/app/command"
	"github.com/talav/talav/pkg/component/zorya"
	"github.com/talav/talav/pkg/module/media/dto"
	"gorm.io/gorm"
)

// UpdateMediaHandler handles HTTP requests for updating media.
type UpdateMediaHandler struct {
	updateMediaHandler *command.UpdateMediaHandler
}

// NewUpdateMediaHandler creates a new UpdateMediaHandler instance.
func NewUpdateMediaHandler(updateMediaHandler *command.UpdateMediaHandler) *UpdateMediaHandler {
	return &UpdateMediaHandler{
		updateMediaHandler: updateMediaHandler,
	}
}

// Handle handles HTTP PATCH requests to update a media entry.
func (h *UpdateMediaHandler) Handle(ctx context.Context, req *dto.UpdateMediaRequest) (*dto.UpdateMediaResponse, error) {
	cmd := &command.UpdateMediaCommand{
		ID:          req.ID,
		Description: req.Body.Description,
	}

	result, err := h.updateMediaHandler.Handle(ctx, cmd)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, zorya.Error404NotFound("Media not found")
		}

		return nil, err
	}

	return dto.FromUpdateResult(result), nil
}
