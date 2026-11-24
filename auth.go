package main

import (
	"net/http"
	"strings"
)

func authenticate(r *http.Request, store *Store) (User, bool) {
	authHeader := r.Header.Get("Authorization")

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return User{}, false
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")

	store.mu.Lock()
	defer store.mu.Unlock()

	userID, exists := store.Tokens[token]
	if !exists {
		return User{}, false
	}

	user, exists := store.Users[userID]
	if !exists {
		return User{}, false
	}

	return user, true
}