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
	database, err := db.NewDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	paymentProvider := buildPaymentProvider()

	userService := services.NewUserService(database.DB)
	agentService := services.NewAgentService(database.DB)
	policyService := services.NewPolicyService(database.DB)
	spendService := services.NewSpendServiceWithProvider(database.DB, paymentProvider)
	adminAuthService := services.NewAdminAuthService(database.DB, getAdminSessionTTLHours())
	paymentWebhookService := services.NewPaymentWebhookService(database.DB)

	authMiddleware := middleware.NewAuthMiddleware(agentService)
	adminAuthMiddleware := middleware.NewAdminAuthMiddleware(adminAuthService)

	userHandler := handlers.NewUserHandler(userService)
	agentHandler := handlers.NewAgentHandler(agentService)
	policyHandler := handlers.NewPolicyHandler(policyService)
	spendHandler := handlers.NewSpendHandler(spendService)
	adminAuthHandler := handlers.NewAdminAuthHandler(adminAuthService)
	adminDashboardService := services.NewAdminDashboardServiceWithProvider(database.DB, paymentProvider)
	adminDashboardHandler := handlers.NewAdminDashboardHandler(adminDashboardService)
	stripeWebhookHandler := handlers.NewStripeWebhookHandler(paymentProvider, paymentWebhookService)

	mux := http.NewServeMux()

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

	mux.HandleFunc("/spend", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			authMiddleware.Authenticate(spendHandler.Spend)(w, r)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

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

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"healthy"}`))
	})

	port := getEnv("PORT", "8080")
	log.Printf("Governor API server starting on port %s", port)
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
