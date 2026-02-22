package handlers

import (
	"io"
	"net/http"

	"agentpay/internal/httpjson"
	"agentpay/internal/payments"
	"agentpay/internal/services"
)

// StripeWebhookHandler ingests Stripe webhook events and updates transaction payment state.
type StripeWebhookHandler struct {
	provider       payments.Provider
	webhookService *services.PaymentWebhookService
}

func NewStripeWebhookHandler(provider payments.Provider, webhookService *services.PaymentWebhookService) *StripeWebhookHandler {
	return &StripeWebhookHandler{provider: provider, webhookService: webhookService}
}

func (h *StripeWebhookHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if h.provider == nil || !h.provider.Enabled() {
		http.Error(w, "payment provider not enabled", http.StatusNotImplemented)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read webhook payload", http.StatusBadRequest)
		return
	}

	event, err := h.provider.ParseWebhook(body, r.Header.Get("Stripe-Signature"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.webhookService.HandleEvent(r.Context(), event); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	httpjson.Write(w, http.StatusOK, map[string]bool{"received": true})
}
