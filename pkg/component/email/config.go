package email

import "time"

// EmailConfig represents the complete email component configuration.
type EmailConfig struct {
	SMTP  SMTPSettings  `config:"smtp"`
	Retry RetrySettings `config:"retry"`
}

// DefaultEmailConfig returns sensible defaults for email configuration.
func DefaultEmailConfig() EmailConfig {
	return EmailConfig{
		SMTP:  DefaultSMTPSettings(),
		Retry: DefaultRetrySettings(),
	}
}

// SMTPSettings represents SMTP server configuration.
type SMTPSettings struct {
	Host           string `config:"host"`
	Port           int    `config:"port"`
	Username       string `config:"username"`
	Password       string `config:"password"`
	From           string `config:"from"`
	FromName       string `config:"from_name"`
	EnableStartTLS bool   `config:"enable_starttls"`
	TimeoutSeconds int    `config:"timeout_seconds"`
}

// DefaultSMTPSettings returns sensible defaults for SMTP configuration.
// Port 587 with STARTTLS is the recommended configuration for modern SMTP.
func DefaultSMTPSettings() SMTPSettings {
	return SMTPSettings{
		Port:           587,  // Standard submission port with STARTTLS
		EnableStartTLS: true, // Always enable TLS for security
		TimeoutSeconds: 30,   // Reasonable timeout for SMTP operations
	}
}

// RetrySettings represents retry configuration for email sending.
type RetrySettings struct {
	MaxAttempts  int           `config:"max_attempts"`
	InitialDelay time.Duration `config:"initial_delay"`
}

// DefaultRetrySettings returns sensible defaults for retry configuration.
func DefaultRetrySettings() RetrySettings {
	return RetrySettings{
		MaxAttempts:  3,
		InitialDelay: time.Second,
	}
}
