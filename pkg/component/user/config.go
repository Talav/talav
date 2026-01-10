package user

// PasswordResetConfig represents password reset configuration.
type PasswordResetConfig struct {
	TokenExpirationMinutes int    `config:"token_expiration_minutes"`
	ResetURLTemplate       string `config:"reset_url_template"`
}
