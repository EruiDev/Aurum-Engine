package domain

import "errors"

var ErrNotFound = errors.New("Payment doesn't exist")
var ErrInvalidTransition = errors.New("Can't perform this transition")
var ErrInvalidAmount = errors.New("Invalid amount, zero or negative not allowed")
var ErrInvalidCurrency = errors.New("Invalid currency code, only ISO 4217 allowed")
var ErrIdempotencyConflict = errors.New("Idempotency conflict encountered, payloads are different")
