package payments

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestStripeProvider_ParseWebhook(t *testing.T) {
	txnID := uuid.New()
	payload := []byte(fmt.Sprintf(`{
		"id":"evt_test_1",
		"type":"checkout.session.completed",
		"data":{
			"object":{
				"id":"cs_test_123",
				"payment_intent":"pi_test_123",
				"metadata":{
					"transaction_id":"%s"
				}
			}
		}
	}`, txnID.String()))

	secret := "whsec_test_secret"
	header := makeStripeSignatureHeader(secret, payload)

	provider := NewStripeProvider(StripeConfig{WebhookSecret: secret})
	event, err := provider.ParseWebhook(payload, header)
	if err != nil {
		t.Fatalf("expected parse success, got error: %v", err)
	}

	if event.EventID != "evt_test_1" {
		t.Fatalf("expected event id evt_test_1, got %s", event.EventID)
	}
	if event.TransactionID != txnID {
		t.Fatalf("expected transaction id %s, got %s", txnID, event.TransactionID)
	}
	if event.ProviderStatus != "checkout_completed" {
		t.Fatalf("expected provider status checkout_completed, got %s", event.ProviderStatus)
	}
}

func TestStripeProvider_ParseWebhookInvalidSignature(t *testing.T) {
	provider := NewStripeProvider(StripeConfig{WebhookSecret: "whsec_test_secret"})
	_, err := provider.ParseWebhook([]byte(`{"id":"evt","type":"checkout.session.completed","data":{"object":{}}}`), "t=123,v1=bad")
	if err == nil {
		t.Fatalf("expected signature validation error")
	}
}

func makeStripeSignatureHeader(secret string, payload []byte) string {
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	signed := timestamp + "." + string(payload)
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(signed))
	signature := hex.EncodeToString(mac.Sum(nil))
	return fmt.Sprintf("t=%s,v1=%s", timestamp, signature)
}
