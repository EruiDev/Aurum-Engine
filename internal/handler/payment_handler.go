package handler

import (
	"aurum/internal/service"
	"encoding/json"
	"log/slog"
	"net/http"
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
	_, err := h.service.CreatePayment(r.Context(), req)
	if err != nil { // TODO add all messages for all errors
		h.writeError(w, http.StatusInternalServerError, "test", "test")
		slog.Error("Error creating the payment: ", err)
		return
	}
	h.writeJSON(w, http.StatusAccepted, req)
}

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
