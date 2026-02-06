SHELL := /bin/bash
.SHELLFLAGS := -e -o pipefail -c

APP := initialsdb

DEV_DIR  := ../dev
PROD_DIR := ../prod

BACKEND  := backend
FRONTEND := frontend
MIGRATIONS := migrations

BIN_DEV  := $(DEV_DIR)/bin/server
BIN_PROD := $(PROD_DIR)/bin/server

# --------------------------------------------------
# Frontend
# --------------------------------------------------

build-frontend:
	@echo "📦 Building frontend..."
	cd $(FRONTEND) && npm install && npm run build

	rm -rf $(DEV_DIR)/web $(PROD_DIR)/web
	mkdir -p $(DEV_DIR)/web $(PROD_DIR)/web

	cp -r $(FRONTEND)/dist/* $(DEV_DIR)/web/
	cp -r $(FRONTEND)/dist/* $(PROD_DIR)/web/

# --------------------------------------------------
# Backend binaries
# --------------------------------------------------

build-backend-dev:
	@echo "🔨 Building backend (dev x86)..."
	mkdir -p $(DEV_DIR)/bin
	cd $(BACKEND) && \
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
	go build -o ../$(DEV_DIR)/bin/server ./server

build-backend-prod:
	@echo "🔨 Building backend (prod arm64)..."
	mkdir -p $(PROD_DIR)/bin
	cd $(BACKEND) && \
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 \
	go build -o ../$(PROD_DIR)/bin/server ./server


# --------------------------------------------------
# Migrations
# --------------------------------------------------

sync-migrations:
	@echo "📂 Syncing migrations..."
	rm -rf $(DEV_DIR)/migrations $(PROD_DIR)/migrations
	mkdir -p $(DEV_DIR)/migrations $(PROD_DIR)/migrations

	cp -r $(MIGRATIONS)/* $(DEV_DIR)/migrations/
	cp -r $(MIGRATIONS)/* $(PROD_DIR)/migrations/

# --------------------------------------------------
# High-level targets
# --------------------------------------------------

build: build-frontend build-backend-dev build-backend-prod sync-migrations
	@echo "✅ Build + distribution complete"

clean:
	rm -rf $(DEV_DIR)/bin $(DEV_DIR)/web $(DEV_DIR)/migrations
	rm -rf $(PROD_DIR)/bin $(PROD_DIR)/web $(PROD_DIR)/migrations

