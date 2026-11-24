package main

import (
	"net/http"
	"strings"
	"time"
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

func reportAlertHandler(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Authenticate SDK using app key
		appKey := r.Header.Get("X-App-Key")

		store.mu.Lock()
		appID, exists := store.AppKeys[appKey]
		_, appExists := store.Apps[appID]
		store.mu.Unlock()

		if appKey == "" || !exists || !appExists {
			writeError(w, http.StatusUnauthorized, "invalid app key")
			return
		}

		// Read request body
		var req struct {
			DeviceID   string `json:"deviceId"`
			ThreatType string `json:"threatType"`
			Severity   int    `json:"severity"`
		}

		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}

		req.DeviceID = strings.TrimSpace(req.DeviceID)
		req.ThreatType = strings.TrimSpace(req.ThreatType)

		if isEmpty(req.DeviceID) ||
			isEmpty(req.ThreatType) ||
			req.Severity < 1 || req.Severity > 5 {

			writeError(w, http.StatusBadRequest, "invalid alert details")
			return
		}

		now := time.Now().UTC()

		// Lock while checking and updating alerts
		store.mu.Lock()
		defer store.mu.Unlock()

		// Check for an existing open alert
		for id, alert := range store.Alerts {

			if alert.AppID == appID &&
				alert.DeviceID == req.DeviceID &&
				alert.ThreatType == req.ThreatType &&
				alert.Status == "open" {

				alert.Count++
				alert.LastSeen = now

				store.Alerts[id] = alert

				writeJSON(w, http.StatusOK, alert)
				return
			}
		}

		// Create a new alert
		alertID, err := generateRandomString()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "could not generate alert ID")
			return
		}

		alert := Alert{
			ID:         alertID,
			AppID:      appID,
			DeviceID:   req.DeviceID,
			ThreatType: req.ThreatType,
			Severity:   req.Severity,
			Count:      1,
			Status:     "open",
			FirstSeen:  now,
			LastSeen:   now,
			ResolvedBy: "",
		}

		store.Alerts[alertID] = alert

		writeJSON(w, http.StatusCreated, alert)
	}
}

func getAlertsHandler(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// 1. Authenticate the user
		user, ok := authenticate(r, store)
		if !ok {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		// 2. Only admin and viewer roles are allowed
		if user.Role != "admin" && user.Role != "viewer" {
			writeError(w, http.StatusForbidden, "forbidden")
			return
		}

		// 3. Read and validate the status filter
		status := r.URL.Query().Get("status")

		if status != "" && status != "open" && status != "resolved" {
			writeError(w, http.StatusBadRequest, "invalid status")
			return
		}

		// 4. Find alerts belonging to this user's company
		alerts := make([]Alert, 0)

		store.mu.Lock()
		defer store.mu.Unlock()

		for _, alert := range store.Alerts {

			app, exists := store.Apps[alert.AppID]
			if !exists || app.OrgID != user.OrgID {
				continue
			}

			// Apply status filter if provided
			if status != "" && alert.Status != status {
				continue
			}

			alerts = append(alerts, alert)
		}

		// 5. Return the alerts
		writeJSON(w, http.StatusOK, map[string]any{
			"alerts": alerts,
		})
	}
}

func getAlertByIDHandler(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// 1. Authenticate user
		user, ok := authenticate(r, store)
		if !ok {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		// 2. Check user role
		if user.Role != "admin" && user.Role != "viewer" {
			writeError(w, http.StatusForbidden, "forbidden")
			return
		}

		// 3. Get alert ID from URL
		alertID := r.PathValue("id")

		// 4. Find the alert
		store.mu.Lock()
		defer store.mu.Unlock()

		alert, exists := store.Alerts[alertID]
		if !exists {
			writeError(w, http.StatusNotFound, "alert not found")
			return
		}

		// 5. Check company ownership
		app, exists := store.Apps[alert.AppID]
		if !exists || app.OrgID != user.OrgID {
			writeError(w, http.StatusNotFound, "alert not found")
			return
		}

		// 6. Return alert
		writeJSON(w, http.StatusOK, alert)
	}
}

func resolveAlertHandler(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// 1. Authenticate user
		user, ok := authenticate(r, store)
		if !ok {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		// 2. Only admins can resolve alerts
		if user.Role != "admin" {
			writeError(w, http.StatusForbidden, "only admin can resolve alerts")
			return
		}

		// 3. Get alert ID
		alertID := r.PathValue("id")

		store.mu.Lock()
		defer store.mu.Unlock()

		// 4. Find alert
		alert, exists := store.Alerts[alertID]
		if !exists {
			writeError(w, http.StatusNotFound, "alert not found")
			return
		}

		// 5. Check company ownership
		app, exists := store.Apps[alert.AppID]
		if !exists || app.OrgID != user.OrgID {
			writeError(w, http.StatusNotFound, "alert not found")
			return
		}

		// 6. Check if already resolved
		if alert.Status == "resolved" {
			writeError(w, http.StatusConflict, "alert already resolved")
			return
		}

		// 7. Resolve alert
		alert.Status = "resolved"
		alert.ResolvedBy = user.UserID

		store.Alerts[alertID] = alert

		writeJSON(w, http.StatusOK, alert)
	}
}