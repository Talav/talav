package schema

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseBodyTag(t *testing.T) {
	tests := []struct {
		name        string
		fieldName   string
		tagValue    string
		want        *BodyMetadata
		wantErr     bool
		errContains string
	}{
		{
			name:      "empty tag defaults to structured",
			fieldName: goFieldBody,
			tagValue:  "",
			want: &BodyMetadata{
				MapKey:   goFieldBody,
				BodyType: BodyTypeStructured,
				Required: false,
			},
		},
		{
			name:      "structured body type",
			fieldName: goFieldBody,
			tagValue:  bodyStructured,
			want: &BodyMetadata{
				MapKey:   goFieldBody,
				BodyType: BodyTypeStructured,
				Required: false,
			},
		},
		{
			name:      "file body type",
			fieldName: goFieldFile,
			tagValue:  bodyFile,
			want: &BodyMetadata{
				MapKey:   goFieldFile,
				BodyType: BodyTypeFile,
				Required: false,
			},
		},
		{
			name:      "multipart body type",
			fieldName: goFieldForm,
			tagValue:  bodyMultipart,
			want: &BodyMetadata{
				MapKey:   goFieldForm,
				BodyType: BodyTypeMultipart,
				Required: false,
			},
		},
		{
			name:      "structured with required true",
			fieldName: goFieldBody,
			tagValue:  "structured,required=true",
			want: &BodyMetadata{
				MapKey:   goFieldBody,
				BodyType: BodyTypeStructured,
				Required: true,
			},
		},
		{
			name:      "structured with required flag form",
			fieldName: goFieldBody,
			tagValue:  "structured,required",
			want: &BodyMetadata{
				MapKey:   goFieldBody,
				BodyType: BodyTypeStructured,
				Required: true,
			},
		},
		{
			name:      "file with required true",
			fieldName: goFieldFile,
			tagValue:  "file,required=true",
			want: &BodyMetadata{
				MapKey:   goFieldFile,
				BodyType: BodyTypeFile,
				Required: true,
			},
		},
		{
			name:      "multipart with required true",
			fieldName: goFieldForm,
			tagValue:  "multipart,required=true",
			want: &BodyMetadata{
				MapKey:   goFieldForm,
				BodyType: BodyTypeMultipart,
				Required: true,
			},
		},
		{
			name:      "structured with required false",
			fieldName: goFieldBody,
			tagValue:  "structured,required=false",
			want: &BodyMetadata{
				MapKey:   goFieldBody,
				BodyType: BodyTypeStructured,
				Required: false,
			},
		},
		{
			name:        msgInvalidBodyType,
			fieldName:   goFieldBody,
			tagValue:    "invalid",
			wantErr:     true,
			errContains: msgInvalidBodyType,
		},
		{
			name:        "invalid body type with error message",
			fieldName:   goFieldBody,
			tagValue:    tagJSON,
			wantErr:     true,
			errContains: "must be 'structured', 'file', or 'multipart'",
		},
		{
			name:      "case sensitive body type",
			fieldName: goFieldBody,
			tagValue:  "Structured",
			wantErr:   true,
		},
		{
			name:      "file with empty required value",
			fieldName: goFieldFile,
			tagValue:  "file,required",
			want: &BodyMetadata{
				MapKey:   goFieldFile,
				BodyType: BodyTypeFile,
				Required: true,
			},
		},
		{
			name:      "multipart with empty required value",
			fieldName: goFieldForm,
			tagValue:  "multipart,required",
			want: &BodyMetadata{
				MapKey:   goFieldForm,
				BodyType: BodyTypeMultipart,
				Required: true,
			},
		},
		{
			name:      "long field name",
			fieldName: "VeryLongFieldNameForBody",
			tagValue:  bodyStructured,
			want: &BodyMetadata{
				MapKey:   "VeryLongFieldNameForBody",
				BodyType: BodyTypeStructured,
				Required: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := reflect.StructField{
				Name: tt.fieldName,
				Type: reflect.TypeFor[string](),
			}

			result, err := ParseBodyTag(field, 0, tt.tagValue)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}

				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)

			meta, ok := result.(*BodyMetadata)
			require.True(t, ok, "result should be *BodyMetadata")

			assert.Equal(t, tt.want.MapKey, meta.MapKey)
			assert.Equal(t, tt.want.BodyType, meta.BodyType)
			assert.Equal(t, tt.want.Required, meta.Required)
		})
	}
}

func TestParseBodyTag_DifferentFieldTypes(t *testing.T) {
	tests := []struct {
		name      string
		fieldName string
		fieldType reflect.Type
		tagValue  string
		want      *BodyMetadata
	}{
		{
			name:      "string field",
			fieldName: goFieldBody,
			fieldType: reflect.TypeFor[string](),
			tagValue:  bodyStructured,
			want: &BodyMetadata{
				MapKey:   goFieldBody,
				BodyType: BodyTypeStructured,
				Required: false,
			},
		},
		{
			name:      "byte slice field for file",
			fieldName: goFieldFile,
			fieldType: reflect.TypeFor[[]byte](),
			tagValue:  bodyFile,
			want: &BodyMetadata{
				MapKey:   goFieldFile,
				BodyType: BodyTypeFile,
				Required: false,
			},
		},
		{
			name:      "struct field for structured",
			fieldName: goFieldBody,
			fieldType: reflect.TypeFor[struct{}](),
			tagValue:  bodyStructured,
			want: &BodyMetadata{
				MapKey:   goFieldBody,
				BodyType: BodyTypeStructured,
				Required: false,
			},
		},
		{
			name:      "interface field",
			fieldName: goFieldBody,
			fieldType: reflect.TypeFor[any](),
			tagValue:  bodyStructured,
			want: &BodyMetadata{
				MapKey:   goFieldBody,
				BodyType: BodyTypeStructured,
				Required: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := reflect.StructField{
				Name: tt.fieldName,
				Type: tt.fieldType,
			}

			result, err := ParseBodyTag(field, 0, tt.tagValue)
			require.NoError(t, err)
			require.NotNil(t, result)

			meta, ok := result.(*BodyMetadata)
			require.True(t, ok)

			assert.Equal(t, tt.want.MapKey, meta.MapKey)
			assert.Equal(t, tt.want.BodyType, meta.BodyType)
			assert.Equal(t, tt.want.Required, meta.Required)
		})
	}
}

func TestParseBodyTag_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		fieldName   string
		tagValue    string
		want        *BodyMetadata
		wantErr     bool
		errContains string
	}{
		{
			name:      "empty string tag",
			fieldName: goFieldBody,
			tagValue:  "",
			want: &BodyMetadata{
				MapKey:   goFieldBody,
				BodyType: BodyTypeStructured,
				Required: false,
			},
		},
		{
			name:      "whitespace only tag",
			fieldName: goFieldBody,
			tagValue:  "   ",
			want: &BodyMetadata{
				MapKey:   goFieldBody,
				BodyType: BodyTypeStructured,
				Required: false,
			},
		},
		{
			name:      "multiple required flags",
			fieldName: goFieldBody,
			tagValue:  "structured,required,required",
			want: &BodyMetadata{
				MapKey:   goFieldBody,
				BodyType: BodyTypeStructured,
				Required: true,
			},
		},
		{
			name:      "required false overrides flag",
			fieldName: goFieldBody,
			tagValue:  "structured,required,required=false",
			want: &BodyMetadata{
				MapKey:   goFieldBody,
				BodyType: BodyTypeStructured,
				Required: false,
			},
		},
		{
			name:        "invalid body type with options",
			fieldName:   goFieldBody,
			tagValue:    "invalid,required=true",
			wantErr:     true,
			errContains: msgInvalidBodyType,
		},
		{
			name:      "structured with unknown option",
			fieldName: goFieldBody,
			tagValue:  "structured,unknown=value",
			want: &BodyMetadata{
				MapKey:   goFieldBody,
				BodyType: BodyTypeStructured,
				Required: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := reflect.StructField{
				Name: tt.fieldName,
				Type: reflect.TypeFor[string](),
			}

			result, err := ParseBodyTag(field, 0, tt.tagValue)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}

				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)

			meta, ok := result.(*BodyMetadata)
			require.True(t, ok)

			assert.Equal(t, tt.want.MapKey, meta.MapKey)
			assert.Equal(t, tt.want.BodyType, meta.BodyType)
			assert.Equal(t, tt.want.Required, meta.Required)
		})
	}
}

func TestParseBodyType(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		want        BodyType
		wantErr     bool
		errContains string
	}{
		{
			name:    "empty string defaults to structured",
			input:   "",
			want:    BodyTypeStructured,
			wantErr: false,
		},
		{
			name:    "structured type",
			input:   bodyStructured,
			want:    BodyTypeStructured,
			wantErr: false,
		},
		{
			name:    "file type",
			input:   bodyFile,
			want:    BodyTypeFile,
			wantErr: false,
		},
		{
			name:    "multipart type",
			input:   bodyMultipart,
			want:    BodyTypeMultipart,
			wantErr: false,
		},
		{
			name:        "invalid type",
			input:       "invalid",
			want:        "",
			wantErr:     true,
			errContains: msgInvalidBodyType,
		},
		{
			name:        "case sensitive - uppercase",
			input:       "STRUCTURED",
			want:        "",
			wantErr:     true,
			errContains: msgInvalidBodyType,
		},
		{
			name:        "case sensitive - mixed case",
			input:       "Structured",
			want:        "",
			wantErr:     true,
			errContains: msgInvalidBodyType,
		},
		{
			name:        "partial match",
			input:       "struct",
			want:        "",
			wantErr:     true,
			errContains: msgInvalidBodyType,
		},
		{
			name:        "json type (invalid)",
			input:       tagJSON,
			want:        "",
			wantErr:     true,
			errContains: "must be 'structured', 'file', or 'multipart'",
		},
		{
			name:        "xml type (invalid)",
			input:       "xml",
			want:        "",
			wantErr:     true,
			errContains: msgInvalidBodyType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseBodyType(tt.input)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestBodyTypeConstants(t *testing.T) {
	tests := []struct {
		name     string
		bodyType BodyType
		wantStr  string
	}{
		{
			name:     "BodyTypeStructured",
			bodyType: BodyTypeStructured,
			wantStr:  bodyStructured,
		},
		{
			name:     "BodyTypeFile",
			bodyType: BodyTypeFile,
			wantStr:  bodyFile,
		},
		{
			name:     "BodyTypeMultipart",
			bodyType: BodyTypeMultipart,
			wantStr:  bodyMultipart,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantStr, string(tt.bodyType))
		})
	}
}
