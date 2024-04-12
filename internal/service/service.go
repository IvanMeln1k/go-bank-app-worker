package service

import (
	"context"
	"errors"

	"github.com/IvanMeln1k/go-bank-app-bank/pkg/tokens"
	"github.com/IvanMeln1k/go-bank-app-worker/internal/repository"
	"github.com/IvanMeln1k/go-bank-app-worker/pkg/email"
	"github.com/google/uuid"
)

var (
	ErrInternal             = errors.New("error internal")
	ErrSendEmailMessage     = errors.New("error send email message")
	ErrMachineStatsNotFound = errors.New("error machine stats not found")
)

type Auth interface {
	SendEmailVerificationMessage(email string) error
}

type Machines interface {
	Cashout(ctx context.Context, id uuid.UUID, email string, accId uuid.UUID,
		amount int, newMoney int) error
	Deposit(ctx context.Context, id uuid.UUID, email string, accId uuid.UUID,
		amount int, newMoney int) error
}

type Service struct {
	Auth
	Machines
}

type Deps struct {
	Repos           repository.Repository
	EmailSender     email.EmailSender
	TokenManager    tokens.TokenManagerInterface
	VerificationURL string
}

func NewService(deps Deps) *Service {
	return &Service{
		Auth:     NewAuthService(deps.EmailSender, deps.TokenManager, deps.VerificationURL),
		Machines: NewMachinesService(deps.Repos.Machines, deps.EmailSender),
	}
}
