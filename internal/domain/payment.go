package domain

import (
	"encoding/json"
	"fmt"
	"time"

	currencies "github.com/bojanz/currency"
	"github.com/google/uuid"
)

type PaymentStatus string

const ( // main statuses
	StatusInitiated  PaymentStatus = "INITIATED"
	StatusAuthorized PaymentStatus = "AUTHORIZED"
	StatusCaptured   PaymentStatus = "CAPTURED"
	StatusSettled    PaymentStatus = "SETTLED"
	StatusFailed     PaymentStatus = "FAILED"   // could happen on authorization
	StatusVoided     PaymentStatus = "VOIDED"   // could happen on capture
	StatusRefunded   PaymentStatus = "REFUNDED" // could happen after capture
)

var allowedTransitions = map[PaymentStatus][]PaymentStatus{
	StatusInitiated:  {StatusAuthorized, StatusFailed},
	StatusAuthorized: {StatusCaptured, StatusVoided},
	StatusCaptured:   {StatusSettled, StatusRefunded},
	StatusSettled:    {StatusRefunded},
}

type Payment struct {
	ID             uuid.UUID
	IdempotencyKey string
	Amount         int64
	Currency       string // should be compatible with ISO 4217, might use library later
	Status         PaymentStatus
	MerchantID     uuid.UUID
	CustomerID     uuid.UUID
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Metadata       map[string]string
}

type OutboxEvent struct {
	ID          uuid.UUID
	AggregateID uuid.UUID
	EventType   string
	Payload     []byte
	Published   bool
	CreatedAt   time.Time
	PublishedAt *time.Time
}

func NewOutboxEvent(payment *Payment, eventType string) (*OutboxEvent, error) {
	payload, err := json.Marshal(payment)
	if err != nil {
		return nil, err
	}

	return &OutboxEvent{
		ID:          uuid.New(),
		AggregateID: payment.ID,
		EventType:   eventType,
		Payload:     payload,
		Published:   false,
		CreatedAt:   time.Now(),
	}, nil
}

func (p *Payment) TransitionTo(to PaymentStatus) error {
	allowed, ok := allowedTransitions[p.Status]
	if !ok {
		return ErrInvalidTransition
	}
	for _, val := range allowed {
		if val == to {
			p.Status = to
			p.UpdatedAt = time.Now()
			return nil
		}
	}
	return fmt.Errorf("%w: %s -> %s", ErrInvalidTransition, p.Status, to)
}

func (p *Payment) IsTerminal() bool {
	switch p.Status {
	case StatusVoided, StatusFailed, StatusRefunded:
		return true
	default:
		return false
	}
}

func NewPayment(
	idempotencyKey string,
	amout int64,
	currency string,
	merchantID uuid.UUID,
	customerID uuid.UUID,
) (*Payment, error) {
	if idempotencyKey == "" {
		return nil, ErrInvalidImportency
	}
	ok := currencies.IsValid(currency)
	if !ok {
		return nil, ErrInvalidCurrency
	}
	if amout <= 0 {
		return nil, ErrInvalidAmount
	}
	time := time.Now()
	return &Payment{
		ID:             uuid.New(),
		IdempotencyKey: idempotencyKey,
		Amount:         amout,
		Currency:       currency,
		Status:         StatusInitiated,
		MerchantID:     merchantID,
		CustomerID:     customerID,
		CreatedAt:      time,
		UpdatedAt:      time,
		Metadata:       make(map[string]string),
	}, nil
}
