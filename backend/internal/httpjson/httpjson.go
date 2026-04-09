package httpjson

import (
	"encoding/json"
	"io"
	"net/http"
)

func DecodeStrict(w http.ResponseWriter, r *http.Request, dst any) bool {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return false
	}
	// Reject trailing tokens.
	var extra any
	if err := dec.Decode(&extra); err == nil || err != io.EOF {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return false
	}
	return true
}

func Write(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

