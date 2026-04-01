package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"agentpay/internal/db"
	"agentpay/internal/handlers"
	"agentpay/internal/logger"
	"agentpay/internal/middleware"
	"agentpay/internal/services"
)

func main() {
	// Initialize structured logger
	logger.Init()

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
	approvalService := services.NewApprovalService(database.DB)

	log.Println("✅ Services initialized")

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(agentService)
	apiKeyLimiter := middleware.NewRateLimiter(100, 100) // 100 req/s per API key
	ipLimiter := middleware.NewRateLimiter(10, 10)       // 10 req/s per IP

	// Initialize handlers (HTTP layer)
	userHandler := handlers.NewUserHandler(userService)
	agentHandler := handlers.NewAgentHandler(agentService)
	policyHandler := handlers.NewPolicyHandler(policyService)
	spendHandler := handlers.NewSpendHandler(spendService)
	transactionHandler := handlers.NewTransactionHandler(transactionService)
	approvalHandler := handlers.NewApprovalHandler(approvalService)

	log.Println("✅ Handlers initialized")

	// Setup routes
	mux := http.NewServeMux()

	// Public routes (IP rate limited)
	mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			ipLimiter.LimitByIP(userHandler.CreateUser)(w, r)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	// GET /users/:id
	mux.HandleFunc("/users/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			ipLimiter.LimitByIP(userHandler.GetUser)(w, r)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	mux.HandleFunc("/agents", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			ipLimiter.LimitByIP(agentHandler.CreateAgent)(w, r)
			return
		}
		if r.Method == http.MethodGet {
			ipLimiter.LimitByIP(agentHandler.ListAgents)(w, r)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	mux.HandleFunc("/policies", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			ipLimiter.LimitByIP(policyHandler.UpsertPolicy)(w, r)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	// GET /policies/:agent_id (authenticated)
	mux.HandleFunc("/policies/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			authMiddleware.Authenticate(apiKeyLimiter.LimitByAPIKey(policyHandler.GetPolicy))(w, r)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	// Protected routes (API key rate limited)
	mux.HandleFunc("/spend", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			authMiddleware.Authenticate(apiKeyLimiter.LimitByAPIKey(spendHandler.Spend))(w, r)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	// GET /transactions (authenticated)
	mux.HandleFunc("/transactions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			authMiddleware.Authenticate(apiKeyLimiter.LimitByAPIKey(transactionHandler.ListTransactions))(w, r)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	// POST /transactions/:id/approve and POST /transactions/:id/deny
	mux.HandleFunc("/transactions/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		path := r.URL.Path
		switch {
		case strings.HasSuffix(path, "/approve"):
			ipLimiter.LimitByIP(approvalHandler.Approve)(w, r)
		case strings.HasSuffix(path, "/deny"):
			ipLimiter.LimitByIP(approvalHandler.Deny)(w, r)
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
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
	log.Println("   GET  /transactions               - List transactions (authenticated)")
	log.Println("   POST /transactions/:id/approve   - Approve pending transaction")
	log.Println("   POST /transactions/:id/deny      - Deny pending transaction")
	log.Println("   GET  /health                     - Health check")

	if err := http.ListenAndServe(":"+port, middleware.RequestLogger(mux)); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
