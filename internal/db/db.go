package db

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"time"

	"embed"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

//go:embed migrations/*.sql
var migrations embed.FS

type DB struct {
	conn *sql.DB
}

func New() (*DB, error) {
	if err := godotenv.Load(); err != nil {
		slog.Info("no .env file found, using environment variables")
	}

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

func (db *DB) Ping(ctx context.Context) error {
	return db.conn.PingContext(ctx)
}

func (db *DB) SetParams(maxOpenConns int, maxIdelConns int, connMaxLifetime time.Duration) {
	db.conn.SetMaxOpenConns(maxOpenConns)
	db.conn.SetMaxIdleConns(maxIdelConns)
	db.conn.SetConnMaxLifetime(connMaxLifetime)
}

func (db *DB) WithTransaction(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := db.conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (db *DB) RunMigrations() error {
	entries, err := migrations.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	for _, entry := range entries {
		sql, err := migrations.ReadFile("migrations/" + entry.Name())
		if err != nil {
			return fmt.Errorf("read migration %s: %w", entry.Name(), err)
		}

		if _, err := db.conn.Exec(string(sql)); err != nil {
			return fmt.Errorf("run migration %s: %w", entry.Name(), err)
		}

		slog.Info("migration applied", "file", entry.Name())
	}

	return nil
}
