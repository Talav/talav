package metadata

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	goFieldAge            = "Age"
	goFieldCode           = "Code"
	goFieldCreditCard     = "CreditCard"
	goFieldEmail          = "Email"
	goFieldGeneric        = "Field"
	goFieldName           = "Name"
	goFieldPrice          = "Price"
	goFieldStatus         = "Status"
	valueActive           = "active"
	keyBillingAddress     = "billing_address"
	caseEmptyTag          = "empty tag"
	valueInactive         = "inactive"
	msgInvalidSyntax      = "invalid syntax"
	caseInvalidTagParsing = "invalid tag parsing"
	valuePending          = "pending"
	valuePlain            = "value"
	keyCustomExtension    = "x-custom"
)

func floatPtr(f float64) *float64 {
	return &f
}

func boolPtr(b bool) *bool {
	return &b
}

//nolint:maintidx // Table-driven test with many cases - acceptable complexity for test function
func TestParseValidateTag(t *testing.T) {
	tests := []struct {
		name        string
		fieldName   string
		tagValue    string
		want        *ValidateMetadata
		wantErr     bool
		errContains string
	}{
		{
			name:      caseEmptyTag,
			fieldName: goFieldGeneric,
			tagValue:  "",
			want:      &ValidateMetadata{},
		},
		{
			name:      "required flag",
			fieldName: "ID",
			tagValue:  "required",
			want: &ValidateMetadata{
				Required: boolPtr(true),
			},
		},
		{
			name:      "required with explicit false",
			fieldName: "ID",
			tagValue:  "required=false",
			want: &ValidateMetadata{
				Required: boolPtr(false),
			},
		},
		{
			name:      "min constraint",
			fieldName: goFieldAge,
			tagValue:  "min=18",
			want: &ValidateMetadata{
				Minimum: floatPtr(18),
			},
		},
		{
			name:      "max constraint",
			fieldName: goFieldAge,
			tagValue:  "max=100",
			want: &ValidateMetadata{
				Maximum: floatPtr(100),
			},
		},
		{
			name:      "min and max together",
			fieldName: goFieldAge,
			tagValue:  "min=18,max=100",
			want: &ValidateMetadata{
				Minimum: floatPtr(18),
				Maximum: floatPtr(100),
			},
		},
		{
			name:      "gte constraint",
			fieldName: "Score",
			tagValue:  "gte=0",
			want: &ValidateMetadata{
				Minimum: floatPtr(0),
			},
		},
		{
			name:      "lte constraint",
			fieldName: "Score",
			tagValue:  "lte=100",
			want: &ValidateMetadata{
				Maximum: floatPtr(100),
			},
		},
		{
			name:      "gt constraint",
			fieldName: goFieldPrice,
			tagValue:  "gt=0",
			want: &ValidateMetadata{
				ExclusiveMinimum: floatPtr(0),
			},
		},
		{
			name:      "lt constraint",
			fieldName: goFieldPrice,
			tagValue:  "lt=1000",
			want: &ValidateMetadata{
				ExclusiveMaximum: floatPtr(1000),
			},
		},
		{
			name:      "multiple_of constraint",
			fieldName: "Quantity",
			tagValue:  "multiple_of=5",
			want: &ValidateMetadata{
				MultipleOf: floatPtr(5),
			},
		},
		{
			name:      "len constraint",
			fieldName: goFieldCode,
			tagValue:  "len=6",
			want: &ValidateMetadata{
				Minimum: floatPtr(6),
				Maximum: floatPtr(6),
			},
		},
		{
			name:      "email format",
			fieldName: goFieldEmail,
			tagValue:  "email",
			want: &ValidateMetadata{
				Format: "email",
			},
		},
		{
			name:      "url format",
			fieldName: "URL",
			tagValue:  "url",
			want: &ValidateMetadata{
				Format: "uri",
			},
		},
		{
			name:      "alpha pattern",
			fieldName: goFieldName,
			tagValue:  "alpha",
			want: &ValidateMetadata{
				Pattern: "^[a-zA-Z]+$",
			},
		},
		{
			name:      "alphanum pattern",
			fieldName: "Username",
			tagValue:  "alphanum",
			want: &ValidateMetadata{
				Pattern: "^[a-zA-Z0-9]+$",
			},
		},
		{
			name:      "alphaunicode pattern",
			fieldName: goFieldName,
			tagValue:  "alphaunicode",
			want: &ValidateMetadata{
				Pattern: "^[\\p{L}]+$",
			},
		},
		{
			name:      "alphanumunicode pattern",
			fieldName: "Username",
			tagValue:  "alphanumunicode",
			want: &ValidateMetadata{
				Pattern: "^[\\p{L}\\p{N}]+$",
			},
		},
		{
			name:      "pattern constraint",
			fieldName: goFieldCode,
			tagValue:  "pattern=^[A-Z0-9]+$",
			want: &ValidateMetadata{
				Pattern: "^[A-Z0-9]+$",
			},
		},
		{
			name:      "pattern with quoted value",
			fieldName: goFieldCode,
			tagValue:  "pattern='^[a-z]+$'",
			want: &ValidateMetadata{
				Pattern: "^[a-z]+$",
			},
		},
		{
			name:      "oneof enum",
			fieldName: goFieldStatus,
			tagValue:  "oneof=active inactive pending",
			want: &ValidateMetadata{
				Enum: []any{valueActive, valueInactive, valuePending},
			},
		},
		{
			name:      "oneof enum with single value",
			fieldName: "Type",
			tagValue:  "oneof=admin",
			want: &ValidateMetadata{
				Enum: []any{"admin"},
			},
		},
		{
			name:      "oneof enum with extra spaces",
			fieldName: goFieldStatus,
			tagValue:  "oneof=  active  inactive  pending  ",
			want: &ValidateMetadata{
				Enum: []any{valueActive, valueInactive, valuePending},
			},
		},
		{
			name:        caseInvalidTagParsing,
			fieldName:   goFieldGeneric,
			tagValue:    "required,'unclosed quote",
			wantErr:     true,
			errContains: "failed to parse validate tag",
		},
		{
			name:        "invalid min value",
			fieldName:   goFieldGeneric,
			tagValue:    "min=invalid",
			wantErr:     true,
			errContains: "invalid min value",
		},
		{
			name:        "invalid len value",
			fieldName:   goFieldGeneric,
			tagValue:    "len=invalid",
			wantErr:     true,
			errContains: "invalid len value",
		},
		{
			name:        "empty oneof",
			fieldName:   goFieldGeneric,
			tagValue:    "oneof=",
			wantErr:     true,
			errContains: "oneof requires at least one value",
		},
		{
			name:      "oneof with empty values filtered",
			fieldName: goFieldStatus,
			tagValue:  "oneof=active  pending",
			want: &ValidateMetadata{
				Enum: []any{valueActive, valuePending},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := reflect.StructField{
				Name: tt.fieldName,
			}

			result, err := ParseValidateTag(field, 0, tt.tagValue)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}

				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)

			vm, ok := result.(*ValidateMetadata)
			require.True(t, ok, "result should be *ValidateMetadata")

			assert.Equal(t, tt.want.Required, vm.Required, "Required mismatch")
			if tt.want.Minimum == nil {
				assert.Nil(t, vm.Minimum, "Minimum should be nil")
			} else {
				require.NotNil(t, vm.Minimum, "Minimum should not be nil")
				assert.Equal(t, *tt.want.Minimum, *vm.Minimum, "Minimum mismatch")
			}
			if tt.want.ExclusiveMinimum == nil {
				assert.Nil(t, vm.ExclusiveMinimum, "ExclusiveMinimum should be nil")
			} else {
				require.NotNil(t, vm.ExclusiveMinimum, "ExclusiveMinimum should not be nil")
				assert.Equal(t, *tt.want.ExclusiveMinimum, *vm.ExclusiveMinimum, "ExclusiveMinimum mismatch")
			}
			if tt.want.Maximum == nil {
				assert.Nil(t, vm.Maximum, "Maximum should be nil")
			} else {
				require.NotNil(t, vm.Maximum, "Maximum should not be nil")
				assert.Equal(t, *tt.want.Maximum, *vm.Maximum, "Maximum mismatch")
			}
			if tt.want.ExclusiveMaximum == nil {
				assert.Nil(t, vm.ExclusiveMaximum, "ExclusiveMaximum should be nil")
			} else {
				require.NotNil(t, vm.ExclusiveMaximum, "ExclusiveMaximum should not be nil")
				assert.Equal(t, *tt.want.ExclusiveMaximum, *vm.ExclusiveMaximum, "ExclusiveMaximum mismatch")
			}
			if tt.want.MultipleOf == nil {
				assert.Nil(t, vm.MultipleOf, "MultipleOf should be nil")
			} else {
				require.NotNil(t, vm.MultipleOf, "MultipleOf should not be nil")
				assert.Equal(t, *tt.want.MultipleOf, *vm.MultipleOf, "MultipleOf mismatch")
			}
			assert.Equal(t, tt.want.Pattern, vm.Pattern, "Pattern mismatch")
			assert.Equal(t, tt.want.Format, vm.Format, "Format mismatch")
			assert.Equal(t, tt.want.Enum, vm.Enum, "Enum mismatch")
		})
	}
}

func TestParseValidateTag_RealWorldScenarios(t *testing.T) {
	t.Run("user registration email", func(t *testing.T) {
		field := reflect.StructField{
			Name: goFieldEmail,
		}

		tagValue := "required,email,max=255"

		result, err := ParseValidateTag(field, 0, tagValue)
		require.NoError(t, err)

		vm, ok := result.(*ValidateMetadata)
		require.True(t, ok)
		assert.Equal(t, boolPtr(true), vm.Required)
		assert.Equal(t, "email", vm.Format)
		assert.Equal(t, floatPtr(255), vm.Maximum)
	})

	t.Run("password with length and pattern", func(t *testing.T) {
		field := reflect.StructField{
			Name: "Password",
		}

		tagValue := "required,min=8,max=128,alphanum"

		result, err := ParseValidateTag(field, 0, tagValue)
		require.NoError(t, err)

		vm, ok := result.(*ValidateMetadata)
		require.True(t, ok)
		assert.Equal(t, boolPtr(true), vm.Required)
		assert.Equal(t, floatPtr(8), vm.Minimum)
		assert.Equal(t, floatPtr(128), vm.Maximum)
		assert.Equal(t, "^[a-zA-Z0-9]+$", vm.Pattern)
	})

	t.Run("age with range", func(t *testing.T) {
		field := reflect.StructField{
			Name: goFieldAge,
		}

		tagValue := "gte=18,lte=120"

		result, err := ParseValidateTag(field, 0, tagValue)
		require.NoError(t, err)

		vm, ok := result.(*ValidateMetadata)
		require.True(t, ok)
		assert.Equal(t, floatPtr(18), vm.Minimum)
		assert.Equal(t, floatPtr(120), vm.Maximum)
	})

	t.Run("status enum", func(t *testing.T) {
		field := reflect.StructField{
			Name: goFieldStatus,
		}

		tagValue := "required,oneof=active inactive suspended deleted"

		result, err := ParseValidateTag(field, 0, tagValue)
		require.NoError(t, err)

		vm, ok := result.(*ValidateMetadata)
		require.True(t, ok)
		assert.Equal(t, boolPtr(true), vm.Required)
		assert.Equal(t, []any{valueActive, valueInactive, "suspended", "deleted"}, vm.Enum)
	})

	t.Run("price with exclusive bounds", func(t *testing.T) {
		field := reflect.StructField{
			Name: goFieldPrice,
		}

		tagValue := "gt=0,lt=10000"

		result, err := ParseValidateTag(field, 0, tagValue)
		require.NoError(t, err)

		vm, ok := result.(*ValidateMetadata)
		require.True(t, ok)
		assert.Equal(t, floatPtr(0), vm.ExclusiveMinimum)
		assert.Equal(t, floatPtr(10000), vm.ExclusiveMaximum)
	})

	t.Run("product code with exact length and pattern", func(t *testing.T) {
		field := reflect.StructField{
			Name: "ProductCode",
		}

		tagValue := "required,len=10,pattern=^[A-Z0-9]+$"

		result, err := ParseValidateTag(field, 0, tagValue)
		require.NoError(t, err)

		vm, ok := result.(*ValidateMetadata)
		require.True(t, ok)
		assert.Equal(t, boolPtr(true), vm.Required)
		assert.Equal(t, floatPtr(10), vm.Minimum)
		assert.Equal(t, floatPtr(10), vm.Maximum)
		assert.Equal(t, "^[A-Z0-9]+$", vm.Pattern)
	})
}
