package main

import (
	"aurum/internal/db"
	"aurum/internal/handler"
	"aurum/internal/repository"
	"aurum/internal/service"
	"log/slog"
	"net/http"
	"os"
	"time"
)

func main() {
	// Setting up the logger to be in JSON format -> might redirect to file
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	database, err := db.New()
	if err != nil {
		os.Exit(1)
	}
	defer database.Close()
	database.SetParams(25, 10, 5*time.Minute)
	err = database.RunMigrations()
	if err != nil {
		slog.Error("migration failed: ", "err", err.Error())
		os.Exit(1)
	}

	paymentRepo := repository.NewPaymentRepository(database.Conn())
	outboxRepo := repository.NewOutboxRepository(database.Conn())
	paymentService := service.NewPaymentService(database, paymentRepo, outboxRepo)
	paymentHandler := handler.NewHandler(paymentService)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /payments", paymentHandler.CreatePayment)
	mux.HandleFunc("GET /payments/{id}", paymentHandler.GetPayment)

	server := &http.Server{
		Addr:         ":" + "8080", // TODO create fallback for env file
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Create proper shutdown of server with signal handling
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("server failed to start", "err", err)
		os.Exit(1)
	}
}
