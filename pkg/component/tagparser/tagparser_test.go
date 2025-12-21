package tagparser

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type M = map[string]string

func TestParseWithName(t *testing.T) {
	tests := []struct {
		testName string
		tag      string
		name     string
		opts     map[string]string
		error    string
	}{
		{`empty`, ``, "", nil, ``},

		{`simple 1`, `alfa`, `alfa`, nil, ``},
		{`simple 2`, `alfa,bravo`, `alfa`, M{"bravo": ""}, ``},

		{`quoted key 1`, `'alfa,bravo'`, `alfa,bravo`, nil, ``},
		{`quoted key 2`, `'alfa=bravo'`, `alfa=bravo`, nil, ``},
		{`quoted key 3`, `'alfa\=bravo'`, `alfa=bravo`, nil, ``},
		{`quoted key 4`, "'alfa=bravo'", `alfa=bravo`, nil, ``},

		{`escaped key 1`, `\ =alfa`, "", M{" ": "alfa"}, ""},
		{`escaped key 2`, `' '=alfa`, "", M{" ": "alfa"}, ""},

		{`no name 1`, `,alfa`, "", M{"alfa": ""}, ``},
		{`no name 2`, `,alfa,bravo`, "", M{"alfa": "", "bravo": ""}, ``},
		{`key with empty value`, `alfa=`, "", M{"alfa": ""}, ``},
		{`key with empty quoted value`, `alfa=''`, "", M{"alfa": ""}, ``},
		{`key-value 1`, `alfa=bravo`, "", M{"alfa": "bravo"}, ``},
		{`key-value 2`, `alfa=bravo,charlie`, "", M{"alfa": "bravo", "charlie": ""}, ``},
		{`key-value 3`, `alfa=bravo,charlie=delta`, "", M{"alfa": "bravo", "charlie": "delta"}, ``},

		{`whitespace 1`, `  alfa  `, "alfa", nil, ``},
		{`whitespace 2`, ` alfa ,  bravo  `, "alfa", M{"bravo": ""}, ``},
		{`whitespace 3`, ` alfa, charlie= delta `, "alfa", M{"charlie": "delta"}, ``},

		{`skipped key`, `alfa,,charlie`, "alfa", M{"charlie": ""}, ``},

		{`quoted value 1`, `alfa='bravo,charlie'`, "", M{"alfa": "bravo,charlie"}, ``},
		{`quoted value 2`, `alfa='bravo,charlie',delta`, "", M{"alfa": "bravo,charlie", "delta": ""}, ``},
		{`quoted value 3`, `alfa='bravo=charlie',delta`, "", M{"alfa": "bravo=charlie", "delta": ""}, ``},
		{`quoted value 4`, `alfa='d\'Elta', bravo=charlie`, "", M{"alfa": "d'Elta", "bravo": "charlie"}, ``},

		{`disallowed quote in the middle 1`, `alfa=bravo', charlie 'delta`, "", nil, `quotes must enclose the entire value (at 11)`},
		{`disallowed quote in the middle 2`, `alfa='bravo 'charlie' delta'`, "", nil, `quotes must enclose the entire value (at 13)`},
		{`disallowed quote in the middle of name`, `bravo' charlie'`, "", nil, `quotes must enclose the entire value (at 6)`},
		{`disallowed quote in the middle of name with options`, `alfa,bravo' charlie'`, "alfa", nil, `quotes must enclose the entire value (at 11)`},
		{`disallowed quote in the middle of key`, `bravo' charlie'= delta`, "", nil, `quotes must enclose the entire value (at 6)`},

		{`disallowed vmihailenco-style parenthesized value`, `alfa=bravo('charlie', 'delta')`, "", nil, `quotes must enclose the entire value (at 12)`},

		{`malformed empty key 1`, `alfa,=bravo`, "alfa", nil, `empty key (at 6)`},
		{`malformed empty key 2`, `,=alfa`, "", nil, `empty key (at 2)`},
		{`malformed empty key 3`, `''=alfa`, "", nil, `empty key (at 1)`},
		{`malformed empty key 4`, ` '' =alfa`, "", nil, `empty key (at 1)`},
		{`duplicate key last wins`, `alfa,bravo=charlie,bravo=delta`, "alfa", M{"bravo": "delta"}, ``},
		{`duplicate key first item last wins`, `foo=bar,foo=boz`, "", M{"foo": "boz"}, ``},
		{`malformed unterminated quote 1`, `alfa,'bravo=charlie`, "alfa", M{"bravo=charlie": ""}, `unterminated quote (at 6)`},
		{`malformed unterminated quote 2`, `alfa,bravo='charlie`, "alfa", M{"bravo": "charlie"}, `unterminated quote (at 12)`},
		{`malformed unterminated quote 3`, `'alfa`, "alfa", nil, `unterminated quote (at 1)`},
		{`malformed escape 1`, `a\lfa`, "alfa", nil, `invalid escape character (at 3)`},
		{`malformed escape 2`, `al\`, "al", nil, `unterminated escape sequence (at 3)`},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			tag, err := ParseWithName(test.tag)

			if test.error != "" {
				require.Error(t, err)
				assert.Equal(t, test.error, err.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.name, tag.Name)
				if test.opts == nil {
					assert.Empty(t, tag.Options)
				} else {
					assert.Equal(t, test.opts, tag.Options)
				}
			}
		})
	}
}

func TestParseWithName_Duplicate(t *testing.T) {
	tag, err := ParseWithName(`foo=bar,foo=boz`)
	require.NoError(t, err)
	assert.Equal(t, "", tag.Name)
	assert.Equal(t, M{"foo": "boz"}, tag.Options) // last value wins
}

func TestParseWithName_Unquoting(t *testing.T) {
	// Test that Go struct tag quoted inputs are automatically unquoted
	tag, err := ParseWithName(`"name=value,other=key"`)
	require.NoError(t, err)
	assert.Equal(t, "", tag.Name) // No name part, just options
	assert.Equal(t, M{"name": "value", "other": "key"}, tag.Options)

	// Test that unquoted inputs work the same way
	tag2, err := ParseWithName(`name=value,other=key`)
	require.NoError(t, err)
	assert.Equal(t, "", tag2.Name)
	assert.Equal(t, M{"name": "value", "other": "key"}, tag2.Options)
}

func TestParse_OptionsMode(t *testing.T) {
	// Test Parse (options mode) - all items are options
	tag, err := Parse(`foo,bar=baz`)
	require.NoError(t, err)
	assert.Equal(t, "", tag.Name)
	assert.Equal(t, M{"foo": "", "bar": "baz"}, tag.Options)

	// First item with equals is also an option
	tag2, err := Parse(`foo=bar,baz`)
	require.NoError(t, err)
	assert.Equal(t, "", tag2.Name)
	assert.Equal(t, M{"foo": "bar", "baz": ""}, tag2.Options)

	// All items treated as options
	tag3, err := Parse(`required,email,min=5`)
	require.NoError(t, err)
	assert.Equal(t, "", tag3.Name)
	assert.Equal(t, M{"required": "", "email": "", "min": "5"}, tag3.Options)
}

func TestParseWithName_NameMode(t *testing.T) {
	// Test ParseWithName (name mode) - first item without equals is name
	tag, err := ParseWithName(`foo,bar=baz`)
	require.NoError(t, err)
	assert.Equal(t, "foo", tag.Name)
	assert.Equal(t, M{"bar": "baz"}, tag.Options)

	// First item with equals is treated as option
	tag2, err := ParseWithName(`foo=bar,baz`)
	require.NoError(t, err)
	assert.Equal(t, "", tag2.Name)
	assert.Equal(t, M{"foo": "bar", "baz": ""}, tag2.Options)
}

var errSimulated = errors.New("simulated error")

func TestParseFuncWithName_CustomErrorInName(t *testing.T) {
	const tag = `foo,bar=boz`
	const expErr = `simulated error (at 1)`
	err := ParseFuncWithName(tag, func(key, value string) error {
		if key == "" {
			return errSimulated
		}

		return nil
	})
	require.Error(t, err)
	assert.Equal(t, expErr, err.Error())

	var errType *Error
	require.True(t, errors.As(err, &errType))
	assert.True(t, errors.Is(errType.Cause, errSimulated))
}

func TestParseFuncWithName_CustomErrorInKey(t *testing.T) {
	const tag = `foo,bar=boz`
	const expErr = `bar: simulated error (at 5)`
	err := ParseFuncWithName(tag, func(key, value string) error {
		if key == "bar" {
			return errSimulated
		}

		return nil
	})
	require.Error(t, err)
	assert.Equal(t, expErr, err.Error())

	var errType *Error
	require.True(t, errors.As(err, &errType))
	assert.True(t, errors.Is(errType.Cause, errSimulated))
}

func TestParseFunc_OptionsMode(t *testing.T) {
	// Test ParseFunc (options mode) - all items are options, no empty keys
	opts := make(M)
	err := ParseFunc(`foo,bar=baz`, func(key, value string) error {
		opts[key] = value

		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, M{"foo": "", "bar": "baz"}, opts)
}

func BenchmarkParseFunc(t *testing.B) {
	slice := make([]string, 0, 20)
	for t.Loop() {
		slice = slice[:0]
		err := ParseFunc(`foo,bar=boz,fubar,bar=zob,oof`, func(key, value string) error {
			slice = append(slice, key, value)

			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}

func BenchmarkParseFuncWithName(t *testing.B) {
	slice := make([]string, 0, 20)
	for t.Loop() {
		slice = slice[:0]
		err := ParseFuncWithName(`foo,bar=boz,fubar,bar=zob,oof`, func(key, value string) error {
			slice = append(slice, key, value)

			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}
