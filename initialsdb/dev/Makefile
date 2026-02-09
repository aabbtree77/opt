SHELL := /bin/bash
.SHELLFLAGS := -e -o pipefail -c

# --------------------------------------------------
# Docker lifecycle
# --------------------------------------------------

up:
	docker compose up -d --build

down:
	docker compose down

logs:
	docker compose logs -f

ps:
	docker compose ps

# --------------------------------------------------
# Soft reset: preserve DB
# --------------------------------------------------

soft-reset:
	@echo "🧹 Stopping containers and cleaning images..."
	docker compose down --remove-orphans
	-docker image rm initialsdb-dev
	@echo "✅ Soft reset done (DB preserved)."

# --------------------------------------------------
# Hard reset: wipe DB
# --------------------------------------------------

hard-reset:
	@echo "🧹 Stopping containers and cleaning images..."
	docker compose down --remove-orphans
	-docker image rm initialsdb-dev
	@echo "💣 Wiping dev database volume..."
	docker volume rm initialsdb_postgres_dev || true
	@echo "✅ Hard reset done (DB fully wiped)."


# ==================================================
# Convenience aliases
# ==================================================

reset: soft-reset

nuke-db: hard-reset

# ==================================================
# Optional: full rebuild target
# ==================================================
# rebuild: hard-reset up
#	@echo "✅ Production environment rebuilt from scratch."