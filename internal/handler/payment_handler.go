package handler

import (
	"aurum/internal/domain"
	"aurum/internal/service"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

type Handler struct {
	service *service.PaymentService
}

func NewHandler(service *service.PaymentService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreatePayment(w http.ResponseWriter, r *http.Request) {
	var req service.CreatePaymentRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "request body is invalid")
		return
	}

	req.IdempotencyKey = r.Header.Get("Idempotency-Key")
	if req.IdempotencyKey == "" {
		h.writeError(w, http.StatusBadRequest, "missing_idempotency_key", "Idempotency-Key header is required")
		return
	}
	res, err := h.service.CreatePayment(r.Context(), req)

	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidAmount):
			h.writeError(w, http.StatusUnprocessableEntity, "invalid_amount", err.Error())
		case errors.Is(err, domain.ErrInvalidCurrency):
			h.writeError(w, http.StatusUnprocessableEntity, "invalid_currency", err.Error())
		case errors.Is(err, domain.ErrInvalidCustomerID):
			h.writeError(w, http.StatusUnprocessableEntity, "invalid_customer_id", err.Error())
		case errors.Is(err, domain.ErrInvalidMerchantID):
			h.writeError(w, http.StatusUnprocessableEntity, "invalid_merchant_id", err.Error())
		default:
			slog.Error("unexpected error creating payment", "err", err, "path", r.URL.Path, "method", r.Method)
			h.writeError(w, http.StatusInternalServerError, "interal_error", "unexpected error")
		}
		return
	}

	h.writeJSON(w, http.StatusCreated, res)
	slog.Info("new payment created with idempotency key:", "idempotency_key:", res.IdempotencyKey)
}

func (h *Handler) GetPayment(w http.ResponseWriter, r *http.Request) {
	id := (r.PathValue("id"))
	err := uuid.Validate(id)

	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_id", "id is not a valid UUID")
		return
	}

	p, err := h.service.GetPayment(r.Context(), id)

	h.writeJSON(w, http.StatusOK, p)
}

// Helper functions

func (h *Handler) writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func (h *Handler) writeError(w http.ResponseWriter, status int, code, message string) {
	h.writeJSON(w, status, service.ErrorResponse{
		Code:    code,
		Message: message,
	})
}
