package schema

import (
	"net/http"
	"reflect"

	"github.com/talav/talav/pkg/component/mapstructure"
)

// Option configures a Codec.
type Option func(*Codec)

// Decoder interface for decoding HTTP request data to maps.
type Decoder interface {
	// Decode decodes HTTP parameters (query, header, cookie, path, body) to map.
	Decode(request *http.Request, routerParams map[string]string, metadata *StructMetadata) (map[string]any, error)
}

// Unmarshaler interface for unmarshaling maps to Go structs.
type Unmarshaler interface {
	// Unmarshal transforms map[string]any into a Go struct pointed to by result.
	Unmarshal(data map[string]any, result any) error
}

// Codec handles encoding and decoding between structs and parameter strings.
// It uses injectable decoder and unmarshaler for request handling.
type Codec struct {
	fieldCache  *StructMetadataCache
	unmarshaler Unmarshaler
	decoder     Decoder
}

// NewCodec creates a new Codec with the given options.
func NewCodec(opts ...Option) *Codec {
	c := &Codec{}

	// Apply options first to allow cache/parser injection
	for _, opt := range opts {
		opt(c)
	}

	// Set defaults for anything not configured
	if c.fieldCache == nil {
		parser := newStructMetadataParser()
		c.fieldCache = NewStructMetadataCache(parser)
	}

	if c.unmarshaler == nil {
		c.unmarshaler = mapstructure.NewDefaultUnmarshaler(nil)
	}

	if c.decoder == nil {
		c.decoder = newDefaultDecoder(c.fieldCache)
	}

	return c
}

// WithDecoder sets a custom decoder for the codec.
func WithDecoder(decoder Decoder) Option {
	return func(c *Codec) {
		c.decoder = decoder
	}
}

// WithUnmarshaler sets a custom unmarshaler for the codec.
func WithUnmarshaler(unmarshaler Unmarshaler) Option {
	return func(c *Codec) {
		c.unmarshaler = unmarshaler
	}
}

// WithFieldCache sets a custom field cache, allowing cache sharing between codecs.
func WithFieldCache(cache *StructMetadataCache) Option {
	return func(c *Codec) {
		c.fieldCache = cache
	}
}

// WithParserOptions configures the parser with custom options.
// Only used if WithFieldCache is not provided.
func WithParserOptions(parserOpts ...StructMetadataParserOption) Option {
	return func(c *Codec) {
		if c.fieldCache == nil {
			parser := newStructMetadataParser(parserOpts...)
			c.fieldCache = NewStructMetadataCache(parser)
		}
	}
}

// DecodeRequest decodes an HTTP request into the provided struct.
// result must be a pointer to the target struct.
func (c *Codec) DecodeRequest(request *http.Request, routerParams map[string]string, result any) error {
	typ := reflect.TypeOf(result)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	metadata, err := c.fieldCache.getStructMetadata(typ)
	if err != nil {
		return err
	}

	// Decode parameters to map
	paramMap, err := c.decoder.Decode(request, routerParams, metadata)
	if err != nil {
		return err
	}

	// Unmarshal map to struct
	return c.unmarshaler.Unmarshal(paramMap, result)
}
