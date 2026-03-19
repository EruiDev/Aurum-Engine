package repository

import (
	"aurum/internal/domain"
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type PaymentRepository struct {
	db *sql.DB
}

func NewPaymentRepository(db *sql.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

func (r *PaymentRepository) Create(ctx context.Context, tx *sql.Tx, p *domain.Payment) error {
	_, err := tx.ExecContext(ctx, `
        INSERT INTO payments (id, idempotency_key, amount, currency, status, merchant_id, customer_id, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, p.ID, p.IdempotencyKey, p.Amount, p.Currency, p.Status, p.MerchantID, p.CustomerID, p.CreatedAt, p.UpdatedAt)
	return err
}

func (r *PaymentRepository) FindByIdempotencyKey(ctx context.Context, key string) (*domain.Payment, error) {
	p := &domain.Payment{}
	err := r.db.QueryRowContext(ctx, `
	SELECT * FROM payments WHERE idempotency_key = $1`, key).Scan(
		&p.ID, &p.IdempotencyKey, &p.Amount, &p.Currency, &p.Status, &p.MerchantID, &p.CustomerID, &p.CreatedAt, &p.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("findByIdempotencyKey: %w", err)
	}
	return p, nil
}
