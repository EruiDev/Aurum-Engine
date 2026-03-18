package handler

import (
	"aurum/internal/service"
	"encoding/json"
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
	}

	req.IdempotencyKey = r.Header.Get("Idempotency-Key")
	if req.IdempotencyKey == "" {
		h.writeError(w, http.StatusBadRequest, "missing_idempotency_key", "Idempotency-Key header is required")
		return
	}
	//payment, err := h.service.CreatePayment(r.Context(), req)
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
