package payments

import (
	"context"
	"fmt"
)

// NoopProvider disables payment integration while keeping service contracts stable.
type NoopProvider struct{}

func NewNoopProvider() *NoopProvider {
	return &NoopProvider{}
}

func (p *NoopProvider) Enabled() bool {
	return false
}

func (p *NoopProvider) Name() string {
	return "noop"
}

func (p *NoopProvider) CreateCheckoutSession(_ context.Context, _ CreateCheckoutRequest) (*CheckoutSession, error) {
	return nil, fmt.Errorf("payment provider is disabled")
}

func (p *NoopProvider) ParseWebhook(_ []byte, _ string) (*WebhookEvent, error) {
	return nil, fmt.Errorf("payment provider is disabled")
}
