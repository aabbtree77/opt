SHELL := /bin/bash
.SHELLFLAGS := -e -o pipefail -c

# --------------------------------------------------
# Docker lifecycle
# --------------------------------------------------

up:
	docker compose up -d

down:
	docker compose down

logs:
	docker compose logs -f

ps:
	docker compose ps

restart:
	docker compose down
	docker compose up -d

clean:
	@echo "🧹 Stopping Caddy..."
	docker compose down --remove-orphans
	@echo "✅ Caddy cleaned."

# --------------------------------------------------
# Network management (edge_net)
# --------------------------------------------------

net:
	@docker network inspect edge_net >/dev/null 2>&1 || \
		docker network create edge_net
	@echo "🌐 edge_net ready"

net-remove:
	@docker network inspect edge_net >/dev/null 2>&1 || exit 0
	@if [ "$$(docker network inspect -f '{{len .Containers}}' edge_net)" -eq 0 ]; then \
		docker network rm edge_net; \
		echo "🗑️ edge_net removed"; \
	else \
		echo "⚠️ edge_net not removed: containers still attached"; \
	fi
