package service

import (
	"aurum/internal/db"
	"aurum/internal/domain"
	"aurum/internal/metrics"
	"aurum/internal/repository"
	"context"
	"database/sql"
	"errors"
	"strings"
)

type PaymentService struct {
	db          *db.DB
	repo        *repository.PaymentRepository
	outbox_repo *repository.OutboxRepository
}

func NewPaymentService(
	db *db.DB,
	payments *repository.PaymentRepository,
	outbox *repository.OutboxRepository,
) *PaymentService {
	return &PaymentService{db: db, repo: payments, outbox_repo: outbox}
}

func (s *PaymentService) CreatePayment(ctx context.Context, req CreatePaymentRequest) (*domain.Payment, error) {
	p, err := s.repo.FindByIdempotencyKey(ctx, req.IdempotencyKey)

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
		if err := s.repo.Create(ctx, tx, p); err != nil {
			return err
		}
		if err := s.outbox_repo.Insert(ctx, tx, event); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	metrics.PaymentsCreatedTotal.WithLabelValues(p.Currency).Inc()
	return p, nil
}

func (s *PaymentService) GetPayment(ctx context.Context, id string) (*domain.Payment, error) {
	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (s *PaymentService) CapturePayment(ctx context.Context, id string) (*domain.Payment, error) {
	return s.transition(ctx, domain.StatusCaptured, id)
}

func (s *PaymentService) AuthorizePayment(ctx context.Context, id string) (*domain.Payment, error) {
	return s.transition(ctx, domain.StatusAuthorized, id)
}

func (s *PaymentService) VoidPayment(ctx context.Context, id string) (*domain.Payment, error) {
	return s.transition(ctx, domain.StatusVoided, id)
}

func (s *PaymentService) RefundPayment(ctx context.Context, id string) (*domain.Payment, error) {
	return s.transition(ctx, domain.StatusRefunded, id)
}

func (s *PaymentService) SettlePayment(ctx context.Context, id string) (*domain.Payment, error) {
	return s.transition(ctx, domain.StatusSettled, id)
}

func (s *PaymentService) transition(ctx context.Context, to domain.PaymentStatus, id string) (*domain.Payment, error) {
	var payment *domain.Payment
	var prevStatus domain.PaymentStatus

	err := s.db.WithTransaction(ctx, func(tx *sql.Tx) error {
		var err error
		payment, err = s.repo.FindByIDForUpdate(ctx, tx, id)
		if err != nil {
			return err
		}
		prevStatus = payment.Status

		if err = payment.TransitionTo(to); err != nil {
			return err
		}

		if err = s.repo.UpdateStatus(ctx, tx, payment); err != nil {
			return err
		}

		event, err := domain.NewOutboxEvent(payment, "payment."+strings.ToLower(string(to)))
		if err != nil {
			return err
		}

		return s.outbox_repo.Insert(ctx, tx, event)
	})

	if err != nil {
		return nil, err
	}

	metrics.PaymentTransitionTotal.WithLabelValues(string(prevStatus), string(to)).Inc()

	return payment, nil
}
