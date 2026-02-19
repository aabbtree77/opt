SHELL := /bin/bash
.SHELLFLAGS := -e -o pipefail -c

ROOT := ..
SRC  := $(ROOT)/src
DEV  := $(ROOT)/dev

BACKEND  := $(SRC)/backend
FRONTEND := $(SRC)/frontend

DEBUG_BIN := bin/server
MIGRATIONS_DIR := migrations

# --------------------------------------------------
# Helper to load environment
# --------------------------------------------------
define load_env
	set -a && \
	. ./.env && \
	. $(DEV)/.secrets && \
	set +a
endef

# --------------------------------------------------
# Database (containerized only)
# --------------------------------------------------
db-up:
	@echo "🐘 Starting Postgres container..."
	cd $(DEV) && docker compose up -d db

db-down:
	@echo "🛑 Stopping Postgres container..."
	cd $(DEV) && docker compose stop db

db-logs:
	cd $(DEV) && docker compose logs -f db

db-nuke:
	@echo "💣 Destroying dev database volume..."
	cd $(DEV) && docker compose down
	docker volume rm initialsdb_postgres_dev || true

# --------------------------------------------------
# Backend (debug build + run)
# --------------------------------------------------
backend-build-debug:
	@echo "🔧 Building backend debug binary..."
	mkdir -p ./bin
	$(load_env) && \
	cd $(BACKEND) && \
	CGO_ENABLED=0 go build -gcflags="all=-N -l" -o ../../debug/$(DEBUG_BIN) ./server

backend-sync-migrations:
	@echo "📂 Syncing migrations for debug..."
	rm -rf ./$(MIGRATIONS_DIR)
	mkdir -p ./$(MIGRATIONS_DIR)
	cp -r ../src/migrations/* ./$(MIGRATIONS_DIR)/

backend-run: backend-build-debug backend-sync-migrations
	@echo "🚀 Running backend (debug mode)..."
	$(load_env) && \
	./$(DEBUG_BIN)

# --------------------------------------------------
# Frontend (Vite dev server)
# --------------------------------------------------
frontend-install:
	cd $(FRONTEND) && npm install

frontend-run:
	@echo "⚛️ Running Vite dev server..."
	cd $(FRONTEND) && npm run dev

# --------------------------------------------------
# Combined
# --------------------------------------------------
up:
	@echo "🧩 Starting debug environment..."
	@echo "Run backend + frontend in separate terminals."
	make db-up

down:
	make db-down
