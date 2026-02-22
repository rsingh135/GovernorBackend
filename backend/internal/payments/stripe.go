package payments

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

const defaultStripeBaseURL = "https://api.stripe.com"

// StripeConfig configures Stripe provider behavior.
type StripeConfig struct {
	SecretKey     string
	WebhookSecret string
	SuccessURL    string
	CancelURL     string
	BaseURL       string
	HTTPClient    *http.Client
}

// StripeProvider implements Provider using Stripe Checkout Sessions.
type StripeProvider struct {
	secretKey     string
	webhookSecret string
	successURL    string
	cancelURL     string
	baseURL       string
	httpClient    *http.Client
}

func NewStripeProvider(cfg StripeConfig) *StripeProvider {
	baseURL := strings.TrimSpace(cfg.BaseURL)
	if baseURL == "" {
		baseURL = defaultStripeBaseURL
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}

	return &StripeProvider{
		secretKey:     strings.TrimSpace(cfg.SecretKey),
		webhookSecret: strings.TrimSpace(cfg.WebhookSecret),
		successURL:    strings.TrimSpace(cfg.SuccessURL),
		cancelURL:     strings.TrimSpace(cfg.CancelURL),
		baseURL:       strings.TrimRight(baseURL, "/"),
		httpClient:    httpClient,
	}
}

func (p *StripeProvider) Enabled() bool {
	return p.secretKey != ""
}

func (p *StripeProvider) Name() string {
	return "stripe"
}

func (p *StripeProvider) CreateCheckoutSession(ctx context.Context, req CreateCheckoutRequest) (*CheckoutSession, error) {
	if !p.Enabled() {
		return nil, fmt.Errorf("stripe provider is not enabled")
	}

	values := url.Values{}
	values.Set("mode", "payment")
	values.Set("success_url", p.successURL)
	values.Set("cancel_url", p.cancelURL)
	values.Set("line_items[0][price_data][currency]", strings.ToLower(strings.TrimSpace(req.Currency)))
	values.Set("line_items[0][price_data][unit_amount]", strconv.FormatInt(req.AmountCents, 10))
	values.Set("line_items[0][price_data][product_data][name]", req.Vendor)
	values.Set("line_items[0][quantity]", "1")
	values.Set("metadata[transaction_id]", req.TransactionID.String())
	values.Set("metadata[agent_id]", req.AgentID.String())
	values.Set("metadata[vendor]", req.Vendor)
	values.Set("payment_intent_data[metadata][transaction_id]", req.TransactionID.String())
	values.Set("payment_intent_data[metadata][agent_id]", req.AgentID.String())

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/v1/checkout/sessions", strings.NewReader(values.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to build stripe request: %w", err)
	}
	httpReq.SetBasicAuth(p.secretKey, "")
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call stripe: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read stripe response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("stripe create checkout failed with status %d: %s", resp.StatusCode, string(respBytes))
	}

	var out struct {
		ID            string `json:"id"`
		URL           string `json:"url"`
		PaymentIntent string `json:"payment_intent"`
	}
	if err := json.Unmarshal(respBytes, &out); err != nil {
		return nil, fmt.Errorf("failed to parse stripe response: %w", err)
	}

	return &CheckoutSession{
		Provider:        p.Name(),
		SessionID:       out.ID,
		CheckoutURL:     out.URL,
		PaymentIntentID: out.PaymentIntent,
		ProviderStatus:  "checkout_created",
	}, nil
}

func (p *StripeProvider) ParseWebhook(payload []byte, signature string) (*WebhookEvent, error) {
	if p.webhookSecret == "" {
		return nil, fmt.Errorf("stripe webhook secret is not configured")
	}

	if err := verifyStripeSignature(payload, signature, p.webhookSecret); err != nil {
		return nil, err
	}

	var envelope struct {
		ID   string `json:"id"`
		Type string `json:"type"`
		Data struct {
			Object map[string]interface{} `json:"object"`
		} `json:"data"`
	}
	if err := json.Unmarshal(payload, &envelope); err != nil {
		return nil, fmt.Errorf("failed to parse stripe webhook payload: %w", err)
	}

	obj := envelope.Data.Object
	metadata := objectMap(obj["metadata"])
	txnRaw := strings.TrimSpace(stringFromAny(metadata["transaction_id"]))
	if txnRaw == "" {
		return nil, fmt.Errorf("stripe webhook missing metadata.transaction_id")
	}

	txnID, err := uuid.Parse(txnRaw)
	if err != nil {
		return nil, fmt.Errorf("invalid transaction_id in webhook metadata: %w", err)
	}

	event := &WebhookEvent{
		EventID:         envelope.ID,
		Provider:        p.Name(),
		TransactionID:   txnID,
		SessionID:       stringFromAny(obj["id"]),
		PaymentIntentID: stringFromAny(obj["payment_intent"]),
	}

	switch envelope.Type {
	case "checkout.session.completed":
		event.ProviderStatus = "checkout_completed"
	case "checkout.session.expired":
		event.ProviderStatus = "checkout_expired"
	case "payment_intent.succeeded":
		event.ProviderStatus = "payment_succeeded"
		event.PaymentIntentID = stringFromAny(obj["id"])
	case "payment_intent.payment_failed":
		event.ProviderStatus = "payment_failed"
		event.PaymentIntentID = stringFromAny(obj["id"])
	default:
		return nil, fmt.Errorf("unsupported stripe event type: %s", envelope.Type)
	}

	return event, nil
}

func verifyStripeSignature(payload []byte, header, secret string) error {
	timestamp, signatures, err := parseStripeSignatureHeader(header)
	if err != nil {
		return err
	}

	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid stripe signature timestamp: %w", err)
	}

	if age := time.Since(time.Unix(ts, 0)); age > 5*time.Minute || age < -5*time.Minute {
		return fmt.Errorf("stripe webhook timestamp outside tolerance")
	}

	signedPayload := timestamp + "." + string(payload)
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(signedPayload))
	expected := hex.EncodeToString(mac.Sum(nil))

	for _, sig := range signatures {
		if hmac.Equal([]byte(expected), []byte(sig)) {
			return nil
		}
	}

	return fmt.Errorf("invalid stripe signature")
}

func parseStripeSignatureHeader(header string) (string, []string, error) {
	if strings.TrimSpace(header) == "" {
		return "", nil, fmt.Errorf("missing stripe signature header")
	}

	parts := strings.Split(header, ",")
	var timestamp string
	sigs := make([]string, 0)

	for _, part := range parts {
		kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(kv) != 2 {
			continue
		}
		switch kv[0] {
		case "t":
			timestamp = kv[1]
		case "v1":
			sigs = append(sigs, kv[1])
		}
	}

	if timestamp == "" || len(sigs) == 0 {
		return "", nil, fmt.Errorf("invalid stripe signature header")
	}

	return timestamp, sigs, nil
}

func objectMap(v interface{}) map[string]interface{} {
	if v == nil {
		return map[string]interface{}{}
	}
	if m, ok := v.(map[string]interface{}); ok {
		return m
	}
	return map[string]interface{}{}
}

func stringFromAny(v interface{}) string {
	switch t := v.(type) {
	case string:
		return t
	default:
		return ""
	}
}
