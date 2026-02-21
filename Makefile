.PHONY: help db-up db-down db-reset db-test-setup backend test test-verbose

help:
	@echo "Available commands:"
	@echo "  make db-up         - Start PostgreSQL database"
	@echo "  make db-down       - Stop PostgreSQL database"
	@echo "  make db-reset      - Reset database (stop, remove volumes, start)"
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

db-test-setup:
	docker exec -it agentpay-db psql -U postgres -c "DROP DATABASE IF EXISTS agentpay_test;"
	docker exec -it agentpay-db psql -U postgres -c "CREATE DATABASE agentpay_test;"
	docker exec -it agentpay-db psql -U postgres -d agentpay_test -f /docker-entrypoint-initdb.d/001_initial_schema.sql
	@echo "Test database created and migrated successfully!"

backend:
	cd backend && go run cmd/server/main.go

test:
	cd backend && DB_NAME_TEST=agentpay_test go test ./internal/handlers -v

test-verbose:
	cd backend && DB_NAME_TEST=agentpay_test go test ./internal/handlers -v -count=1
