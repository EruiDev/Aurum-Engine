package handler

import (
	"aurum/internal/domain"
	"aurum/internal/service"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

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
			slog.Error("unexpected error creating payment", "err", err, "path", sanitize(r.URL.Path), "method", r.Method)
			h.writeError(w, http.StatusInternalServerError, "interal_error", "unexpected error")
		}
		return
	}

	h.writeJSON(w, http.StatusCreated, res)
	slog.Info("new payment created with idempotency key:", "idempotency_key:", res.IdempotencyKey)
}

func (h *Handler) GetPayment(w http.ResponseWriter, r *http.Request) {
	id, err := h.validateID(w, r)
	if err != nil {
		return
	}

	p, err := h.service.GetPayment(r.Context(), id)

	if errors.Is(err, domain.ErrNotFound) {
		h.writeError(w, http.StatusNotFound, "id_not_found", "the id provided doesn't have an attached payment")
		return
	} else if err != nil {
		slog.Error("unexpected error getting payment", "err", err)
		h.writeError(w, http.StatusInternalServerError, "interal_error", "unexpected error")
	}

	h.writeJSON(w, http.StatusOK, p)
}

func (h *Handler) TransitionPayment(w http.ResponseWriter, r *http.Request) {
	id, err := h.validateID(w, r)
	if err != nil {
		return
	}

	action := r.PathValue("action")

	var payment *domain.Payment

	switch action {
	case "authorize":
		payment, err = h.service.AuthorizePayment(r.Context(), id)
	case "capture":
		payment, err = h.service.CapturePayment(r.Context(), id)
	case "void":
		payment, err = h.service.VoidPayment(r.Context(), id)
	//case "refund":
	//payment, err = h.service.RefundPayment(r.Context(), id)
	default:
		h.writeError(w, http.StatusNotFound, "unknown_action", "unknown action: "+action)
		return
	}

	if err != nil {
		switch {
		case errors.Is(err, domain.ErrNotFound):
			h.writeError(w, http.StatusNotFound, "not_found", "payment not found")
		case errors.Is(err, domain.ErrInvalidTransition):
			h.writeError(w, http.StatusUnprocessableEntity, "invalid_transition", err.Error())
		default:
			slog.Error("failed to transition payment", "err", err, "action", sanitize(action), "id", sanitize(id))
			h.writeError(w, http.StatusInternalServerError, "internal_error", "unexpected error")
		}
		return
	}

	h.writeJSON(w, http.StatusOK, payment)
}

// Helper functions

func (h *Handler) validateID(w http.ResponseWriter, r *http.Request) (string, error) {
	id := (r.PathValue("id"))
	err := uuid.Validate(id)

	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_id", "id is not a valid UUID")
		return "", err
	}
	return id, nil
}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("failed to encode response", "err", err)
	}
}

func (h *Handler) writeError(w http.ResponseWriter, status int, code, message string) {
	h.writeJSON(w, status, service.ErrorResponse{
		Code:    code,
		Message: message,
	})
}

func sanitize(s string) string {
	return strings.NewReplacer(
		"\n", "",
		"\r", "",
		"\t", "",
	).Replace(s)
}
