package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func setupTestStore() *Store {
	store := NewStore()

	// Companies
	store.Companies["org1"] = Company{OrgID: "org1", Name: "Acme"}
	store.Companies["org2"] = Company{OrgID: "org2", Name: "Beta"}

	// Users
	store.Users["admin1"] = User{
		UserID: "admin1", OrgID: "org1", Role: "admin",
	}
	store.Users["viewer1"] = User{
		UserID: "viewer1", OrgID: "org1", Role: "viewer",
	}
	store.Users["admin2"] = User{
		UserID: "admin2", OrgID: "org2", Role: "admin",
	}

	// Authentication tokens
	store.Tokens["admin-token"] = "admin1"
	store.Tokens["viewer-token"] = "viewer1"
	store.Tokens["other-token"] = "admin2"

	// App
	store.Apps["app1"] = App{
		AppID: "app1",
		OrgID: "org1",
		Name: "Test App",
		Platform: "android",
		AppKey: "app-key",
	}

	store.AppKeys["app-key"] = "app1"

	// Existing alert
	store.Alerts["alert1"] = Alert{
		ID: "alert1",
		AppID: "app1",
		DeviceID: "device1",
		ThreatType: "malware",
		Severity: 3,
		Count: 1,
		Status: "open",
	}

	return store
}

// Test 1: Viewer cannot resolve an alert.
func TestViewerCannotResolveAlert(t *testing.T) {
	store := setupTestStore()

	mux := http.NewServeMux()
	mux.HandleFunc("POST /alerts/{id}/resolve", resolveAlertHandler(store))

	req := httptest.NewRequest("POST", "/alerts/alert1/resolve", nil)
	req.Header.Set("Authorization", "Bearer viewer-token")

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", rec.Code)
	}
}

// Test 2: Another company's user cannot access the alert.
func TestCrossCompanyAlertAccess(t *testing.T) {
	store := setupTestStore()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /alerts/{id}", getAlertByIDHandler(store))

	req := httptest.NewRequest("GET", "/alerts/alert1", nil)
	req.Header.Set("Authorization", "Bearer other-token")

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}
}

// Test 3: 50 simultaneous identical alerts create one alert with count 50.
func TestConcurrentIdenticalAlerts(t *testing.T) {
	store := setupTestStore()

	handler := reportAlertHandler(store)

	const totalRequests = 50

	var wg sync.WaitGroup

	for i := 0; i < totalRequests; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			body := `{
				"deviceId": "device-concurrent",
				"threatType": "malware",
				"severity": 4
			}`

			req := httptest.NewRequest(
				"POST",
				"/alerts",
				strings.NewReader(body),
			)

			req.Header.Set("X-App-Key", "app-key")

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusCreated &&
				rec.Code != http.StatusOK {
				t.Errorf("unexpected status code: %d", rec.Code)
			}
		}()
	}

	wg.Wait()

	store.mu.Lock()
	defer store.mu.Unlock()

	if len(store.Alerts) != 2 {
		t.Fatalf("expected 2 total alerts, got %d", len(store.Alerts))
	}

	var concurrentAlert Alert

	for _, alert := range store.Alerts {
		if alert.DeviceID == "device-concurrent" {
			concurrentAlert = alert
			break
		}
	}

	if concurrentAlert.ID == "" {
		t.Fatal("concurrent alert was not created")
	}

	if concurrentAlert.Count != totalRequests {
		t.Fatalf(
			"expected count %d, got %d",
			totalRequests,
			concurrentAlert.Count,
		)
	}
}