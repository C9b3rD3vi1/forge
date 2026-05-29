package utils

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/smtp"
	"os"
	"path/filepath"
	"time"
)

type EmailConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	From     string
}

type EmailData struct {
	RecipientName  string
	RecipientEmail string
	MessageBody    template.HTML
	Subject        string
	Services       string
	LinkURL        string
	SiteName       string
	SiteURL        string
	Year           int
}

func GetEmailConfig() EmailConfig {
	cfg := EmailConfig{
		Host:     getEnv("SMTP_HOST", "smtp.gmail.com"),
		Port:     getEnv("SMTP_PORT", "587"),
		User:     os.Getenv("SMTP_USER"),
		Password: os.Getenv("SMTP_PASSWORD"),
		From:     os.Getenv("SMTP_FROM"),
	}
	if cfg.From == "" {
		cfg.From = cfg.User
	}
	return cfg
}

func IsSMTPConfigured() bool {
	cfg := GetEmailConfig()
	return cfg.User != "" && cfg.Password != ""
}

func SendHTMLEmail(to, subject, templateFile string, data EmailData) error {
	if !IsSMTPConfigured() {
		log.Printf("SMTP not configured — skipping email to %s (%s)", to, subject)
		return nil
	}

	if data.SiteName == "" {
		data.SiteName = getEnv("SITE_NAME", "Forge.Hub")
	}
	if data.SiteURL == "" {
		data.SiteURL = getEnv("SITE_URL", "http://localhost:3031")
	}
	if data.Year == 0 {
		data.Year = time.Now().Year()
	}
	if data.RecipientEmail == "" {
		data.RecipientEmail = to
	}

	tmplPath := filepath.Join("templates", "email", templateFile)
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		return fmt.Errorf("parse email template %s: %w", tmplPath, err)
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return fmt.Errorf("execute email template %s: %w", tmplPath, err)
	}

	cfg := GetEmailConfig()

	var msg bytes.Buffer
	msg.WriteString(fmt.Sprintf("From: %s\r\n", cfg.From))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", to))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")
	msg.WriteString("Content-Transfer-Encoding: quoted-printable\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(body.String())

	auth := smtp.PlainAuth("", cfg.User, cfg.Password, cfg.Host)
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	return smtp.SendMail(addr, auth, cfg.User, []string{to}, msg.Bytes())
}

func SendEmailAsync(to, subject, templateFile string, data EmailData) {
	go func() {
		if err := SendHTMLEmail(to, subject, templateFile, data); err != nil {
			log.Printf("Email send error (%s): %v", subject, err)
		}
	}()
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func GetEnv(key, fallback string) string {
	return getEnv(key, fallback)
}

func EscapeHTML(s string) string {
	var buf bytes.Buffer
	template.HTMLEscape(&buf, []byte(s))
	return buf.String()
}
