package workers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/IvanMeln1k/go-bank-app-worker/internal/domain"
	"github.com/IvanMeln1k/go-bank-app-worker/internal/service"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

const (
	emailVerificationQueue = "queue:verification:email"
	cashoutQueue           = "queue:cashout"
	depositQueue           = "queue:deposit"
)

var (
	ErrGetTask = errors.New("error get task")
	ErrDoTask  = errors.New("error do task")
)

type Workers struct {
	rdb      *redis.Client
	services service.Service
}

type Deps struct {
	Rdb      *redis.Client
	Services service.Service
}

type Config struct {
	EmailVerificationSenderCnt int
}

func NewWorkers(deps Deps) *Workers {
	return &Workers{
		rdb:      deps.Rdb,
		services: deps.Services,
	}
}

func (w *Workers) Run(ctx context.Context, cfg Config) {
	for i := 1; i <= cfg.EmailVerificationSenderCnt; i++ {
		go w.EmailVerificationSender(ctx, i)
	}
	go w.CashoutHandler(ctx, 1)
	go w.DepositHandler(ctx, 1)
}

func (w *Workers) getTask(ctx context.Context, key string, scanObj interface{}) error {
	task, err := w.rdb.LPop(ctx, key).Result()
	if err != nil {
		if err.Error() == "redis: nil" {
			return ErrGetTask
		}
		logrus.Errorf("error getting task from redis by key %s", key)
		return ErrGetTask
	}
	err = json.Unmarshal([]byte(task), scanObj)
	if err != nil {
		logrus.Errorf("error unmarshalling json task by key %s", key)
		return ErrGetTask
	}
	return nil
}

func (w *Workers) DepositHandler(ctx context.Context, id int) {
	ticker := time.NewTicker(5 * time.Second)
LOOP:
	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			break LOOP
		case <-ticker.C:
			for {
				w.doDepositTask(ctx)
			}
		}
	}
	fmt.Printf("worker deposit #%d stopped", id)
}

func (w *Workers) doDepositTask(ctx context.Context) error {
	var depositTask domain.DepositTask
	err := w.getTask(ctx, depositQueue, &depositTask)
	if err != nil {
		return err
	}
	err = w.services.Deposit(ctx, depositTask.MachineId, depositTask.Email,
		depositTask.AccId, depositTask.Amount, depositTask.NewMoney)
	if err != nil {
		logrus.Errorf("error deposit service deposit worker: %s", err)
		return ErrDoTask
	}
	return nil
}

func (w *Workers) CashoutHandler(ctx context.Context, id int) {
	ticker := time.NewTicker(5 * time.Second)
LOOP:
	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			break LOOP
		case <-ticker.C:
			for {
				w.doCashoutTask(ctx)
			}
		}
	}
	logrus.Printf("worker cashout #%d stopped", id)
}

func (w *Workers) doCashoutTask(ctx context.Context) error {
	var cashoutTask domain.CashoutTask
	err := w.getTask(ctx, cashoutQueue, &cashoutTask)
	if err != nil {
		return err
	}
	err = w.services.Cashout(ctx, cashoutTask.MachineId, cashoutTask.Email,
		cashoutTask.AccId, cashoutTask.Amount, cashoutTask.NewMoney)
	if err != nil {
		logrus.Errorf("error cashout service cashout worker: %s", err)
		return ErrDoTask
	}
	return nil
}

func (w *Workers) EmailVerificationSender(ctx context.Context, emailSenderId int) {
	ticker := time.NewTicker(5 * time.Second)
LOOP:
	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			break LOOP
		case <-ticker.C:
			for {
				w.doEmailVerificationSenderTask(ctx)
			}
		}
	}
}

func (w *Workers) doEmailVerificationSenderTask(ctx context.Context) error {
	var emailVerificationTask domain.EmailVerificationTask
	err := w.getTask(ctx, emailVerificationQueue, &emailVerificationTask)
	if err != nil {
		return err
	}
	err = w.services.SendEmailVerificationMessage(emailVerificationTask.Email)
	if err != nil {
		logrus.Errorf("error worker email verification sender sending email verification message: %s", err)
		return ErrDoTask
	}
	return nil
}
