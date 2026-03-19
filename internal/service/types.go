package service

import "github.com/google/uuid"

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
