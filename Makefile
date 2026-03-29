.PHONY: help db-up db-down db-logs db-shell migrate-up migrate-down migrate-status migrate-create

# Default database URL
DB_URL ?= postgres://trazawork:trazawork@localhost:5432/trazawork?sslmode=disable

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# =============================================================================
# Database
# =============================================================================

db-up: ## Start PostgreSQL database
	docker compose up -d db
	@echo "Waiting for database to be ready..."
	@docker compose exec db pg_isready -U trazawork -d trazawork --timeout=30 || (echo "Retrying..." && timeout /t 3 /nobreak >nul && docker compose exec db pg_isready -U trazawork -d trazawork)
	@echo "Database ready at localhost:5432"

db-down: ## Stop PostgreSQL database
	docker compose down

db-logs: ## Show database logs
	docker compose logs -f db

db-shell: ## Open psql shell
	docker exec -it trazawork-db psql -U trazawork -d trazawork

db-reset: db-down ## Reset database (destroy all data)
	rm -rf .docker/postgres
	$(MAKE) db-up

db-clean: ## Remove database data folder only
	rm -rf .docker/postgres

db-backup: ## Backup database folder
	mkdir -p backups
	tar -czf backups/db-backup-$$(date +%Y%m%d-%H%M%S).tar.gz .docker/postgres

db-size: ## Show database folder size
	du -sh .docker/postgres 2>/dev/null || echo "Database folder does not exist yet"

# =============================================================================
# Migrations
# =============================================================================

migrate-up: ## Run all migrations
	migrate -path migrations -database "$(DB_URL)" up

migrate-down: ## Rollback last migration
	migrate -path migrations -database "$(DB_URL)" down 1

migrate-down-all: ## Rollback all migrations
	migrate -path migrations -database "$(DB_URL)" down -all

migrate-status: ## Show migration version
	migrate -path migrations -database "$(DB_URL)" version

migrate-force: ## Force migration version (usage: make migrate-force V=1)
	migrate -path migrations -database "$(DB_URL)" force $(V)

migrate-create: ## Create new migration (usage: make migrate-create NAME=create_users)
	migrate create -ext sql -dir migrations -seq $(NAME)

# =============================================================================
# Development
# =============================================================================

dev-setup: db-up ## Start database (app runs migrations on startup)
	@echo "Database ready. Run the app to apply migrations."

dev-reset: db-reset ## Reset database (app will re-apply migrations on next start)
	@echo "✅ Database reset complete. Folder .docker/postgres removed."

# =============================================================================
# App (Docker)
# =============================================================================

up: ## Start full stack (db + migrate + app)
	docker compose up -d --build

down: ## Stop full stack
	docker compose down

app-logs: ## Tail app logs
	docker compose logs -f app

app-build: ## Build app image
	docker compose build app
