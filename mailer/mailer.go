package mailer

import (
	"bytes"
	"context"
	"log/slog"
	"os"
	"strconv"

	"github.com/a-h/templ"
	gomail "gopkg.in/gomail.v2"
)

type Message struct {
	To      string
	Subject string
	Body    templ.Component
}

func Send(ctx context.Context, msg Message) error {
	host := os.Getenv("SMTP_HOST")
	portStr := os.Getenv("SMTP_PORT")
	user := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASS")
	from := os.Getenv("SMTP_FROM")

	port, err := strconv.Atoi(portStr)
	if err != nil {
		port = 587
	}

	var buf bytes.Buffer
	if err := msg.Body.Render(ctx, &buf); err != nil {
		return err
	}

	m := gomail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", msg.To)
	m.SetHeader("Subject", msg.Subject)
	m.SetBody("text/html", buf.String())

	d := gomail.NewDialer(host, port, user, pass)
	if err := d.DialAndSend(m); err != nil {
		slog.Error("mailer: send failed", "to", msg.To, "error", err)
		return err
	}

	slog.Info("mailer: sent", "to", msg.To, "subject", msg.Subject)
	return nil
}
