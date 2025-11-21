package schema

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// defaultMarshaler returns a default Marshaler instance for testing.
func defaultMarshaler() Marshaler {
	fieldCache := NewFieldCache()

	return NewDefaultMarshaler("schema", fieldCache)
}

// TestMarshal_BasicTypes tests basic type marshaling.
func TestMarshal_BasicTypes(t *testing.T) {
	type TestStruct struct {
		F01 int     `schema:"f01"`
		F02 int     `schema:"-"`
		F03 string  `schema:"f03"`
		F04 string  `schema:"f04,omitempty"`
		F05 bool    `schema:"f05"`
		F06 bool    `schema:"f06"`
		F07 *string `schema:"f07"`
		F08 *int8   `schema:"f08"`
		F09 float64 `schema:"f09"`
	}

	f07 := "seven"
	var f08 int8 = 8
	s := &TestStruct{
		F01: 1,
		F02: 2,
		F03: "three",
		F04: "four",
		F05: true,
		F06: false,
		F07: &f07,
		F08: &f08,
		F09: 1.618,
	}

	marshaler := defaultMarshaler()
	result, err := marshaler.Marshal(s)
	require.NoError(t, err)

	f01Val, ok1 := result["f01"]
	require.True(t, ok1)
	f01, ok2 := f01Val.(int)
	if !ok2 {
		t.Fatalf("expected int, got %T", f01Val)
	}
	assert.Equal(t, 1, f01)
	assert.NotContains(t, result, "f02")
	f03Val, ok3 := result["f03"]
	if !ok3 {
		t.Fatalf("expected f03 in result")
	}
	assert.Equal(t, "three", f03Val)
	f04Val, ok4 := result["f04"]
	require.True(t, ok4)
	assert.Equal(t, "four", f04Val)
	assert.Equal(t, true, result["f05"])
	assert.Equal(t, false, result["f06"])
	assert.Equal(t, "seven", result["f07"])
	f08Val, ok5 := result["f08"]
	require.True(t, ok5)
	f08, ok6 := f08Val.(int8)
	require.True(t, ok6)
	assert.Equal(t, int8(8), f08)
	assert.InDelta(t, 1.618, result["f09"], 0.0001)
}

// TestMarshal_EmbeddedStruct tests embedded struct marshaling.
func TestMarshal_EmbeddedStruct(t *testing.T) {
	type inner struct {
		F12 int
	}

	type TestStruct struct {
		F01 int
		inner
	}

	s := &TestStruct{
		F01:   1,
		inner: inner{F12: 12},
	}

	marshaler := defaultMarshaler()
	result, err := marshaler.Marshal(s)
	require.NoError(t, err)

	f01Val, ok := result["F01"].(int)
	require.True(t, ok)
	assert.Equal(t, 1, f01Val)
	f12Val, ok := result["F12"].(int)
	require.True(t, ok)
	assert.Equal(t, 12, f12Val)
}

// TestMarshal_AllNumericTypes tests all numeric type variants.
func TestMarshal_AllNumericTypes(t *testing.T) {
	type TestStruct struct {
		F01 bool    `schema:"f01"`
		F02 float32 `schema:"f02"`
		F03 float64 `schema:"f03"`
		F04 int     `schema:"f04"`
		F05 int8    `schema:"f05"`
		F06 int16   `schema:"f06"`
		F07 int32   `schema:"f07"`
		F08 int64   `schema:"f08"`
		F09 string  `schema:"f09"`
		F10 uint    `schema:"f10"`
		F11 uint8   `schema:"f11"`
		F12 uint16  `schema:"f12"`
		F13 uint32  `schema:"f13"`
		F14 uint64  `schema:"f14"`
	}

	src := &TestStruct{
		F01: true,
		F02: 4.2,
		F03: 4.3,
		F04: -42,
		F05: -43,
		F06: -44,
		F07: -45,
		F08: -46,
		F09: "foo",
		F10: 42,
		F11: 43,
		F12: 44,
		F13: 45,
		F14: 46,
	}

	marshaler := defaultMarshaler()
	result, err := marshaler.Marshal(src)
	require.NoError(t, err)

	assert.Equal(t, true, result["f01"])
	f02Val, ok := result["f02"]
	require.True(t, ok)
	f02, ok := f02Val.(float32)
	require.True(t, ok)
	assert.InDelta(t, 4.2, f02, 0.0001)
	f03Val, ok := result["f03"]
	require.True(t, ok)
	f03, ok := f03Val.(float64)
	require.True(t, ok)
	assert.InDelta(t, 4.3, f03, 0.0001)
	f04Val, ok := result["f04"]
	require.True(t, ok)
	f04, ok := f04Val.(int)
	require.True(t, ok)
	assert.Equal(t, -42, f04)
	f05Val, ok := result["f05"]
	require.True(t, ok)
	f05, ok := f05Val.(int8)
	require.True(t, ok)
	assert.Equal(t, int8(-43), f05)
	f06Val, ok := result["f06"].(int16)
	require.True(t, ok)
	assert.Equal(t, int16(-44), f06Val)
	f07Val, ok := result["f07"].(int32)
	require.True(t, ok)
	assert.Equal(t, int32(-45), f07Val)
	f08Val, ok := result["f08"].(int64)
	require.True(t, ok)
	assert.Equal(t, int64(-46), f08Val)
	assert.Equal(t, "foo", result["f09"])
	f10Val, ok := result["f10"].(uint)
	require.True(t, ok)
	assert.Equal(t, uint(42), f10Val)
	f11Val, ok := result["f11"].(uint8)
	require.True(t, ok)
	assert.Equal(t, uint8(43), f11Val)
	f12Val, ok := result["f12"].(uint16)
	require.True(t, ok)
	assert.Equal(t, uint16(44), f12Val)
	f13Val, ok := result["f13"].(uint32)
	require.True(t, ok)
	assert.Equal(t, uint32(45), f13Val)
	f14Val, ok := result["f14"].(uint64)
	require.True(t, ok)
	assert.Equal(t, uint64(46), f14Val)
}

// TestMarshal_Empty tests marshaling with empty/zero values.
func TestMarshal_Empty(t *testing.T) {
	type TestStruct struct {
		F01 int    `schema:"f01"`
		F02 int    `schema:"-"`
		F03 string `schema:"f03"`
		F04 string `schema:"f04,omitempty"`
	}

	s := &TestStruct{
		F01: 1,
		F02: 2,
		F03: "three",
	}

	marshaler := defaultMarshaler()
	result, err := marshaler.Marshal(s)
	require.NoError(t, err)

	f01Val, ok := result["f01"].(int)
	require.True(t, ok)
	assert.Equal(t, 1, f01Val)
	assert.NotContains(t, result, "f02")
	f03Val, ok := result["f03"].(string)
	require.True(t, ok)
	assert.Equal(t, "three", f03Val)
	assert.NotContains(t, result, "f04")
}

// TestMarshal_NonStruct tests marshaling non-struct types.
func TestMarshal_NonStruct(t *testing.T) {
	marshaler := defaultMarshaler()
	_, err := marshaler.Marshal("hello world")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot marshal non-struct")
}

// TestMarshal_Slices tests slice marshaling.
func TestMarshal_Slices(t *testing.T) {
	type TestStruct struct {
		Ints     []int     `schema:"ints"`
		Nonempty []int     `schema:"nonempty"`
		Empty    []int     `schema:"empty,omitempty"`
		Strings  []string  `schema:"strings"`
		Bools    []bool    `schema:"bools"`
		Floats   []float64 `schema:"floats"`
		Ptrs     []*int    `schema:"ptrs"`
		NilSlice []int     `schema:"nilslice,omitempty"`
	}

	ptr1 := 1
	ptr2 := 2
	s1 := &TestStruct{
		Ints:     []int{1, 1},
		Nonempty: []int{},
		Strings:  []string{"a", "b"},
		Bools:    []bool{true, false},
		Floats:   []float64{1.1, 2.2},
		Ptrs:     []*int{&ptr1, &ptr2},
	}

	marshaler := defaultMarshaler()
	result, err := marshaler.Marshal(s1)
	require.NoError(t, err)

	assert.Equal(t, []any{1, 1}, result["ints"])
	assert.Equal(t, []any{}, result["nonempty"])
	assert.NotContains(t, result, "empty")
	assert.Equal(t, []any{"a", "b"}, result["strings"])
	assert.Equal(t, []any{true, false}, result["bools"])
	floatsVal, ok := result["floats"].([]any)
	require.True(t, ok)
	require.Len(t, floatsVal, 2)
	f1, ok := floatsVal[0].(float64)
	require.True(t, ok)
	f2, ok := floatsVal[1].(float64)
	require.True(t, ok)
	assert.InDeltaSlice(t, []float64{1.1, 2.2}, []float64{f1, f2}, 0.0001)
	assert.Equal(t, []any{1, 2}, result["ptrs"])
	assert.NotContains(t, result, "nilslice")
}

// TestMarshal_SliceOfStructs tests slice of structs marshaling.
func TestMarshal_SliceOfStructs(t *testing.T) {
	type Item struct {
		ID   int
		Name string
	}

	type TestStruct struct {
		Items []Item `schema:"items"`
	}

	s := &TestStruct{
		Items: []Item{
			{ID: 1, Name: "first"},
			{ID: 2, Name: "second"},
		},
	}

	marshaler := defaultMarshaler()
	result, err := marshaler.Marshal(s)
	require.NoError(t, err)

	items, ok := result["items"].([]any)
	require.True(t, ok)
	require.Len(t, items, 2)

	item1, ok := items[0].(map[string]any)
	require.True(t, ok)
	idVal, ok := item1["ID"].(int)
	require.True(t, ok)
	assert.Equal(t, 1, idVal)
	assert.Equal(t, "first", item1["Name"])

	item2, ok := items[1].(map[string]any)
	require.True(t, ok)
	idVal2, ok := item2["ID"].(int)
	require.True(t, ok)
	assert.Equal(t, 2, idVal2)
	assert.Equal(t, "second", item2["Name"])
}

// TestMarshal_NestedStruct tests nested struct marshaling.
func TestMarshal_NestedStruct(t *testing.T) {
	type Nested struct {
		Field int
	}

	type TestStruct struct {
		Nested Nested `schema:"nested"`
	}

	s := &TestStruct{
		Nested: Nested{Field: 42},
	}

	marshaler := defaultMarshaler()
	result, err := marshaler.Marshal(s)
	require.NoError(t, err)

	nested, ok := result["nested"].(map[string]any)
	require.True(t, ok)
	fieldVal, ok := nested["Field"].(int)
	require.True(t, ok)
	assert.Equal(t, 42, fieldVal)
}

// TestMarshal_PointerToStruct tests pointer to struct marshaling.
func TestMarshal_PointerToStruct(t *testing.T) {
	type Nested struct {
		Value int
	}

	type TestStruct struct {
		Nested *Nested `schema:"nested"`
		NilPtr *Nested `schema:"nilptr,omitempty"`
	}

	nested := &Nested{Value: 42}
	s := &TestStruct{
		Nested: nested,
		NilPtr: nil,
	}

	marshaler := defaultMarshaler()
	result, err := marshaler.Marshal(s)
	require.NoError(t, err)

	nestedMap, ok := result["nested"].(map[string]any)
	require.True(t, ok)
	valueVal, ok := nestedMap["Value"].(int)
	require.True(t, ok)
	assert.Equal(t, 42, valueVal)
	assert.NotContains(t, result, "nilptr")
}

// TestMarshal_OmitEmpty tests omitempty tag functionality.
func TestMarshal_OmitEmpty(t *testing.T) {
	type Nested struct {
		F0601 string `schema:"f0601,omitempty"`
	}

	type TestStruct struct {
		F01 int      `schema:"f01,omitempty"`
		F02 string   `schema:"f02,omitempty"`
		F03 *string  `schema:"f03,omitempty"`
		F04 *int8    `schema:"f04,omitempty"`
		F05 float64  `schema:"f05,omitempty"`
		F06 Nested   `schema:"f06,omitempty"`
		F07 Nested   `schema:"f07,omitempty"`
		F08 []string `schema:"f08,omitempty"`
		F09 []string `schema:"f09,omitempty"`
	}

	s := TestStruct{
		F02: "test",
		F07: Nested{
			F0601: "test",
		},
		F09: []string{"test"},
	}

	marshaler := defaultMarshaler()
	result, err := marshaler.Marshal(s)
	require.NoError(t, err)

	assert.NotContains(t, result, "f01")
	assert.Equal(t, "test", result["f02"])
	assert.NotContains(t, result, "f03")
	assert.NotContains(t, result, "f04")
	assert.NotContains(t, result, "f05")
	assert.NotContains(t, result, "f06")
	f07, ok := result["f07"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "test", f07["f0601"])
	assert.NotContains(t, result, "f08")
	assert.Equal(t, []any{"test"}, result["f09"])
}

// TestMarshal_TypeAliases tests type alias marshaling.
func TestMarshal_TypeAliases(t *testing.T) {
	type IntAlias int
	type StringAlias string

	type TestStruct struct {
		IntVal    IntAlias    `schema:"intval"`
		StringVal StringAlias `schema:"stringval"`
	}

	s := TestStruct{
		IntVal:    IntAlias(42),
		StringVal: StringAlias("hello"),
	}

	marshaler := defaultMarshaler()
	result, err := marshaler.Marshal(s)
	require.NoError(t, err)

	// Type aliases preserve their type
	intVal, ok := result["intval"].(IntAlias)
	require.True(t, ok)
	assert.Equal(t, IntAlias(42), intVal)
	assert.Equal(t, "hello", result["stringval"])
}

// TestMarshal_ComplexNested tests complex nested structures.
func TestMarshal_ComplexNested(t *testing.T) {
	type Tag struct {
		Name string
	}

	type Metadata struct {
		Tags []Tag
	}

	type Item struct {
		ID       int
		Metadata Metadata
	}

	type Container struct {
		Items *[]*Item
	}

	type TestStruct struct {
		Container Container
	}

	item1 := &Item{
		ID: 1,
		Metadata: Metadata{
			Tags: []Tag{{Name: "tag1"}},
		},
	}
	item2 := &Item{
		ID: 2,
		Metadata: Metadata{
			Tags: []Tag{{Name: "tag2"}},
		},
	}
	items := []*Item{item1, item2}

	s := TestStruct{
		Container: Container{
			Items: &items,
		},
	}

	marshaler := defaultMarshaler()
	result, err := marshaler.Marshal(s)
	require.NoError(t, err)

	container, ok := result["Container"].(map[string]any)
	require.True(t, ok)
	itemsArr, ok := container["Items"].([]any)
	require.True(t, ok)
	require.Len(t, itemsArr, 2)

	item1Map, ok := itemsArr[0].(map[string]any)
	require.True(t, ok)
	idValMap, ok := item1Map["ID"].(int)
	require.True(t, ok)
	assert.Equal(t, 1, idValMap)

	metadata1, ok := item1Map["Metadata"].(map[string]any)
	require.True(t, ok)
	tags1, ok := metadata1["Tags"].([]any)
	require.True(t, ok)
	require.Len(t, tags1, 1)
	tag1, ok := tags1[0].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "tag1", tag1["Name"])
}

// TestMarshal_NilPointer tests nil pointer handling.
func TestMarshal_NilPointer(t *testing.T) {
	type TestStruct struct {
		Value *int `schema:"value"`
		Nil   *int `schema:"nil,omitempty"`
	}

	val := 42
	s := TestStruct{
		Value: &val,
		Nil:   nil,
	}

	marshaler := defaultMarshaler()
	result, err := marshaler.Marshal(s)
	require.NoError(t, err)

	valueVal, ok := result["value"].(int)
	require.True(t, ok)
	assert.Equal(t, 42, valueVal)
	assert.NotContains(t, result, "nil")
}

// TestMarshal_NilPointerExplicit tests explicit nil pointer.
func TestMarshal_NilPointerExplicit(t *testing.T) {
	type TestStruct struct {
		Value *int `schema:"value"`
	}

	s := TestStruct{
		Value: nil,
	}

	marshaler := defaultMarshaler()
	result, err := marshaler.Marshal(s)
	require.NoError(t, err)

	assert.Nil(t, result["value"])
}

// TestMarshal_AnonymousEmbedded tests anonymous embedded struct.
func TestMarshal_AnonymousEmbedded(t *testing.T) {
	type Inner struct {
		F12 int
	}

	type TestStruct struct {
		F01 int
		Inner
	}

	s := TestStruct{
		F01:   1,
		Inner: Inner{F12: 12},
	}

	marshaler := defaultMarshaler()
	result, err := marshaler.Marshal(s)
	require.NoError(t, err)

	f01Val2, ok := result["F01"].(int)
	require.True(t, ok)
	assert.Equal(t, 1, f01Val2)
	f12Val2, ok := result["F12"].(int)
	require.True(t, ok)
	assert.Equal(t, 12, f12Val2)
}

// TestMarshal_NamedEmbedded tests named embedded struct.
func TestMarshal_NamedEmbedded(t *testing.T) {
	type Inner struct {
		F12 int
	}

	type TestStruct struct {
		F01   int
		Inner Inner
	}

	s := TestStruct{
		F01:   1,
		Inner: Inner{F12: 12},
	}

	marshaler := defaultMarshaler()
	result, err := marshaler.Marshal(s)
	require.NoError(t, err)

	f01Val3, ok := result["F01"].(int)
	require.True(t, ok)
	assert.Equal(t, 1, f01Val3)
	inner, ok := result["Inner"].(map[string]any)
	require.True(t, ok)
	f12Val3, ok := inner["F12"].(int)
	require.True(t, ok)
	assert.Equal(t, 12, f12Val3)
}

// TestMarshal_IntVariants tests all int variants.
func TestMarshal_IntVariants(t *testing.T) {
	type TestStruct struct {
		IntVal   int
		Int8Val  int8
		Int16Val int16
		Int32Val int32
		Int64Val int64
	}

	s := TestStruct{
		IntVal:   42,
		Int8Val:  127,
		Int16Val: 32767,
		Int32Val: 2147483647,
		Int64Val: 9223372036854775807,
	}

	marshaler := defaultMarshaler()
	result, err := marshaler.Marshal(s)
	require.NoError(t, err)

	intVal, ok := result["IntVal"].(int)
	require.True(t, ok)
	assert.Equal(t, 42, intVal)
	int8Val, ok := result["Int8Val"].(int8)
	require.True(t, ok)
	assert.Equal(t, int8(127), int8Val)
	int16Val, ok := result["Int16Val"].(int16)
	require.True(t, ok)
	assert.Equal(t, int16(32767), int16Val)
	int32Val, ok := result["Int32Val"].(int32)
	require.True(t, ok)
	assert.Equal(t, int32(2147483647), int32Val)
	int64Val, ok := result["Int64Val"].(int64)
	require.True(t, ok)
	assert.Equal(t, int64(9223372036854775807), int64Val)
}

// TestMarshal_UintVariants tests all uint variants.
func TestMarshal_UintVariants(t *testing.T) {
	type TestStruct struct {
		UintVal   uint
		Uint8Val  uint8
		Uint16Val uint16
		Uint32Val uint32
		Uint64Val uint64
	}

	s := TestStruct{
		UintVal:   42,
		Uint8Val:  255,
		Uint16Val: 65535,
		Uint32Val: 4294967295,
		Uint64Val: 18446744073709551615,
	}

	marshaler := defaultMarshaler()
	result, err := marshaler.Marshal(s)
	require.NoError(t, err)

	uintVal, ok := result["UintVal"].(uint)
	require.True(t, ok)
	assert.Equal(t, uint(42), uintVal)
	uint8Val, ok := result["Uint8Val"].(uint8)
	require.True(t, ok)
	assert.Equal(t, uint8(255), uint8Val)
	uint16Val, ok := result["Uint16Val"].(uint16)
	require.True(t, ok)
	assert.Equal(t, uint16(65535), uint16Val)
	uint32Val, ok := result["Uint32Val"].(uint32)
	require.True(t, ok)
	assert.Equal(t, uint32(4294967295), uint32Val)
	uint64Val, ok := result["Uint64Val"].(uint64)
	require.True(t, ok)
	assert.Equal(t, uint64(18446744073709551615), uint64Val)
}

// TestMarshal_FloatVariants tests float variants.
func TestMarshal_FloatVariants(t *testing.T) {
	type TestStruct struct {
		Float32Val float32
		Float64Val float64
	}

	s := TestStruct{
		Float32Val: 3.14,
		Float64Val: 2.718,
	}

	marshaler := defaultMarshaler()
	result, err := marshaler.Marshal(s)
	require.NoError(t, err)

	float32Val, ok := result["Float32Val"].(float32)
	require.True(t, ok)
	assert.InDelta(t, 3.14, float32Val, 0.0001)
	float64Val, ok := result["Float64Val"].(float64)
	require.True(t, ok)
	assert.InDelta(t, 2.718, float64Val, 0.0001)
}

// TestMarshal_SliceOfPointers tests slice of pointers.
func TestMarshal_SliceOfPointers(t *testing.T) {
	type TestStruct struct {
		Values []*int
	}

	val1 := 1
	val2 := 2
	val3 := 3
	s := TestStruct{
		Values: []*int{&val1, &val2, &val3},
	}

	marshaler := defaultMarshaler()
	result, err := marshaler.Marshal(s)
	require.NoError(t, err)

	values, ok := result["Values"].([]any)
	require.True(t, ok)
	require.Len(t, values, 3)
	assert.Equal(t, 1, values[0])
	assert.Equal(t, 2, values[1])
	assert.Equal(t, 3, values[2])
}

// TestMarshal_PointerToSlice tests pointer to slice.
func TestMarshal_PointerToSlice(t *testing.T) {
	type TestStruct struct {
		Value *[]int
	}

	vals := []int{1, 2, 3}
	s := TestStruct{
		Value: &vals,
	}

	marshaler := defaultMarshaler()
	result, err := marshaler.Marshal(s)
	require.NoError(t, err)

	value, ok := result["Value"].([]any)
	require.True(t, ok)
	assert.Equal(t, []any{1, 2, 3}, value)
}

// TestMarshal_NilPointerToSlice tests nil pointer to slice.
func TestMarshal_NilPointerToSlice(t *testing.T) {
	type TestStruct struct {
		Value *[]int `schema:"value,omitempty"`
	}

	s := TestStruct{
		Value: nil,
	}

	marshaler := defaultMarshaler()
	result, err := marshaler.Marshal(s)
	require.NoError(t, err)

	assert.NotContains(t, result, "value")
}

// TestMarshal_DeeplyNested tests deeply nested structures.
func TestMarshal_DeeplyNested(t *testing.T) {
	type Level3 struct {
		Value int
	}

	type Level2 struct {
		Level3 Level3
	}

	type Level1 struct {
		Level2 Level2
	}

	type TestStruct struct {
		Level1 Level1
	}

	s := TestStruct{
		Level1: Level1{
			Level2: Level2{
				Level3: Level3{Value: 42},
			},
		},
	}

	marshaler := defaultMarshaler()
	result, err := marshaler.Marshal(s)
	require.NoError(t, err)

	level1, ok := result["Level1"].(map[string]any)
	require.True(t, ok)
	level2, ok := level1["Level2"].(map[string]any)
	require.True(t, ok)
	level3, ok := level2["Level3"].(map[string]any)
	require.True(t, ok)
	valueVal2, ok := level3["Value"].(int)
	require.True(t, ok)
	assert.Equal(t, 42, valueVal2)
}

// TestMarshal_NilPointerToStruct tests nil pointer to struct.
func TestMarshal_NilPointerToStruct(t *testing.T) {
	type Nested struct {
		Value int
	}

	type TestStruct struct {
		Nested *Nested `schema:"nested,omitempty"`
	}

	s := TestStruct{
		Nested: nil,
	}

	marshaler := defaultMarshaler()
	result, err := marshaler.Marshal(s)
	require.NoError(t, err)

	assert.NotContains(t, result, "nested")
}

// TestMarshal_ZeroValues tests zero value handling.
func TestMarshal_ZeroValues(t *testing.T) {
	type TestStruct struct {
		EmptyString string  `schema:"emptystring"`
		ZeroInt     int     `schema:"zeroint"`
		ZeroFloat   float32 `schema:"zerofloat"`
		FalseBool   bool    `schema:"falsebool"`
		NilSlice    []int   `schema:"nilslice,omitempty"`
		EmptySlice  []int   `schema:"emptyslice"`
	}

	s := TestStruct{}

	marshaler := defaultMarshaler()
	result, err := marshaler.Marshal(s)
	require.NoError(t, err)

	assert.Equal(t, "", result["emptystring"])
	zeroIntVal, ok := result["zeroint"].(int)
	require.True(t, ok)
	assert.Equal(t, 0, zeroIntVal)
	zeroFloatVal, ok := result["zerofloat"].(float32)
	require.True(t, ok)
	assert.InDelta(t, 0.0, zeroFloatVal, 0.0001)
	assert.Equal(t, false, result["falsebool"])
	assert.NotContains(t, result, "nilslice")
	assert.Equal(t, []any{}, result["emptyslice"])
}

// TestMarshal_StructPointerWithPointerFields tests pointer to struct with pointer fields.
func TestMarshal_StructPointerWithPointerFields(t *testing.T) {
	type Nested struct {
		Value *int
	}

	type TestStruct struct {
		Nested *Nested
	}

	val := 42
	s := TestStruct{
		Nested: &Nested{Value: &val},
	}

	marshaler := defaultMarshaler()
	result, err := marshaler.Marshal(s)
	require.NoError(t, err)

	nested, ok := result["Nested"].(map[string]any)
	require.True(t, ok)
	valueVal3, ok := nested["Value"].(int)
	require.True(t, ok)
	assert.Equal(t, 42, valueVal3)
}

func TestUnmarshal_Converter_ShouldNotApply_ToNestedMap(t *testing.T) {
	type Point struct {
		X, Y int
	}

	fieldCache := NewFieldCache()
	converters := NewConverterRegistry()
	// Register default converters for int
	converters.Register(reflect.TypeOf(int(0)), convertInt)
	converters.Register(reflect.TypeOf(Point{}), func(s string) (reflect.Value, error) {
		return reflect.Value{}, fmt.Errorf("converter should not be called")
	})
	unmarshaler := NewDefaultUnmarshaler("schema", fieldCache, converters)

	type Config struct {
		Origin Point `schema:"origin"`
	}

	// User sends nested object (map), not string
	data := map[string]any{
		"origin": map[string]any{"X": "10", "Y": "20"},
	}

	var cfg Config
	err := unmarshaler.Unmarshal(data, &cfg)
	require.NoError(t, err) // Currently FAILS: tries to use converter on map

	assert.Equal(t, Point{X: 10, Y: 20}, cfg.Origin)
}
