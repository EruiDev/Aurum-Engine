package service

import (
	"aurum/internal/db"
	"aurum/internal/domain"
	"aurum/internal/repository"
	"context"
	"database/sql"
	"errors"
)

type PaymentService struct {
	db       *db.DB
	payments *repository.PaymentRepository
	outbox   *repository.OutboxRepository
}

func NewPaymentService(
	db *db.DB,
	payments *repository.PaymentRepository,
	outbox *repository.OutboxRepository,
) *PaymentService {
	return &PaymentService{db: db, payments: payments, outbox: outbox}
}

func (s *PaymentService) CreatePayment(ctx context.Context, req CreatePaymentRequest) (*domain.Payment, error) {
	p, err := s.payments.FindByIdempotencyKey(ctx, req.IdempotencyKey)

	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return nil, err
	}
	if p != nil {
		return p, nil
	}
	p, err = domain.NewPayment(
		req.IdempotencyKey,
		req.Amount,
		req.Currency,
		req.MerchantID,
		req.CustomerID,
	)
	if err != nil {
		return nil, err
	}

	event, err := domain.NewOutboxEvent(p, "payment.initiated")

	if err != nil {
		return nil, err
	}

	err = s.db.WithTransaction(ctx, func(tx *sql.Tx) error {
		if err := s.payments.Create(ctx, tx, p); err != nil {
			return err
		}
		if err := s.outbox.Insert(ctx, tx, event); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return p, nil
}
