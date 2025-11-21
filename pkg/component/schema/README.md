# schema

A Go package for OpenAPI 3.0 parameter serialization and deserialization. Convert between parameter strings (query, path, header, cookie) and Go structs with full support for all OpenAPI 3.0 serialization styles.

## Features

- ✅ Full OpenAPI 3.0 parameter serialization support
- ✅ All parameter locations: query, path, header, cookie
- ✅ All serialization styles: form, simple, matrix, label, spaceDelimited, pipeDelimited, deepObject
- ✅ Explode parameter support
- ✅ Struct ↔ Parameter string conversion
- ✅ Custom type converters
- ✅ Field caching for performance
- ✅ Extensible architecture (custom decoders/encoders/marshalers)

## Installation

```bash
go get github.com/talav/talav/pkg/component/schema
```

## Quick Start

### Decode Query Parameters to Struct

```go
package main

import (
    "fmt"
    "github.com/talav/talav/pkg/component/schema"
)

type User struct {
    Name string `schema:"name"`
    Age  int    `schema:"age"`
}

func main() {
    codec := schema.NewCodec()
    opts, _ := schema.NewOptions(schema.LocationQuery, schema.StyleForm)
    
    var user User
    err := codec.Decode("name=John&age=30", opts, &user)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("User: %+v\n", user)
    // Output: User: {Name:John Age:30}
}
```

### Encode Struct to Query Parameters

```go
user := User{Name: "John", Age: 30}
encoded, err := codec.Encode(user, opts)
if err != nil {
    panic(err)
}

fmt.Println(encoded)
// Output: ?name=John&age=30
```

## Core Concepts

### Data Flow

The package uses a two-stage conversion process:

**Decoding (Parameter String → Struct):**
```
Parameter String (e.g., "name=John&age=30")
    ↓ [Decoder]
Map[string]any (e.g., {"name": "John", "age": "30"})
    ↓ [Unmarshaler]
Go Struct (e.g., User{Name: "John", Age: 30})
```

**Encoding (Struct → Parameter String):**
```
Go Struct
    ↓ [Marshaler]
Map[string]any
    ↓ [Encoder]
Parameter String
```

### Parameter Locations

- **`LocationQuery`** - Query string parameters (e.g., `?name=John`)
- **`LocationPath`** - Path parameters (e.g., `/users/{id}`)
- **`LocationHeader`** - HTTP header parameters
- **`LocationCookie`** - Cookie parameters

### Serialization Styles

| Style | Location | Description | Example |
|-------|----------|-------------|---------|
| `StyleForm` | query, cookie | Form-style serialization (default for query/cookie) | `name=John&age=30` |
| `StyleSimple` | path, header | Simple style without formatting (default for path/header) | `John,30` |
| `StyleMatrix` | path | Matrix style with semicolon prefix | `;name=John;age=30` |
| `StyleLabel` | path | Label style with period prefix | `.John.30` |
| `StyleSpaceDelimited` | query | Space-delimited arrays | `ids=1 2 3` |
| `StylePipeDelimited` | query | Pipe-delimited arrays | `ids=1\|2\|3` |
| `StyleDeepObject` | query | Deep object style | `filter[type]=car&filter[color]=red` |

### Explode Parameter

The `explode` parameter controls how arrays and objects are serialized:

- **`explode=true`**: Arrays/objects are expanded into multiple parameters
  - Array: `ids=1&ids=2&ids=3`
  - Object: `filter.type=car&filter.color=red`
  
- **`explode=false`**: Arrays/objects are serialized as single values
  - Array: `ids=1,2,3`
  - Object: `filter=type,car,color,red`

## Usage Examples

### Basic Usage - Query Parameters

```go
type User struct {
    Name string `schema:"name"`
    Age  int    `schema:"age"`
}

codec := schema.NewCodec()
opts, _ := schema.NewOptions(schema.LocationQuery, schema.StyleForm)

// Decode
var user User
err := codec.Decode("name=John&age=30", opts, &user)

// Encode
encoded, err := codec.Encode(user, opts)
// Result: "?name=John&age=30"
```

### Path Parameters

#### Matrix Style

```go
opts, _ := schema.NewOptions(schema.LocationPath, schema.StyleMatrix)

// Encode
values := map[string]any{"ids": []any{"1", "2", "3"}}
encoder := schema.NewDefaultEncoder()
encoded, _ := encoder.Encode(values, opts)
// Result: ";ids=1;ids=2;ids=3"

// Decode
decoder := schema.NewDefaultDecoder()
decoded, _ := decoder.Decode(";ids=1;ids=2;ids=3", opts)
```

#### Label Style

```go
opts, _ := schema.NewOptions(schema.LocationPath, schema.StyleLabel)

// Encode
values := map[string]any{"ids": []any{"1", "2", "3"}}
encoded, _ := encoder.Encode(values, opts)
// Result: ".1.2.3"
```

#### Simple Style

```go
opts, _ := schema.NewOptions(schema.LocationPath, schema.StyleSimple)

// Encode
values := map[string]any{"ids": []any{"1", "2", "3"}}
encoded, _ := encoder.Encode(values, opts)
// Result: "1,2,3"
```

### Header Parameters

```go
opts, _ := schema.NewOptions(schema.LocationHeader, schema.StyleSimple)

// Encode
values := map[string]any{"x-api-key": "secret123"}
encoded, _ := encoder.Encode(values, opts)
// Result: "secret123"

// Decode
decoded, _ := decoder.Decode("secret123", opts)
```

### Cookie Parameters

```go
opts, _ := schema.NewOptions(schema.LocationCookie, schema.StyleForm)

// Encode
values := map[string]any{"session": "abc123", "user": "john"}
encoded, _ := encoder.Encode(values, opts)
// Result: "?session=abc123&user=john"
```

### Arrays and Collections

#### Exploded Arrays

```go
opts, _ := schema.NewOptions(schema.LocationQuery, schema.StyleForm, true)

values := map[string]any{"ids": []any{"1", "2", "3"}}
encoded, _ := encoder.Encode(values, opts)
// Result: "?ids=1&ids=2&ids=3"
```

#### Non-Exploded Arrays

```go
opts, _ := schema.NewOptions(schema.LocationQuery, schema.StyleForm, false)

values := map[string]any{"ids": []any{"1", "2", "3"}}
encoded, _ := encoder.Encode(values, opts)
// Result: "?ids=1,2,3"
```

#### Space-Delimited Arrays

```go
opts, _ := schema.NewOptions(schema.LocationQuery, schema.StyleSpaceDelimited)

values := map[string]any{"ids": []any{"1", "2", "3"}}
encoded, _ := encoder.Encode(values, opts)
// Result: "?ids=1 2 3"
```

#### Pipe-Delimited Arrays

```go
opts, _ := schema.NewOptions(schema.LocationQuery, schema.StylePipeDelimited)

values := map[string]any{"ids": []any{"1", "2", "3"}}
encoded, _ := encoder.Encode(values, opts)
// Result: "?ids=1|2|3"
```

### Nested Objects

#### Deep Object Style

```go
opts, _ := schema.NewOptions(schema.LocationQuery, schema.StyleDeepObject)

values := map[string]any{
    "filter": map[string]any{
        "type":  "car",
        "color": "red",
    },
}
encoded, _ := encoder.Encode(values, opts)
// Result: "?filter[type]=car&filter[color]=red"
```

#### Dotted Notation (Exploded Objects)

```go
opts, _ := schema.NewOptions(schema.LocationQuery, schema.StyleForm, true)

values := map[string]any{
    "filter": map[string]any{
        "type":  "car",
        "color": "red",
    },
}
encoded, _ := encoder.Encode(values, opts)
// Result: "?filter.type=car&filter.color=red"
```

#### Nested Structs

```go
type Filter struct {
    Type  string `schema:"type"`
    Color string `schema:"color"`
}

type Query struct {
    Filter Filter `schema:"filter"`
    Limit  int    `schema:"limit"`
}

codec := schema.NewCodec()
opts, _ := schema.NewOptions(schema.LocationQuery, schema.StyleDeepObject)

query := Query{
    Filter: Filter{Type: "car", Color: "red"},
    Limit:  10,
}

encoded, _ := codec.Encode(query, opts)
// Result: "?filter[type]=car&filter[color]=red&limit=10"
```

### Custom Converters

Register custom type converters for user-defined types:

```go
import (
    "reflect"
    "strconv"
)

type UserID int

codec := schema.NewCodec(
    schema.WithConverter(UserID(0), func(s string) (reflect.Value, error) {
        id, err := strconv.Atoi(s)
        if err != nil {
            return reflect.Value{}, err
        }
        return reflect.ValueOf(UserID(id)), nil
    }),
)

type User struct {
    ID   UserID `schema:"id"`
    Name string `schema:"name"`
}

opts, _ := schema.NewOptions(schema.LocationQuery, schema.StyleForm)
var user User
err := codec.Decode("id=42&name=John", opts, &user)
// user.ID will be UserID(42)
```

## API Reference

### Codec (Main Entry Point)

The `Codec` is the high-level API that handles the complete conversion between structs and parameter strings.

#### `NewCodec(opts ...Option) *Codec`

Creates a new Codec instance with optional configuration.

```go
codec := schema.NewCodec()
```

#### Options

- **`WithDecoder(decoder Decoder) Option`** - Set a custom decoder
- **`WithEncoder(encoder Encoder) Option`** - Set a custom encoder
- **`WithMarshaler(marshaler Marshaler) Option`** - Set a custom marshaler
- **`WithUnmarshaler(unmarshaler Unmarshaler) Option`** - Set a custom unmarshaler
- **`WithConverter(typ any, conv Converter) Option`** - Register a custom type converter

```go
codec := schema.NewCodec(
    schema.WithConverter(MyType(0), myConverter),
    schema.WithDecoder(myDecoder),
)
```

#### `codec.Decode(value string, opts Options, result any) error`

Decodes a parameter string into the provided struct. `result` must be a pointer to the target struct.

```go
var user User
err := codec.Decode("name=John&age=30", opts, &user)
```

#### `codec.Encode(v any, opts Options) (string, error)`

Encodes a struct into a parameter string.

```go
encoded, err := codec.Encode(user, opts)
```

### Options

Options configure how parameters are encoded/decoded.

#### `NewOptions(location ParameterLocation, style Style, explode ...bool) (Options, error)`

Creates validated options for the given location and style. Returns an error if the style is not allowed for the location.
The `explode` parameter is variadic: omit it to use the default, or provide one `bool` value to set it explicitly.

```go
opts, err := schema.NewOptions(schema.LocationQuery, schema.StyleForm) // Uses default explode
opts, err := schema.NewOptions(schema.LocationQuery, schema.StyleForm, true) // Explicit explode=true
opts, err := schema.NewOptions(schema.LocationQuery, schema.StyleForm, false) // Explicit explode=false
```

#### `DefaultOptions(location ParameterLocation) Options`

Returns default options for the given parameter location. Uses the default style for the location.

```go
opts := schema.DefaultOptions(schema.LocationQuery)
```

#### `opts.Location() ParameterLocation`

Returns the parameter location.

#### `opts.Style() Style`

Returns the validated style.

#### `DefaultStyle(location ParameterLocation) Style`

Returns the default style for the given location:
- `LocationQuery` → `StyleForm`
- `LocationPath` → `StyleSimple`
- `LocationHeader` → `StyleSimple`
- `LocationCookie` → `StyleForm`

#### `AllowedStyles(location ParameterLocation) []Style`

Returns all allowed styles for the given location.

#### `ValidateStyle(location ParameterLocation, style Style) (Style, error)`

Validates that the given style is allowed for the specified location. If style is empty, returns the default style.

### Decoder (Low-level)

The `Decoder` interface handles conversion from parameter strings to maps.

#### `NewDefaultDecoder() Decoder`

Creates a new default decoder.

```go
decoder := schema.NewDefaultDecoder()
```

#### `decoder.Decode(value string, opts Options) (map[string]any, error)`

Decodes a parameter string into a map.

```go
result, err := decoder.Decode("name=John&age=30", opts)
// result: map[string]any{"name": "John", "age": "30"}
```

### Encoder (Low-level)

The `Encoder` interface handles conversion from maps to parameter strings.

#### `NewDefaultEncoder() Encoder`

Creates a new default encoder.

```go
encoder := schema.NewDefaultEncoder()
```

#### `encoder.Encode(values map[string]any, opts Options) (string, error)`

Encodes a map into a parameter string.

```go
values := map[string]any{"name": "John", "age": 30}
encoded, err := encoder.Encode(values, opts)
```

### Marshaler (Low-level)

The `Marshaler` interface handles conversion from Go structs to maps.

#### `NewDefaultMarshaler(tagName string, fieldCache *FieldCache) Marshaler`

Creates a new default marshaler.

```go
fieldCache := schema.NewFieldCache()
marshaler := schema.NewDefaultMarshaler("schema", fieldCache)
```

#### `marshaler.Marshal(v any) (map[string]any, error)`

Marshals a Go struct into a map.

```go
user := User{Name: "John", Age: 30}
result, err := marshaler.Marshal(user)
// result: map[string]any{"name": "John", "age": 30}
```

### Unmarshaler (Low-level)

The `Unmarshaler` interface handles conversion from maps to Go structs.

#### `NewDefaultUnmarshaler(tagName string, fieldCache *FieldCache, converters *ConverterRegistry) Unmarshaler`

Creates a new default unmarshaler.

```go
fieldCache := schema.NewFieldCache()
converters := schema.NewConverterRegistry()
unmarshaler := schema.NewDefaultUnmarshaler("schema", fieldCache, converters)
```

#### `unmarshaler.Unmarshal(data map[string]any, result any) error`

Unmarshals a map into a Go struct. `result` must be a pointer.

```go
data := map[string]any{"name": "John", "age": 30}
var user User
err := unmarshaler.Unmarshal(data, &user)
```

### Converter

A `Converter` is a function that converts a string value to a `reflect.Value` of a specific type.

```go
type Converter func(value string) (reflect.Value, error)
```

Converters are used during unmarshaling to convert individual field values from strings to the target type.

### FieldCache

`FieldCache` provides caching for struct field metadata to improve performance.

#### `NewFieldCache() *FieldCache`

Creates a new field cache.

```go
fieldCache := schema.NewFieldCache()
```

The cache is automatically used by the default marshaler and unmarshaler.

## Advanced Topics

### Custom Decoders/Encoders

You can implement custom decoders and encoders for specialized serialization formats.

#### Custom Decoder

```go
type MyDecoder struct{}

func (d *MyDecoder) Decode(value string, opts schema.Options) (map[string]any, error) {
    // Custom decoding logic
    return map[string]any{"key": "value"}, nil
}

codec := schema.NewCodec(schema.WithDecoder(&MyDecoder{}))
```

#### Custom Encoder

```go
type MyEncoder struct{}

func (e *MyEncoder) Encode(values map[string]any, opts schema.Options) (string, error) {
    // Custom encoding logic
    return "custom format", nil
}

codec := schema.NewCodec(schema.WithEncoder(&MyEncoder{}))
```

### Custom Marshalers/Unmarshalers

Implement custom marshalers/unmarshalers for specialized struct mapping logic.

#### Custom Marshaler

```go
type MyMarshaler struct{}

func (m *MyMarshaler) Marshal(v any) (map[string]any, error) {
    // Custom marshaling logic
    return map[string]any{"key": "value"}, nil
}

codec := schema.NewCodec(schema.WithMarshaler(&MyMarshaler{}))
```

#### Custom Unmarshaler

```go
type MyUnmarshaler struct{}

func (u *MyUnmarshaler) Unmarshal(data map[string]any, result any) error {
    // Custom unmarshaling logic
    return nil
}

codec := schema.NewCodec(schema.WithUnmarshaler(&MyUnmarshaler{}))
```

### Type Conversion

#### Default Converters

The package includes default converters for all Go primitive types:
- `bool`, `string`
- `int`, `int8`, `int16`, `int32`, `int64`
- `uint`, `uint8`, `uint16`, `uint32`, `uint64`
- `float32`, `float64`

#### Custom Converters

Register custom converters for user-defined types:

```go
type Status string

const (
    StatusActive   Status = "active"
    StatusInactive Status = "inactive"
)

codec := schema.NewCodec(
    schema.WithConverter(Status(""), func(s string) (reflect.Value, error) {
        status := Status(s)
        if status != StatusActive && status != StatusInactive {
            return reflect.Value{}, fmt.Errorf("invalid status: %s", s)
        }
        return reflect.ValueOf(status), nil
    }),
)
```

### Struct Tags

Use struct tags to control field mapping:

#### Field Name Mapping

```go
type User struct {
    Name string `schema:"user_name"`  // Maps to "user_name" in parameter string
    Age  int    `schema:"age"`
}
```

#### Skip Field

```go
type User struct {
    Name     string `schema:"name"`
    Password string `schema:"-"`  // Field is skipped during encoding/decoding
}
```

#### Default Behavior

If no tag is provided, the field name (lowercased) is used:

```go
type User struct {
    Name string  // Maps to "name" in parameter string
    Age  int     // Maps to "age" in parameter string
}
```

### Performance Considerations

#### Field Caching

The package uses field caching to improve performance. Field metadata is cached per struct type, so repeated marshaling/unmarshaling of the same struct type is faster.

#### Reusing Codec Instances

Create a single `Codec` instance and reuse it:

```go
// Good: Reuse codec instance
codec := schema.NewCodec()
for _, user := range users {
    encoded, _ := codec.Encode(user, opts)
}

// Bad: Creating new codec for each operation
for _, user := range users {
    codec := schema.NewCodec()  // Inefficient
    encoded, _ := codec.Encode(user, opts)
}
```

#### Low-level APIs vs Codec

Use low-level APIs (`Decoder`, `Encoder`, `Marshaler`, `Unmarshaler`) when you only need part of the conversion pipeline:

```go
// If you only need map ↔ parameter string conversion
decoder := schema.NewDefaultDecoder()
encoder := schema.NewDefaultEncoder()

// If you need full struct ↔ parameter string conversion
codec := schema.NewCodec()
```

## Error Handling

The package defines several error types for consistent error handling:

- **`ErrUnsupportedLocation`** - Unsupported parameter location
- **`ErrUnsupportedStyle`** - Unsupported style for a given location
- **`ErrInvalidStyle`** - Invalid style value
- **`ErrInvalidFormat`** - Invalid format in input data
- **`ErrUnsupportedType`** - Unsupported type for marshaling
- **`ErrInvalidElement`** - Invalid element value
- **`ErrUnsupportedSliceElementType`** - Unsupported slice element type
- **`ErrInvalidOptions`** - Invalid options configuration

Errors are wrapped using `%w` for error chain inspection:

```go
var user User
err := codec.Decode("invalid", opts, &user)
if errors.Is(err, schema.ErrInvalidFormat) {
    // Handle invalid format error
}
```

### Common Error Scenarios

#### Invalid Style for Location

```go
opts, err := schema.NewOptions(schema.LocationQuery, schema.StyleMatrix)
// Returns error: style "matrix" is not allowed for parameter location "query"
```

#### Invalid Parameter Format

```go
decoder := schema.NewDefaultDecoder()
opts, _ := schema.NewOptions(schema.LocationPath, schema.StyleMatrix)
_, err := decoder.Decode("invalid format", opts)
// May return ErrInvalidFormat if format is invalid
```

## OpenAPI 3.0 Compatibility

### Supported Features

The package fully supports OpenAPI 3.0 parameter serialization:

- ✅ All parameter locations (query, path, header, cookie)
- ✅ All serialization styles
- ✅ Explode parameter
- ✅ Arrays and objects
- ✅ Nested structures

### Style/Location Compatibility

| Location | Allowed Styles | Default Style |
|----------|---------------|---------------|
| query | form, spaceDelimited, pipeDelimited, deepObject | form |
| path | simple, label, matrix | simple |
| header | simple | simple |
| cookie | form | form |

### Default Behaviors

- **Default explode values:**
  - `StyleForm`: `explode=true`
  - All other styles: `explode=false`

- **Array serialization:**
  - Exploded: Multiple parameters (e.g., `ids=1&ids=2`)
  - Non-exploded: Single parameter with delimiter (e.g., `ids=1,2`)

- **Object serialization:**
  - Exploded: Dotted notation (e.g., `filter.type=car`)
  - Non-exploded: Comma-separated (e.g., `filter=type,car`)

## Examples by Use Case

### HTTP Request Handling

#### Parsing Query Parameters

```go
import (
    "net/http"
    "github.com/talav/talav/pkg/component/schema"
)

func handleRequest(w http.ResponseWriter, r *http.Request) {
    type QueryParams struct {
        Page  int    `schema:"page"`
        Limit int    `schema:"limit"`
        Sort  string `schema:"sort"`
    }
    
    codec := schema.NewCodec()
    opts, _ := schema.NewOptions(schema.LocationQuery, schema.StyleForm)
    
    var params QueryParams
    err := codec.Decode(r.URL.RawQuery, opts, &params)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    // Use params.Page, params.Limit, params.Sort
}
```

#### Parsing Path Parameters

```go
func handlePathParams(path string) {
    type PathParams struct {
        ID   int    `schema:"id"`
        Name string `schema:"name"`
    }
    
    codec := schema.NewCodec()
    opts, _ := schema.NewOptions(schema.LocationPath, schema.StyleSimple)
    
    var params PathParams
    err := codec.Decode(path, opts, &params)
    // Handle params
}
```

#### Parsing Headers

```go
func handleHeaders(headerValue string) {
    type HeaderParams struct {
        APIKey string `schema:"x-api-key"`
    }
    
    codec := schema.NewCodec()
    opts, _ := schema.NewOptions(schema.LocationHeader, schema.StyleSimple)
    
    var params HeaderParams
    err := codec.Decode(headerValue, opts, &params)
    // Handle params
}
```

### API Client Generation

#### Building Query Strings

```go
type SearchParams struct {
    Query  string   `schema:"q"`
    Tags   []string `schema:"tags"`
    Limit  int      `schema:"limit"`
    Offset int      `schema:"offset"`
}

func buildSearchURL(baseURL string, params SearchParams) string {
    codec := schema.NewCodec()
    opts, _ := schema.NewOptions(schema.LocationQuery, schema.StyleForm, true)
    
    encoded, _ := codec.Encode(params, opts)
    return baseURL + encoded
}

// Usage
params := SearchParams{
    Query:  "golang",
    Tags:   []string{"api", "rest"},
    Limit:  10,
    Offset: 0,
}
url := buildSearchURL("https://api.example.com/search", params)
// Result: "https://api.example.com/search?q=golang&tags=api&tags=rest&limit=10&offset=0"
```

#### Building Path Parameters

```go
type ResourceParams struct {
    ID   int    `schema:"id"`
    Type string `schema:"type"`
}

func buildResourcePath(params ResourceParams) string {
    codec := schema.NewCodec()
    opts, _ := schema.NewOptions(schema.LocationPath, schema.StyleMatrix)
    
    encoded, _ := codec.Encode(params, opts)
    return "/resources" + encoded
}

// Usage
params := ResourceParams{ID: 123, Type: "user"}
path := buildResourcePath(params)
// Result: "/resources;id=123;type=user"
```

### OpenAPI Code Generation

When working with generated OpenAPI clients, map OpenAPI parameter definitions to schema options:

```go
// OpenAPI parameter definition:
// {
//   "name": "filter",
//   "in": "query",
//   "style": "deepObject",
//   "explode": true
// }

opts, _ := schema.NewOptions(
    schema.LocationQuery,
    schema.StyleDeepObject,
    true,
)

type Filter struct {
    Type  string `schema:"type"`
    Color string `schema:"color"`
}

type Query struct {
    Filter Filter `schema:"filter"`
}

codec := schema.NewCodec()
query := Query{Filter: Filter{Type: "car", Color: "red"}}
encoded, _ := codec.Encode(query, opts)
// Result: "?filter[type]=car&filter[color]=red"
```

## Testing

### Testing Encoding/Decoding

Test encoding and decoding operations:

```go
func TestEncodeDecode(t *testing.T) {
    type User struct {
        Name string `schema:"name"`
        Age  int    `schema:"age"`
    }
    
    codec := schema.NewCodec()
    opts, _ := schema.NewOptions(schema.LocationQuery, schema.StyleForm)
    
    original := User{Name: "John", Age: 30}
    
    // Encode
    encoded, err := codec.Encode(original, opts)
    require.NoError(t, err)
    
    // Decode
    var decoded User
    err = codec.Decode(encoded, opts, &decoded)
    require.NoError(t, err)
    
    assert.Equal(t, original, decoded)
}
```

### Round-Trip Testing

Test that encoding and decoding are inverse operations:

```go
func TestRoundTrip(t *testing.T) {
    decoder := schema.NewDefaultDecoder()
    encoder := schema.NewDefaultEncoder()
    opts, _ := schema.NewOptions(schema.LocationQuery, schema.StyleForm)
    
    original := map[string]any{
        "ids":  []any{"1", "2", "3"},
        "name": "John",
    }
    
    // Encode then decode
    encoded, _ := encoder.Encode(original, opts)
    decoded, _ := decoder.Decode(encoded, opts)
    
    assert.Equal(t, original["name"], decoded["name"])
    assert.Equal(t, original["ids"], decoded["ids"])
}
```

### Mocking Decoders/Encoders

For testing, you can create mock decoders/encoders:

```go
type mockDecoder struct {
    result map[string]any
    err    error
}

func (m *mockDecoder) Decode(value string, opts schema.Options) (map[string]any, error) {
    return m.result, m.err
}

func TestWithMockDecoder(t *testing.T) {
    mock := &mockDecoder{
        result: map[string]any{"test": "value"},
    }
    
    codec := schema.NewCodec(schema.WithDecoder(mock))
    // Test with mocked decoder
}
```

## Contributing

### Code Structure

The package is organized into several components:

- **`codec.go`** - High-level Codec API
- **`decoder.go`** - Parameter string → map conversion
- **`encoder.go`** - Map → parameter string conversion
- **`marshaler.go`** - Struct → map conversion
- **`unmarshaler.go`** - Map → struct conversion
- **`options.go`** - Options and validation
- **`schema.go`** - Location and style definitions
- **`converter.go`** - Type conversion utilities
- **`converter_registry.go`** - Converter registration
- **`cache.go`** - Field metadata caching
- **`errors.go`** - Error type definitions

### Adding New Styles/Locations

To add support for new styles or locations:

1. Add the style/location constant to `schema.go`
2. Update `AllowedStyles()` to include the new style for the location
3. Implement encoding logic in `encoder.go`
4. Implement decoding logic in `decoder.go`
5. Add tests in `*_test.go` files

### Testing Requirements

- All new features must include tests
- Tests should cover both encoding and decoding
- Include round-trip tests where applicable
- Test error cases and edge cases
- Follow existing test patterns
