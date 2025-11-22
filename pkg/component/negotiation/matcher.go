package negotiation

import (
	"strings"
)

// Match represents a match between an accept header and a priority.
type matchResult struct {
	Quality float64
	Score   int
	Index   int
}

// matcher determines if an accept header matches a priority.
type matcher func(accept, priority *Header, index int) *matchResult

// matchMediaType matches media types with support for wildcards and plus-segments.
func matchMediaType(accept, priority *Header, index int) *matchResult {
	if !paramsMatch(accept.Parameters, priority.Parameters) {
		return nil
	}

	acceptSubPart, acceptSuffix := splitSubPart(accept.SubPart)
	prioritySubPart, prioritySuffix := splitSubPart(priority.SubPart)

	if !matchesBase(accept.BasePart, priority.BasePart) {
		return nil
	}

	if !matchesSubtype(acceptSubPart, prioritySubPart) {
		return nil
	}

	if !matchesSuffix(acceptSuffix, prioritySuffix) {
		return nil
	}

	score := calculateMediaTypeScore(
		accept.BasePart, priority.BasePart,
		acceptSubPart, prioritySubPart,
		acceptSuffix, prioritySuffix,
	)

	return &matchResult{
		Quality: accept.Quality * priority.Quality,
		Score:   score,
		Index:   index,
	}
}

// matchesBase checks if base parts match (with wildcard support).
func matchesBase(acceptBase, priorityBase string) bool {
	return acceptBase == "*" || strings.EqualFold(acceptBase, priorityBase)
}

// matchesSubtype checks if subtype parts match (with wildcard support).
func matchesSubtype(acceptSubPart, prioritySubPart string) bool {
	return acceptSubPart == "*" || prioritySubPart == "*" ||
		strings.EqualFold(acceptSubPart, prioritySubPart)
}

// matchesSuffix checks if suffix parts match (RFC 6839).
func matchesSuffix(acceptSuffix, prioritySuffix string) bool {
	return (acceptSuffix == "" && prioritySuffix == "") ||
		acceptSuffix == "*" || prioritySuffix == "*" ||
		strings.EqualFold(acceptSuffix, prioritySuffix)
}

// calculateMediaTypeScore calculates the match score for media types.
func calculateMediaTypeScore(acceptBase, priorityBase, acceptSubPart, prioritySubPart, acceptSuffix, prioritySuffix string) int {
	baseEqual := strings.EqualFold(acceptBase, priorityBase)
	score := 100 * boolToInt(baseEqual)

	subMatches := matchesSubtype(acceptSubPart, prioritySubPart)
	if subMatches && acceptSubPart != "*" && prioritySubPart != "*" {
		score += 10
	}

	suffixMatches := matchesSuffix(acceptSuffix, prioritySuffix)
	if suffixMatches && acceptSuffix != "" && prioritySuffix != "" &&
		acceptSuffix != "*" && prioritySuffix != "*" {
		score += 1
	}

	return score
}

// splitSubPart splits a subpart into the subpart and "plus" suffix.
// For media-types like "application/vnd.example+json", this splits into
// ("vnd.example", "json") to allow wildcard matching.
// Uses strings.LastIndex to handle multiple "+" correctly (RFC 6839).
func splitSubPart(subPart string) (string, string) {
	if idx := strings.LastIndex(subPart, "+"); idx >= 0 {
		return subPart[:idx], subPart[idx+1:]
	}

	return subPart, ""
}

// MatchLanguage matches languages with support for base/sub matching and fallback.
func matchLanguage(accept, priority *Header, index int) *matchResult {
	ab := accept.BasePart
	pb := priority.BasePart
	as := accept.SubPart
	ps := priority.SubPart

	baseEqual := strings.EqualFold(ab, pb)
	subEqual := strings.EqualFold(as, ps)

	// Match if base parts match (or accept is wildcard) and sub parts match or are nil
	if (ab == "*" || baseEqual) && (as == "" || subEqual || ps == "") {
		score := 10*boolToInt(baseEqual) + boolToInt(subEqual)

		return &matchResult{
			Quality: accept.Quality * priority.Quality,
			Score:   score,
			Index:   index,
		}
	}

	return nil
}

// MatchSimple matches simple string types (charset, encoding) with wildcard support.
func matchSimple(accept, priority *Header, index int) *matchResult {
	ac := accept.Type
	pc := priority.Type

	equal := strings.EqualFold(ac, pc)

	if equal || ac == "*" {
		score := boolToInt(equal)

		return &matchResult{
			Quality: accept.Quality * priority.Quality,
			Score:   score,
			Index:   index,
		}
	}

	return nil
}

// paramsMatch checks that all accept parameters are satisfied by priority parameters.
// Per RFC 7231: priority (server) must satisfy all accept (client) parameter requirements.
func paramsMatch(acceptParams, priorityParams map[string]string) bool {
	for k, acceptValue := range acceptParams {
		priorityValue, ok := priorityParams[k]
		if !ok || !strings.EqualFold(acceptValue, priorityValue) {
			return false
		}
	}

	return true
}

// boolToInt converts a boolean to an integer (1 for true, 0 for false).
func boolToInt(b bool) int {
	if b {
		return 1
	}

	return 0
}
