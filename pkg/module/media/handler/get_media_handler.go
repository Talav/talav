package handler

import (
	"context"
	"errors"

	"github.com/talav/talav/pkg/component/media/app/query"
	"github.com/talav/talav/pkg/component/zorya"
	"github.com/talav/talav/pkg/module/media/dto"
	"gorm.io/gorm"
)

// GetMediaHandler handles HTTP requests for retrieving media data.
type GetMediaHandler struct {
	getMediaQueryHandler *query.GetMediaQueryHandler
}

// NewGetMediaHandler creates a new GetMediaHandler instance.
func NewGetMediaHandler(getMediaQueryHandler *query.GetMediaQueryHandler) *GetMediaHandler {
	return &GetMediaHandler{
		getMediaQueryHandler: getMediaQueryHandler,
	}
}

// Handle handles HTTP GET requests to retrieve a media entry by ID.
func (h *GetMediaHandler) Handle(ctx context.Context, req *dto.GetMediaRequest) (*dto.GetMediaResponse, error) {
	q := &query.GetMediaQuery{ID: req.ID}
	result, err := h.getMediaQueryHandler.Handle(ctx, q)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, zorya.Error404NotFound("Media not found")
		}

		return nil, err
	}

	resp := &dto.GetMediaResponse{}
	resp.Body = dto.ToMediaResponse(result.Media)

	return resp, nil
}
