package mapstructure

import "reflect"

// FieldMetadata holds cached struct field information.
type FieldMetadata struct {
	StructFieldName string       // Go field name
	MapKey          string       // Key to lookup in map
	Index           int          // Field index for reflection
	Type            reflect.Type // Field type
	Embedded        bool         // Anonymous/embedded struct
}

// StructMetadata holds cached metadata for a struct type.
type StructMetadata struct {
	Fields []FieldMetadata
}
