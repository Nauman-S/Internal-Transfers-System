package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/Nauman-S/Internal-Transfers-System/api"
	"github.com/Nauman-S/Internal-Transfers-System/config"
	"github.com/Nauman-S/Internal-Transfers-System/storage"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.Info("Starting payments transfers service")

	appConfig := &config.ApplicationConfig{
		Ctx: context.Background(),
	}

	if err := initializeStorage(appConfig); err != nil {
		log.Fatalf("failed to initialize storage: %v", err)
	}

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", 8080),
		Handler:      api.InitRouter(appConfig),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Info("Shutting down server...")

		shutdownCtx, cancel := context.WithTimeout(appConfig.Ctx, 30*time.Second)
		defer cancel()

		appConfig.DB.Close()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Error("Error shutting down server:", err)
		}

		log.Info("Gracefully shutdown")
		os.Exit(0)
	}()

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Panicf("server listen and serve err: %s", err.Error())
	}

	log.Info("Server stopped")
}


func initializeStorage(appConfig *config.ApplicationConfig) error {
	dbPort, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		dbPort = 5432
	}

	db, err := storage.InitDB(appConfig.Ctx, storage.Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     dbPort,
		Database: os.Getenv("DB_NAME"),
		Username: os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		SSLMode:  os.Getenv("DB_SSL_MODE"),
	})

	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}

	if err := db.Start(appConfig.Ctx); err != nil {
		return fmt.Errorf("failed to start storage: %w", err)
	}

	appConfig.DB = db

	appConfig.AccountRepository = storage.NewAccountRepository(db)
	appConfig.TransferRepository = storage.NewTransferRepository(db)

	return nil
}