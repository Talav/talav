# Email Component

A framework-agnostic email service with SMTP support, retry logic, and automatic transactional email headers. Built on top of [go-mail](https://github.com/wneessen/go-mail).

## Features

- **SMTP Support**: Send emails via SMTP with TLS/STARTTLS
- **Retry Logic**: Automatic retry with exponential backoff
- **Context Support**: Full context cancellation support
- **HTML & Text**: Support for HTML, plain text, or multipart emails
- **Transactional Headers**: Automatic headers to prevent auto-replies
- **Configurable**: Flexible configuration for different SMTP providers
- **Framework Agnostic**: No dependency injection dependencies

## Installation

```bash
go get github.com/talav/talav/pkg/component/email
```

## Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "log"
    "log/slog"
    
    "github.com/talav/talav/pkg/component/email"
)

func main() {
    cfg := email.SMTPConfig{
        Host:           "smtp.example.com",
        Port:           587,
        Username:       "user@example.com",
        Password:       "password",
        From:           "noreply@example.com",
        FromName:       "My App",
        EnableStartTLS: true,
        TimeoutSeconds: 30,
    }
    
    logger := slog.Default()
    service, err := email.NewEmailService(cfg, logger)
    if err != nil {
        log.Fatal(err)
    }
    defer service.Close()
    
    // Send HTML email
    ctx := context.Background()
    err = service.SendHTML(
        ctx,
        "recipient@example.com",
        "Welcome!",
        "<h1>Welcome to our service!</h1><p>Thanks for signing up.</p>",
    )
    if err != nil {
        log.Fatal(err)
    }
}
```

### Send Multipart Email

```go
err := service.SendEmail(ctx, email.Email{
    To:       "user@example.com",
    Subject:  "Account Confirmation",
    HTMLBody: "<h1>Confirm Your Account</h1><p>Click the link below...</p>",
    TextBody: "Confirm Your Account\n\nClick the link below...",
})
```

### Send Text-Only Email

```go
err := service.SendText(
    ctx,
    "user@example.com",
    "Password Reset",
    "Your password reset code is: 123456",
)
```

## Configuration

```go
type SMTPConfig struct {
    Host           string // SMTP server hostname
    Port           int    // SMTP server port (25, 465, 587)
    Username       string // SMTP authentication username
    Password       string // SMTP authentication password
    From           string // From email address
    FromName       string // From display name (optional)
    EnableStartTLS bool   // Enable STARTTLS (recommended)
    TimeoutSeconds int    // Connection timeout in seconds
}
```

### Common SMTP Configurations

**Gmail**:
```go
SMTPConfig{
    Host:           "smtp.gmail.com",
    Port:           587,
    Username:       "your-email@gmail.com",
    Password:       "app-specific-password",
    From:           "your-email@gmail.com",
    EnableStartTLS: true,
    TimeoutSeconds: 30,
}
```

**SendGrid**:
```go
SMTPConfig{
    Host:           "smtp.sendgrid.net",
    Port:           587,
    Username:       "apikey",
    Password:       "your-api-key",
    From:           "verified-sender@yourdomain.com",
    EnableStartTLS: true,
    TimeoutSeconds: 30,
}
```

**Mailgun**:
```go
SMTPConfig{
    Host:           "smtp.mailgun.org",
    Port:           587,
    Username:       "postmaster@yourdomain.mailgun.org",
    Password:       "your-password",
    From:           "noreply@yourdomain.com",
    EnableStartTLS: true,
    TimeoutSeconds: 30,
}
```

## Retry Logic

The service automatically retries failed sends up to 3 times with exponential backoff:

- **Attempt 1**: Immediate
- **Attempt 2**: 1 second delay
- **Attempt 3**: 2 second delay
- **Attempt 4**: 4 second delay

Retries respect context cancellation - if the context is cancelled, retries stop immediately.

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

// Will retry but stop if context times out
err := service.SendHTML(ctx, to, subject, body)
```

## Transactional Email Headers

The service automatically sets headers to prevent auto-replies:

- `Precedence: bulk` - Prevents auto-replies and out-of-office messages
- `X-Auto-Response-Suppress: All` - Suppresses auto-responders
- `Reply-To` - Set to the From address by default
- `Importance: Normal` - Explicitly set for consistency

These headers make the emails suitable for transactional use (password resets, confirmations, notifications).

## Error Handling

```go
err := service.SendEmail(ctx, email.Email{
    To:       "invalid@",
    Subject:  "Test",
    HTMLBody: "<p>Test</p>",
})

if err != nil {
    // Errors include:
    // - Invalid email addresses
    // - SMTP connection failures
    // - Authentication failures
    // - Network timeouts
    // - Empty body
    log.Printf("Failed to send email: %v", err)
}
```

## Best Practices

### 1. Use Context with Timeout

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

err := service.SendEmail(ctx, email)
```

### 2. Provide Both HTML and Text

```go
email.Email{
    To:       "user@example.com",
    Subject:  "Welcome",
    HTMLBody: "<h1>Welcome!</h1>",
    TextBody: "Welcome!", // Fallback for email clients that don't support HTML
}
```

### 3. Close Service on Shutdown

```go
defer service.Close()
```

### 4. Use App-Specific Passwords

For Gmail and similar providers, use app-specific passwords instead of your main password.

### 5. Verify Sender Addresses

Ensure your `From` address is verified with your SMTP provider to avoid delivery issues.

## API Reference

### Types

```go
type Email struct {
    To       string
    Subject  string
    HTMLBody string
    TextBody string
}

type SMTPConfig struct {
    Host           string
    Port           int
    Username       string
    Password       string
    From           string
    FromName       string
    EnableStartTLS bool
    TimeoutSeconds int
}

type EmailService struct {
    // Internal fields
}
```

### Constructors

```go
// NewEmailService creates a new EmailService instance
func NewEmailService(cfg SMTPConfig, logger *slog.Logger) (*EmailService, error)
```

### Methods

```go
// SendEmail sends an email with retry logic
func (s *EmailService) SendEmail(ctx context.Context, e Email) error

// SendHTML sends an HTML email
func (s *EmailService) SendHTML(ctx context.Context, to, subject, htmlBody string) error

// SendText sends a plain text email
func (s *EmailService) SendText(ctx context.Context, to, subject, textBody string) error

// Close closes the email service client
func (s *EmailService) Close() error
```

## Testing

For testing, consider using a mock SMTP server or services like [MailHog](https://github.com/mailhog/MailHog) or [MailCatcher](https://mailcatcher.me/).

```go
// Test configuration
cfg := email.SMTPConfig{
    Host:           "localhost",
    Port:           1025, // MailHog default port
    EnableStartTLS: false,
    From:           "test@example.com",
}
```

## Security Considerations

1. **Never commit passwords**: Store SMTP credentials in environment variables or secure vaults
2. **Use STARTTLS**: Always enable `EnableStartTLS` in production
3. **Validate recipients**: Sanitize and validate recipient addresses
4. **Rate limiting**: Implement rate limiting to prevent abuse
5. **SPF/DKIM/DMARC**: Configure proper DNS records for your domain

## Limitations

- **Single recipient**: Each email supports one `To` address. For multiple recipients, send multiple emails.
- **No CC/BCC**: Transactional design focuses on single-recipient emails
- **No attachments**: Current version doesn't support attachments

## Dependencies

- [github.com/wneessen/go-mail](https://github.com/wneessen/go-mail) - Modern Go SMTP client library
