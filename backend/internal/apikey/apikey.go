package apikey

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"strings"
)

// Generate returns (apiKey, sha256(apiKey), prefix).
// apiKey is safe to show once at creation time; only the hash should be persisted.
func Generate() (string, []byte, string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", nil, "", err
	}
	// URL-safe, no padding.
	token := base64.RawURLEncoding.EncodeToString(b)
	apiKey := "sk_agent_" + token
	sum := sha256.Sum256([]byte(apiKey))
	return apiKey, sum[:], Prefix(apiKey), nil
}

func Hash(apiKey string) []byte {
	sum := sha256.Sum256([]byte(apiKey))
	return sum[:]
}

func Prefix(apiKey string) string {
	const max = 16
	p := apiKey
	if len(p) > max {
		p = p[:max]
	}
	return strings.TrimSpace(p)
}

