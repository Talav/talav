package email

import (
	"fmt"
	"strings"
)

// Validate checks if the email has required fields.
func (e Email) Validate() error {
	if e.To == "" {
		return fmt.Errorf("recipient address is required")
	}

	if !isValidEmailFormat(e.To) {
		return fmt.Errorf("invalid recipient email format: %s", e.To)
	}

	if e.Subject == "" {
		return fmt.Errorf("subject is required")
	}

	if e.HTMLBody == "" && e.TextBody == "" {
		return fmt.Errorf("email body is required")
	}

	return nil
}

// isValidEmailFormat performs basic email format validation.
func isValidEmailFormat(email string) bool {
	// Basic validation: contains @ and has parts before and after
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}

	localPart := parts[0]
	domain := parts[1]

	if localPart == "" || domain == "" {
		return false
	}

	// Domain should have at least one dot
	if !strings.Contains(domain, ".") {
		return false
	}

	return true
}
