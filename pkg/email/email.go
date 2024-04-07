package email

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"net/smtp"
	"regexp"

	"github.com/sirupsen/logrus"
)

type EmailSender interface {
	SendMessage(path string, to string, subject string, data interface{}) error
}

type SMTPSender struct {
	Email string
	Pass  string
	Host  string
	Port  string
	auth  smtp.Auth
}

type Config struct {
	Email string
	Pass  string
	Host  string
	Port  string
}

func NewSMTPSender(cfg Config) (*SMTPSender, error) {
	if err := validateEmail(cfg.Email); err != nil {
		return nil, ErrInvalidEmail
	}
	auth := smtp.PlainAuth("", cfg.Email, cfg.Pass, cfg.Host)
	return &SMTPSender{
		Email: cfg.Email,
		Pass:  cfg.Pass,
		Host:  cfg.Host,
		Port:  cfg.Port,
		auth:  auth,
	}, nil
}

var (
	ErrInvalidEmail = errors.New("invalid email")
)

func validateEmail(email string) error {
	if len(email) < 3 || len(email) > 1024 {
		return ErrInvalidEmail
	}

	matched, _ := regexp.MatchString("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$",
		email)
	if !matched {
		return ErrInvalidEmail
	}

	return nil
}

func (s *SMTPSender) SendMessage(path string, to string, subject string, data interface{}) error {
	t, err := template.ParseFiles(path)
	if err != nil {
		logrus.Errorf("error parse file with path %s when sending email message: %s", path, err)
		return err
	}
	buf := new(bytes.Buffer)
	if err := t.Execute(buf, data); err != nil {
		logrus.Errorf("error execute template with data %#v when sending email message: %s",
			data, err)
		return err
	}

	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"

	msg := fmt.Sprintf("To: %s\nSubject: %s\n", to, subject) + mime + buf.String()
	tos := []string{
		to,
	}

	err = smtp.SendMail(s.Host+":"+s.Port, s.auth, s.Email, tos, []byte(msg))
	if err != nil {
		logrus.Errorf("error send email message when senging email message: %s", err)
		return err
	}

	return nil
}
