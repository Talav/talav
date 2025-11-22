package negotiation

import "strings"

// newHeaderAccept is the single shared implementation for all Accept-* headers.
func newHeaderAccept(value string, parseType func(string) (string, string, string, error)) (*Header, error) {
	typ, params, q, err := parseAcceptValue(value)
	if err != nil {
		return nil, err
	}

	// parseType returns: normalizedType, base, sub, error
	typ, base, sub, err := parseType(typ)
	if err != nil {
		return nil, err
	}

	return newHeader(value, typ, base, sub, q, params), nil
}

// newMedia creates a new Header for a media type from a header value.
func newMedia(value string) (*Header, error) {
	return newHeaderAccept(value, func(typ string) (string, string, string, error) {
		if typ == "*" {
			typ = "*/*"
		}
		parts := strings.SplitN(typ, "/", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return "", "", "", &InvalidMediaTypeError{}
		}

		return typ, parts[0], parts[1], nil
	})
}

// newLanguage creates a new Header for a language from a header value.
func newLanguage(value string) (*Header, error) {
	return newHeaderAccept(value, func(typ string) (string, string, string, error) {
		parts := strings.Split(typ, "-")
		switch len(parts) {
		case 1:
			return typ, parts[0], "", nil
		case 2:
			return typ, parts[0], parts[1], nil
		case 3: // zh-Hans-CN
			return typ, parts[0], parts[2], nil
		default:
			return "", "", "", &InvalidLanguageError{}
		}
	})
}

// newCharset creates a new Header for a charset from a header value.
func newCharset(value string) (*Header, error) {
	return newHeaderAccept(value, func(typ string) (string, string, string, error) {
		return typ, "", "", nil
	})
}

// newEncoding creates a new Header for an encoding from a header value.
func newEncoding(value string) (*Header, error) {
	return newHeaderAccept(value, func(typ string) (string, string, string, error) {
		return typ, "", "", nil
	})
}
