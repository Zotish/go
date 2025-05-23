package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type User struct {
	ID        string    `json:"id"`
	PublicKey string    `json:"publicKey"`
	CreatedAt time.Time `json:"createdAt"`
}

type AISession struct {
	SessionID string    `json:"sessionId"`
	UserID    string    `json:"userId"`
	ExpiresAt time.Time `json:"expiresAt"`
}

var (
	users    = make(map[string]User)
	sessions = make(map[string]AISession)
)

func main() {
	r := mux.NewRouter()

	// CORS setup
	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})

	// Auth endpoints
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Phantom Backend is running"))
	}).Methods("GET")
	r.HandleFunc("/auth/phantom", phantomAuthHandler).Methods("POST")
	r.HandleFunc("/ai/session", createAISessionHandler).Methods("POST")
	// Add to main.go
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", handlers.CORS(originsOk, headersOk, methodsOk)(r)))
}
func phantomAuthHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PublicKey string `json:"publicKey"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Generate user ID
	userId := generateID()
	users[userId] = User{
		ID:        userId,
		PublicKey: req.PublicKey,
		CreatedAt: time.Now(),
	}

	json.NewEncoder(w).Encode(map[string]string{
		"userId": userId,
	})
}

// Update the createAISessionHandler to return proper JSON
func createAISessionHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserId string `json:"userId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Verify user exists
	if _, exists := users[req.UserId]; !exists {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Create session
	sessionId := generateID()
	expiration := time.Now().Add(24 * time.Hour)

	session := AISession{
		SessionID: sessionId,
		UserID:    req.UserId,
		ExpiresAt: expiration,
	}
	sessions[sessionId] = session

	// Return complete session info
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sessionId": session.SessionID,
		"userId":    session.UserID,
		"expiresAt": session.ExpiresAt.Format(time.RFC3339),
	})
}

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
