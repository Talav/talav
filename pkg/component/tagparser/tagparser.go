package tagparser

import (
	"errors"
	"fmt"
	"strings"
)

// ErrDuplicateKey is returned as Error.Cause for duplicate tag keys.
var ErrDuplicateKey = errors.New("duplicate option key")

const (
	errQuotesMustEnclose  = "quotes must enclose the entire value"
	errUnterminatedQuote  = "unterminated quote"
	errEmptyKey           = "empty key"
	errUnterminatedEscape = "unterminated escape sequence"
	errInvalidEscape      = "invalid escape character"
	errInvalidQuote       = "invalid quote"
)

// Error is the type of error returned by parse funcs in this package.
type Error struct {
	Tag   string // Original tag string
	Pos   int    // 0-based position of error
	Msg   string // Error message
	Cause error  // Optional underlying error
}

func (e *Error) Error() string {
	if e.Cause != nil {
		if e.Msg != "" {
			return fmt.Sprintf("%s: %v (at %d)", e.Msg, e.Cause, e.Pos+1)
		}

		return fmt.Sprintf("%v (at %d)", e.Cause, e.Pos+1)
	}

	return fmt.Sprintf("%s (at %d)", e.Msg, e.Pos+1)
}

func (e *Error) Unwrap() error { return e.Cause }

// Tag represents a parsed struct tag.
type Tag struct {
	Name    string
	Options map[string]string
}

// Parse parses a tag treating the first item as a name.
func Parse(tag string) (*Tag, error) {
	result := &Tag{Options: make(map[string]string)}
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
// Format: name,key1,key2:value2,key3:'quoted, value',key4
//
// Rules:
//   - Items are comma-separated; key:value pairs use colon
//   - Values can be bare words or single-quoted strings
//   - Backslash escapes special characters
//   - Leading/trailing ASCII whitespace is trimmed
//   - First item without colon becomes the name (empty key)
//   - Empty keys are not allowed for normal items
func ParseFunc(tag string, callback func(key, value string) error) error {
	p := parser{tag: tag, callback: callback}

	return p.parse()
}

type parser struct {
	tag      string
	callback func(key, value string) error
	pos      int
	start    int
	keyStart int
	key      string
	inValue  bool
	inQuote  bool
	count    int
}

func (p *parser) parse() error {
	for p.pos < len(p.tag) {
		c := p.tag[p.pos]
		if p.inQuote {
			if err := p.handleQuoted(c); err != nil {
				return err
			}
		} else {
			if err := p.handleUnquoted(c); err != nil {
				return err
			}
		}
		p.pos++
	}

	if p.inQuote {
		return &Error{p.tag, p.start, errUnterminatedQuote, nil}
	}

	return p.emitItem()
}

func (p *parser) handleQuoted(c byte) error {
	switch c {
	case '\'':
		p.inQuote = false
	case '\\':
		if err := p.consumeEscape(); err != nil {
			return err
		}
	}

	return nil
}

func (p *parser) handleUnquoted(c byte) error {
	switch c {
	case '\'':
		p.inQuote = true
	case '\\':
		return p.consumeEscape()
	case ':':
		if !p.inValue {
			return p.setKey()
		}
	case ',':
		if err := p.emitItem(); err != nil {
			return err
		}
		p.start = p.pos + 1
		p.inValue = false
	}

	return nil
}

func (p *parser) consumeEscape() error {
	next := p.pos + 1
	if next >= len(p.tag) {
		return &Error{p.tag, p.pos, errUnterminatedEscape, nil}
	}
	c := p.tag[next]
	if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
		return &Error{p.tag, next, errInvalidEscape, nil}
	}
	p.pos = next

	return nil
}

func (p *parser) setKey() error {
	keyStr := p.tag[p.start:p.pos]
	key, err := unquoteTrim(keyStr)
	if err != nil {
		return p.wrapUnquoteError(err, p.start)
	}
	if key == "" {
		return &Error{p.tag, p.start, errEmptyKey, nil}
	}
	p.key = keyStr
	p.keyStart = p.start
	p.start = p.pos + 1
	p.inValue = true

	return nil
}

func (p *parser) emitItem() error {
	p.count++

	// Skip empty items after first
	if p.start >= p.pos && p.count > 1 {
		return nil
	}

	key, value, err := p.getKeyValue()
	if err != nil {
		return err
	}

	if err := p.callback(key, value); err != nil {
		return &Error{p.tag, p.keyStart, p.key, err}
	}

	return nil
}

func (p *parser) getKeyValue() (string, string, error) {
	switch {
	case p.count == 1 && !p.inValue:
		// First item is the name
		value, err := unquoteTrim(p.tag[p.start:p.pos])
		if err != nil {
			return "", "", p.wrapUnquoteError(err, p.start)
		}

		return "", value, nil

	case p.inValue:
		// Key-value pair
		key, err := unquoteTrim(p.key)
		if err != nil {
			return "", "", p.wrapUnquoteError(err, p.keyStart)
		}
		value, err := unquoteTrim(p.tag[p.start:p.pos])
		if err != nil {
			return "", "", p.wrapUnquoteError(err, p.start)
		}

		return key, value, nil

	case p.start < p.pos:
		// Key-only item
		key, err := unquoteTrim(p.tag[p.start:p.pos])
		if err != nil {
			return "", "", p.wrapUnquoteError(err, p.start)
		}
		if key == "" {
			return "", "", &Error{p.tag, p.start, errEmptyKey, nil}
		}

		return key, "", nil
	}

	return "", "", nil
}

func (p *parser) wrapUnquoteError(err error, offset int) error {
	var ue *unquoteError
	if errors.As(err, &ue) {
		return &Error{p.tag, offset + ue.pos, ue.msg, nil}
	}

	return &Error{p.tag, offset, err.Error(), nil}
}

// unquoteError represents an error during unquoting.
type unquoteError struct {
	msg string
	pos int
}

func (e *unquoteError) Error() string { return e.msg }

var asciiSpace = [256]uint8{'\t': 1, '\n': 1, '\v': 1, '\f': 1, '\r': 1, ' ': 1}

// unquoteTrim trims whitespace, processes escapes, and removes quotes.
func unquoteTrim(s string) (string, error) {
	start, end := trimWhitespace(s)
	if start >= end {
		return "", nil
	}

	// Fast path: no escapes or quotes
	if strings.IndexByte(s, '\\') < 0 && strings.IndexByte(s, '\'') < 0 {
		return s[start:end], nil
	}

	return processQuotedString(s, start, end)
}

func processQuotedString(s string, start, end int) (string, error) {
	hasQuotes := s[start] == '\'' && s[end-1] == '\''
	b := make([]byte, 0, end-start)
	quoteCount := 0
	firstQuotePos := -1

	for i := start; i < end; i++ {
		c := s[i]
		switch c {
		case '\\':
			if i+1 < end {
				b = append(b, s[i+1])
				i++
			}
		case '\'':
			quoteCount++
			if firstQuotePos < 0 {
				firstQuotePos = i
			}
			if err := validateQuoteAt(quoteCount, i, start, end); err != nil {
				return string(b), err
			}
		default:
			b = append(b, c)
		}
	}

	if err := validateFinalQuotes(hasQuotes, quoteCount, firstQuotePos); err != nil {
		return string(b), err
	}

	return string(b), nil
}

func validateQuoteAt(quoteCount, pos, start, end int) error {
	switch quoteCount {
	case 1:
		if pos != start {
			return &unquoteError{errQuotesMustEnclose, pos}
		}
	case 2:
		if pos != end-1 {
			return &unquoteError{errQuotesMustEnclose, pos}
		}
	default:
		return &unquoteError{errInvalidQuote, pos}
	}

	return nil
}

func validateFinalQuotes(hasQuotes bool, quoteCount, firstQuotePos int) error {
	if hasQuotes && quoteCount != 2 {
		return &unquoteError{errQuotesMustEnclose, firstQuotePos}
	}
	if !hasQuotes && quoteCount > 0 {
		return &unquoteError{errQuotesMustEnclose, firstQuotePos}
	}

	return nil
}

func trimWhitespace(s string) (start, end int) {
	n := len(s)
	for start < n && asciiSpace[s[start]] != 0 {
		start++
	}
	end = n
	for end > start && asciiSpace[s[end-1]] != 0 {
		// Check if space is escaped
		backslashes := 0
		for j := end - 2; j >= start && s[j] == '\\'; j-- {
			backslashes++
		}
		if backslashes%2 == 1 {
			break
		}
		end--
	}

	return start, end
}
