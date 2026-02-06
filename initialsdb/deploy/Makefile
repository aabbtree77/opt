SHELL := /bin/bash
.SHELLFLAGS := -e -o pipefail -c

# --------------------------------------------------
# Derive app name from folder
# --------------------------------------------------
APP_NAME := initialsdb

# --------------------------------------------------
# Load deploy config (VPS only)
# --------------------------------------------------
-include .env
export

# --------------------------------------------------
# Local paths (relative to deploy/)
# --------------------------------------------------
LOCAL_APP_PROD := ../prod
LOCAL_CADDY    := ../../caddy

# --------------------------------------------------
# Remote paths (flattened /opt)
# --------------------------------------------------
REMOTE_APP_DIR := /opt/$(APP_NAME)
REMOTE_CADDY   := $(VPS_CADDY_DIR)

# --------------------------------------------------
# App copy
# --------------------------------------------------
copy-app:
	@echo "📦 Copying $(APP_NAME) to $(REMOTE_APP_DIR)"
	ssh $(VPS_USER)@$(VPS_HOST) "mkdir -p $(REMOTE_APP_DIR)"
	scp -r \
		$(LOCAL_APP_PROD)/* \
		$(LOCAL_APP_PROD)/.[!.]* \
		$(VPS_USER)@$(VPS_HOST):$(REMOTE_APP_DIR)/
	@echo "✅ App copied"

# --------------------------------------------------
# Caddy copy
# --------------------------------------------------
copy-caddy:
	@echo "📦 Copying Caddy to $(REMOTE_CADDY)"
	ssh $(VPS_USER)@$(VPS_HOST) "\
		mkdir -p $(REMOTE_CADDY)/data $(REMOTE_CADDY)/config \
	"
	scp -r \
		$(LOCAL_CADDY)/* \
		$(VPS_USER)@$(VPS_HOST):$(REMOTE_CADDY)/
	ssh $(VPS_USER)@$(VPS_HOST) "\
		chmod 700 $(REMOTE_CADDY)/data $(REMOTE_CADDY)/config \
	"
	@echo "✅ Caddy copied"

# --------------------------------------------------
# Combined
# --------------------------------------------------
copy: copy-app copy-caddy
	@echo "🚀 $(APP_NAME) + Caddy copied to VPS"

# --------------------------------------------------
# Backup script
# --------------------------------------------------
copy-backup-script:
	@echo "📦 Deploying backup script for $(APP_NAME)"
	ssh $(VPS_USER)@$(VPS_HOST) "\
		mkdir -p \
			$(REMOTE_APP_DIR)/backups \
			$(REMOTE_APP_DIR)/scripts \
	"
	scp \
		./backup.sh \
		$(VPS_USER)@$(VPS_HOST):$(REMOTE_APP_DIR)/scripts/backup.sh
	ssh $(VPS_USER)@$(VPS_HOST) "\
		chmod +x $(REMOTE_APP_DIR)/scripts/backup.sh \
	"
	@echo "✅ Backup script installed"

# --------------------------------------------------
# Cron automation
# --------------------------------------------------

install-backup-cron:
	@echo "⏱ Installing backup cron for $(APP_NAME)"
	ssh $(VPS_USER)@$(VPS_HOST) "\
		( crontab -l 2>/dev/null | grep -v '^0 3 \* \* \* /opt/$(APP_NAME)/scripts/backup.sh' ; \
		  echo '0 3 * * * /opt/$(APP_NAME)/scripts/backup.sh >> /opt/$(APP_NAME)/scripts/backup.log 2>&1' ) | crontab - \
	"
	@echo "✅ Cron installed"


remove-backup-cron:
	@echo "🧹 Removing backup cron for $(APP_NAME)"
	ssh $(VPS_USER)@$(VPS_HOST) "\
		crontab -l 2>/dev/null | grep -v '/opt/$(APP_NAME)/scripts/backup.sh' | crontab - \
	"
	@echo "✅ Cron removed"
