package tagparser

import (
	"errors"
	"reflect"
	"testing"
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

			checkError(t, test.tag, err, test.error)
			if err != nil {
				return
			}

			checkName(t, test.tag, tag.Name, test.name)
			checkOptions(t, test.tag, tag.Options, test.opts)
		})
	}
}

func checkError(t *testing.T, tag string, err error, wantErr string) {
	t.Helper()
	if err != nil {
		ae := err.Error()
		if wantErr == "" {
			t.Errorf("** Parse(%q) error %q, wanted no error", tag, ae)
		} else if ae != wantErr {
			t.Errorf("** Parse(%q) error %q, wanted error %q", tag, ae, wantErr)
		}
	} else if wantErr != "" {
		t.Errorf("** Parse(%q) no error, wanted error %q", tag, wantErr)
	}
}

func checkName(t *testing.T, tag string, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("** Parse(%q) name = %q, wanted %q", tag, got, want)
	}
}

func checkOptions(t *testing.T, tag string, got, want map[string]string) {
	t.Helper()
	if reflect.DeepEqual(got, want) {
		return
	}

	for k, ev := range want {
		av, ok := got[k]
		if !ok {
			t.Errorf("** Parse(%q) missing option %q = %q", tag, k, ev)
		} else if av != ev {
			t.Fatalf("** Parse(%q) option %q = %q, wanted %q", tag, k, av, ev)
		}
	}
	for k, av := range got {
		_, ok := want[k]
		if !ok {
			t.Errorf("** Parse(%q) extra option %q = %q", tag, k, av)
		}
	}
}

func TestParse_duplicate(t *testing.T) {
	_, err := Parse(`foo:bar,foo:boz`)
	if !errors.Is(err, ErrDuplicateKey) {
		t.Errorf("** err = %v, wanted %v", err, ErrDuplicateKey)
	}
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
	if err.Error() != expErr {
		t.Errorf("** error = %v, wanted %v", err, expErr)
	}
	var errType *Error
	if !errors.As(err, &errType) || !errors.Is(errType.Cause, errSimulated) {
		t.Errorf("** Cause = %v, wanted %v", err, expErr)
	}
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
	if err.Error() != expErr {
		t.Errorf("** error = %v, wanted %v", err, expErr)
	}
	var errType *Error
	if !errors.As(err, &errType) || !errors.Is(errType.Cause, errSimulated) {
		t.Errorf("** Cause = %v, wanted %v", err, expErr)
	}
}

func BenchmarkParseFunc(t *testing.B) {
	slice := make([]string, 0, 20)
	for i := 0; i < t.N; i++ {
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
