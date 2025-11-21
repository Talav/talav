package schema

import (
	"reflect"
)

// Converter converts a string value to a reflect.Value of a specific type.
type Converter func(value string) (reflect.Value, error)

// Option configures a Codec.
type Option func(*Codec)

// Decoder interface for decoding parameter strings to maps.
type Decoder interface {
	Decode(value string, opts Options) (map[string]any, error)
}

// Encoder interface for encoding maps to parameter strings.
type Encoder interface {
	Encode(values map[string]any, opts Options) (string, error)
}

const defaultTagName = "schema"

// Codec handles encoding and decoding between structs and parameter strings.
// It uses injectable decoder and encoder for parameter string handling.
type Codec struct {
	fieldCache  *FieldCache
	converters  *ConverterRegistry
	marshaler   Marshaler
	unmarshaler Unmarshaler
	decoder     Decoder
	encoder     Encoder
}

// NewCodec creates a new Codec with the given options.
func NewCodec(opts ...Option) *Codec {
	fieldCache := NewFieldCache()
	converters := NewConverterRegistry()

	marshaler := NewDefaultMarshaler(defaultTagName, fieldCache)
	unmarshaler := NewDefaultUnmarshaler(defaultTagName, fieldCache, converters)

	c := &Codec{
		marshaler:   marshaler,
		unmarshaler: unmarshaler,
		fieldCache:  fieldCache,
		converters:  converters,
		decoder:     NewDefaultDecoder(),
		encoder:     NewDefaultEncoder(),
	}

	// Register default primitive converters
	c.registerDefaultConverters()

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// WithDecoder sets a custom decoder for the codec.
func WithDecoder(decoder Decoder) Option {
	return func(c *Codec) {
		c.decoder = decoder
	}
}

// WithEncoder sets a custom encoder for the codec.
func WithEncoder(encoder Encoder) Option {
	return func(c *Codec) {
		c.encoder = encoder
	}
}

// WithMarshaler sets a custom marshaler for the codec.
func WithMarshaler(marshaler Marshaler) Option {
	return func(c *Codec) {
		c.marshaler = marshaler
	}
}

// WithUnmarshaler sets a custom unmarshaler for the codec.
func WithUnmarshaler(unmarshaler Unmarshaler) Option {
	return func(c *Codec) {
		c.unmarshaler = unmarshaler
	}
}

// WithConverter registers a converter for the given type as an option.
// This is a convenience wrapper for registerConverter that can be used during Codec creation.
func WithConverter(typ any, conv Converter) Option {
	return func(c *Codec) {
		c.registerConverter(typ, conv)
	}
}

// Decode decodes a parameter string into the provided struct.
// result must be a pointer to the target struct.
func (c *Codec) Decode(value string, opts Options, result any) error {
	// Decode parameter string to map
	data, err := c.decoder.Decode(value, opts)
	if err != nil {
		return err
	}

	// Unmarshal map to struct
	return c.unmarshaler.Unmarshal(data, result)
}

// Encode encodes a struct into a parameter string.
func (c *Codec) Encode(v any, opts Options) (string, error) {
	// Marshal struct to map
	values, err := c.marshaler.Marshal(v)
	if err != nil {
		return "", err
	}

	// Encode map to parameter string
	return c.encoder.Encode(values, opts)
}

// registerConverter registers a custom converter for the given type.
// Converters are field-level type converters used during unmarshaling.
// They convert individual field values from strings to the target type.
func (c *Codec) registerConverter(typ any, conv Converter) {
	c.registerConverterByType(reflect.TypeOf(typ), conv)
}

// registerConverterByType registers a converter for the given reflect.Type.
func (c *Codec) registerConverterByType(typ reflect.Type, conv Converter) {
	c.converters.Register(typ, conv)
}

// registerDefaultConverters registers default converters for primitive types.
func (c *Codec) registerDefaultConverters() {
	c.registerConverterByType(reflect.TypeOf(bool(false)), convertBool)
	c.registerConverterByType(reflect.TypeOf(string("")), convertString)
	c.registerConverterByType(reflect.TypeOf(int(0)), convertInt)
	c.registerConverterByType(reflect.TypeOf(int8(0)), convertInt8)
	c.registerConverterByType(reflect.TypeOf(int16(0)), convertInt16)
	c.registerConverterByType(reflect.TypeOf(int32(0)), convertInt32)
	c.registerConverterByType(reflect.TypeOf(int64(0)), convertInt64)
	c.registerConverterByType(reflect.TypeOf(uint(0)), convertUint)
	c.registerConverterByType(reflect.TypeOf(uint8(0)), convertUint8)
	c.registerConverterByType(reflect.TypeOf(uint16(0)), convertUint16)
	c.registerConverterByType(reflect.TypeOf(uint32(0)), convertUint32)
	c.registerConverterByType(reflect.TypeOf(uint64(0)), convertUint64)
	c.registerConverterByType(reflect.TypeOf(float32(0)), convertFloat32)
	c.registerConverterByType(reflect.TypeOf(float64(0)), convertFloat64)
}
