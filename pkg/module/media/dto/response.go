package dto

import (
	"github.com/talav/talav/pkg/component/media/app/command"
	"github.com/talav/talav/pkg/component/media/app/query"
	"github.com/talav/talav/pkg/component/media/domain"
)

// MediaResponse represents the HTTP response for media data.
type MediaResponse struct {
	ID               string                       `json:"id"`
	Preset           string                       `json:"preset"`
	Type             string                       `json:"type"`
	OriginalFileName string                       `json:"original_file_name"`
	FileName         string                       `json:"file_name"`
	URL              string                       `json:"url"`
	FileSize         int64                        `json:"file_size"`
	MimeType         string                       `json:"mime_type"`
	Width            int                          `json:"width"`
	Height           int                          `json:"height"`
	Description      string                       `json:"description"`
	Thumbnails       map[string]ThumbnailResponse `json:"thumbnails,omitempty"`
	CreatedAt        int64                        `json:"created_at"`
	UpdatedAt        int64                        `json:"updated_at"`
}

// ThumbnailResponse represents thumbnail metadata in the HTTP response.
type ThumbnailResponse struct {
	URL       string `json:"url"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	FileSize  int64  `json:"file_size"`
	Extension string `json:"extension"`
}

// CreateMediaResponse represents the HTTP response for media creation.
type CreateMediaResponse struct {
	Body struct {
		Media MediaResponse `json:"media"`
	} `body:"structured"`
}

// GetMediaResponse represents the HTTP response for getting media.
type GetMediaResponse struct {
	Body MediaResponse `body:"structured"`
}

// UpdateMediaResponse represents the HTTP response for media update.
type UpdateMediaResponse struct {
	Body struct {
		Media MediaResponse `json:"media"`
	} `body:"structured"`
}

// ListMediaResponse represents the HTTP response for listing media.
type ListMediaResponse struct {
	Body struct {
		Media      []MediaResponse `json:"media"`
		NextCursor string          `json:"next_cursor,omitempty"`
		HasMore    bool            `json:"has_more"`
	} `body:"structured"`
}

// DeleteMediaResponse represents the HTTP response for media deletion.
type DeleteMediaResponse struct {
	Status int `status:""`
}

// ToMediaResponse converts a domain.Media to MediaResponse.
func ToMediaResponse(m *domain.Media) MediaResponse {
	thumbnails := make(map[string]ThumbnailResponse)
	for name, thumb := range m.Thumbnails {
		thumbnails[name] = ThumbnailResponse{
			URL:       thumb.URL,
			Width:     thumb.Width,
			Height:    thumb.Height,
			FileSize:  thumb.FileSize,
			Extension: thumb.Extension,
		}
	}

	return MediaResponse{
		ID:               m.ID,
		Preset:           m.Preset,
		Type:             string(m.Type),
		OriginalFileName: m.OriginalFileName,
		FileName:         m.FileName,
		URL:              m.URL,
		FileSize:         m.Metadata.FileSize,
		MimeType:         m.Metadata.MimeType,
		Width:            m.Metadata.Width,
		Height:           m.Metadata.Height,
		Description:      m.Description,
		Thumbnails:       thumbnails,
		CreatedAt:        m.CreatedAt,
		UpdatedAt:        m.UpdatedAt,
	}
}

// ToMediaResponseFromView converts a domain.MediaView to MediaResponse.
func ToMediaResponseFromView(m *domain.MediaView) MediaResponse {
	resp := ToMediaResponse(m.Media)
	// URL is already handled in ToMediaResponse from m.Media.URL
	return resp
}

// FromCreateResult converts command.CreateMediaResult to CreateMediaResponse.
func FromCreateResult(result *command.CreateMediaResult) *CreateMediaResponse {
	resp := &CreateMediaResponse{}
	resp.Body.Media = ToMediaResponseFromView(&result.MediaView)

	return resp
}

// FromUpdateResult converts command.UpdateMediaResult to UpdateMediaResponse.
func FromUpdateResult(result *command.UpdateMediaResult) *UpdateMediaResponse {
	resp := &UpdateMediaResponse{}
	resp.Body.Media = ToMediaResponseFromView(&result.MediaView)

	return resp
}

// FromListResult converts query.ListMediaResult to ListMediaResponse.
func FromListResult(result *query.ListMediaResult) *ListMediaResponse {
	resp := &ListMediaResponse{}
	resp.Body.Media = make([]MediaResponse, len(result.Media))
	for i, m := range result.Media {
		resp.Body.Media[i] = ToMediaResponse(m)
	}
	resp.Body.NextCursor = result.NextCursor
	resp.Body.HasMore = result.HasMore

	return resp
}
