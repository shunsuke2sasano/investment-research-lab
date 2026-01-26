package handlers

import (
	"encoding/json"
	"net/http"
)

func DecodeJSON(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}

func WriteJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func WriteError(w http.ResponseWriter, status int, msg string) {
	WriteJSON(w, status, map[string]string{"error": msg})
}

func AuthOK(r *http.Request, apiKey string) bool {
	if apiKey == "" {
		return true
	}
	if r.Header.Get("X-API-Key") == apiKey {
		return true
	}
	auth := r.Header.Get("Authorization")
	return auth == "Bearer "+apiKey
}
