package schema

import (
	"fmt"
	"reflect"
	"sync"
)

// StructMetadataCache provides caching for struct field metadata.
type StructMetadataCache struct {
	cache  sync.Map // map[reflect.Type][]cachedField
	parser *structMetadataParser
}

// NewSchemaCache creates a new schema cache.
func NewStructMetadataCache(parser *structMetadataParser) *StructMetadataCache {
	return &StructMetadataCache{
		parser: parser,
	}
}

// GetCachedFields retrieves or builds cached struct field metadata.
// Uses TagConfig to determine which tag names to parse.
func (fc *StructMetadataCache) getStructMetadata(typ reflect.Type) (*StructMetadata, error) {
	// Check cache first
	if cached, ok := fc.cache.Load(typ); ok {
		if fields, ok := cached.(*StructMetadata); ok {
			return fields, nil
		}
	}

	// Build cache
	fields, err := fc.parser.BuildStructMetadata(typ)
	if err != nil {
		return nil, fmt.Errorf("failed to build struct metadata: %w", err)
	}

	// Store in cache (or get existing if another goroutine stored it first)
	actual, _ := fc.cache.LoadOrStore(typ, fields)
	fields, _ = actual.(*StructMetadata)

	return fields, nil
}
