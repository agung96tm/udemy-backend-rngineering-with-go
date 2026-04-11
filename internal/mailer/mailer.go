package mailer

import "embed"

const (
	FromEmail = "noreply@example.com"
	FromName  = "Example"
)

const (
	UserWelcomeTemplate = "user_invitation.templ"
)

// FS berisi template email yang di-embed (templates/*.templ).
//
//go:embed "templates/*"
var FS embed.FS

type Client interface {
	Send(templateFile string, username, email string, data any, isSandbox bool) error
}
