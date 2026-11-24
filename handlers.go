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

func createUserHandler(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		currentUser, ok := authenticate(r, store)
		if !ok {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		if currentUser.Role != "admin" {
			writeError(w, http.StatusForbidden, "admin access required")
			return
		}

		var req struct {
			Name  string `json:"name"`
			Email string `json:"email"`
			Role  string `json:"role"`
		}

		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}

		req.Name = strings.TrimSpace(req.Name)
		req.Email = strings.TrimSpace(req.Email)
		req.Role = strings.TrimSpace(req.Role)

		if isEmpty(req.Name) || !isValidEmail(req.Email) ||
			(req.Role != "admin" && req.Role != "viewer") {
			writeError(w, http.StatusBadRequest, "invalid user details")
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

		user := User{
			UserID: userID,
			OrgID:  currentUser.OrgID,
			Name:   req.Name,
			Email:  req.Email,
			Role:   req.Role,
			Token:  token,
		}

		store.mu.Lock()
		store.Users[userID] = user
		store.Tokens[token] = userID
		store.mu.Unlock()

		writeJSON(w, http.StatusCreated, user)
	}
}

func createAppHandler(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		currentUser, ok := authenticate(r, store)
		if !ok {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		if currentUser.Role != "admin" {
			writeError(w, http.StatusForbidden, "admin access required")
			return
		}

		var req struct {
			Name     string `json:"name"`
			Platform string `json:"platform"`
		}

		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}

		req.Name = strings.TrimSpace(req.Name)
		req.Platform = strings.TrimSpace(req.Platform)

		if isEmpty(req.Name) ||
			(req.Platform != "android" && req.Platform != "ios") {
			writeError(w, http.StatusBadRequest, "invalid app details")
			return
		}

		appID, err := generateRandomString()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "could not generate app ID")
			return
		}

		appKey, err := generateRandomString()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "could not generate app key")
			return
		}

		app := App{
			AppID:    appID,
			OrgID:    currentUser.OrgID,
			Name:     req.Name,
			Platform: req.Platform,
			AppKey:   appKey,
		}

		store.mu.Lock()
		store.Apps[appID] = app
		store.AppKeys[appKey] = appID
		store.mu.Unlock()

		writeJSON(w, http.StatusCreated, app)
	}
}