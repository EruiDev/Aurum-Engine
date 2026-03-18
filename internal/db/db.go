package db

import (
	"database/sql"
	"log/slog"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

type DB struct {
	conn *sql.DB
}

func New() (*DB, error) {
	if err := godotenv.Load(); err != nil {
		slog.Warn("no .env file found, using environment variables")
	} // TODO when changed into docker just get it from env

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

	return &DB{conn: conn}, nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}

func (db *DB) Conn() *sql.DB {
	return db.conn
}

func (db *DB) SetParams(maxOpenConns int, maxIdelConns int, connMaxLifetime time.Duration) {
	db.conn.SetMaxOpenConns(maxOpenConns)
	db.conn.SetMaxIdleConns(maxIdelConns)
	db.conn.SetConnMaxLifetime(connMaxLifetime)
}
