package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
)

func generateRandomString() (string, error) {
	bytes := make([]byte, 16)

	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{
		"error": message,
	})
}

func decodeJSON(r *http.Request, data any) error {
	return json.NewDecoder(r.Body).Decode(data)
}

func isEmpty(value string) bool {
	return strings.TrimSpace(value) == ""
}

func isValidEmail(email string) bool {
	email = strings.TrimSpace(email)
	return email != "" && strings.Contains(email, "@")
}