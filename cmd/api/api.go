package main

import (
	"aurum/internal/db"
	"aurum/internal/handler"
	"aurum/internal/repository"
	"aurum/internal/service"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func healthHandler(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := "ok"
		dbStatus := "ok"
		code := http.StatusOK

		if err := database.Ping(r.Context()); err != nil {
			status = "degraded"
			dbStatus = "unreachable"
			code = http.StatusServiceUnavailable
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		json.NewEncoder(w).Encode(struct {
			Status string `json:"status"`
			DB     string `json:"db"`
		}{
			Status: status,
			DB:     dbStatus,
		})
	}
}

func getPort() string {
	if port := os.Getenv("PORT"); port != "" {
		return port
	}
	return "8080"
}

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
	mux.HandleFunc("POST /payments/{id}/{action}", paymentHandler.TransitionPayment)
	// mux.HandleFunc("GET /payments",               )
	mux.HandleFunc("GET /health", healthHandler(database))
	// mux.HandleFunc("GET /metrics",				 ) Might add for Prometheus handling, potential graphana

	server := &http.Server{
		Addr:         ":" + getPort(),
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("server starting", "port", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server failed to start", "err", err)
			os.Exit(1)
		}
	}()

	sig := <-quit
	slog.Info("shutdown signal received", "signal", sig)
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("server shutdown failed", "err", err)
	}

	slog.Info("shutdown complete")
}
