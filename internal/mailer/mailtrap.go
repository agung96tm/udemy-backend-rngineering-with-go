package mailer

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
)

// MailTrap menyimpan konfigurasi SMTP Mailtrap
type MailTrap struct {
	Host     string
	Username string
	Password string
	Port     string
	auth     smtp.Auth
}

// NewMailtrap membuat instance baru MailTrap
func NewMailtrap(host, username, password, port string) *MailTrap {
	auth := smtp.PlainAuth("", username, password, host)
	return &MailTrap{
		auth:     auth,
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
	}
}

func (m *MailTrap) Send(templateFile, username, email string, data any, isSandbox bool) error {
	to := []string{email}

	tmpl, err := template.ParseFS(FS, "templates/"+templateFile)
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}

	// Build subject
	subjectBuf := new(bytes.Buffer)
	if err := tmpl.ExecuteTemplate(subjectBuf, "subject", data); err != nil {
		return fmt.Errorf("execute subject template: %w", err)
	}

	// Build body
	bodyBuf := new(bytes.Buffer)
	if err := tmpl.ExecuteTemplate(bodyBuf, "body", data); err != nil {
		return fmt.Errorf("execute body template: %w", err)
	}

	// Email headers (use CRLF per RFC 5322)
	fromHeader := fmt.Sprintf("From: %s <%s>\r\n", FromName, FromEmail)
	toHeader := fmt.Sprintf("To: %s <%s>\r\n", username, email)
	subjectHeader := fmt.Sprintf("Subject: %s\r\n", subjectBuf.String())
	mimeHeader := "MIME-Version: 1.0\r\nContent-Type: text/html; charset=\"UTF-8\"\r\n\r\n"

	// Combine all into one message
	message := []byte(fromHeader + toHeader + subjectHeader + mimeHeader + bodyBuf.String())

	// Send email
	err = smtp.SendMail(m.Host+":"+m.Port, m.auth, FromEmail, to, message)
	if err != nil {
		return fmt.Errorf("send mail: %w", err)
	}

	fmt.Println("✅ Email sent successfully!")
	return nil
}
