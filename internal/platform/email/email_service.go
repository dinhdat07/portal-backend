package email

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"html"
	appLogger "log"
	"net"
	"net/smtp"
	"portal-system/internal/config"
	"portal-system/internal/services"
	"strings"
	"time"
)

type SMTPEmailService struct {
	cfg config.SMTPConfig
}

func NewSMTPEmailService(cfg config.SMTPConfig) services.EmailSender {
	return &SMTPEmailService{cfg: cfg}
}

func (s *SMTPEmailService) SendVerificationEmail(ctx context.Context, to, name, verifyURL string) error {
	subject := "Verify your email"

	textBody := fmt.Sprintf(
		"Hello %s,\n\nPlease verify your email by clicking the link below:\n%s\n\nIf you did not create this account, you can ignore this email.\n",
		fallbackName(name),
		verifyURL,
	)

	htmlBody := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
		<meta charset="UTF-8" />
		<title>Verify your email</title>
		</head>
		<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #222;">
		<p>Hello %s,</p>
		<p>Please verify your email by clicking the button below:</p>
		<p>
			<a href="%s" style="display:inline-block;padding:10px 16px;background:#2563eb;color:#fff;text-decoration:none;border-radius:6px;">
			Verify Email
			</a>
		</p>
		<p>Or open this link manually:</p>
		<p><a href="%s">%s</a></p>
		<p>If you did not create this account, you can ignore this email.</p>
		</body>
		</html>`,
		html.EscapeString(fallbackName(name)),
		html.EscapeString(verifyURL),
		html.EscapeString(verifyURL),
		html.EscapeString(verifyURL),
	)

	return s.send(ctx, to, subject, textBody, htmlBody)
}

func (s *SMTPEmailService) SendResetPasswordEmail(ctx context.Context, to, name, resetURL string) error {
	subject := "Reset your password"

	textBody := fmt.Sprintf(
		"Hello %s,\n\nYou requested to reset your password.\nClick the link below to continue:\n%s\n\nIf you did not request this, you can ignore this email.\n",
		fallbackName(name),
		resetURL,
	)

	htmlBody := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
		<meta charset="UTF-8" />
		<title>Reset your password</title>
		</head>
		<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #222;">
		<p>Hello %s,</p>
		<p>You requested to reset your password.</p>
		<p>
			<a href="%s" style="display:inline-block;padding:10px 16px;background:#dc2626;color:#fff;text-decoration:none;border-radius:6px;">
			Reset Password
			</a>
		</p>
		<p>Or open this link manually:</p>
		<p><a href="%s">%s</a></p>
		<p>If you did not request this, you can ignore this email.</p>
		</body>
		</html>`,
		html.EscapeString(fallbackName(name)),
		html.EscapeString(resetURL),
		html.EscapeString(resetURL),
		html.EscapeString(resetURL),
	)

	return s.send(ctx, to, subject, textBody, htmlBody)
}

func (s *SMTPEmailService) SendSetPasswordEmail(ctx context.Context, to, name, setPasswordURL string) error {
	subject := "Set your password"

	textBody := fmt.Sprintf(
		"Hello %s,\n\nPlease set your password by clicking the link below:\n%s\n\nIf you were not expecting this email, you can ignore it.\n",
		fallbackName(name),
		setPasswordURL,
	)

	htmlBody := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
		<meta charset="UTF-8" />
		<title>Set your password</title>
		</head>
		<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #222;">
		<p>Hello %s,</p>
		<p>Please set your password by clicking the button below:</p>
		<p>
			<a href="%s" style="display:inline-block;padding:10px 16px;background:#16a34a;color:#fff;text-decoration:none;border-radius:6px;">
			Set Password
			</a>
		</p>
		<p>Or open this link manually:</p>
		<p><a href="%s">%s</a></p>
		<p>If you were not expecting this email, you can ignore it.</p>
		</body>
		</html>`,
		html.EscapeString(fallbackName(name)),
		html.EscapeString(setPasswordURL),
		html.EscapeString(setPasswordURL),
		html.EscapeString(setPasswordURL),
	)

	return s.send(ctx, to, subject, textBody, htmlBody)
}

func (s *SMTPEmailService) send(ctx context.Context, to, subject, textBody, htmlBody string) error {
	fromHeader := s.cfg.From
	if strings.TrimSpace(s.cfg.FromName) != "" {
		fromHeader = fmt.Sprintf("%s <%s>", mimeHeaderEncode(s.cfg.FromName), s.cfg.From)
	}

	msg := buildMultipartMessage(fromHeader, to, subject, textBody, htmlBody)

	addr := net.JoinHostPort(s.cfg.Host, s.cfg.Port)

	dialer := &net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("dial smtp server: %w", err)
	}

	client, err := smtp.NewClient(conn, s.cfg.Host)
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("create smtp client: %w", err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			appLogger.Println("failed to close smtp client connection", "error", err)
		}
	}()

	if s.cfg.UseTLS {
		tlsConfig := &tls.Config{
			ServerName: s.cfg.Host,
		}
		if err := client.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("start tls: %w", err)
		}
	}

	if s.cfg.UseAuth {
		auth := smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
	}

	if err := client.Mail(s.cfg.From); err != nil {
		return fmt.Errorf("smtp mail from: %w", err)
	}

	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("smtp rcpt to: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}

	if _, err := w.Write([]byte(msg)); err != nil {
		_ = w.Close()
		return fmt.Errorf("write message: %w", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("close message writer: %w", err)
	}

	if err := client.Quit(); err != nil {
		return fmt.Errorf("smtp quit: %w", err)
	}

	return nil
}

func buildMultipartMessage(from, to, subject, textBody, htmlBody string) string {
	boundary := fmt.Sprintf("portal-boundary-%d", time.Now().UnixNano())

	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("From: %s\r\n", from))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", to))
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", mimeHeaderEncode(subject)))
	buf.WriteString("MIME-Version: 1.0\r\n")
	buf.WriteString(fmt.Sprintf(`Content-Type: multipart/alternative; boundary="%s"`+"\r\n", boundary))
	buf.WriteString("\r\n")

	buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	buf.WriteString(`Content-Type: text/plain; charset="UTF-8"` + "\r\n")
	buf.WriteString("Content-Transfer-Encoding: 8bit\r\n")
	buf.WriteString("\r\n")
	buf.WriteString(textBody)
	buf.WriteString("\r\n")

	buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	buf.WriteString(`Content-Type: text/html; charset="UTF-8"` + "\r\n")
	buf.WriteString("Content-Transfer-Encoding: 8bit\r\n")
	buf.WriteString("\r\n")
	buf.WriteString(htmlBody)
	buf.WriteString("\r\n")

	buf.WriteString(fmt.Sprintf("--%s--\r\n", boundary))

	return buf.String()
}

func fallbackName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "there"
	}
	return name
}

// simple for local/MailHog
func mimeHeaderEncode(s string) string {
	return s
}
