package repository

import (
	"aurum/internal/domain"
	"context"
	"database/sql"
)

type PaymentRepositoy struct {
	db *sql.DB
}

func NewPaymentRepository(db *sql.DB) *PaymentRepositoy {
	return &PaymentRepositoy{db: db}
}

func (r *PaymentRepositoy) Create(ctx context.Context, tx *sql.Tx, p *domain.Payment) error {
	_, err := tx.ExecContext(ctx, `
        INSERT INTO payments (id, idempotency_key, amount, currency, status, merchant_id, customer_id, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, p.ID, p.IdempotencyKey, p.Amount, p.Currency, p.Status, p.MerchantID, p.CustomerID, p.CreatedAt, p.UpdatedAt)
	return err
}
