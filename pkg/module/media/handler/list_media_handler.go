package handler

import (
	"context"

	"github.com/talav/talav/pkg/component/media/app/query"
	"github.com/talav/talav/pkg/module/media/dto"
)

// ListMediaHandler handles HTTP requests for listing media.
type ListMediaHandler struct {
	listMediaQueryHandler *query.ListMediaQueryHandler
}

// NewListMediaHandler creates a new ListMediaHandler instance.
func NewListMediaHandler(listMediaQueryHandler *query.ListMediaQueryHandler) *ListMediaHandler {
	return &ListMediaHandler{
		listMediaQueryHandler: listMediaQueryHandler,
	}
}

// Handle handles HTTP GET requests to list media entries with pagination and filtering.
func (h *ListMediaHandler) Handle(ctx context.Context, req *dto.ListMediaRequest) (*dto.ListMediaResponse, error) {
	// Set default limit if not provided
	limit := req.Limit
	if limit <= 0 {
		limit = 10
	}

	q := &query.ListMediaQuery{
		Limit:  limit,
		Cursor: req.Cursor,
		Preset: req.Preset,
		Type:   req.Type,
	}

	result, err := h.listMediaQueryHandler.Handle(ctx, q)
	if err != nil {
		return nil, err
	}

	return dto.FromListResult(result), nil
}
