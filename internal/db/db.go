package db

import (
	"database/sql"
	"log/slog"
	"os"
	"time"
)

type DB struct {
	Conn *sql.DB
}

func New() (*DB, error) {
	conn, err := sql.Open("pgx", os.Getenv("DATABASE_URL"))
	if err != nil {
		slog.Error("failed to open db", "err", err)
		return nil, err
	}

	err = conn.Ping()
	if err != nil {
		slog.Error("failed to ping db", "err", err)
		conn.Close()
		return nil, err
	}

	return &DB{Conn: conn}, nil
}

func (db *DB) SetParams(maxOpenConns int, maxIdelConns int, connMaxLifetime time.Duration) {
	db.Conn.SetMaxOpenConns(maxOpenConns)
	db.Conn.SetMaxIdleConns(maxIdelConns)
	db.Conn.SetConnMaxLifetime(connMaxLifetime)
}
