package main

import (
	"aurum/internal/db"
	"aurum/internal/domain"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/google/uuid"
)

func main() {
	// Setting up the logger to be in JSON format -> might redirect to file
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Create db with SQL module
	database, err := db.New()
	if err != nil {
		os.Exit(1)
	}
	defer database.Conn.Close()
	database.SetParams(25, 10, 5*time.Minute)

	state, err := domain.NewPayment("test", 10, "EUR", uuid.New(), uuid.New())
	if err != nil {
		slog.Error("failed to open db", "err", err)
		os.Exit(1)
	}

	fmt.Println(state)
}
