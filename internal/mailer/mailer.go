package mailer

import "embed"

const (
	FromName  = "Admin <admin@example.com>"
	FromEmail = "admin@example.com"

	UserWelcomeTemplate = "user_invitation.templ"
)

//go:embed "templates"
var FS embed.FS

type Client interface {
	Send(templateFile, username, email string, data any, isSandbox bool) error
}
