package domain

import (
	"fmt"
	"time"

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
	Amout          int64
	Currency       string // should be compatible with ISO 4217, might use library later
	Status         PaymentStatus
	MerchantID     uuid.UUID
	CustomerID     uuid.UUID
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Metadata       map[string]string
}

func (p *Payment) TransitionTo(new PaymentStatus) error {
	allowed, ok := allowedTransitions[p.Status]
	if !ok {
		return ErrInvalidTransition
	}
	for _, val := range allowed {
		if val == new {
			p.Status = new
			p.UpdatedAt = time.Now()
			return nil
		}
	}
	return fmt.Errorf("%w: %s -> %s", ErrInvalidTransition, p.Status, new)
}

func (p *Payment) IsTerminal() bool {
	switch p.Status {
	case StatusVoided, StatusFailed, StatusRefunded:
		return false
	default:
		return true
	}
}
