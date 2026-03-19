package repository

import (
	"aurum/internal/domain"
	"context"
	"database/sql"
	"fmt"
)

type OutboxRepository struct {
	db *sql.DB
}

func NewOutboxRepository(db *sql.DB) *OutboxRepository {
	return &OutboxRepository{db: db}
}

func (r *OutboxRepository) Insert(ctx context.Context, tx *sql.Tx, event *domain.OutboxEvent) error {
	_, err := tx.ExecContext(ctx, `
	INSERT INTO outbox_events (id, aggregate_id, event_type, payload, published, created_at)
	VALUES ($1, $2, $3, $4, $5, $6)`,
		event.ID, event.AggregateID, event.EventType, event.Payload, event.Published, event.CreatedAt)
	if err != nil {
		return fmt.Errorf("outbox insert: %w", err)
	}
	return nil
}
