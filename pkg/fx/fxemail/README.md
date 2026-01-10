# FxEmail Module

Uber FX integration for the email component. Provides automatic SMTP configuration loading and email service registration.

## Features

- **Automatic Configuration**: Loads SMTP settings from config files
- **Dependency Injection**: Integrates with FX lifecycle
- **Simple Setup**: One module to add email capabilities

## Installation

```bash
go get github.com/talav/talav/pkg/fx/fxemail
```

## Quick Start

### 1. Add Module to Application

```go
package main

import (
    "github.com/talav/talav/pkg/fx/fxconfig"
    "github.com/talav/talav/pkg/fx/fxemail"
    "github.com/talav/talav/pkg/fx/fxlogger"
    "go.uber.org/fx"
)

func main() {
    app := fx.New(
        fxconfig.FxConfigModule,
        fxlogger.FxLoggerModule,
        fxemail.FxEmailModule,  // Add email module
        // Your modules...
    )
    app.Run()
}
```

### 2. Configure SMTP Settings

Create `config.yaml`:

```yaml
smtp:
  host: smtp.gmail.com
  port: 587
  username: ${SMTP_USERNAME}
  password: ${SMTP_PASSWORD}
  from: noreply@example.com
  from_name: "My Application"
  enable_starttls: true
  timeout_seconds: 30
```

### 3. Inject and Use Email Service

```go
package myservice

import (
    "context"
    
    "github.com/talav/talav/pkg/component/email"
)

type MyService struct {
    emailService *email.EmailService
}

func NewMyService(emailService *email.EmailService) *MyService {
    return &MyService{emailService: emailService}
}

func (s *MyService) SendWelcomeEmail(ctx context.Context, userEmail string) error {
    return s.emailService.SendHTML(
        ctx,
        userEmail,
        "Welcome!",
        "<h1>Welcome to our service!</h1>",
    )
}
```

## Module Configuration

The module provides:

1. **SMTPConfig**: Loaded from `smtp` config key
2. **EmailService**: Fully configured email service

```go
var FxEmailModule = fx.Module(
    ModuleName,
    fxconfig.AsConfig("smtp", email.SMTPConfig{}),
    fx.Provide(NewFxEmailService),
)
```

## Configuration Reference

### Config Structure

```yaml
smtp:
  host: smtp.example.com           # SMTP server hostname
  port: 587                         # SMTP server port
  username: user@example.com        # SMTP auth username
  password: secret                  # SMTP auth password
  from: noreply@example.com         # From address
  from_name: "My App"               # From display name
  enable_starttls: true             # Enable STARTTLS
  timeout_seconds: 30               # Connection timeout
```

### Environment Variables

Use environment variable substitution for sensitive data:

```yaml
smtp:
  host: ${SMTP_HOST}
  port: ${SMTP_PORT:587}            # Default to 587
  username: ${SMTP_USERNAME}
  password: ${SMTP_PASSWORD}
  from: ${SMTP_FROM}
  from_name: ${SMTP_FROM_NAME}
  enable_starttls: ${SMTP_STARTTLS:true}
  timeout_seconds: ${SMTP_TIMEOUT:30}
```

Set environment variables:
```bash
export SMTP_HOST=smtp.gmail.com
export SMTP_USERNAME=myapp@gmail.com
export SMTP_PASSWORD=app-specific-password
export SMTP_FROM=noreply@myapp.com
export SMTP_FROM_NAME="My Application"
```

## Complete Example

### Project Structure

```
myapp/
├── main.go
├── config.yaml
└── internal/
    └── notification/
        └── service.go
```

### main.go

```go
package main

import (
    "github.com/talav/talav/pkg/fx/fxconfig"
    "github.com/talav/talav/pkg/fx/fxemail"
    "github.com/talav/talav/pkg/fx/fxlogger"
    "myapp/internal/notification"
    "go.uber.org/fx"
)

func main() {
    app := fx.New(
        // Infrastructure
        fxlogger.FxLoggerModule,
        fxconfig.FxConfigModule,
        fxemail.FxEmailModule,
        
        // Application
        fx.Provide(notification.NewService),
        
        // Start services
        fx.Invoke(func(*notification.Service) {}),
    )
    app.Run()
}
```

### config.yaml

```yaml
smtp:
  host: smtp.gmail.com
  port: 587
  username: ${SMTP_USERNAME}
  password: ${SMTP_PASSWORD}
  from: noreply@myapp.com
  from_name: "My Application"
  enable_starttls: true
  timeout_seconds: 30
```

### internal/notification/service.go

```go
package notification

import (
    "context"
    "fmt"
    "log/slog"
    
    "github.com/talav/talav/pkg/component/email"
)

type Service struct {
    email  *email.EmailService
    logger *slog.Logger
}

func NewService(email *email.EmailService, logger *slog.Logger) *Service {
    return &Service{
        email:  email,
        logger: logger,
    }
}

func (s *Service) SendPasswordReset(ctx context.Context, userEmail, token string) error {
    htmlBody := fmt.Sprintf(`
        <h1>Password Reset Request</h1>
        <p>Click the link below to reset your password:</p>
        <a href="https://myapp.com/reset?token=%s">Reset Password</a>
        <p>This link expires in 1 hour.</p>
    `, token)
    
    textBody := fmt.Sprintf(`
        Password Reset Request
        
        Use this token to reset your password: %s
        
        This token expires in 1 hour.
    `, token)
    
    err := s.email.SendEmail(ctx, email.Email{
        To:       userEmail,
        Subject:  "Password Reset Request",
        HTMLBody: htmlBody,
        TextBody: textBody,
    })
    
    if err != nil {
        s.logger.ErrorContext(ctx, "Failed to send password reset email",
            "error", err,
            "email", userEmail,
        )
        return err
    }
    
    s.logger.InfoContext(ctx, "Password reset email sent",
        "email", userEmail,
    )
    return nil
}

func (s *Service) SendVerificationEmail(ctx context.Context, userEmail, code string) error {
    return s.email.SendHTML(
        ctx,
        userEmail,
        "Verify Your Email",
        fmt.Sprintf(`
            <h1>Verify Your Email</h1>
            <p>Your verification code is: <strong>%s</strong></p>
        `, code),
    )
}
```

## Testing

For integration tests, you can replace the module with a test configuration:

```go
func TestMyService_Integration(t *testing.T) {
    var service *MyService
    
    // Test SMTP config (use MailHog or similar)
    testConfig := email.SMTPConfig{
        Host: "localhost",
        Port: 1025,
        From: "test@example.com",
    }
    
    app := fxtest.New(
        t,
        fx.NopLogger,
        fxlogger.FxLoggerModule,
        fx.Provide(
            func() email.SMTPConfig { return testConfig },
            fxemail.NewFxEmailService,
        ),
        fx.Provide(NewMyService),
        fx.Populate(&service),
    ).RequireStart()
    defer app.RequireStop()
    
    // Test email sending
    err := service.SendWelcomeEmail(context.Background(), "test@example.com")
    require.NoError(t, err)
}
```

For unit tests, mock the email service:

```go
type MockEmailService struct {
    mock.Mock
}

func (m *MockEmailService) SendEmail(ctx context.Context, e email.Email) error {
    args := m.Called(ctx, e)
    return args.Error(0)
}

func TestMyService_Unit(t *testing.T) {
    mockEmail := new(MockEmailService)
    mockEmail.On("SendEmail", mock.Anything, mock.Anything).Return(nil)
    
    service := NewMyService(mockEmail, slog.Default())
    
    err := service.SendWelcomeEmail(context.Background(), "test@example.com")
    require.NoError(t, err)
    mockEmail.AssertExpectations(t)
}
```

## API Reference

### Module

```go
var FxEmailModule = fx.Module(
    ModuleName,
    fxconfig.AsConfig("smtp", email.SMTPConfig{}),
    fx.Provide(NewFxEmailService),
)
```

### Constructor

```go
func NewFxEmailService(p FxEmailServiceParam) (*email.EmailService, error)
```

### Module Providers

The module automatically provides:

- `email.SMTPConfig` - SMTP configuration from config files
- `*email.EmailService` - Configured email service

## Best Practices

### 1. Use Environment Variables for Secrets

Never commit SMTP credentials:

```yaml
smtp:
  username: ${SMTP_USERNAME}
  password: ${SMTP_PASSWORD}
```

### 2. Configure Per Environment

Use environment-specific config files:

```
config/
├── config.yaml          # Base config
├── config_dev.yaml      # Dev SMTP (MailHog)
├── config_test.yaml     # Test SMTP
└── config_prod.yaml     # Production SMTP
```

```yaml
# config_dev.yaml
smtp:
  host: localhost
  port: 1025
  enable_starttls: false
```

```yaml
# config_prod.yaml  
smtp:
  host: smtp.sendgrid.net
  port: 587
  enable_starttls: true
```

### 3. Handle Errors Gracefully

```go
err := emailService.SendEmail(ctx, email)
if err != nil {
    logger.ErrorContext(ctx, "Email failed", "error", err)
    // Don't fail critical operations due to email failures
    // Consider queuing for retry
}
```

### 4. Use Template Engines

For complex emails, use HTML templates:

```go
import "html/template"

tmpl, _ := template.ParseFiles("templates/welcome.html")
var buf bytes.Buffer
tmpl.Execute(&buf, data)

emailService.SendHTML(ctx, user.Email, "Welcome", buf.String())
```

### 5. Rate Limit Email Sending

Implement rate limiting to respect SMTP provider limits and prevent abuse.

## Common Configurations

### Gmail

```yaml
smtp:
  host: smtp.gmail.com
  port: 587
  username: ${GMAIL_USERNAME}
  password: ${GMAIL_APP_PASSWORD}  # App-specific password
  enable_starttls: true
```

### SendGrid

```yaml
smtp:
  host: smtp.sendgrid.net
  port: 587
  username: apikey
  password: ${SENDGRID_API_KEY}
  enable_starttls: true
```

### Amazon SES

```yaml
smtp:
  host: email-smtp.us-east-1.amazonaws.com
  port: 587
  username: ${AWS_SES_USERNAME}
  password: ${AWS_SES_PASSWORD}
  enable_starttls: true
```

### Development (MailHog)

```yaml
smtp:
  host: localhost
  port: 1025
  enable_starttls: false
```

## Security

- Store credentials in environment variables or secret managers
- Use STARTTLS in production
- Verify sender addresses with your SMTP provider
- Implement rate limiting
- Validate and sanitize recipient addresses
- Configure SPF, DKIM, and DMARC DNS records

## Dependencies

- `github.com/talav/talav/pkg/component/email` - Email component
- `github.com/talav/talav/pkg/fx/fxconfig` - Config module
- `go.uber.org/fx` - Dependency injection framework
