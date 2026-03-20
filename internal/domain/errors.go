package domain

import "errors"

var (
	ErrNotFound            = errors.New("Payment doesn't exist")
	ErrInvalidTransition   = errors.New("Can't perform this transition")
	ErrInvalidAmount       = errors.New("Invalid amount, zero or negative not allowed")
	ErrInvalidCurrency     = errors.New("Invalid currency code, only ISO 4217 allowed")
	ErrIdempotencyConflict = errors.New("Idempotency conflict encountered, payloads are different")
	ErrInvalidImportency   = errors.New("Invalid impotency key")
	ErrInvalidMerchantID   = errors.New("Merchant ID was invalid")
	ErrInvalidCustomerID   = errors.New("Customer ID was invalid")
)
