package repository

import (
	"aurum/internal/domain"
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
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

func (r *OutboxRepository) FetchUnpublished(ctx context.Context, amount int) ([]domain.OutboxEvent, error) {
	// Have to add transaction for Locked
	rows, err := r.db.QueryContext(ctx, `
	SELECT id, aggregate_id, event_type, payload, created_at
	FROM outbox_events WHERE published = false
	ORDER BY created_at ASC
	FOR UPDATE SKIP LOCKED
	LIMIT $1`, amount)
	if err != nil {
		return nil, fmt.Errorf("outbox_repository: fetch published %w", err)
	}
	defer rows.Close()

	var events []domain.OutboxEvent
	for rows.Next() {
		var curr domain.OutboxEvent
		err = rows.Scan(
			&curr.ID,
			&curr.AggregateID,
			&curr.EventType,
			&curr.Payload,
			&curr.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("outbox_repository: scan event: %w", err)
		}
		events = append(events, curr)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("outbox_repository: rows error: %w", err)
	}
	return events, nil
}

func (r *OutboxRepository) MarkPublished(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `
	UPDATE outbox_events
	SET published = true,
		published_at = NOW()
	WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("outbox_repository: error marking as published: %w", err)
	}
	return nil
}
