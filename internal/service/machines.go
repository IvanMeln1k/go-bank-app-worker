package service

import (
	"context"
	"errors"

	"github.com/IvanMeln1k/go-bank-app-worker/internal/domain"
	"github.com/IvanMeln1k/go-bank-app-worker/internal/repository"
	"github.com/IvanMeln1k/go-bank-app-worker/pkg/email"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type MachinesService struct {
	machinesRepo repository.Machines
	emailSender  email.EmailSender
}

func NewMachinesService(machinesRepo repository.Machines, emailSender email.EmailSender) *MachinesService {
	return &MachinesService{
		machinesRepo: machinesRepo,
		emailSender:  emailSender,
	}
}

func (r *MachinesService) Cashout(ctx context.Context, id uuid.UUID, email string, accId uuid.UUID,
	amount int, newMoney int) error {
	err := r.emailSender.SendMessage("templates/cashout.html",
		email, "Обналичивание средств в GO-BANK-APP!", map[string]interface{}{
			"Id":     accId,
			"Amount": amount,
			"Money":  newMoney,
		})
	if err != nil {
		logrus.Errorf("error sending email when cashout: %s", err)
	}

	stats, err := r.machinesRepo.GetStats(ctx, id)
	if err != nil {
		logrus.Errorf("error getting stats from machines repo: %s", err)
		if errors.Is(repository.ErrMachinesStatsNotFound, err) {
			return ErrMachineStatsNotFound
		}
		return ErrInternal
	}

	newCashout := stats.Cashout + amount
	newStats := domain.MachineStatsUpdate{
		Cashout: &newCashout,
	}
	_, err = r.machinesRepo.UpdateStats(ctx, id, newStats)
	if err != nil {
		logrus.Errorf("error updating machine stats when cashout: %s", err)
		return ErrInternal
	}

	return nil
}

func (r *MachinesService) Deposit(ctx context.Context, id uuid.UUID, email string, accId uuid.UUID,
	amount int, newMoney int) error {
	err := r.emailSender.SendMessage("templates/deposit.html",
		email, "Пополнение счета в GO-BANK-APP!", map[string]interface{}{
			"Id":     accId,
			"Amount": amount,
			"Money":  newMoney,
		})
	if err != nil {
		logrus.Errorf("error sending email when deposit: %s", err)
	}

	stats, err := r.machinesRepo.GetStats(ctx, id)
	if err != nil {
		logrus.Errorf("error getting stats from machines repo: %s", err)
		if errors.Is(repository.ErrMachinesStatsNotFound, err) {
			return ErrMachineStatsNotFound
		}
		return ErrInternal
	}

	newDeposit := stats.Deposit + amount
	newStats := domain.MachineStatsUpdate{
		Deposit: &newDeposit,
	}
	_, err = r.machinesRepo.UpdateStats(ctx, id, newStats)
	if err != nil {
		logrus.Errorf("error updating machine stats when deposit: %s", err)
		return ErrInternal
	}

	return nil
}
