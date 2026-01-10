package dto

// CreateMediaRequest represents the HTTP request to create media.
type CreateMediaRequest struct {
	Body struct {
		File        []byte `json:"file" openapi:"title=File,description=File content to upload,format=binary"`
		Preset      string `json:"preset" openapi:"title=Preset,description=Processing preset to apply"`
		Provider    string `json:"provider" openapi:"title=Provider,description=Storage provider to use"`
		Description string `json:"description" openapi:"title=Description,description=Optional file description"`
	} `body:"multipart,required"`
}

// GetMediaRequest represents the HTTP request to get media by ID.
type GetMediaRequest struct {
	ID string `schema:"id,location=path,required=true" openapi:"title=Media ID,description=Unique identifier of the media"`
}

// ListMediaRequest represents the HTTP request to list media with pagination and filtering.
type ListMediaRequest struct {
	Cursor string `schema:"cursor,location=query" openapi:"title=Cursor,description=Cursor for pagination"`
	Limit  int    `schema:"limit,location=query" openapi:"title=Limit,description=Number of media per page (default: 10)"`
	Preset string `schema:"preset,location=query" openapi:"title=Preset,description=Filter by preset"`
	Type   string `schema:"type,location=query" openapi:"title=Type,description=Filter by type (image/video/file)"`
}

// UpdateMediaRequest represents the HTTP request to update media.
type UpdateMediaRequest struct {
	ID   string `schema:"id,location=path,required=true" openapi:"title=Media ID,description=Unique identifier of the media"`
	Body struct {
		Description string `json:"description" openapi:"title=Description,description=New description for the media"`
	} `body:"structured,required"`
}

// DeleteMediaRequest represents the HTTP request to delete media.
type DeleteMediaRequest struct {
	ID string `schema:"id,location=path,required=true" openapi:"title=Media ID,description=Unique identifier of the media"`
}
