package service

import (
	"fmt"

	"github.com/IvanMeln1k/go-bank-app-bank/pkg/tokens"
	"github.com/IvanMeln1k/go-bank-app-worker/pkg/email"
	"github.com/sirupsen/logrus"
)

type AuthService struct {
	emailSender     email.EmailSender
	tokenManager    tokens.TokenManagerInterface
	verificationURL string
}

func NewAuthService(emailSender email.EmailSender,
	tokenManager tokens.TokenManagerInterface, verificationURL string) *AuthService {
	return &AuthService{
		emailSender:     emailSender,
		tokenManager:    tokenManager,
		verificationURL: verificationURL,
	}
}

func (s *AuthService) SendEmailVerificationMessage(email string) error {
	emailToken, err := s.tokenManager.CreateEmailToken(email)
	if err != nil {
		logrus.Errorf("error")
		return ErrInternal
	}

	err = s.emailSender.SendMessage("templates/email-verification.html",
		email, "Добро пожаловать в GO-BANK-APP", map[string]string{
			"Link": fmt.Sprintf("%s?token=%s", s.verificationURL, emailToken),
		})
	if err != nil {
		logrus.Errorf("error")
		return ErrSendEmailMessage
	}

	return nil
}
