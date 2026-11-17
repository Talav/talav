package media

import (
	"github.com/talav/talav/pkg/module/media/handler"
	"github.com/talav/zorya"
)

// tagMedia groups the media routes in the OpenAPI document.
const tagMedia = "Media"

// RegisterRoutes registers all media-related routes.
// Called automatically by the FX module during application startup.
func RegisterRoutes(
	api zorya.API,
	createMedia *handler.CreateMediaHandler,
	getMedia *handler.GetMediaHandler,
	listMedia *handler.ListMediaHandler,
	updateMedia *handler.UpdateMediaHandler,
	deleteMedia *handler.DeleteMediaHandler,
) {
	// Media group for media management endpoints
	mediaGroup := zorya.NewGroup(api, "/media")

	// POST /media - Create media (multipart upload)
	zorya.Post(mediaGroup, "", createMedia.Handle,
		func(r *zorya.BaseRoute) {
			r.Operation = &zorya.Operation{
				Summary:     "Create Media",
				Description: "Upload a new media file (image, video, or document)",
				Tags:        []string{tagMedia},
				OperationID: "createMedia",
			}
		},
	)

	// GET /media - List media
	zorya.Get(mediaGroup, "", listMedia.Handle,
		func(r *zorya.BaseRoute) {
			r.Operation = &zorya.Operation{
				Summary:     "List Media",
				Description: "Retrieve a paginated list of media entries with optional filtering",
				Tags:        []string{tagMedia},
				OperationID: "listMedia",
			}
		},
	)

	// GET /media/{id} - Get media by ID
	zorya.Get(mediaGroup, "/{id}", getMedia.Handle,
		func(r *zorya.BaseRoute) {
			r.Operation = &zorya.Operation{
				Summary:     "Get Media by ID",
				Description: "Retrieve a single media entry by its unique ID",
				Tags:        []string{tagMedia},
				OperationID: "getMedia",
			}
		},
	)

	// PATCH /media/{id} - Update media
	zorya.Patch(mediaGroup, "/{id}", updateMedia.Handle,
		func(r *zorya.BaseRoute) {
			r.Operation = &zorya.Operation{
				Summary:     "Update Media",
				Description: "Update media description (only mutable field)",
				Tags:        []string{tagMedia},
				OperationID: "updateMedia",
			}
		},
	)

	// DELETE /media/{id} - Delete media
	zorya.Delete(mediaGroup, "/{id}", deleteMedia.Handle,
		func(r *zorya.BaseRoute) {
			r.Operation = &zorya.Operation{
				Summary:     "Delete Media",
				Description: "Delete a media entry and its associated file",
				Tags:        []string{tagMedia},
				OperationID: "deleteMedia",
			}
		},
	)
}
