package tagparser

import (
	"errors"
	"fmt"
	"strings"
	"unsafe"
)

// ErrDuplicateKey is returned as Error.Cause for duplicate tag keys.
var ErrDuplicateKey = errors.New("duplicate option key")

// Error is the type of error returned by parse funcs in this package.
type Error struct {
	// Tag is the original tag string that has a syntax error.
	Tag string
	// Pos is a 0-based position within the Tag string appropriate to report
	// as errorneous.
	Pos int
	// Msg is an error message, or an optional prefix to the error message of
	// the Cause.
	Msg string
	// Cause is an optional underlying error returned by ParseFunc callback, or
	// ErrDuplicateKey.
	Cause error
}

type Tag struct {
	Name    string
	Options map[string]string
}

func (e *Error) Error() string {
	if e.Cause != nil {
		if e.Msg != "" {
			return fmt.Sprintf("%s: %v (at %d)", e.Msg, e.Cause, e.Pos+1)
		} else {
			return fmt.Sprintf("%v (at %d)", e.Cause, e.Pos+1)
		}
	} else {
		return fmt.Sprintf("%s (at %d)", e.Msg, e.Pos+1)
	}
}

func (e *Error) Unwrap() error {
	return e.Cause
}

// Parse parses a tag treating the first item as a name. See ParseFunc for
// the full syntax and details.
func Parse(tag string) (*Tag, error) {
	result := &Tag{
		Options: make(map[string]string),
	}
	err := ParseFunc(tag, func(key, value string) error {
		if key == "" {
			result.Name = value
		} else {
			if _, ok := result.Options[key]; ok {
				return ErrDuplicateKey
			}
			result.Options[key] = value
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ParseFunc enumerates fields of a tag formatted as a list of keys and/or
// key-value pairs, treating the first item as a name.
//
// The format of the tag is:
//
//	name,key1,key2:value2,key3:'quoted, value',key4
//
// Tag syntax:
//
//  1. A tag is a list of comma-separated items.
//
//  2. An item is either a key:value pair or just a single string.
//
//  3. Both keys and values can be bare words (`foo: bar`) or single-quoted
//     strings (`foo: 'bar: boz, buzz and fubar'`). Quotes, if present, must
//     enclose the entire value after trimming whitespace. Mixed quoting like
//     `foo'bar'` is not allowed.
//
//  4. Both keys and values can use a backslash to escape special characters
//     (`foo\ bar`, `foo\:bar`, `foo\,bar`, `'foo\'n\'bar'`). In bare strings,
//     escape colons and commas. In quoted strings, escape quotes and backslashes.
//     Examples:
//     - Bare: `foo\:bar` → "foo:bar"
//     - Quoted: `'foo\'bar'` → "foo'bar"
//     - Quoted: `'foo\\bar'` → "foo\bar"
//     The escapes are processed and removed from the values (so `foo:\:\,\!` is
//     returned as `map[string]string{"foo": ":,!"}`); you can escape any
//     non-alphabetical characters.
//
//  5. Non-escaped unquoted leading and trailing ASCII whitespace is trimmed
//     from keys and values. (There seems to be no reason to handle Unicode
//     whitespace within struct tags.)
//
//  6. Parse and ParseFunc give special treatment to the first item of
//     the tag if it does not have a colon. Such an item is returned as Tag.Name
//     by Parse / as a value with an empty key by ParseFunc. If the first
//     item does have a colon, it is treated as a normal key; Parse returns an
//     empty Tag.Name, and ParseFunc reports a normal item and does not report
//     an item with an empty key.
//
//  7. For normal items, empty key names are not allowed. Empty values are
//     allowed (e.g., `key:` is valid and represents an empty string value).
//
// The error, if present, is *Error. If your callback returns an error, it will
// be wrapped in an Error with your error stored in Error.Cause.
func ParseFunc(tag string, callback func(key, value string) error) error {
	return parseFunc(tag, callback)
}

func parseFunc(tag string, callback func(key, value string) error) error {
	var parseErr error
	fail := func(i int, msg string, cause error) {
		if parseErr == nil {
			parseErr = &Error{tag, i, msg, cause}
		}
	}

	var count int
	var inValue bool
	var start int
	var key string
	var keyStart int

	flush := func(i int) {
		count++
		var value string

		parseValue := func(s string, pos int) (string, bool) {
			v, msg, p, quoted := unquoteTrim(s)
			if msg != "" {
				fail(pos+p, msg, nil)
			}
			return v, quoted
		}

		if count == 1 && !inValue {
			key = ""
			keyStart = start
			value, _ = parseValue(tag[start:i], start)
		} else {
			if inValue {
				key, _ = parseValue(key, keyStart)
				valueStr := tag[start:i]
				value, _ = parseValue(valueStr, start)
			} else if start < i {
				keyStart = start
				key, _ = parseValue(tag[start:i], start)
			} else {
				return
			}
			if key == "" {
				fail(keyStart, "empty key", nil)
				return
			}
		}

		if parseErr != nil {
			return // Early return on parse error
		}

		err := callback(key, value)
		if err != nil {
			fail(keyStart, key, err)
		}
	}

	n := len(tag)

	checkEscape := func(i int) {
		if i >= n {
			fail(i-1, "unterminated escape sequence", nil)
			return
		}
		c := tag[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
			fail(i, "invalid escape character", nil)
		}
	}

	var quoteStart int = -1
	for i := 0; i < n; i++ {
		if quoteStart >= 0 {
			switch tag[i] {
			case '\'':
				quoteStart = -1
			case '\\':
				i++
				checkEscape(i)
			}
		} else {
			switch tag[i] {
			case '\'':
				quoteStart = i
			case '\\':
				i++
				checkEscape(i)
			case ':':
				if !inValue {
					key = tag[start:i]
					keyStart = start
					start = i + 1
					inValue = true
				}
			case ',':
				flush(i)
				start = i + 1
				inValue = false
			}
		}
	}
	if quoteStart >= 0 {
		fail(quoteStart, "unterminated quote", nil)
	}
	if start < n || inValue {
		flush(n)
	}
	return parseErr
}

var asciiSpace = [256]uint8{'\t': 1, '\n': 1, '\v': 1, '\f': 1, '\r': 1, ' ': 1}

// unquoteTrim trims leading and trailing unescaped ASCII whitespace, processes
// escape sequences within the string and removes single quotes.
// Quotes, if present, must enclose the entire value after trimming whitespace.
// Returns whether the input was quoted (after trimming).
func unquoteTrim(s string) (result string, parseErr string, errPos int, wasQuoted bool) {
	n := len(s)

	// Trim leading unescaped whitespace
	var start int
	for start < n && asciiSpace[s[start]] != 0 {
		start++
	}

	// Trim trailing unescaped whitespace
	// Need to check for escapes to avoid trimming escaped spaces
	var end int = n
	for end > start && asciiSpace[s[end-1]] != 0 {
		// Check if this whitespace is escaped
		// Count preceding backslashes
		numBackslashes := 0
		for j := end - 2; j >= start && s[j] == '\\'; j-- {
			numBackslashes++
		}
		// If odd number of backslashes, the space is escaped
		if numBackslashes%2 == 1 {
			break
		}
		end--
	}

	// Check if value starts/ends with quotes (after trimming)
	hasQuotes := start < end && s[start] == '\'' && s[end-1] == '\''
	hasAnyQuote := strings.IndexByte(s, '\'') >= 0

	// Fast path: no escapes and no quotes at all
	if strings.IndexByte(s, '\\') < 0 && !hasAnyQuote {
		return s[start:end], "", 0, false
	}

	b := make([]byte, 0, n)
	var quoteCount int
	var firstQuotePos int = -1

mainLoop:
	for i := start; i < end; i++ {
		c := s[i]
		switch c {
		case '\\':
			if i+1 < end {
				b = append(b, s[i+1])
				i++
			}
			continue mainLoop
		case '\'':
			quoteCount++
			switch quoteCount {
			case 1:
				firstQuotePos = i
				// First quote must be at start (after trimming)
				if i != start {
					if parseErr == "" {
						parseErr, errPos = "quotes must enclose the entire value", i
					}
				}
			case 2:
				// Second quote must be at end (after trimming)
				if i != end-1 {
					if parseErr == "" {
						parseErr, errPos = "quotes must enclose the entire value", i
					}
				}
			default:
				// More than 2 quotes
				if parseErr == "" {
					parseErr, errPos = "invalid quote", i
				}
			}
			continue mainLoop
		}
		b = append(b, c)
	}

	// Validate strict quoting: if quotes present, must be exactly 2 and enclose entire value
	if hasQuotes {
		if quoteCount != 2 {
			if parseErr == "" {
				parseErr, errPos = "quotes must enclose the entire value", firstQuotePos
			}
		}
		wasQuoted = quoteCount == 2
	} else if quoteCount > 0 {
		// Quotes found but not at boundaries
		if parseErr == "" {
			parseErr, errPos = "quotes must enclose the entire value", firstQuotePos
		}
	}

	if len(b) > 0 {
		result = unsafe.String(&b[0], len(b))
	}
	return
}
