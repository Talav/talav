package negotiation

import (
	"maps"
	"slices"
	"sort"
)

// headerFactory creates Header instances from string values.
type headerFactory func(value string) (*Header, error)

// Negotiator handles all negotiation logic.
type Negotiator struct {
	factory headerFactory
	matcher matcher
}

// NewCharsetNegotiator creates a new Negotiator for charsets.
func NewCharsetNegotiator() *Negotiator {
	return newNegotiator(newCharset, matchSimple)
}

// NewEncodingNegotiator creates a new Negotiator for encodings.
func NewEncodingNegotiator() *Negotiator {
	return newNegotiator(newEncoding, matchSimple)
}

// NewLanguageNegotiator creates a new Negotiator for languages.
func NewLanguageNegotiator() *Negotiator {
	return newNegotiator(newLanguage, matchLanguage)
}

// NewMediaNegotiator creates a new Negotiator for media types.
func NewMediaNegotiator() *Negotiator {
	return newNegotiator(newMedia, matchMediaType)
}

// newNegotiator creates a new Negotiator with the given factory and matcher.
func newNegotiator(factory headerFactory, matcher matcher) *Negotiator {
	return &Negotiator{
		factory: factory,
		matcher: matcher,
	}
}

// GetBest returns the best matching accept header from priorities based on the header.
// If strict is true, returns errors for invalid headers; otherwise skips invalid entries.
func (c *Negotiator) Negotiate(header string, priorities []string, strict bool) (*Header, error) {
	if len(priorities) == 0 {
		return nil, &InvalidArgumentError{Message: "a set of server priorities should be given"}
	}

	if header == "" {
		return nil, &InvalidArgumentError{Message: "the header string should not be empty"}
	}

	// Parse accept headers once (performance critical)
	acceptedHeaders, err := c.parseAcceptHeaders(header, strict)
	if err != nil {
		return nil, err
	}

	// Parse priorities
	acceptedPriorities := make([]*Header, 0, len(priorities))
	for _, p := range priorities {
		acc, err := c.factory(p)
		if err != nil {
			if strict {
				return nil, err
			}

			continue
		}
		acceptedPriorities = append(acceptedPriorities, acc)
	}

	matches := c.findMatches(acceptedHeaders, acceptedPriorities)
	specificMatches := c.reduceMatches(matches)

	if len(specificMatches) == 0 {
		return nil, ErrNoMatch
	}

	sort.Slice(specificMatches, func(i, j int) bool {
		mi, mj := specificMatches[i], specificMatches[j]
		if mi.Quality != mj.Quality {
			return mi.Quality > mj.Quality
		}

		return mi.Index < mj.Index
	})

	bestMatch := specificMatches[0]

	return acceptedPriorities[bestMatch.Index], nil
}

// GetOrderedElements returns all accept header elements ordered by quality.
func (c *Negotiator) GetOrderedElements(header string) ([]*Header, error) {
	if header == "" {
		return nil, &InvalidArgumentError{Message: "the header string should not be empty"}
	}

	// Parse once (performance critical)
	elements, err := c.parseAcceptHeaders(header, false)
	if err != nil {
		return nil, err
	}

	sort.Slice(elements, func(i, j int) bool {
		if elements[i].Quality != elements[j].Quality {
			return elements[i].Quality > elements[j].Quality
		}

		return elements[i].originalIndex < elements[j].originalIndex
	})

	return elements, nil
}

// parseAcceptHeaders parses an Accept* header string into Header instances.
// Parses once to avoid redundant parsing (performance critical).
func (c *Negotiator) parseAcceptHeaders(header string, strict bool) ([]*Header, error) {
	parts, err := parseHeader(header)
	if err != nil {
		if strict {
			return nil, err
		}

		return []*Header{}, nil
	}

	headers := make([]*Header, 0, len(parts))
	for i, part := range parts {
		h, err := c.factory(part)
		if err != nil {
			if strict {
				return nil, err
			}

			continue
		}
		h.originalIndex = i
		headers = append(headers, h)
	}

	return headers, nil
}

// findMatches finds all matches between headers and priorities.
// Both arguments are already parsed Header instances (no redundant parsing).
func (c *Negotiator) findMatches(headers, priorities []*Header) []*matchResult {
	matches := make([]*matchResult, 0)

	for i, priority := range priorities {
		for _, accept := range headers {
			if match := c.matcher(accept, priority, i); match != nil {
				matches = append(matches, match)
			}
		}
	}

	return matches
}

// reduceMatches reduces matches to the best match per priority index.
func (c *Negotiator) reduceMatches(matches []*matchResult) []*matchResult {
	bestByIndex := make(map[int]*matchResult)

	for _, match := range matches {
		if existing, ok := bestByIndex[match.Index]; !ok || existing.Score < match.Score {
			bestByIndex[match.Index] = match
		}
	}

	return slices.Collect(maps.Values(bestByIndex))
}
