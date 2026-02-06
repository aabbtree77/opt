SHELL := /bin/bash
.SHELLFLAGS := -e -o pipefail -c

up:
	docker compose up -d --build

down:
	docker compose down

logs:
	docker compose logs -f

ps:
	docker compose ps

reset:
	docker compose down
	-docker image rm initialsdb-dev		

nuke-db:
	docker compose down
	docker volume rm initialsdb_postgres_dev || true


clean:
	@echo "🧹 Stopping and removing all containers, images, volumes..."
	docker compose down --volumes --remove-orphans
	@echo "🧹 Removing dev image..."
	-docker image rm initialsdb-dev
	@echo "🧹 Removing dev postgres volume..."
	-docker volume rm initialsdb_postgres_dev
	@echo "✅ Dev environment is clean."
