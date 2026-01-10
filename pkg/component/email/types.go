package email

// Email represents an email message.
type Email struct {
	To       string
	Subject  string
	HTMLBody string
	TextBody string
}
