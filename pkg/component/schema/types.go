package schema

import "reflect"

// Local constants for reflect.Kind to satisfy linter requirements.
const (
	kindPtr     = reflect.Ptr
	kindSlice   = reflect.Slice
	kindStruct  = reflect.Struct
	kindBool    = reflect.Bool
	kindString  = reflect.String
	kindInt     = reflect.Int
	kindInt8    = reflect.Int8
	kindInt16   = reflect.Int16
	kindInt32   = reflect.Int32
	kindInt64   = reflect.Int64
	kindUint    = reflect.Uint
	kindUint8   = reflect.Uint8
	kindUint16  = reflect.Uint16
	kindUint32  = reflect.Uint32
	kindUint64  = reflect.Uint64
	kindFloat32 = reflect.Float32
	kindFloat64 = reflect.Float64
)
