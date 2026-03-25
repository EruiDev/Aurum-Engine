package worker

import (
	"aurum/internal/publisher"
	"aurum/internal/repository"
	"context"
	"log/slog"
	"time"
)

type OutboxWorker struct {
	outbox    *repository.OutboxRepository
	publisher *publisher.KafkaPublisher
}

func NewOutboxWorker(
	outbox *repository.OutboxRepository,
	publisher *publisher.KafkaPublisher,
) *OutboxWorker {
	return &OutboxWorker{
		outbox:    outbox,
		publisher: publisher,
	}
}

func (w *OutboxWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("outbox worker shutting down")
			return
		case <-ticker.C:
			w.processEvents(ctx)
		}
	}
}

func (w *OutboxWorker) processEvents(ctx context.Context) {
	batchCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	events, err := w.outbox.FetchUnpublished(batchCtx, 100)

	if err != nil {
		slog.Error("outbox: failed to fetch events", "error", err)
		return
	}

	for _, event := range events {
		if err = w.publisher.Publish(batchCtx, event); err != nil {
			slog.Error("error publishing event on kafka:",
				"err", err,
				"event_id", event.ID,
			)
			continue
		}
		if err = w.outbox.MarkPublished(batchCtx, event.ID); err != nil {
			slog.Error("error marking as published event on db:",
				"err", err,
				"event_id", event.ID,
			)
			// Sadly we can't really do much about the error here, outbox pattern only problem
		}
	}
}
