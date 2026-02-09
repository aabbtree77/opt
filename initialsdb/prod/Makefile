SHELL := /bin/bash
.SHELLFLAGS := -e -o pipefail -c

# ==================================================
# Docker Compose lifecycle (VPS / Production)
# ==================================================

up:
	docker compose up -d --build

down:
	docker compose down

logs:
	docker compose logs -f

ps:
	docker compose ps

# ==================================================
# Soft reset: rebuild containers and images, preserve DB
# ==================================================

soft-reset:
	@echo "🧹 Stopping containers and cleaning images..."
	docker compose down --remove-orphans
	-@docker image rm initialsdb-prod || true
	@echo "✅ Soft reset done (DB preserved)."

# ==================================================
# Hard reset: full wipe including DB
# ==================================================

hard-reset: soft-reset
	@echo "💣 Wiping prod database volume..."
	-docker volume rm initialsdb_postgres_prod || true
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
