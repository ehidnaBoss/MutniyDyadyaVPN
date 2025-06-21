package payments

import (
	"encoding/json"
	"io"
	"net/http"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// CreatePayment создает новый платеж
func (h *Handler) CreatePayment(w http.ResponseWriter, r *http.Request) {
	var req CreatePaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	payment, err := h.service.CreatePayment(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := PaymentResponse{
		ID:            payment.ID,
		Status:        payment.Status,
		Amount:        payment.Amount,
		PaymentMethod: payment.PaymentMethod,
		CreatedAt:     payment.CreatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// WebhookCloudPayments обрабатывает webhook от CloudPayments
func (h *Handler) WebhookCloudPayments(w http.ResponseWriter, r *http.Request) {
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	// Проверяем подпись (опционально)
	signature := r.Header.Get("Content-HMAC")
	if signature != "" {
		if !h.service.provider.ValidateWebhook(payload, signature) {
			http.Error(w, "Invalid signature", http.StatusUnauthorized)
			return
		}
	}

	if err := h.service.ProcessWebhook(r.Context(), payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetPayment получает информацию о платеже
func (h *Handler) GetPayment(w http.ResponseWriter, r *http.Request) {
}