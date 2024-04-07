package service

import (
	"errors"

	"github.com/IvanMeln1k/go-bank-app-bank/pkg/tokens"
	"github.com/IvanMeln1k/go-bank-app-worker/pkg/email"
)

var (
	ErrInternal         = errors.New("error internal")
	ErrSendEmailMessage = errors.New("errorr send email message")
)

type Auth interface {
	SendEmailVerificationMessage(email string) error
}

type Service struct {
	Auth
}

type Deps struct {
	EmailSender     email.EmailSender
	TokenManager    tokens.TokenManagerInterface
	VerificationURL string
}

func NewService(deps Deps) *Service {
	return &Service{
		Auth: NewAuthService(deps.EmailSender, deps.TokenManager, deps.VerificationURL),
	}
}
