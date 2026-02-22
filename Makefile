.PHONY: help db-up db-down db-reset migrate db-test-setup backend test test-verbose

help:
	@echo "Available commands:"
	@echo "  make db-up         - Start PostgreSQL database"
	@echo "  make db-down       - Stop PostgreSQL database"
	@echo "  make db-reset      - Reset database (stop, remove volumes, start)"
	@echo "  make migrate       - Apply all migrations to main database"
	@echo "  make db-test-setup - Create test database"
	@echo "  make backend       - Run Go backend server"
	@echo "  make test          - Run test suite"
	@echo "  make test-verbose  - Run test suite with verbose output"

db-up:
	docker-compose up -d
	@echo "Waiting for PostgreSQL to be ready..."
	@sleep 5

db-down:
	docker-compose down

db-reset:
	docker-compose down -v
	docker-compose up -d
	@echo "Waiting for PostgreSQL to be ready..."
	@sleep 5

migrate:
	docker exec agentpay-db psql -U postgres -d agentpay -f /docker-entrypoint-initdb.d/002_add_users_table.sql
	docker exec agentpay-db psql -U postgres -d agentpay -f /docker-entrypoint-initdb.d/003_add_admin_auth.sql
	docker exec agentpay-db psql -U postgres -d agentpay -f /docker-entrypoint-initdb.d/004_add_payment_provider_fields.sql
	@echo "Main database incremental migrations applied successfully!"

db-test-setup:
	docker exec agentpay-db psql -U postgres -c "DROP DATABASE IF EXISTS agentpay_test;"
	docker exec agentpay-db psql -U postgres -c "CREATE DATABASE agentpay_test;"
	docker exec agentpay-db psql -U postgres -d agentpay_test -f /docker-entrypoint-initdb.d/001_initial_schema.sql
	docker exec agentpay-db psql -U postgres -d agentpay_test -f /docker-entrypoint-initdb.d/002_add_users_table.sql
	docker exec agentpay-db psql -U postgres -d agentpay_test -f /docker-entrypoint-initdb.d/003_add_admin_auth.sql
	docker exec agentpay-db psql -U postgres -d agentpay_test -f /docker-entrypoint-initdb.d/004_add_payment_provider_fields.sql
	@echo "Test database created and migrated successfully!"

backend:
	cd backend && go run cmd/server/main.go

test:
	cd backend && DB_NAME_TEST=agentpay_test go test ./... -v

test-verbose:
	cd backend && DB_NAME_TEST=agentpay_test go test ./... -v -count=1
