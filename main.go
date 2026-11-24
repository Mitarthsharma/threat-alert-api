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