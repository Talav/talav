# schema

A Go package for decoding HTTP requests into Go structs with full support for OpenAPI 3.0 parameter serialization. Handles query, path, header, cookie parameters and request bodies (JSON, XML, forms, multipart, files) in a single unified API.

## Features

- Full OpenAPI 3.0 parameter serialization support
- All parameter locations: query, path, header, cookie
- All serialization styles: form, simple, matrix, label, spaceDelimited, pipeDelimited, deepObject
- Explode parameter support
- Request body decoding: JSON, XML, URL-encoded forms, multipart forms, file uploads
- Struct tag-based configuration
- Metadata caching for performance
- Extensible architecture (custom decoders/unmarshalers)

## Installation

```bash
go get github.com/talav/talav/pkg/component/schema
```

## Quick Start

```go
package main

import (
    "net/http"
    "github.com/talav/talav/pkg/component/schema"
)

type CreateUserRequest struct {
    // Query parameter
    Version string `schema:"version,location=query"`
    // Path parameter (from router)
    OrgID string `schema:"org_id,location=path"`
    // Header parameter
    APIKey string `schema:"X-Api-Key,location=header"`
    // Request body
    Body struct {
        Name  string `schema:"name"`
        Email string `schema:"email"`
    } `body:"structured"`
}

func handler(w http.ResponseWriter, r *http.Request) {
    codec := schema.NewCodec()
    
    // Router params come from your router (chi, gorilla, etc.)
    routerParams := map[string]string{"org_id": "123"}
    
    var req CreateUserRequest
    if err := codec.DecodeRequest(r, routerParams, &req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    // Use req.Version, req.OrgID, req.APIKey, req.Body.Name, req.Body.Email
}
```

## Struct Tags

### Parameter Tag (`schema`)

The `schema` tag configures how fields are extracted from HTTP request parameters.

```go
type Request struct {
    // Basic: field name used as parameter name, defaults to query location
    Name string `schema:"name"`
    
    // With location
    ID string `schema:"id,location=path"`
    
    // With style and explode
    IDs []string `schema:"ids,location=query,style=form,explode=true"`
    
    // Required field
    Token string `schema:"token,location=header,required"`
    
    // Full specification
    Filter map[string]string `schema:"filter,location=query,style=deepObject,explode=true,required=true"`
}
```

**Tag Options:**

| Option | Values | Default | Description |
|--------|--------|---------|-------------|
| `location` | `query`, `path`, `header`, `cookie` | `query` | Parameter location |
| `style` | See styles table | Location default | Serialization style |
| `explode` | `true`, `false` | Style default | Explode arrays/objects |
| `required` | `true`, `false` | `false` (`true` for path) | Mark as required |

**Skip a field:**

```go
type Request struct {
    Internal string `schema:"-"` // Skipped during decoding
}
```

### Body Tag (`body`)

The `body` tag configures how the request body is decoded.

```go
type Request struct {
    // JSON or XML body (auto-detected from Content-Type)
    Body UserData `body:"structured"`
    
    // File upload (raw bytes)
    File []byte `body:"file"`
    
    // Multipart form
    Form FormData `body:"multipart"`
}
```

**Body Types:**

| Type | Content-Types | Description |
|------|---------------|-------------|
| `structured` | `application/json`, `application/xml`, `text/xml`, `application/x-www-form-urlencoded` | Structured data |
| `file` | `application/octet-stream` | Raw file bytes |
| `multipart` | `multipart/form-data` | Multipart form with files |

**Body field types:**

```go
// Structured body - decoded into struct
Body struct {
    Name string `schema:"name"`
} `body:"structured"`

// File as bytes
Body []byte `body:"file"`

// File as reader
Body io.ReadCloser `body:"file"`

// Multipart with file upload
Body struct {
    Title    string        `schema:"title"`
    Document io.ReadCloser `schema:"document"` // File field
} `body:"multipart"`

// Multipart with multiple files
Body struct {
    Files []io.ReadCloser `schema:"files"`
} `body:"multipart"`
```

## Parameter Locations

| Location | Description | Default Style |
|----------|-------------|---------------|
| `query` | Query string parameters (`?name=value`) | `form` |
| `path` | Path parameters from router (`/users/{id}`) | `simple` |
| `header` | HTTP headers (`X-Api-Key: value`) | `simple` |
| `cookie` | Cookies (`Cookie: session=abc`) | `form` |

## Serialization Styles

### Query Parameters

| Style | Example | Description |
|-------|---------|-------------|
| `form` (default) | `?ids=1&ids=2` (explode) or `?ids=1,2` | Standard form encoding |
| `spaceDelimited` | `?ids=1%202%203` | Space-separated values |
| `pipeDelimited` | `?ids=1\|2\|3` | Pipe-separated values |
| `deepObject` | `?filter[type]=car&filter[color]=red` | Nested object notation |

### Path Parameters

| Style | Example | Description |
|-------|---------|-------------|
| `simple` (default) | `1,2,3` | Comma-separated |
| `label` | `.1.2.3` | Period-prefixed |
| `matrix` | `;ids=1;ids=2` | Semicolon-prefixed key-value |

### Header & Cookie Parameters

| Location | Style | Description |
|----------|-------|-------------|
| `header` | `simple` only | Comma-separated values |
| `cookie` | `form` only | Standard cookie format |

## Explode Parameter

Controls how arrays and objects are serialized:

**`explode=true` (default for form/deepObject):**
```
Arrays:  ?ids=1&ids=2&ids=3
Objects: ?filter[type]=car&filter[color]=red (deepObject)
         ?type=car&color=red (form)
```

**`explode=false`:**
```
Arrays:  ?ids=1,2,3
Objects: ?filter=type,car,color,red
```

## Usage Examples

### Query Parameters with Different Styles

```go
type SearchRequest struct {
    // Form style (default) - ?tags=go&tags=api or ?tags=go,api
    Tags []string `schema:"tags,location=query,style=form"`
    
    // Space delimited - ?ids=1%202%203
    IDs []int `schema:"ids,location=query,style=spaceDelimited"`
    
    // Pipe delimited - ?colors=red|green|blue
    Colors []string `schema:"colors,location=query,style=pipeDelimited"`
    
    // Deep object - ?filter[status]=active&filter[type]=user
    Filter struct {
        Status string `schema:"status"`
        Type   string `schema:"type"`
    } `schema:"filter,location=query,style=deepObject"`
}
```

### Path Parameters

```go
type GetUserRequest struct {
    // Simple style (default) - /users/123
    UserID string `schema:"user_id,location=path"`
    
    // Label style - /resources/.1.2.3
    Values []string `schema:"values,location=path,style=label,explode=true"`
    
    // Matrix style - /items;id=1;id=2
    IDs []string `schema:"ids,location=path,style=matrix,explode=true"`
}
```

### Headers and Cookies

```go
type AuthenticatedRequest struct {
    // Header parameter
    Authorization string `schema:"Authorization,location=header"`
    RequestID     string `schema:"X-Request-ID,location=header"`
    
    // Cookie parameter
    SessionToken string `schema:"session,location=cookie"`
}
```

### JSON Body

```go
type CreatePostRequest struct {
    // Query param for API version
    Version string `schema:"v,location=query"`
    
    // JSON body
    Body struct {
        Title   string   `schema:"title"`
        Content string   `schema:"content"`
        Tags    []string `schema:"tags"`
    } `body:"structured"`
}

// Handles: POST /posts?v=2
// Body: {"title": "Hello", "content": "World", "tags": ["go", "api"]}
```

### XML Body

```go
type XMLRequest struct {
    Body struct {
        XMLName xml.Name `xml:"user"`
        Name    string   `xml:"name"`
        Email   string   `xml:"email"`
    } `body:"structured"`
}

// Handles: POST /users (Content-Type: application/xml)
// Body: <user><name>John</name><email>john@example.com</email></user>
```

### URL-Encoded Form Body

```go
type FormRequest struct {
    Body struct {
        Username string `schema:"username"`
        Password string `schema:"password"`
    } `body:"structured"`
}

// Handles: POST /login (Content-Type: application/x-www-form-urlencoded)
// Body: username=john&password=secret
```

### File Upload

```go
// As bytes
type FileUploadRequest struct {
    Body []byte `body:"file"`
}

// As reader (for streaming large files)
type StreamingUploadRequest struct {
    Body io.ReadCloser `body:"file"`
}

// With query parameters
type VersionedUploadRequest struct {
    Version string `schema:"version,location=query"`
    Body    []byte `body:"file"`
}
```

### Multipart Form with Files

```go
type DocumentUploadRequest struct {
    Body struct {
        Title       string        `schema:"title"`
        Description string        `schema:"description"`
        Document    io.ReadCloser `schema:"document"`    // Single file
        Attachments []io.ReadCloser `schema:"attachments"` // Multiple files
    } `body:"multipart"`
}

func handler(w http.ResponseWriter, r *http.Request) {
    codec := schema.NewCodec()
    
    var req DocumentUploadRequest
    if err := codec.DecodeRequest(r, nil, &req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    // Don't forget to close file readers
    if req.Body.Document != nil {
        defer req.Body.Document.Close()
    }
    for _, f := range req.Body.Attachments {
        defer f.Close()
    }
    
    // Process files...
}
```

### Mixed Parameters and Body

```go
type CompleteRequest struct {
    // Path parameter
    ResourceID string `schema:"id,location=path"`
    
    // Query parameters
    Format   string `schema:"format,location=query"`
    Page     int    `schema:"page,location=query"`
    PageSize int    `schema:"page_size,location=query"`
    
    // Header parameters
    Authorization string `schema:"Authorization,location=header"`
    RequestID     string `schema:"X-Request-ID,location=header"`
    
    // Cookie parameters
    SessionID string `schema:"session_id,location=cookie"`
    
    // Body
    Body struct {
        Action string            `schema:"action"`
        Data   map[string]string `schema:"data"`
    } `body:"structured"`
}
```

## API Reference

### Codec

The `Codec` is the main entry point for decoding HTTP requests.

#### `NewCodec(opts ...Option) *Codec`

Creates a new Codec with optional configuration.

```go
codec := schema.NewCodec()
```

#### Options

```go
// Custom decoder
codec := schema.NewCodec(schema.WithDecoder(myDecoder))

// Custom unmarshaler
codec := schema.NewCodec(schema.WithUnmarshaler(myUnmarshaler))

// Share field cache between codecs
cache := schema.NewStructMetadataCache(schema.NewStructMetadataParser())
codec1 := schema.NewCodec(schema.WithFieldCache(cache))
codec2 := schema.NewCodec(schema.WithFieldCache(cache))

// Custom parser options (tag names)
codec := schema.NewCodec(schema.WithParserOptions(
    schema.WithSchemaTag("param"),  // Use `param` instead of `schema`
    schema.WithBodyTag("content"),  // Use `content` instead of `body`
))
```

#### `codec.DecodeRequest(request *http.Request, routerParams map[string]string, result any) error`

Decodes an HTTP request into the provided struct.

- `request`: The HTTP request to decode
- `routerParams`: Path parameters from your router (can be `nil`)
- `result`: Pointer to the target struct

```go
var req MyRequest
err := codec.DecodeRequest(r, routerParams, &req)
```

### Decoder Interface

```go
type Decoder interface {
    Decode(request *http.Request, routerParams map[string]string, metadata *StructMetadata) (map[string]any, error)
}
```

Implement this interface for custom decoding logic:

```go
type MyDecoder struct{}

func (d *MyDecoder) Decode(request *http.Request, routerParams map[string]string, metadata *StructMetadata) (map[string]any, error) {
    // Custom decoding logic
    return map[string]any{"key": "value"}, nil
}

codec := schema.NewCodec(schema.WithDecoder(&MyDecoder{}))
```

### Unmarshaler Interface

```go
type Unmarshaler interface {
    Unmarshal(data map[string]any, result any) error
}
```

The default unmarshaler uses `mapstructure`. Implement this interface for custom unmarshaling:

```go
type MyUnmarshaler struct{}

func (u *MyUnmarshaler) Unmarshal(data map[string]any, result any) error {
    // Custom unmarshaling logic
    return nil
}

codec := schema.NewCodec(schema.WithUnmarshaler(&MyUnmarshaler{}))
```

### StructMetadataParser

Parses struct tags to extract field metadata.

```go
parser := schema.NewStructMetadataParser()

// With custom tag names
parser := schema.NewStructMetadataParser(
    schema.WithSchemaTag("param"),
    schema.WithBodyTag("content"),
)
```

### StructMetadataCache

Caches parsed struct metadata for performance.

```go
parser := schema.NewStructMetadataParser()
cache := schema.NewStructMetadataCache(parser)

// Share between codecs
codec1 := schema.NewCodec(schema.WithFieldCache(cache))
codec2 := schema.NewCodec(schema.WithFieldCache(cache))
```

## Style/Location Compatibility

| Location | Allowed Styles | Default Style | Default Explode |
|----------|---------------|---------------|-----------------|
| query | form, spaceDelimited, pipeDelimited, deepObject | form | true (form, deepObject), false (others) |
| path | simple, label, matrix | simple | false |
| header | simple | simple | false |
| cookie | form | form | true |

## Default Behaviors

### Fields Without Tags

Fields without `schema` or `body` tags are treated as query parameters with form style:

```go
type Request struct {
    Name string // Equivalent to: `schema:"Name,location=query,style=form,explode=true"`
}
```

### Path Parameters

Path parameters are automatically marked as required:

```go
type Request struct {
    ID string `schema:"id,location=path"` // Always required
}
```

### Body Content-Type Detection

The body decoder automatically detects content type:

- `application/json` → JSON decoding
- `application/xml`, `text/xml` → XML decoding
- `application/x-www-form-urlencoded` → Form decoding
- `multipart/form-data` → Multipart form decoding (requires `body:"multipart"`)
- `application/octet-stream` → Raw file bytes (requires `body:"file"`)

## Performance

### Metadata Caching

Struct metadata is cached per type. The first decode for a struct type parses and caches the metadata; subsequent decodes reuse the cache.

```go
// Create codec once, reuse for all requests
codec := schema.NewCodec()

// Efficient: metadata cached after first call
for _, r := range requests {
    var req MyRequest
    codec.DecodeRequest(r, nil, &req)
}
```

### Sharing Cache

Share cache between multiple codecs:

```go
parser := schema.NewStructMetadataParser()
cache := schema.NewStructMetadataCache(parser)

// All codecs share the same metadata cache
codec1 := schema.NewCodec(schema.WithFieldCache(cache))
codec2 := schema.NewCodec(schema.WithFieldCache(cache))
```

## Error Handling

Common error scenarios:

```go
var req MyRequest
err := codec.DecodeRequest(r, routerParams, &req)
if err != nil {
    // Possible errors:
    // - Failed to parse query string
    // - Failed to parse multipart form
    // - Failed to unmarshal JSON/XML body
    // - Invalid style for location
    // - Type conversion errors
}
```

## Integration with Routers

### Chi

```go
import "github.com/go-chi/chi/v5"

func handler(w http.ResponseWriter, r *http.Request) {
    codec := schema.NewCodec()
    
    routerParams := map[string]string{
        "id": chi.URLParam(r, "id"),
    }
    
    var req MyRequest
    codec.DecodeRequest(r, routerParams, &req)
}
```

### Gorilla Mux

```go
import "github.com/gorilla/mux"

func handler(w http.ResponseWriter, r *http.Request) {
    codec := schema.NewCodec()
    
    var req MyRequest
    codec.DecodeRequest(r, mux.Vars(r), &req)
}
```

### Standard Library (Go 1.22+)

```go
func handler(w http.ResponseWriter, r *http.Request) {
    codec := schema.NewCodec()
    
    routerParams := map[string]string{
        "id": r.PathValue("id"),
    }
    
    var req MyRequest
    codec.DecodeRequest(r, routerParams, &req)
}
```

## Code Structure

| File | Description |
|------|-------------|
| `codec.go` | High-level Codec API |
| `decoder.go` | HTTP request decoding |
| `decoder_body.go` | Body content decoding (JSON, XML, forms, files) |
| `decoder_styles.go` | OpenAPI style-specific decoding |
| `parser.go` | Struct tag parsing |
| `cache.go` | Metadata caching |
| `schema.go` | Types, constants, validation |
| `types.go` | Reflect kind constants |
| `util.go` | Utility functions |
