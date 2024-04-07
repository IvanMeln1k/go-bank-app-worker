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
			// logrus.Printf("Worker email verification sender #%d start do task at %s", emailSenderId, time)
			tasksLen, err := w.rdb.LLen(ctx, "queue:verification:email").Result()
			if err != nil {
				logrus.Errorf("error worker email verification sender #%d getting tasksLen from rdb: %s", emailSenderId, err)
				break SELECT
			}
			if tasksLen == 0 {
				break SELECT
			}
			for i := 0; i < int(tasksLen); i++ {
				task, err := w.rdb.LPop(ctx, "queue:verification:email").Result()
				if err != nil {
					logrus.Errorf("error worker email verificatoin sender #%d getting task from rdb: %s", emailSenderId, err)
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
