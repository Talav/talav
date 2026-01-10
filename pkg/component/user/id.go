package user

import (
	"crypto/rand"
	"encoding/base64"
	"strings"
)

// GenerateID generates a unique ID with the given prefix.
// Format: prefix_base64url (URL-safe, no padding).
func GenerateID(prefix string) string {
	// Generate 12 random bytes (96 bits)
	b := make([]byte, 12)
	_, _ = rand.Read(b)

	// Encode to base64url (URL-safe, no padding)
	encoded := base64.RawURLEncoding.EncodeToString(b)

	return prefix + "_" + encoded
}

// ExtractPrefix extracts the prefix from an ID.
// Returns empty string if ID doesn't contain an underscore.
func ExtractPrefix(id string) string {
	prefix, _, found := strings.Cut(id, "_")
	if !found {
		return ""
	}

	return prefix
}
