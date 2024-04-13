package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IvanMeln1k/go-bank-app-bank/pkg/postgres"
	"github.com/IvanMeln1k/go-bank-app-bank/pkg/redisdb"
	"github.com/IvanMeln1k/go-bank-app-bank/pkg/tokens"
	"github.com/IvanMeln1k/go-bank-app-worker/internal/repository"
	"github.com/IvanMeln1k/go-bank-app-worker/internal/service"
	"github.com/IvanMeln1k/go-bank-app-worker/internal/workers"
	"github.com/IvanMeln1k/go-bank-app-worker/pkg/email"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {
	if err := initConfig(); err != nil {
		logrus.Fatalf("error initializing configs: %s", err)
	}

	if err := godotenv.Load(); err != nil {
		logrus.Fatalf("error loading env variables: %s", err)
	}

	db, err := postgres.NewPostgresDB(postgres.Config{
		Host:     viper.GetString("db.host"),
		Port:     viper.GetString("db.port"),
		User:     viper.GetString("db.user"),
		Password: os.Getenv("POSTGRES_PASS"),
		DBName:   viper.GetString("db.name"),
		SSLMode:  viper.GetString("db.sslmode"),
	})
	if err != nil {
		logrus.Fatalf("error connecting to db: %s", err)
	}

	rdb := redisdb.NewRedisDB(redisdb.Config{
		Host:     viper.GetString("rdb.host"),
		Port:     viper.GetString("rdb.port"),
		Password: os.Getenv("REDIS_PASS"),
		DB:       viper.GetInt("rdb.db"),
	})

	emailSender, err := email.NewSMTPSender(email.Config{
		Email: viper.GetString("smtp.email"),
		Pass:  os.Getenv("SMTP_PASS"),
		Host:  viper.GetString("smtp.host"),
		Port:  viper.GetString("smtp.port"),
	})

	if err != nil {
		logrus.Fatalf("error creating email sender: %s", err)
	}
	accessTTL, err := time.ParseDuration(viper.GetString("tokens.accessTTL"))
	if err != nil {
		logrus.Fatalf("error parsing accessTTL from config: %s", err)
	}
	emailTTL, err := time.ParseDuration(viper.GetString("tokens.emailTTL"))
	if err != nil {
		logrus.Fatalf("error parsing emailTTL from config: %s", err)
	}
	tokenManager := tokens.NewTokenManager(tokens.Config{
		SecretKey: os.Getenv("SECRET_KEY"),
		AccessTTL: accessTTL,
		EmailTTL:  emailTTL,
	})

	repos := repository.NewRepository(repository.Deps{
		DB: db,
	})

	services := service.NewService(service.Deps{
		Repos:           *repos,
		EmailSender:     emailSender,
		TokenManager:    tokenManager,
		VerificationURL: viper.GetString("verification.url"),
	})
	worker := workers.NewWorkers(workers.Deps{
		Rdb:      rdb,
		Services: *services,
	})

	ctx, cancel := context.WithCancel(context.Background())
	worker.Run(ctx, workers.Config{
		EmailVerificationSenderCnt: 1,
	})

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	logrus.Print("Server shutting down...")

	cancel()

	if err := db.Close(); err != nil {
		logrus.Fatalf("error clode db connect: %s", err)
	}
	if err := rdb.Close(); err != nil {
		logrus.Fatalf("error close redisdb connect: %s", err)
	}

	logrus.Print("Server stoped")
}

func initConfig() error {
	viper.AddConfigPath("configs")
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}
