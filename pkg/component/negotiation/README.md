# Negotiation

A Go implementation of content negotiation based on RFC 7231. This package provides tools for implementing content negotiation in your application, supporting media types, languages, charsets, and encodings.

This is a Go port of the [PHP Negotiation library](https://github.com/willdurand/Negotiation) by William Durand, rewritten following Go best practices with no embedded structs, DRY principles, and idiomatic Go code.

## Features

- **Media Type Negotiation** - Negotiate based on `Accept` headers
- **Language Negotiation** - Negotiate based on `Accept-Language` headers
- **Charset Negotiation** - Negotiate based on `Accept-Charset` headers
- **Encoding Negotiation** - Negotiate based on `Accept-Encoding` headers
- **RFC 7231 Compliant** - Follows HTTP content negotiation standards
- **Quality Value Support** - Handles q-values for preference ordering
- **Wildcard Support** - Supports wildcard matching (`*/*`, `text/*`, etc.)
- **Parameter Matching** - Matches media type parameters (e.g., `charset=UTF-8`)
- **Plus-Segment Matching** - Supports media types with plus segments (e.g., `application/vnd.api+json`)

## Installation

```bash
go get github.com/talav/talav/pkg/component/negotiation
```

## Usage

### Media Type Negotiation

```go
package main

import (
    "fmt"
    "github.com/talav/talav/pkg/component/negotiation"
)

func main() {
    negotiator := negotiation.NewMediaNegotiator()
    
    acceptHeader := "text/html, application/xhtml+xml, application/xml;q=0.9, */*;q=0.8"
    priorities := []string{"application/json", "application/xml", "text/html"}
    
    best, err := negotiator.GetBest(acceptHeader, priorities, false)
    if err != nil {
        panic(err)
    }
    
    if best != nil {
        fmt.Printf("Best match: %s\n", best.Type)
        // Output: Best match: text/html
    }
}
```

### Language Negotiation

```go
negotiator := negotiation.NewLanguageNegotiator()

acceptLanguageHeader := "en; q=0.1, fr; q=0.4, fu; q=0.9, de; q=0.2"
priorities := []string{"en", "fu", "de"}

best, err := negotiator.GetBest(acceptLanguageHeader, priorities, false)
if err != nil {
    panic(err)
}

if best != nil {
    fmt.Printf("Best language: %s\n", best.Type)
    // Output: Best language: fu
    fmt.Printf("Quality: %f\n", best.Quality)
    // Output: Quality: 0.900000
}
```

### Charset Negotiation

```go
negotiator := negotiation.NewCharsetNegotiator()

acceptCharsetHeader := "ISO-8859-1, UTF-8; q=0.9"
priorities := []string{"iso-8859-1;q=0.3", "utf-8;q=0.9", "utf-16;q=1.0"}

best, err := negotiator.GetBest(acceptCharsetHeader, priorities, false)
if err != nil {
    panic(err)
}

if best != nil {
    fmt.Printf("Best charset: %s\n", best.Type)
    // Output: Best charset: utf-8
}
```

### Encoding Negotiation

```go
negotiator := negotiation.NewEncodingNegotiator()

acceptEncodingHeader := "gzip;q=1.0, identity; q=0.5, *;q=0"
priorities := []string{"identity", "gzip"}

best, err := negotiator.GetBest(acceptEncodingHeader, priorities, false)
if err != nil {
    panic(err)
}

if best != nil {
    fmt.Printf("Best encoding: %s\n", best.Type)
    // Output: Best encoding: identity
}
```

### Getting Ordered Elements

You can also get all accept header elements ordered by quality:

```go
negotiator := negotiation.NewMediaNegotiator()

elements, err := negotiator.GetOrderedElements("text/html;q=0.3, text/html;q=0.7")
if err != nil {
    panic(err)
}

for _, elem := range elements {
    fmt.Printf("%s (q=%f)\n", elem.Value, elem.Quality)
}
// Output:
// text/html;q=0.7 (q=0.700000)
// text/html;q=0.3 (q=0.300000)
```

## Error Handling

The package defines several error types:

- `ErrInvalidArgument` - Invalid argument provided
- `ErrInvalidHeader` - Header cannot be parsed
- `ErrInvalidMediaType` - Invalid media type format
- `ErrInvalidLanguage` - Invalid language tag format

## Testing

The package includes comprehensive tests covering all functionality from the original PHP library:

```bash
go test ./pkg/component/negotiation/...
```


## Credits

- Original PHP library: [willdurand/Negotiation](https://github.com/willdurand/Negotiation)

