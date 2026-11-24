package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	store := NewStore()

	mux := http.NewServeMux()

	mux.HandleFunc("POST /register", registerHandler(store))
	mux.HandleFunc("POST /users", createUserHandler(store))
mux.HandleFunc("POST /apps", createAppHandler(store))
mux.HandleFunc("POST /alerts", reportAlertHandler(store))
mux.HandleFunc("GET /alerts", getAlertsHandler(store))
mux.HandleFunc("GET /alerts/{id}", getAlertByIDHandler(store))
mux.HandleFunc("POST /alerts/{id}/resolve", resolveAlertHandler(store))
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("Server running on port", port)

	err := http.ListenAndServe(":"+port, mux)
	if err != nil {
		log.Fatal(err)
	}
}