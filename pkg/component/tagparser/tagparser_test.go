package tagparser

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type M = map[string]string

func TestParse(t *testing.T) {
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
		{`quoted key 2`, `'alfa:bravo'`, `alfa:bravo`, nil, ``},
		{`quoted key 3`, `'alfa\:bravo'`, `alfa:bravo`, nil, ``},
		{`quoted key 4`, "'alfa:bravo'", `alfa:bravo`, nil, ``},

		{`escaped key 1`, `\ :alfa`, "", M{" ": "alfa"}, ""},
		{`escaped key 2`, `' ':alfa`, "", M{" ": "alfa"}, ""},

		{`no name 1`, `,alfa`, "", M{"alfa": ""}, ``},
		{`no name 2`, `,alfa,bravo`, "", M{"alfa": "", "bravo": ""}, ``},
		{`key with empty value`, `alfa:`, "", M{"alfa": ""}, ``},
		{`key with empty quoted value`, `alfa:''`, "", M{"alfa": ""}, ``},
		{`key-value 1`, `alfa:bravo`, "", M{"alfa": "bravo"}, ``},
		{`key-value 2`, `alfa:bravo,charlie`, "", M{"alfa": "bravo", "charlie": ""}, ``},
		{`key-value 3`, `alfa:bravo,charlie:delta`, "", M{"alfa": "bravo", "charlie": "delta"}, ``},

		{`whitespace 1`, `  alfa  `, "alfa", nil, ``},
		{`whitespace 2`, ` alfa ,  bravo  `, "alfa", M{"bravo": ""}, ``},
		{`whitespace 3`, ` alfa, charlie: delta `, "alfa", M{"charlie": "delta"}, ``},

		{`skipped key`, `alfa,,charlie`, "alfa", M{"charlie": ""}, ``},

		{`quoted value 1`, `alfa:'bravo,charlie'`, "", M{"alfa": "bravo,charlie"}, ``},
		{`quoted value 2`, `alfa:'bravo,charlie',delta`, "", M{"alfa": "bravo,charlie", "delta": ""}, ``},
		{`quoted value 3`, `alfa:'bravo:charlie',delta`, "", M{"alfa": "bravo:charlie", "delta": ""}, ``},
		{`quoted value 4`, `alfa:'d\'Elta', bravo:charlie`, "", M{"alfa": "d'Elta", "bravo": "charlie"}, ``},

		{`disallowed quote in the middle 1`, `alfa:bravo', charlie 'delta`, "", nil, `quotes must enclose the entire value (at 11)`},
		{`disallowed quote in the middle 2`, `alfa:'bravo 'charlie' delta'`, "", nil, `quotes must enclose the entire value (at 13)`},
		{`disallowed quote in the middle of name`, `bravo' charlie'`, "", nil, `quotes must enclose the entire value (at 6)`},
		{`disallowed quote in the middle of name with options`, `alfa,bravo' charlie'`, "alfa", nil, `quotes must enclose the entire value (at 11)`},
		{`disallowed quote in the middle of key`, `bravo' charlie': delta`, "", nil, `quotes must enclose the entire value (at 6)`},

		{`disallowed vmihailenco-style parenthesized value`, `alfa:bravo('charlie', 'delta')`, "", nil, `quotes must enclose the entire value (at 12)`},

		{`malformed empty key 1`, `alfa,:bravo`, "alfa", nil, `empty key (at 6)`},
		{`malformed empty key 2`, `,:alfa`, "", nil, `empty key (at 2)`},
		{`malformed empty key 3`, `'':alfa`, "", nil, `empty key (at 1)`},
		{`malformed empty key 4`, ` '' :alfa`, "", nil, `empty key (at 1)`},
		{`malformed duplicate key`, `alfa,bravo:charlie,bravo:delta`, "alfa", M{"bravo": "charlie"}, `bravo: duplicate option key (at 20)`},
		{`malformed duplicate key first item`, `foo:bar,foo:boz`, "", M{"foo": "bar"}, `foo: duplicate option key (at 9)`},
		{`malformed unterminated quote 1`, `alfa,'bravo:charlie`, "alfa", M{"bravo:charlie": ""}, `unterminated quote (at 6)`},
		{`malformed unterminated quote 2`, `alfa,bravo:'charlie`, "alfa", M{"bravo": "charlie"}, `unterminated quote (at 12)`},
		{`malformed unterminated quote 3`, `'alfa`, "alfa", nil, `unterminated quote (at 1)`},
		{`malformed escape 1`, `a\lfa`, "alfa", nil, `invalid escape character (at 3)`},
		{`malformed escape 2`, `al\`, "al", nil, `unterminated escape sequence (at 3)`},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			tag, err := Parse(test.tag)

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

func TestParse_duplicate(t *testing.T) {
	_, err := Parse(`foo:bar,foo:boz`)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrDuplicateKey))
}

var errSimulated = errors.New("simulated error")

func TestParseFunc_custom_error_in_name(t *testing.T) {
	const tag = `foo,bar:boz`
	const expErr = `simulated error (at 1)`
	err := ParseFunc(tag, func(key, value string) error {
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

func TestParseFunc_custom_error_in_key(t *testing.T) {
	const tag = `foo,bar:boz`
	const expErr = `bar: simulated error (at 5)`
	err := ParseFunc(tag, func(key, value string) error {
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

func BenchmarkParseFunc(t *testing.B) {
	slice := make([]string, 0, 20)
	for t.Loop() {
		slice = slice[:0]
		err := ParseFunc(`foo,bar:boz,fubar,bar:zob,oof`, func(key, value string) error {
			slice = append(slice, key, value)

			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}
