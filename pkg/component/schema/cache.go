package schema

import (
	"reflect"
	"strings"
	"sync"
)

// cachedField represents a cached struct field metadata.
type cachedField struct {
	name     string
	tagName  string
	mapKey   string // Precomputed: tagName if present, otherwise exact field name
	index    int
	kind     reflect.Kind
	embedded bool
}

// FieldCache provides caching for struct field metadata.
type FieldCache struct {
	cache sync.Map // map[reflect.Type][]cachedField
}

// NewFieldCache creates a new field cache.
func NewFieldCache() *FieldCache {
	return &FieldCache{}
}

// GetCachedFields retrieves or builds cached struct field metadata.
func (fc *FieldCache) GetCachedFields(typ reflect.Type, tagName string) []cachedField {
	// Check cache first
	if cached, ok := fc.cache.Load(typ); ok {
		if fields, ok := cached.([]cachedField); ok {
			return fields
		}
	}

	// Build cache
	fields := fc.buildFieldCache(typ, tagName)

	// Store in cache (or get existing if another goroutine stored it first)
	actual, _ := fc.cache.LoadOrStore(typ, fields)
	fields, _ = actual.([]cachedField)

	return fields
}

// buildFieldCache builds field cache with tag parsing and embedded detection.
func (fc *FieldCache) buildFieldCache(typ reflect.Type, tagName string) []cachedField {
	var fields []cachedField

	for i := 0; i < typ.NumField(); i++ {
		ft := typ.Field(i)

		// Parse struct tag
		tag := ft.Tag.Get(tagName)
		if tag == "-" {
			continue // Skip fields with "-" tag
		}

		tagNameValue := fc.getFieldName(tag)
		fieldName := ft.Name

		// Precompute map key: tagName if present, otherwise exact field name
		mapKey := tagNameValue
		if mapKey == "" {
			mapKey = fieldName
		}

		// Determine if embedded
		embedded := ft.Anonymous

		fields = append(fields, cachedField{
			name:     fieldName,
			tagName:  tagNameValue,
			mapKey:   mapKey,
			index:    i,
			kind:     ft.Type.Kind(),
			embedded: embedded,
		})
	}

	return fields
}

// getFieldName extracts field name from tag.
func (fc *FieldCache) getFieldName(tag string) string {
	if tag == "" {
		return ""
	}

	// Parse tag: "name,option1,option2" -> "name"
	parts := strings.Split(tag, ",")
	if len(parts) > 0 && parts[0] != "" {
		return parts[0]
	}

	return ""
}
