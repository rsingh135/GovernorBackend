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
	"agentpay/internal/repository"
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
	requestIDMiddleware := middleware.NewRequestIDMiddleware()
	loggingMiddleware := middleware.NewLoggingMiddleware()
	loginRateLimiter := middleware.NewRateLimiterMiddleware(getEnvInt("ADMIN_LOGIN_RATE_LIMIT_PER_MINUTE", 20), time.Minute)
	spendRateLimiter := middleware.NewRateLimiterMiddleware(getEnvInt("SPEND_RATE_LIMIT_PER_MINUTE", 120), time.Minute)
	reviewRateLimiter := middleware.NewRateLimiterMiddleware(getEnvInt("ADMIN_REVIEW_RATE_LIMIT_PER_MINUTE", 60), time.Minute)

	// Initialize handlers (HTTP layer)
	userHandler := handlers.NewUserHandler(userService)
	agentHandler := handlers.NewAgentHandler(agentService)
	policyHandler := handlers.NewPolicyHandler(policyService)
	spendHandler := handlers.NewSpendHandler(spendService)
	proxyHandler := handlers.NewProxyHandler(repository.NewPolicyRepository(database.DB))
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
	mux.HandleFunc("/proxy/browse", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			spendRateLimiter.Limit(authMiddleware.Authenticate(proxyHandler.Browse))(w, r)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	mux.HandleFunc("/spend", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			spendRateLimiter.Limit(authMiddleware.Authenticate(spendHandler.Spend))(w, r)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	// Admin auth routes (dashboard)
	mux.HandleFunc("/admin/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			loginRateLimiter.Limit(adminAuthHandler.Login)(w, r)
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
		if r.Method == http.MethodPost {
			if strings.HasSuffix(r.URL.Path, "/freeze") {
				adminAuthMiddleware.Authenticate(adminDashboardHandler.FreezeUser)(w, r)
				return
			}
			if strings.HasSuffix(r.URL.Path, "/unfreeze") {
				adminAuthMiddleware.Authenticate(adminDashboardHandler.UnfreezeUser)(w, r)
				return
			}
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
		if r.Method == http.MethodPost {
			if strings.HasSuffix(r.URL.Path, "/freeze") {
				adminAuthMiddleware.Authenticate(adminDashboardHandler.FreezeAgent)(w, r)
				return
			}
			if strings.HasSuffix(r.URL.Path, "/unfreeze") {
				adminAuthMiddleware.Authenticate(adminDashboardHandler.UnfreezeAgent)(w, r)
				return
			}
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	mux.HandleFunc("/admin/policies", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			adminAuthMiddleware.Authenticate(adminDashboardHandler.GetPolicyByAgent)(w, r)
			return
		}
		if r.Method == http.MethodPost {
			adminAuthMiddleware.Authenticate(policyHandler.UpsertPolicy)(w, r)
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
				reviewRateLimiter.Limit(adminAuthMiddleware.Authenticate(adminDashboardHandler.ApproveTransaction))(w, r)
				return
			}
			if strings.HasSuffix(r.URL.Path, "/deny") {
				reviewRateLimiter.Limit(adminAuthMiddleware.Authenticate(adminDashboardHandler.DenyTransaction))(w, r)
				return
			}
			http.Error(w, "Not found", http.StatusNotFound)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/admin/audit/approvals", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			adminAuthMiddleware.Authenticate(adminDashboardHandler.ListApprovalAuditLogs)(w, r)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
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
	log.Println("   GET  /proxy/browse?url={encoded_url} - Agent browser sidecar proxy")
	log.Println("   POST /spend      - Process spending request (authenticated)")
	log.Println("   POST /admin/login - Admin session login")
	log.Println("   GET  /admin/me    - Get authenticated admin profile")
	log.Println("   GET  /admin/users - List users")
	log.Println("   GET  /admin/users/{id} - Get user")
	log.Println("   POST /admin/users/{id}/freeze - Freeze user org")
	log.Println("   POST /admin/users/{id}/unfreeze - Unfreeze user org")
	log.Println("   GET  /admin/agents - List agents")
	log.Println("   GET  /admin/agents/{id} - Get agent")
	log.Println("   GET  /admin/agents/{id}/history - Agent recent transactions")
	log.Println("   POST /admin/agents/{id}/freeze - Freeze agent")
	log.Println("   POST /admin/agents/{id}/unfreeze - Unfreeze agent")
	log.Println("   GET  /admin/policies?agent_id={id} - Get policy for agent")
	log.Println("   POST /admin/policies - Update policy for agent")
	log.Println("   GET  /admin/transactions - List transactions")
	log.Println("   GET  /admin/transactions/pending - List pending approvals")
	log.Println("   GET  /admin/transactions/{id} - Get transaction")
	log.Println("   POST /admin/transactions/{id}/approve - Approve pending transaction")
	log.Println("   POST /admin/transactions/{id}/deny - Deny pending transaction")
	log.Println("   GET  /admin/audit/approvals - List approval audit logs")
	log.Println("   POST /webhooks/stripe - Stripe webhook receiver")
	log.Println("   GET  /health     - Health check")

	rootHandler := requestIDMiddleware.AddRequestID(loggingMiddleware.Log(mux))
	if err := http.ListenAndServe(":"+port, rootHandler); err != nil {
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

func getEnvInt(key string, defaultValue int) int {
	raw := strings.TrimSpace(getEnv(key, ""))
	if raw == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return defaultValue
	}
	return value
}

func buildPaymentProvider() payments.Provider {
	stripeSecret := strings.TrimSpace(os.Getenv("STRIPE_SECRET_KEY"))
	if stripeSecret == "" {
		return payments.NewNoopProvider()
	}

	successURL := getEnv("STRIPE_SUCCESS_URL", "http://localhost:5173/checkout/success?session_id={CHECKOUT_SESSION_ID}")
	cancelURL := getEnv("STRIPE_CANCEL_URL", "http://localhost:5173/checkout/cancel")

	return payments.NewStripeProvider(payments.StripeConfig{
		SecretKey:     stripeSecret,
		WebhookSecret: getEnv("STRIPE_WEBHOOK_SECRET", ""),
		SuccessURL:    successURL,
		CancelURL:     cancelURL,
		BaseURL:       getEnv("STRIPE_BASE_URL", ""),
	})
}
