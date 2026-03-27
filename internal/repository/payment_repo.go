package repository

import (
	"aurum/internal/domain"
	"aurum/internal/metrics"
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
	defer metrics.TrackDB("payment.create")()
	_, err := tx.ExecContext(ctx, `
        INSERT INTO payments (id, idempotency_key, amount, currency, status, merchant_id, customer_id, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, p.ID, p.IdempotencyKey, p.Amount, p.Currency, p.Status, p.MerchantID, p.CustomerID, p.CreatedAt, p.UpdatedAt)
	return err
}

func (r *PaymentRepository) FindByIdempotencyKey(ctx context.Context, key string) (*domain.Payment, error) {
	defer metrics.TrackDB("payment.find_by_idempotency_key")()
	p := &domain.Payment{}
	err := r.db.QueryRowContext(ctx, `
	SELECT id, idempotency_key, amount, currency, status, merchant_id, customer_id, created_at, updated_at
	FROM payments WHERE idempotency_key = $1`, key).Scan(
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

func (r *PaymentRepository) FindByID(ctx context.Context, id string) (*domain.Payment, error) {
	defer metrics.TrackDB("payment.find_by_id")()
	p := &domain.Payment{}
	err := r.db.QueryRowContext(ctx, `
	SELECT id, idempotency_key, amount, currency, status, merchant_id, customer_id, created_at, updated_at
	FROM payments WHERE id = $1`, id).Scan(
		&p.ID, &p.IdempotencyKey, &p.Amount, &p.Currency, &p.Status, &p.MerchantID, &p.CustomerID, &p.CreatedAt, &p.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("FindByID: %w", err)
	}
	return p, nil
}

func (r *PaymentRepository) FindByIDForUpdate(ctx context.Context, tx *sql.Tx, id string) (*domain.Payment, error) {
	defer metrics.TrackDB("payment.find_by_id_for_update")()
	p := &domain.Payment{}
	err := tx.QueryRowContext(ctx, `
	SELECT id, idempotency_key, amount, currency, status, merchant_id, customer_id, created_at, updated_at
	FROM payments WHERE id = $1 FOR UPDATE`, id).Scan(
		&p.ID, &p.IdempotencyKey, &p.Amount, &p.Currency, &p.Status, &p.MerchantID, &p.CustomerID, &p.CreatedAt, &p.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("FindByIDForUpdate: %w", err)
	}
	return p, nil
}

func (r *PaymentRepository) UpdateStatus(ctx context.Context, tx *sql.Tx, p *domain.Payment) error {
	defer metrics.TrackDB("payment.update_status")()
	result, err := tx.ExecContext(ctx, `
		UPDATE payments
		SET status = $1, updated_at = $2
		WHERE id = $3
	`, p.Status, p.UpdatedAt, p.ID)

	if err != nil {
		return fmt.Errorf("updateStatus: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("updateStatus rowsAffected: %w", err)
	}
	if rows == 0 {
		return domain.ErrNotFound
	}

	return nil
}
