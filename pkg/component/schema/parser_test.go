package schema

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// assertFieldMetadata compares got field metadata against want field metadata.
func assertFieldMetadata(t *testing.T, fieldName string, gotField, wantField FieldMetadata) {
	t.Helper()
	if wantField.StructFieldName != "" {
		assert.Equal(t, wantField.StructFieldName, gotField.StructFieldName)
	}
	if wantField.ParamName != "" {
		assert.Equal(t, wantField.ParamName, gotField.ParamName)
	}
	if wantField.MapKey != "" {
		assert.Equal(t, wantField.MapKey, gotField.MapKey)
	}
	if wantField.Location != "" {
		assert.Equal(t, wantField.Location, gotField.Location)
	}
	if wantField.Style != "" {
		assert.Equal(t, wantField.Style, gotField.Style)
	}
	if wantField.BodyType != "" {
		assert.Equal(t, wantField.BodyType, gotField.BodyType)
	}
	assert.Equal(t, wantField.IsParameter, gotField.IsParameter, "IsParameter mismatch for %s", fieldName)
	assert.Equal(t, wantField.IsBody, gotField.IsBody, "IsBody mismatch for %s", fieldName)
	assert.Equal(t, wantField.Explode, gotField.Explode, "Explode mismatch for %s", fieldName)
	assert.Equal(t, wantField.Required, gotField.Required, "Required mismatch for %s", fieldName)
	assert.Equal(t, wantField.Embedded, gotField.Embedded, "Embedded mismatch for %s", fieldName)
}

// runBuildStructMetadataTests runs table-driven tests for BuildStructMetadata.
func runBuildStructMetadataTests(t *testing.T, tests []struct {
	name string
	s    any
	want map[string]FieldMetadata
},
) {
	t.Helper()

	parser := newStructMetadataParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata, err := parser.BuildStructMetadata(reflect.TypeOf(tt.s))

			require.NoError(t, err)
			require.NotNil(t, metadata)
			require.Len(t, metadata.Fields, len(tt.want))

			fieldMap := make(map[string]FieldMetadata)
			for _, field := range metadata.Fields {
				fieldMap[field.StructFieldName] = field
			}

			for fieldName, wantField := range tt.want {
				gotField, ok := fieldMap[fieldName]
				require.True(t, ok, "field %s not found", fieldName)
				assertFieldMetadata(t, fieldName, gotField, wantField)
			}
		})
	}
}

func TestBuildStructMetadata_Basic(t *testing.T) {
	tests := []struct {
		name string
		s    any
		want map[string]FieldMetadata
	}{
		{
			name: "basic schema tags",
			s: struct {
				ID   int    `schema:"id"`
				Name string `schema:"name"`
				Age  int    // No tag - gets default metadata
			}{},
			want: map[string]FieldMetadata{
				"ID": {
					StructFieldName: "ID",
					ParamName:       "id",
					MapKey:          "id",
					Location:        LocationQuery,
					Style:           StyleForm,
					Explode:         true,
					Required:        false,
					IsParameter:     true,
				},
				"Name": {
					StructFieldName: "Name",
					ParamName:       "name",
					MapKey:          "name",
					Location:        LocationQuery,
					Style:           StyleForm,
					Explode:         true,
					IsParameter:     true,
				},
				"Age": {
					StructFieldName: "Age",
					ParamName:       "Age",
					MapKey:          "Age",
					Location:        LocationQuery,
					Style:           StyleForm,
					Explode:         true,
					IsParameter:     true,
				},
			},
		},
		{
			name: "empty tag name uses field name",
			s: struct {
				FieldName string `schema:",required"`
				Another   string `schema:",location=path"`
			}{},
			want: map[string]FieldMetadata{
				"FieldName": {
					StructFieldName: "FieldName",
					ParamName:       "FieldName",
					MapKey:          "FieldName",
					Location:        LocationQuery,
					Style:           StyleForm,
					Explode:         true,
					Required:        true,
					IsParameter:     true,
				},
				"Another": {
					StructFieldName: "Another",
					ParamName:       "Another",
					MapKey:          "Another",
					Location:        LocationPath,
					Style:           StyleSimple,
					Explode:         false,
					Required:        true,
					IsParameter:     true,
				},
			},
		},
	}

	runBuildStructMetadataTests(t, tests)
}

func TestBuildStructMetadata_Locations(t *testing.T) {
	tests := []struct {
		name string
		s    any
		want map[string]FieldMetadata
	}{
		{
			name: "all locations",
			s: struct {
				QueryParam  string `schema:"q,location=query"`
				PathParam   int    `schema:"id,location=path"`
				HeaderParam string `schema:"authorization,location=header"`
				CookieParam string `schema:"session,location=cookie"`
			}{},
			want: map[string]FieldMetadata{
				"QueryParam": {
					Location:    LocationQuery,
					Style:       StyleForm,
					Explode:     true,
					IsParameter: true,
				},
				"PathParam": {
					Location:    LocationPath,
					Style:       StyleSimple,
					Explode:     false,
					Required:    true,
					IsParameter: true,
				},
				"HeaderParam": {
					Location:    LocationHeader,
					Style:       StyleSimple,
					Explode:     false,
					IsParameter: true,
				},
				"CookieParam": {
					Location:    LocationCookie,
					Style:       StyleForm,
					Explode:     true,
					IsParameter: true,
				},
			},
		},
	}

	runBuildStructMetadataTests(t, tests)
}

func TestBuildStructMetadata_Styles(t *testing.T) {
	tests := []struct {
		name string
		s    any
		want map[string]FieldMetadata
	}{
		{
			name: "all styles",
			s: struct {
				FormStyle       map[string]string `schema:"filter,location=query,style=form"`
				DeepObjectStyle map[string]string `schema:"filter,location=query,style=deepObject"`
				SimpleStyle     []int             `schema:"ids,location=path,style=simple"`
				SpaceDelimited  []int             `schema:"tags,location=query,style=spaceDelimited"`
				PipeDelimited   []int             `schema:"items,location=query,style=pipeDelimited"`
			}{},
			want: map[string]FieldMetadata{
				"FormStyle": {
					Style:       StyleForm,
					Explode:     true,
					IsParameter: true,
				},
				"DeepObjectStyle": {
					Style:       StyleDeepObject,
					Explode:     true,
					IsParameter: true,
				},
				"SimpleStyle": {
					Style:       StyleSimple,
					Explode:     false,
					Required:    true, // Path params always required
					IsParameter: true,
				},
				"SpaceDelimited": {
					Style:       StyleSpaceDelimited,
					Explode:     false,
					IsParameter: true,
				},
				"PipeDelimited": {
					Style:       StylePipeDelimited,
					Explode:     false,
					IsParameter: true,
				},
			},
		},
	}

	runBuildStructMetadataTests(t, tests)
}

func TestBuildStructMetadata_Explode(t *testing.T) {
	tests := []struct {
		name string
		s    any
		want map[string]FieldMetadata
	}{
		{
			name: "explode options",
			s: struct {
				ExplodeFlag      []int `schema:"ids,explode"`
				ExplodeTrue      []int `schema:"ids,explode=true"`
				ExplodeFalse     []int `schema:"ids,explode=false"`
				DefaultExplode   []int `schema:"tags,location=query,style=form"` // true
				DefaultNoExplode []int `schema:"ids,location=path,style=simple"` // false
			}{},
			want: map[string]FieldMetadata{
				"ExplodeFlag": {
					Explode:     true,
					IsParameter: true,
				},
				"ExplodeTrue": {
					Explode:     true,
					IsParameter: true,
				},
				"ExplodeFalse": {
					Explode:     false,
					IsParameter: true,
				},
				"DefaultExplode": {
					Explode:     true,
					IsParameter: true,
				},
				"DefaultNoExplode": {
					Explode:     false,
					Required:    true, // Path params always required
					IsParameter: true,
				},
			},
		},
	}

	runBuildStructMetadataTests(t, tests)
}

func TestBuildStructMetadata_Required(t *testing.T) {
	tests := []struct {
		name string
		s    any
		want map[string]FieldMetadata
	}{
		{
			name: "required flag",
			s: struct {
				RequiredFlag     string `schema:"name,required"`
				RequiredTrue     string `schema:"name,required=true"`
				RequiredFalse    string `schema:"filter,required=false"`
				PathAutoRequired int    `schema:"id,location=path"` // Always required
				QueryOptional    string `schema:"filter,location=query"`
			}{},
			want: map[string]FieldMetadata{
				"RequiredFlag": {
					Required:    true,
					Explode:     true, // Default for form style
					IsParameter: true,
				},
				"RequiredTrue": {
					Required:    true,
					Explode:     true, // Default for form style
					IsParameter: true,
				},
				"RequiredFalse": {
					Required:    false,
					Explode:     true, // Default for form style
					IsParameter: true,
				},
				"PathAutoRequired": {
					Required:    true,
					Explode:     false, // Default for simple style
					IsParameter: true,
				},
				"QueryOptional": {
					Required:    false,
					Explode:     true, // Default for form style
					IsParameter: true,
				},
			},
		},
	}

	runBuildStructMetadataTests(t, tests)
}

func TestBuildStructMetadata_BodyTags(t *testing.T) {
	tests := []struct {
		name string
		s    any
		want map[string]FieldMetadata
	}{
		{
			name: "body tags",
			s: struct {
				StructuredBody map[string]any    `body:""`
				FileBody       []byte            `body:"file"`
				MultipartBody  map[string]string `body:"multipart"`
				RequiredBody   map[string]any    `body:"structured,required"`
			}{},
			want: map[string]FieldMetadata{
				"StructuredBody": {
					IsBody:   true,
					BodyType: BodyTypeStructured,
					Required: false,
				},
				"FileBody": {
					IsBody:   true,
					BodyType: BodyTypeFile,
				},
				"MultipartBody": {
					IsBody:   true,
					BodyType: BodyTypeMultipart,
				},
				"RequiredBody": {
					IsBody:   true,
					Required: true,
				},
			},
		},
	}

	runBuildStructMetadataTests(t, tests)
}

func TestBuildStructMetadata_SkipFields(t *testing.T) {
	tests := []struct {
		name string
		s    any
		want map[string]FieldMetadata
	}{
		{
			name: "skip fields",
			s: struct {
				Included  string `schema:"name"`
				Skipped1  string `schema:"-"`
				Skipped2  string `body:"-"`
				Included2 int    // Included (exported, no skip tag)
			}{},
			want: map[string]FieldMetadata{
				"Included": {
					StructFieldName: "Included",
					Explode:         true,
					IsParameter:     true,
				},
				"Included2": {
					StructFieldName: "Included2",
					Explode:         true,
					IsParameter:     true,
				},
			},
		},
	}

	runBuildStructMetadataTests(t, tests)
}

func TestBuildStructMetadata_EmbeddedStructs(t *testing.T) {
	tests := []struct {
		name string
		s    any
		want map[string]FieldMetadata
	}{
		{
			name: "embedded structs",
			s: struct {
				Embedded struct {
					Field1 int
					Field2 string
				}
				Field3 int
			}{},
			want: map[string]FieldMetadata{
				"Embedded": {
					StructFieldName: "Embedded",
					Embedded:        false, // Named field, not embedded
					Explode:         true,
					IsParameter:     true,
				},
				"Field3": {
					StructFieldName: "Field3",
					Embedded:        false,
					Explode:         true,
					IsParameter:     true,
				},
			},
		},
	}

	runBuildStructMetadataTests(t, tests)
}

func TestBuildStructMetadata_Conflict(t *testing.T) {
	tests := []struct {
		name     string
		s        any
		wantErr  bool
		errorMsg string
	}{
		{
			name: "conflict between schema and body tags",
			s: struct {
				Conflict string `schema:"id" body:""`
			}{},
			wantErr:  true,
			errorMsg: "cannot have both schema and body tags",
		},
	}

	parser := newStructMetadataParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata, err := parser.BuildStructMetadata(reflect.TypeOf(tt.s))

			if tt.wantErr {
				require.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				assert.Nil(t, metadata)
			} else {
				require.NoError(t, err)
				require.NotNil(t, metadata)
			}
		})
	}
}

func TestBuildStructMetadata_Errors(t *testing.T) {
	tests := []struct {
		name     string
		s        any
		wantErr  bool
		errorMsg string
	}{
		{
			name: "invalid style for location",
			s: struct {
				Field int `schema:"id,location=query,style=matrix"`
			}{},
			wantErr:  true,
			errorMsg: "invalid style",
		},
		{
			name: "invalid body type",
			s: struct {
				Field map[string]any `body:"invalid"`
			}{},
			wantErr:  true,
			errorMsg: "invalid body type",
		},
	}

	parser := newStructMetadataParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata, err := parser.BuildStructMetadata(reflect.TypeOf(tt.s))

			if tt.wantErr {
				require.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				assert.Nil(t, metadata)
			} else {
				require.NoError(t, err)
				require.NotNil(t, metadata)
			}
		})
	}
}

func TestBuildStructMetadata_EdgeCases(t *testing.T) {
	tests := []struct {
		name string
		s    any
		want map[string]FieldMetadata
	}{
		{
			name: "edge cases - whitespace and empty values",
			s: struct {
				WhitespaceName string `schema:"  user_name  ,location=query"`
				EmptyLocation  int    `schema:"id,location="`
				EmptyStyle     int    `schema:"id,style:   "`
			}{},
			want: map[string]FieldMetadata{
				"WhitespaceName": {
					ParamName:   "user_name",
					Explode:     true,
					IsParameter: true,
				},
				"EmptyLocation": {
					Location:    LocationQuery,
					Style:       StyleForm,
					Explode:     true,
					IsParameter: true,
				},
				"EmptyStyle": {
					Style:       StyleForm,
					Explode:     true,
					IsParameter: true,
				},
			},
		},
	}

	runBuildStructMetadataTests(t, tests)
}

func TestBuildStructMetadata_DuplicateOptions(t *testing.T) {
	tests := []struct {
		name string
		s    any
		want map[string]FieldMetadata
	}{
		{
			name: "duplicate options",
			s: struct {
				MultipleRequired string `schema:"name,required,required"`
				LastWinsRequired string `schema:"name,required=false,required"`
				LastWinsExplode  []int  `schema:"ids,explode=false,explode"`
			}{},
			want: map[string]FieldMetadata{
				"MultipleRequired": {
					Required:    true,
					Explode:     true,
					IsParameter: true,
				},
				"LastWinsRequired": {
					Required:    true,
					Explode:     true,
					IsParameter: true,
				},
				"LastWinsExplode": {
					Explode:     true,
					IsParameter: true,
				},
			},
		},
	}

	runBuildStructMetadataTests(t, tests)
}

func TestBuildStructMetadata_CustomTagNames(t *testing.T) {
	tests := []struct {
		name       string
		s          any
		want       map[string]FieldMetadata
		parserOpts []StructMetadataParserOption
	}{
		{
			name: "custom tag names",
			s: struct {
				Param string         `param:"custom"`
				Body  map[string]any `request:""`
			}{},
			parserOpts: []StructMetadataParserOption{
				WithSchemaTag("param"),
				WithBodyTag("request"),
			},
			want: map[string]FieldMetadata{
				"Param": {
					ParamName:   "custom",
					Explode:     true,
					IsParameter: true,
				},
				"Body": {
					IsBody:   true,
					BodyType: BodyTypeStructured,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := newStructMetadataParser(tt.parserOpts...)
			metadata, err := parser.BuildStructMetadata(reflect.TypeOf(tt.s))

			require.NoError(t, err)
			require.NotNil(t, metadata)
			require.Len(t, metadata.Fields, len(tt.want))

			fieldMap := make(map[string]FieldMetadata)
			for _, field := range metadata.Fields {
				fieldMap[field.StructFieldName] = field
			}

			for fieldName, wantField := range tt.want {
				gotField, ok := fieldMap[fieldName]
				require.True(t, ok, "field %s not found", fieldName)
				assertFieldMetadata(t, fieldName, gotField, wantField)
			}
		})
	}
}

func TestBuildStructMetadata_ComplexStruct(t *testing.T) {
	tests := []struct {
		name string
		s    any
		want map[string]FieldMetadata
	}{
		{
			name: "complex struct",
			s: struct {
				QueryParam    string            `schema:"q,location=query"`
				PathParam     int               `schema:"id,location=path"`
				HeaderParam   string            `schema:"auth,location=header,required"`
				CookieParam   string            `schema:"session,location=cookie"`
				Filter        map[string]string `schema:"filter,location=query,style=deepObject"`
				Tags          []string          `schema:"tags,location=query,explode"`
				OptionalQuery string            `schema:"optional,location=query"`
				DefaultField  int               // No tag
				Body          map[string]any    `body:"structured,required"`
				SkipField     string            `schema:"-"`
			}{},
			want: map[string]FieldMetadata{
				"QueryParam": {
					Location:    LocationQuery,
					Style:       StyleForm,
					Explode:     true,
					IsParameter: true,
				},
				"PathParam": {
					Location:    LocationPath,
					Style:       StyleSimple,
					Explode:     false,
					Required:    true,
					IsParameter: true,
				},
				"HeaderParam": {
					Location:    LocationHeader,
					Style:       StyleSimple,
					Explode:     false,
					Required:    true,
					IsParameter: true,
				},
				"CookieParam": {
					Location:    LocationCookie,
					Style:       StyleForm,
					Explode:     true,
					IsParameter: true,
				},
				"Filter": {
					Style:       StyleDeepObject,
					Explode:     true,
					IsParameter: true,
				},
				"Tags": {
					Style:       StyleForm,
					Explode:     true,
					IsParameter: true,
				},
				"OptionalQuery": {
					Location:    LocationQuery,
					Style:       StyleForm,
					Explode:     true,
					IsParameter: true,
				},
				"DefaultField": {
					ParamName:   "DefaultField",
					Explode:     true,
					IsParameter: true,
				},
				"Body": {
					IsBody:   true,
					Required: true,
				},
			},
		},
	}

	runBuildStructMetadataTests(t, tests)
}
