# mapstructure

A Go library for decoding `map[string]any` values into Go structs with type conversion and struct tag support.

## Why?

When decoding data from JSON, YAML, or other formats, you often get `map[string]any` as an intermediate representation. This library converts those maps into strongly-typed Go structs with:

- Automatic type conversion (string → int, float → bool, etc.)
- Struct tag support for field name mapping
- Nested struct and embedded field handling
- Custom type converters

## Installation

```bash
go get github.com/talav/talav/pkg/component/mapstructure
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/talav/talav/pkg/component/mapstructure"
)

type Person struct {
    Name string `schema:"name"`
    Age  int    `schema:"age"`
}

func main() {
    data := map[string]any{
        "name": "Alice",
        "age":  "30", // string automatically converted to int
    }

    var person Person
    if err := mapstructure.Unmarshal(data, &person); err != nil {
        panic(err)
    }

    fmt.Printf("%+v\n", person) // {Name:Alice Age:30}
}
```

## Struct Tags

By default, the `schema` tag is used for field mapping:

```go
type Config struct {
    ServerHost string `schema:"server_host"`
    ServerPort int    `schema:"port"`
    Debug      bool   `schema:"debug"`
    Ignored    string `schema:"-"` // Skip this field
}
```

| Tag | Behavior |
|-----|----------|
| `schema:"name"` | Use "name" as the map key |
| `schema:"-"` | Skip field entirely |
| No tag | Use Go field name |

## Custom Tag Name

Use a different tag (e.g., `json`):

```go
cache := mapstructure.NewStructMetadataCache(
    mapstructure.NewTagCacheBuilder("json"),
)
converters := mapstructure.NewDefaultConverterRegistry(nil)
unmarshaler := mapstructure.NewUnmarshaler(cache, converters)

type User struct {
    Name string `json:"name"`
}

var user User
unmarshaler.Unmarshal(data, &user)
```

## Type Conversion

Built-in converters handle common type conversions:

| Target Type | Accepted Input Types |
|-------------|---------------------|
| `string` | string, bool, int, uint, float, []byte |
| `bool` | bool, int, uint, float, string ("true", "false", "1", "0") |
| `int`, `int8`...`int64` | int, uint, float, bool, string |
| `uint`, `uint8`...`uint64` | int, uint, float, bool, string |
| `float32`, `float64` | int, uint, float, bool, string |
| `[]byte` | []byte, string, []any, io.Reader |
| `io.ReadCloser` | io.ReadCloser, io.Reader, []byte, string |

## Custom Converters

Register converters for custom types:

```go
import (
    "reflect"
    "time"
)

timeConverter := func(value any) (reflect.Value, error) {
    s, ok := value.(string)
    if !ok {
        return reflect.Value{}, fmt.Errorf("expected string")
    }
    t, err := time.Parse(time.RFC3339, s)
    if err != nil {
        return reflect.Value{}, err
    }
    return reflect.ValueOf(t), nil
}

converters := mapstructure.NewDefaultConverterRegistry(map[reflect.Type]mapstructure.Converter{
    reflect.TypeOf(time.Time{}): timeConverter,
})

cache := mapstructure.NewStructMetadataCache(nil)
unmarshaler := mapstructure.NewUnmarshaler(cache, converters)
```

## Nested Structs

Nested structs are handled automatically:

```go
type Address struct {
    City    string `schema:"city"`
    Country string `schema:"country"`
}

type Person struct {
    Name    string  `schema:"name"`
    Address Address `schema:"address"`
}

data := map[string]any{
    "name": "Alice",
    "address": map[string]any{
        "city":    "New York",
        "country": "USA",
    },
}

var person Person
mapstructure.Unmarshal(data, &person)
```

## Embedded Structs

Embedded structs support both promoted and named field access:

```go
type Timestamps struct {
    CreatedAt string `schema:"created_at"`
    UpdatedAt string `schema:"updated_at"`
}

type User struct {
    Timestamps        // Embedded - fields promoted to parent
    Name       string `schema:"name"`
}

// Both work:
data1 := map[string]any{
    "name":       "Alice",
    "created_at": "2024-01-01", // Promoted field
    "updated_at": "2024-01-02",
}

data2 := map[string]any{
    "name": "Alice",
    "Timestamps": map[string]any{ // Named access
        "created_at": "2024-01-01",
        "updated_at": "2024-01-02",
    },
}
```

## Pointers and Slices

```go
type Config struct {
    Tags    []string `schema:"tags"`
    Count   *int     `schema:"count"`
    Data    []byte   `schema:"data"`
}

data := map[string]any{
    "tags":  []any{"go", "api"},
    "count": 42,
    "data":  []any{72, 101, 108, 108, 111}, // Converts to []byte("Hello")
}
```

## API Reference

### Functions

- `Unmarshal(data map[string]any, result any) error` - Simple API using defaults

### Types

- `Unmarshaler` - Configurable unmarshaler instance
- `ConverterRegistry` - Type converter registry
- `StructMetadataCache` - Cached struct field metadata
- `Converter` - Function type: `func(any) (reflect.Value, error)`

### Constructors

- `NewUnmarshaler(cache, converters)` - Create custom unmarshaler
- `NewStructMetadataCache(builder)` - Create metadata cache (nil = default "schema" tag)
- `NewTagCacheBuilder(tagName)` - Create cache builder for specific tag
- `NewDefaultConverterRegistry(additional)` - Create registry with standard converters
- `NewConverterRegistry(converters)` - Create registry with only specified converters
