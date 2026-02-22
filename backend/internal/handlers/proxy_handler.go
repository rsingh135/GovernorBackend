package handlers

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"agentpay/internal/middleware"
	"agentpay/internal/repository"
)

// ProxyHandler allows agents to browse vendor pages via Governor-controlled proxy.
type ProxyHandler struct {
	policyRepo *repository.PolicyRepository
	client     *http.Client
}

func NewProxyHandler(policyRepo *repository.PolicyRepository) *ProxyHandler {
	return &ProxyHandler{
		policyRepo: policyRepo,
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// Browse proxies GET requests while enforcing vendor allowlists.
func (h *ProxyHandler) Browse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	agent, ok := middleware.GetAgentFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	rawURL := strings.TrimSpace(r.URL.Query().Get("url"))
	if rawURL == "" {
		http.Error(w, "url is required", http.StatusBadRequest)
		return
	}

	target, err := url.Parse(rawURL)
	if err != nil || target.Host == "" {
		http.Error(w, "invalid url", http.StatusBadRequest)
		return
	}
	if target.Scheme != "http" && target.Scheme != "https" {
		http.Error(w, "unsupported url scheme", http.StatusBadRequest)
		return
	}

	policy, err := h.policyRepo.GetByAgentID(r.Context(), agent.ID)
	if err != nil {
		http.Error(w, "policy not found", http.StatusForbidden)
		return
	}

	vendor := strings.ToLower(strings.TrimPrefix(target.Hostname(), "www."))
	if !isAllowedProxyVendor(vendor, policy.AllowedVendors) {
		http.Error(w, "vendor not allowed", http.StatusForbidden)
		return
	}

	// Keep checkout credential handling server-side by forcing payment execution through /spend.
	if strings.Contains(vendor, "checkout.stripe.com") || strings.Contains(vendor, "stripe.com") {
		http.Error(w, "checkout pages must be initiated via /spend", http.StatusForbidden)
		return
	}

	proxyReq, err := http.NewRequestWithContext(r.Context(), http.MethodGet, target.String(), nil)
	if err != nil {
		http.Error(w, "failed to construct upstream request", http.StatusInternalServerError)
		return
	}
	proxyReq.Header.Set("User-Agent", "GovernorSidecar/1.0")

	resp, err := h.client.Do(proxyReq)
	if err != nil {
		http.Error(w, "failed to fetch upstream page", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if contentType := resp.Header.Get("Content-Type"); contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = io.CopyN(w, resp.Body, 2<<20) // 2 MiB response cap for MVP
}

func isAllowedProxyVendor(vendor string, allowed []string) bool {
	for _, entry := range allowed {
		if strings.ToLower(strings.TrimSpace(entry)) == vendor {
			return true
		}
	}
	return false
}
