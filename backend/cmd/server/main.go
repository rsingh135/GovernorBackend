package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"agentpay/internal/db"
	"agentpay/internal/handlers"
	"agentpay/internal/middleware"
	"agentpay/internal/payments"
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

	paymentProvider := buildPaymentProvider()

	// Initialize services (business logic layer)
	userService := services.NewUserService(database.DB)
	agentService := services.NewAgentService(database.DB)
	policyService := services.NewPolicyService(database.DB)
	spendService := services.NewSpendServiceWithProvider(database.DB, paymentProvider)
	adminAuthService := services.NewAdminAuthService(database.DB, getAdminSessionTTLHours())
	paymentWebhookService := services.NewPaymentWebhookService(database.DB)

	log.Println("✅ Services initialized")

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(agentService)
	adminAuthMiddleware := middleware.NewAdminAuthMiddleware(adminAuthService)

	// Initialize handlers (HTTP layer)
	userHandler := handlers.NewUserHandler(userService)
	agentHandler := handlers.NewAgentHandler(agentService)
	policyHandler := handlers.NewPolicyHandler(policyService)
	spendHandler := handlers.NewSpendHandler(spendService)
	adminAuthHandler := handlers.NewAdminAuthHandler(adminAuthService)
	adminDashboardService := services.NewAdminDashboardServiceWithProvider(database.DB, paymentProvider)
	adminDashboardHandler := handlers.NewAdminDashboardHandler(adminDashboardService)
	stripeWebhookHandler := handlers.NewStripeWebhookHandler(paymentProvider, paymentWebhookService)

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

	mux.HandleFunc("/agents", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			agentHandler.CreateAgent(w, r)
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

	// Protected routes (require API key authentication)
	mux.HandleFunc("/spend", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			authMiddleware.Authenticate(spendHandler.Spend)(w, r)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	// Admin auth routes (dashboard)
	mux.HandleFunc("/admin/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			adminAuthHandler.Login(w, r)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	mux.HandleFunc("/admin/me", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			adminAuthMiddleware.Authenticate(adminAuthHandler.Me)(w, r)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	mux.HandleFunc("/admin/users", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			adminAuthMiddleware.Authenticate(adminDashboardHandler.ListUsers)(w, r)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	mux.HandleFunc("/admin/users/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			adminAuthMiddleware.Authenticate(adminDashboardHandler.GetUser)(w, r)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	mux.HandleFunc("/admin/agents", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			adminAuthMiddleware.Authenticate(adminDashboardHandler.ListAgents)(w, r)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	mux.HandleFunc("/admin/agents/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			if strings.HasSuffix(r.URL.Path, "/history") {
				adminAuthMiddleware.Authenticate(adminDashboardHandler.GetAgentHistory)(w, r)
				return
			}
			adminAuthMiddleware.Authenticate(adminDashboardHandler.GetAgent)(w, r)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	mux.HandleFunc("/admin/policies", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			adminAuthMiddleware.Authenticate(adminDashboardHandler.GetPolicyByAgent)(w, r)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	mux.HandleFunc("/admin/transactions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			adminAuthMiddleware.Authenticate(adminDashboardHandler.ListTransactions)(w, r)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	mux.HandleFunc("/admin/transactions/pending", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			adminAuthMiddleware.Authenticate(adminDashboardHandler.ListPendingTransactions)(w, r)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	mux.HandleFunc("/admin/transactions/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			adminAuthMiddleware.Authenticate(adminDashboardHandler.GetTransaction)(w, r)
		case http.MethodPost:
			if strings.HasSuffix(r.URL.Path, "/approve") {
				adminAuthMiddleware.Authenticate(adminDashboardHandler.ApproveTransaction)(w, r)
				return
			}
			if strings.HasSuffix(r.URL.Path, "/deny") {
				adminAuthMiddleware.Authenticate(adminDashboardHandler.DenyTransaction)(w, r)
				return
			}
			http.Error(w, "Not found", http.StatusNotFound)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/webhooks/stripe", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			stripeWebhookHandler.Handle(w, r)
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
	log.Println("   POST /users      - Create user account")
	log.Println("   POST /agents     - Provision agent")
	log.Println("   POST /policies   - Manage spending policies")
	log.Println("   POST /spend      - Process spending request (authenticated)")
	log.Println("   POST /admin/login - Admin session login")
	log.Println("   GET  /admin/me    - Get authenticated admin profile")
	log.Println("   GET  /admin/users - List users")
	log.Println("   GET  /admin/users/{id} - Get user")
	log.Println("   GET  /admin/agents - List agents")
	log.Println("   GET  /admin/agents/{id} - Get agent")
	log.Println("   GET  /admin/agents/{id}/history - Agent recent transactions")
	log.Println("   GET  /admin/policies?agent_id={id} - Get policy for agent")
	log.Println("   GET  /admin/transactions - List transactions")
	log.Println("   GET  /admin/transactions/pending - List pending approvals")
	log.Println("   GET  /admin/transactions/{id} - Get transaction")
	log.Println("   POST /admin/transactions/{id}/approve - Approve pending transaction")
	log.Println("   POST /admin/transactions/{id}/deny - Deny pending transaction")
	log.Println("   POST /webhooks/stripe - Stripe webhook receiver")
	log.Println("   GET  /health     - Health check")

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

func getAdminSessionTTLHours() time.Duration {
	hoursRaw := getEnv("ADMIN_SESSION_TTL_HOURS", "24")
	hours, err := strconv.Atoi(hoursRaw)
	if err != nil || hours <= 0 {
		return 24 * time.Hour
	}
	return time.Duration(hours) * time.Hour
}

func buildPaymentProvider() payments.Provider {
	stripeSecret := strings.TrimSpace(os.Getenv("STRIPE_SECRET_KEY"))
	if stripeSecret == "" {
		return payments.NewNoopProvider()
	}

	successURL := getEnv("STRIPE_SUCCESS_URL", "http://localhost:3000/checkout/success?session_id={CHECKOUT_SESSION_ID}")
	cancelURL := getEnv("STRIPE_CANCEL_URL", "http://localhost:3000/checkout/cancel")

	return payments.NewStripeProvider(payments.StripeConfig{
		SecretKey:     stripeSecret,
		WebhookSecret: getEnv("STRIPE_WEBHOOK_SECRET", ""),
		SuccessURL:    successURL,
		CancelURL:     cancelURL,
		BaseURL:       getEnv("STRIPE_BASE_URL", ""),
	})
}
