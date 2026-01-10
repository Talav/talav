package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/talav/talav/pkg/component/media/app/command"
	"github.com/talav/talav/pkg/component/zorya"
	"github.com/talav/talav/pkg/module/media/dto"
	"gorm.io/gorm"
)

// DeleteMediaHandler handles HTTP requests for deleting media.
type DeleteMediaHandler struct {
	deleteMediaHandler *command.DeleteMediaHandler
}

// NewDeleteMediaHandler creates a new DeleteMediaHandler instance.
func NewDeleteMediaHandler(deleteMediaHandler *command.DeleteMediaHandler) *DeleteMediaHandler {
	return &DeleteMediaHandler{
		deleteMediaHandler: deleteMediaHandler,
	}
}

// Handle handles HTTP DELETE requests to delete a media entry.
func (h *DeleteMediaHandler) Handle(ctx context.Context, req *dto.DeleteMediaRequest) (*dto.DeleteMediaResponse, error) {
	cmd := &command.DeleteMediaCommand{
		ID: req.ID,
	}

	_, err := h.deleteMediaHandler.Handle(ctx, cmd)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, zorya.Error404NotFound("Media not found")
		}

		return nil, err
	}

	return &dto.DeleteMediaResponse{Status: http.StatusNoContent}, nil
}
