package email

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/wneessen/go-mail"
)

// EmailService handles email sending with retry logic.
type EmailService struct {
	client      *mail.Client
	smtpConfig  SMTPSettings
	retryConfig RetrySettings
	logger      *slog.Logger
}

// NewEmailService creates a new EmailService instance.
func NewEmailService(cfg SMTPSettings, logger *slog.Logger) (*EmailService, error) {
	return NewEmailServiceWithRetry(cfg, DefaultRetrySettings(), logger)
}

// NewEmailServiceWithRetry creates a new EmailService instance with custom retry configuration.
func NewEmailServiceWithRetry(smtpConfig SMTPSettings, retryConfig RetrySettings, logger *slog.Logger) (*EmailService, error) {
	client, err := createMailClient(smtpConfig)
	if err != nil {
		return nil, err
	}

	return &EmailService{
		client:      client,
		smtpConfig:  smtpConfig,
		retryConfig: retryConfig,
		logger:      logger,
	}, nil
}

// NewEmailServiceFromConfig creates a new EmailService from complete email configuration.
func NewEmailServiceFromConfig(cfg EmailConfig, logger *slog.Logger) (*EmailService, error) {
	return NewEmailServiceWithRetry(cfg.SMTP, cfg.Retry, logger)
}

// SendEmail sends an email with retry logic.
func (s *EmailService) SendEmail(ctx context.Context, e Email) error {
	if err := e.Validate(); err != nil {
		return fmt.Errorf("invalid email: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt < s.retryConfig.MaxAttempts; attempt++ {
		if attempt > 0 {
			if err := s.waitForRetry(ctx, attempt); err != nil {
				return err
			}
		}

		if err := s.send(ctx, e); err == nil {
			s.logSuccess(ctx, e.To, attempt)

			return nil
		} else {
			lastErr = err
			s.logFailure(ctx, e.To, attempt, err)
		}
	}

	return fmt.Errorf("failed to send email after %d attempts: %w", s.retryConfig.MaxAttempts, lastErr)
}

// SendHTML sends an HTML email.
func (s *EmailService) SendHTML(ctx context.Context, to, subject, htmlBody string) error {
	return s.SendEmail(ctx, Email{
		To:       to,
		Subject:  subject,
		HTMLBody: htmlBody,
	})
}

// SendText sends a plain text email.
func (s *EmailService) SendText(ctx context.Context, to, subject, textBody string) error {
	return s.SendEmail(ctx, Email{
		To:       to,
		Subject:  subject,
		TextBody: textBody,
	})
}

// Close closes the email service client.
// Note: DialAndSend creates new connections each time, so this is primarily for interface compatibility.
func (s *EmailService) Close() error {
	if s.client != nil {
		return s.client.Close()
	}

	return nil
}

// send performs the actual email send operation.
func (s *EmailService) send(ctx context.Context, e Email) error {
	// Check context before starting
	if err := ctx.Err(); err != nil {
		return err
	}

	msg, err := s.buildMessage(e)
	if err != nil {
		return err
	}

	// Note: go-mail's DialAndSend doesn't directly accept context
	// Timeout is handled via WithTimeout option in client creation
	if err := s.client.DialAndSend(msg); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// waitForRetry waits before retrying with exponential backoff.
func (s *EmailService) waitForRetry(ctx context.Context, attempt int) error {
	// Exponential backoff: 1s, 2s, 4s, 8s...
	// Use bit shift only with non-negative values to avoid overflow
	exponent := max(0, attempt-1)
	delay := s.retryConfig.InitialDelay * (1 << exponent)

	s.logger.WarnContext(ctx, "Retrying email send",
		"attempt", attempt+1,
		"delay", delay,
	)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(delay):
		return nil
	}
}

// logSuccess logs successful email delivery.
func (s *EmailService) logSuccess(ctx context.Context, to string, attempt int) {
	if attempt > 0 {
		s.logger.InfoContext(ctx, "Email sent successfully after retry",
			"attempt", attempt+1,
			"to", to,
		)
	} else {
		s.logger.InfoContext(ctx, "Email sent successfully",
			"to", to,
		)
	}
}

// logFailure logs failed email delivery attempt.
func (s *EmailService) logFailure(ctx context.Context, to string, attempt int, err error) {
	s.logger.WarnContext(ctx, "Failed to send email",
		"attempt", attempt+1,
		"error", err,
		"to", to,
	)
}

// buildMessage constructs the email message.
func (s *EmailService) buildMessage(e Email) (*mail.Msg, error) {
	msg := mail.NewMsg()

	if err := s.setFromAddress(msg); err != nil {
		return nil, err
	}

	if err := msg.To(e.To); err != nil {
		return nil, fmt.Errorf("failed to set To address: %w", err)
	}

	if err := msg.ReplyTo(s.smtpConfig.From); err != nil {
		return nil, fmt.Errorf("failed to set Reply-To address: %w", err)
	}

	msg.Subject(e.Subject)
	s.setTransactionalHeaders(msg)
	s.setMessageBody(msg, e)

	return msg, nil
}

// setFromAddress sets the From address with optional display name.
func (s *EmailService) setFromAddress(msg *mail.Msg) error {
	fromAddr := s.smtpConfig.From
	if s.smtpConfig.FromName != "" {
		fromAddr = fmt.Sprintf("%s <%s>", s.smtpConfig.FromName, s.smtpConfig.From)
	}

	if err := msg.From(fromAddr); err != nil {
		return fmt.Errorf("failed to set From address: %w", err)
	}

	return nil
}

// setTransactionalHeaders sets headers for transactional emails to prevent auto-replies.
func (s *EmailService) setTransactionalHeaders(msg *mail.Msg) {
	// Precedence: bulk prevents auto-replies and out-of-office messages
	msg.SetGenHeader(mail.HeaderPrecedence, "bulk")
	// X-Auto-Response-Suppress prevents auto-responders
	msg.SetGenHeader(mail.HeaderXAutoResponseSuppress, "All")
	// Set priority to normal (explicitly set for consistency)
	msg.SetImportance(mail.ImportanceNormal)
}

// setMessageBody sets the email body based on available content.
// Note: Email.Validate() ensures at least one body type is present.
func (s *EmailService) setMessageBody(msg *mail.Msg, e Email) {
	// Multipart: both HTML and text
	if e.HTMLBody != "" && e.TextBody != "" {
		msg.SetBodyString(mail.TypeTextHTML, e.HTMLBody)
		msg.AddAlternativeString(mail.TypeTextPlain, e.TextBody)

		return
	}

	// HTML only
	if e.HTMLBody != "" {
		msg.SetBodyString(mail.TypeTextHTML, e.HTMLBody)

		return
	}

	// Text only (guaranteed by validation)
	msg.SetBodyString(mail.TypeTextPlain, e.TextBody)
}

// createMailClient creates and configures a mail client.
func createMailClient(cfg SMTPSettings) (*mail.Client, error) {
	opts := buildMailOptions(cfg)

	client, err := mail.NewClient(cfg.Host, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create mail client: %w", err)
	}

	return client, nil
}

// buildMailOptions constructs mail client options from configuration.
func buildMailOptions(cfg SMTPSettings) []mail.Option {
	opts := []mail.Option{
		mail.WithPort(cfg.Port),
		configureTLS(cfg.EnableStartTLS),
	}

	if cfg.Username != "" && cfg.Password != "" {
		opts = append(opts,
			mail.WithSMTPAuth(mail.SMTPAuthPlain),
			mail.WithUsername(cfg.Username),
			mail.WithPassword(cfg.Password),
		)
	}

	if cfg.TimeoutSeconds > 0 {
		opts = append(opts, mail.WithTimeout(time.Duration(cfg.TimeoutSeconds)*time.Second))
	}

	return opts
}

// configureTLS returns the appropriate TLS policy option.
func configureTLS(enableStartTLS bool) mail.Option {
	if enableStartTLS {
		return mail.WithTLSPolicy(mail.TLSMandatory)
	}

	return mail.WithTLSPolicy(mail.TLSOpportunistic)
}
