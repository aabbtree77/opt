SHELL := /bin/bash
OPT_DIR := $(HOME)/opt

.PHONY: clone-app

# Usage:
# make clone-app SRC=initialsdb DST=mynewapp

clone-app:
ifndef SRC
	$(error SRC is required)
endif
ifndef DST
	$(error DST is required)
endif

	rsync -a \
		--exclude '.git' \
		--exclude 'node_modules/' \
		--exclude '.secrets*' \
		--include '.secrets*.example' \
		--exclude 'dev/volumes/*' \
		--exclude 'prod/volumes/*' \
		--exclude 'dev/bin/*' \
		--exclude 'prod/bin/*' \
		$(OPT_DIR)/$(SRC)/ $(OPT_DIR)/$(DST)/

	mkdir -p \
		$(OPT_DIR)/$(DST)/dev/volumes \
		$(OPT_DIR)/$(DST)/prod/volumes \
		$(OPT_DIR)/$(DST)/dev/bin \
		$(OPT_DIR)/$(DST)/prod/bin

	grep -rl '$(SRC)' $(OPT_DIR)/$(DST) \
		--exclude-dir=node_modules \
		--exclude-dir=volumes \
	| xargs sed -i 's/$(SRC)/$(DST)/g'

	@echo "✔ Cloned $(SRC) → $(DST)"
