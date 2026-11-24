package main

import (
	"net/http"
	"strings"
)

func registerHandler(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var req struct {
			OrgName string `json:"orgName"`
			Name    string `json:"name"`
			Email   string `json:"email"`
		}

		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}

		req.OrgName = strings.TrimSpace(req.OrgName)
		req.Name = strings.TrimSpace(req.Name)
		req.Email = strings.TrimSpace(req.Email)

		if isEmpty(req.OrgName) || isEmpty(req.Name) || !isValidEmail(req.Email) {
			writeError(w, http.StatusBadRequest, "invalid registration details")
			return
		}

		orgID, err := generateRandomString()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "could not generate company ID")
			return
		}

		userID, err := generateRandomString()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "could not generate user ID")
			return
		}

		token, err := generateRandomString()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "could not generate token")
			return
		}

		company := Company{
			OrgID: orgID,
			Name:  req.OrgName,
		}

		user := User{
			UserID: userID,
			OrgID:  orgID,
			Name:   req.Name,
			Email:  req.Email,
			Role:   "admin",
			Token:  token,
		}

		store.mu.Lock()

		store.Companies[orgID] = company
		store.Users[userID] = user
		store.Tokens[token] = userID

		store.mu.Unlock()

		writeJSON(w, http.StatusCreated, user)
	}
}