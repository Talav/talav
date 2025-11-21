package schema

import (
	"reflect"
	"sync"
)

// ConverterRegistry manages type converters.
type ConverterRegistry struct {
	converters sync.Map // map[reflect.Type]Converter
}

// NewConverterRegistry creates a new converter registry.
func NewConverterRegistry() *ConverterRegistry {
	return &ConverterRegistry{}
}

// Register registers a converter for the given type.
func (r *ConverterRegistry) Register(typ reflect.Type, conv Converter) {
	r.converters.Store(typ, conv)
}

// Find finds a converter for the given type.
func (r *ConverterRegistry) Find(typ reflect.Type) (Converter, bool) {
	if converter, ok := r.converters.Load(typ); ok {
		if conv, ok := converter.(Converter); ok {
			return conv, true
		}
	}

	return nil, false
}
