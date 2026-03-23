package domain

import "errors"

var (
	ErrNotFound            = errors.New("payment doesn't exist")
	ErrInvalidTransition   = errors.New("can't perform this transition")
	ErrInvalidAmount       = errors.New("invalid amount, zero or negative not allowed")
	ErrInvalidCurrency     = errors.New("invalid currency code, only ISO 4217 allowed")
	ErrIdempotencyConflict = errors.New("idempotency conflict encountered, payloads are different")
	ErrInvalidImportency   = errors.New("invalid impotency key")
	ErrInvalidMerchantID   = errors.New("merchant ID was invalid")
	ErrInvalidCustomerID   = errors.New("customer ID was invalid")
)
