package service

import (
	"time"

	"github.com/google/uuid"
)

type CreatePaymentRequest struct {
	IdempotencyKey string    `json:"-"` // comes from header, not body
	Amount         int64     `json:"amount"`
	Currency       string    `json:"currency"`
	MerchantID     uuid.UUID `json:"merchant_id"`
	CustomerID     uuid.UUID `json:"customer_id"`
}

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type PaymentResponse struct {
	ID         uuid.UUID `json:"id"`
	Amount     int64     `json:"amount"`
	Currency   string    `json:"currency"`
	Status     string    `json:"status"`
	MerchantID uuid.UUID `json:"merchant_id"`
	CustomerID uuid.UUID `json:"customer_id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type ListPaymentsResponse struct {
	Data   []*PaymentResponse `json:"data"`
	Total  int                `json:"total"`
	Limit  int                `json:"limit"`
	Offset int                `json:"offset"`
}
