package main

import (
	"log"
	"net/http"
	"os"

	"agentpay/internal/db"
	"agentpay/internal/handlers"
	"agentpay/internal/middleware"
	"agentpay/internal/services"
)

func main() {
	// Initialize database connection
	database, err := db.NewDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	log.Println("✅ Database connected successfully")

	// Initialize services (business logic layer)
	userService := services.NewUserService(database.DB)
	agentService := services.NewAgentService(database.DB)
	policyService := services.NewPolicyService(database.DB)
	spendService := services.NewSpendService(database.DB)
	transactionService := services.NewTransactionService(database.DB)

	log.Println("✅ Services initialized")

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(agentService)

	// Initialize handlers (HTTP layer)
	userHandler := handlers.NewUserHandler(userService)
	agentHandler := handlers.NewAgentHandler(agentService)
	policyHandler := handlers.NewPolicyHandler(policyService)
	spendHandler := handlers.NewSpendHandler(spendService)
	transactionHandler := handlers.NewTransactionHandler(transactionService)

	log.Println("✅ Handlers initialized")

	// Setup routes
	mux := http.NewServeMux()

	// Public routes (no auth required)
	mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			userHandler.CreateUser(w, r)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	// GET /users/:id
	mux.HandleFunc("/users/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			userHandler.GetUser(w, r)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	mux.HandleFunc("/agents", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			agentHandler.CreateAgent(w, r)
			return
		}
		if r.Method == http.MethodGet {
			agentHandler.ListAgents(w, r)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	mux.HandleFunc("/policies", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			policyHandler.UpsertPolicy(w, r)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	// GET /policies/:agent_id (authenticated)
	mux.HandleFunc("/policies/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			authMiddleware.Authenticate(policyHandler.GetPolicy)(w, r)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	// Protected routes (require API key authentication)
	mux.HandleFunc("/spend", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			authMiddleware.Authenticate(spendHandler.Spend)(w, r)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	// GET /transactions (authenticated)
	mux.HandleFunc("/transactions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			authMiddleware.Authenticate(transactionHandler.ListTransactions)(w, r)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	})

	port := getEnv("PORT", "8080")
	log.Printf("🚀 Governor API server starting on port %s", port)
	log.Println("📋 Available endpoints:")
	log.Println("   POST /users          - Create user account")
	log.Println("   GET  /users/:id      - Get user by ID")
	log.Println("   POST /agents         - Provision agent")
	log.Println("   GET  /agents         - List agents (with filters)")
	log.Println("   POST /policies       - Manage spending policies")
	log.Println("   GET  /policies/:id   - Get policy by agent ID (authenticated)")
	log.Println("   POST /spend          - Process spending request (authenticated)")
	log.Println("   GET  /transactions   - List transactions (authenticated)")
	log.Println("   GET  /health         - Health check")

	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
