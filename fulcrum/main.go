package main

import (
	"context"
	"crypto/rand"
	"fulcrum/api"
	"fulcrum/gamestore"
	"fulcrum/rules"
	"fulcrum/signaling"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

func WithCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "86400")

		// Gestion du preflight
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// @title Game API
// @version 1.0
// @description API Server for Game Application
// @host localhost:8086
// @BasePath /
func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Initialize the store
	redisAddr := os.Getenv("REDIS_ADDRESS")
	if redisAddr == "" {
		redisAddr = "localhost:6379" // Default fallback
	}

	serverAddr := os.Getenv("SERVER_ADDRESS")
	if serverAddr == "" {
		serverAddr = ":8086" // Default fallback
	}
	serverDomain := os.Getenv("SERVER_DOMAIN")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	store, err := gamestore.NewGameStore(ctx, redisAddr)
	if err != nil {
		log.Fatalf("Failed to connect to Redis at %s: %v", redisAddr, err)
	}

	notifier := signaling.NewHub()
	go notifier.Run()

	rules := gamerules.GetRules()

	adminPassword := os.Getenv("ADMIN_PASSWORD")

	apiHandler := api.NewAPI(store, notifier, rules, adminPassword, rand.Reader)

	if serverDomain == "" {
		log.Printf("Starting http server on %s", serverAddr)
		log.Fatal(http.ListenAndServe(serverAddr, WithCORS(apiHandler)))
	} else {
		server := &http.Server{
			Addr:    serverAddr,
			Handler: WithCORS(apiHandler),
		}
		certFile := "fullchain.pem"
		keyFile := "privkey.pem"
		log.Printf("Starting https server on %s", serverAddr)
		log.Fatal(server.ListenAndServeTLS(certFile, keyFile))
	}
}
