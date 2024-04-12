package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/IvanMeln1k/go-bank-app-worker/internal/domain"
	"github.com/IvanMeln1k/go-bank-app-worker/internal/service"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

var (
	emailVerificationQueue = "queue:verification:email"
	cashoutQueue           = "queue:cashout"
	depositQueue           = "queue:deposit"
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

func (w *Workers) DepositHandler(ctx context.Context, id int) {
	ticker := time.NewTicker(5 * time.Second)
LOOP:
	for {
	SELECT:
		select {
		case <-ctx.Done():
			ticker.Stop()
			break LOOP
		case <-ticker.C:
			for {
				task, err := w.rdb.LPop(ctx, depositQueue).Result()
				if err != nil {
					if err.Error() == "redis: nil" {
						break SELECT
					}
					logrus.Errorf("error worker deposit #%d getting task from rdb: %s", id, err)
					break SELECT
				}
				if task == "" {
					break SELECT
				}
				var depositTask domain.DepositTask
				err = json.Unmarshal([]byte(task), &depositTask)
				if err != nil {
					logrus.Errorf("error worker cashout #%d unmarshaling json: %s",
						id, err)
					break SELECT
				}
				err = w.services.Deposit(ctx, depositTask.MachineId, depositTask.Email,
					depositTask.AccId, depositTask.Amount, depositTask.NewMoney)
				if err != nil {
					logrus.Errorf("error deposit service deposit worker #%d: %s", id, err)
					break SELECT
				}
			}
		}
	}
	fmt.Printf("worker deposit #%d stopped", id)
}

func (w *Workers) CashoutHandler(ctx context.Context, id int) {
	ticker := time.NewTicker(5 * time.Second)
LOOP:
	for {
	SELECT:
		select {
		case <-ctx.Done():
			ticker.Stop()
			break LOOP
		case <-ticker.C:
			for {
				task, err := w.rdb.LPop(ctx, cashoutQueue).Result()
				if err != nil {
					if err.Error() == "redis: nil" {
						break SELECT
					}
					logrus.Errorf("error worker cashout #%d getting task from rdb: %s", id, err)
					break SELECT
				}
				var cashoutTask domain.CashoutTask
				err = json.Unmarshal([]byte(task), &cashoutTask)
				if err != nil {
					logrus.Errorf("error worker cashout #%d unmarshaling json: %s",
						id, err)
					break SELECT
				}
				err = w.services.Cashout(ctx, cashoutTask.MachineId, cashoutTask.Email,
					cashoutTask.AccId, cashoutTask.Amount, cashoutTask.NewMoney)
				if err != nil {
					logrus.Errorf("error cashout service cashout worker #%d: %s", id, err)
					break SELECT
				}
			}
		}
	}
	fmt.Printf("worker cashout #%d stopped", id)
}

func (w *Workers) EmailVerificationSender(ctx context.Context, emailSenderId int) {
	ticker := time.NewTicker(5 * time.Second)
LOOP:
	for {
	SELECT:
		select {
		case <-ctx.Done():
			ticker.Stop()
			break LOOP
		case <-ticker.C:
			for {
				task, err := w.rdb.LPop(ctx, emailVerificationQueue).Result()
				if err != nil {
					if err.Error() == "redis: nil" {
						break SELECT
					}
					logrus.Errorf("error worker email verificatoin sender #%d getting task from rdb: %s",
						emailSenderId, err)
					break SELECT
				}
				if task == "" {
					break SELECT
				}
				var emailVerificationTask domain.EmailVerificationTask
				err = json.Unmarshal([]byte(task), &emailVerificationTask)
				if err != nil {
					logrus.Errorf("error worker email verification sender #%d unmarshaling json: %s",
						emailSenderId, err)
					break SELECT
				}
				err = w.services.SendEmailVerificationMessage(emailVerificationTask.Email)
				if err != nil {
					logrus.Errorf("error worker email verification sender #%d sending email verification message: %s", emailSenderId, err)
					break SELECT
				}
			}
		}
	}
	fmt.Printf("worker email verification sender #%d stopped", emailSenderId)
}
