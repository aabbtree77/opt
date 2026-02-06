SHELL := /bin/bash
.SHELLFLAGS := -e -o pipefail -c

# --------------------------------------------------
# Docker lifecycle (used ON VPS)
# --------------------------------------------------

up:
	docker compose up -d --build

down:
	docker compose down

logs:
	docker compose logs -f

ps:
	docker compose ps

nuke-db:
	docker compose down
	docker volume rm initialsdb_postgres_prod || true

clean:
	@echo "🧹 Cleaning prod environment..."
	docker compose down --volumes --remove-orphans
	-docker image rm initialsdb-prod
	-docker volume rm initialsdb_postgres_prod
	@echo "✅ Prod environment is clean."
