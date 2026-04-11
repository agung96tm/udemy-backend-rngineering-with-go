package mailer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"
)

const (
	mailtrapProdURL    = "https://send.api.mailtrap.io/api/send"
	mailtrapSandboxURL = "https://sandbox.api.mailtrap.io/api/send"
)

// MailtrapMailer mengirim email via Mailtrap Transactional API (production) atau Sandbox.
// Template selalu dibaca dari embed FS (mailer.FS).
type MailtrapMailer struct {
	fromEmail      string
	fromName       string
	apiKey         string
	sandboxInboxID string
	client         *http.Client
}

// MailtrapConfig konfigurasi untuk MailtrapMailer.
type MailtrapConfig struct {
	FromEmail      string // email pengirim
	FromName       string // nama pengirim
	APIKey         string // Mailtrap API token
	SandboxInboxID string // inbox ID untuk sandbox (contoh: 231735), dipakai saat isSandbox=true
}

// NewMailtrapMailer membuat MailtrapMailer dengan fromEmail dan apiKey.
// Deprecated: gunakan NewMailtrapMailerWithConfig untuk opsi fromName.
func NewMailtrapMailer(fromEmail, apiKey string) *MailtrapMailer {
	return NewMailtrapMailerWithConfig(MailtrapConfig{
		FromEmail: fromEmail,
		FromName:  fromEmail,
		APIKey:    apiKey,
	})
}

// NewMailtrapMailerWithConfig membuat MailtrapMailer dari config.
func NewMailtrapMailerWithConfig(cfg MailtrapConfig) *MailtrapMailer {
	if cfg.FromName == "" {
		cfg.FromName = cfg.FromEmail
	}
	return &MailtrapMailer{
		fromEmail:      cfg.FromEmail,
		fromName:       cfg.FromName,
		apiKey:         cfg.APIKey,
		sandboxInboxID: cfg.SandboxInboxID,
		client:         &http.Client{},
	}
}

// mailtrapSendRequest format body request Mailtrap API (production & sandbox).
type mailtrapSendRequest struct {
	From     mailtrapFrom        `json:"from"`
	To       []mailtrapRecipient `json:"to"`
	Subject  string              `json:"subject"`
	HTML     string              `json:"html,omitempty"`
	Text     string              `json:"text,omitempty"`
	Category string              `json:"category,omitempty"`
}

type mailtrapFrom struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type mailtrapRecipient struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

// Send mengirim email menggunakan template. templateFile nama file di templates/ (mis. user_invitation.templ).
// Template bisa berisi {{ define "subject" }} dan {{ define "body" }}; data dipakai untuk render.
// isSandbox=true kirim ke Mailtrap Sandbox (inbox testing).
func (m *MailtrapMailer) Send(templateFile string, username, email string, data any, isSandbox bool) error {
	body, subject, err := m.renderTemplate(templateFile, data)
	if err != nil {
		return fmt.Errorf("mailtrap: render template: %w", err)
	}

	req := mailtrapSendRequest{
		From:    mailtrapFrom{Email: m.fromEmail, Name: m.fromName},
		To:      []mailtrapRecipient{{Email: email, Name: username}},
		Subject: subject,
		HTML:    body,
		Text:    body,
	}
	if isSandbox {
		req.Category = "Integration Test"
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("mailtrap: marshal request: %w", err)
	}

	url := mailtrapProdURL
	if isSandbox {
		if m.sandboxInboxID == "" {
			return fmt.Errorf("mailtrap: sandbox inbox ID required (set MailtrapConfig.SandboxInboxID)")
		}
		url = mailtrapSandboxURL + "/" + m.sandboxInboxID
	}

	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("mailtrap: new request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if isSandbox {
		httpReq.Header.Set("Authorization", "Bearer "+m.apiKey)
	} else {
		httpReq.Header.Set("Api-Token", m.apiKey)
	}

	resp, err := m.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("mailtrap: do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("mailtrap: send failed status=%d", resp.StatusCode)
	}

	return nil
}

// renderTemplate membaca template dari embed FS (mailer.FS), path templates/<templateFile>.
// Template format: {{ define "subject" }}...{{ end }} dan {{ define "body" }}...{{ end }}.
// Mengembalikan (bodyHTML, subject, error).
func (m *MailtrapMailer) renderTemplate(templateFile string, data any) (body string, subject string, err error) {
	if templateFile == "" {
		return "", "", fmt.Errorf("template file is required")
	}
	name := filepath.Join("templates", filepath.Clean(templateFile))
	raw, err := fs.ReadFile(FS, name)
	if err != nil {
		return "", "", err
	}
	tpl, err := template.New(name).Parse(string(raw))
	if err != nil {
		return "", "", err
	}
	var buf bytes.Buffer
	if tpl.Lookup("body") != nil {
		if err := tpl.ExecuteTemplate(&buf, "body", data); err != nil {
			return "", "", err
		}
		body = buf.String()
		buf.Reset()
		if tpl.Lookup("subject") != nil {
			_ = tpl.ExecuteTemplate(&buf, "subject", data)
			subject = buf.String()
		}
	} else {
		if err := tpl.Execute(&buf, data); err != nil {
			return "", "", err
		}
		body = buf.String()
	}
	return body, subject, nil
}
