package email

import (
	"fmt"
	"net/smtp"
	"strings"
)

// EmailService handles email sending
type EmailService struct {
	smtpHost     string
	smtpPort     string
	smtpUsername string
	smtpPassword string
	fromEmail    string
	fromName     string
}

// NewEmailService creates a new email service
func NewEmailService(host, port, username, password, fromEmail, fromName string) *EmailService {
	return &EmailService{
		smtpHost:     host,
		smtpPort:     port,
		smtpUsername: username,
		smtpPassword: password,
		fromEmail:    fromEmail,
		fromName:     fromName,
	}
}

// SendEmail sends an email
func (s *EmailService) SendEmail(to, subject, body string) error {
	// Create email message
	from := fmt.Sprintf("%s <%s>", s.fromName, s.fromEmail)
	msg := []byte(fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s", from, to, subject, body))

	// SMTP auth
	auth := smtp.PlainAuth("", s.smtpUsername, s.smtpPassword, s.smtpHost)

	// Send email
	addr := fmt.Sprintf("%s:%s", s.smtpHost, s.smtpPort)
	err := smtp.SendMail(addr, auth, s.fromEmail, []string{to}, msg)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// SendBulkEmail sends email to multiple recipients
func (s *EmailService) SendBulkEmail(recipients []string, subject, body string) (int, int, error) {
	sent := 0
	failed := 0

	for _, recipient := range recipients {
		err := s.SendEmail(recipient, subject, body)
		if err != nil {
			failed++
		} else {
			sent++
		}
	}

	return sent, failed, nil
}

// RenderTemplate renders a template with variables
func (s *EmailService) RenderTemplate(template string, variables map[string]string) string {
	result := template
	for key, value := range variables {
		placeholder := fmt.Sprintf("{{%s}}", key)
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return result
}
