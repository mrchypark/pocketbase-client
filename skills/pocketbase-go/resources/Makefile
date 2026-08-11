PB_URL ?= http://127.0.0.1:8090
PB_ADMIN_EMAIL ?= admin@example.com
PB_ADMIN_PASSWORD ?= admin1234
PB_DIR ?= .pb
PB_BIN ?= $(PB_DIR)/pocketbase
PB_VERSION ?= 0.36.6
PBC_GEN_VERSION ?= v0.3.9
SCHEMA_FILE ?= pb_schema.json
MODELS_OUT ?= models/models.gen.go
MODELS_PKG ?= models

PB_MIGRATIONS_DIR ?= ./pb_migrations
PB_HOOKS_DIR ?= ./pb_hooks

.PHONY: all init serve migrate dump codegen sync clean

all: serve

init:
	@echo "Installing PocketBase..."
	@mkdir -p $(PB_DIR)
	@if [ ! -f "$(PB_BIN)" ]; then \
		OS=$$(uname -s | tr '[:upper:]' '[:lower:]'); \
		ARCH=$$(uname -m); \
		case $$ARCH in \
			x86_64) ARCH=amd64 ;; \
			arm64|aarch64) ARCH=arm64 ;; \
			*) echo "Unsupported arch: $$ARCH"; exit 1 ;; \
		esac; \
		PB_ZIP="pocketbase_$(PB_VERSION)_$${OS}_$${ARCH}.zip"; \
		PB_ZIP_PATH="$(PB_DIR)/$${PB_ZIP}"; \
		echo "Detected OS: $$OS, Arch: $$ARCH"; \
		curl -L -o "$$PB_ZIP_PATH" "https://github.com/pocketbase/pocketbase/releases/download/v$(PB_VERSION)/$${PB_ZIP}"; \
		curl -L -o "$(PB_DIR)/checksums.txt" "https://github.com/pocketbase/pocketbase/releases/download/v$(PB_VERSION)/checksums.txt"; \
		EXPECTED_SHA=$$(awk '/ '"$${PB_ZIP}"'$$/ {print $$1}' "$(PB_DIR)/checksums.txt"); \
		if [ -z "$$EXPECTED_SHA" ]; then \
			echo "Missing checksum entry for $${PB_ZIP}"; \
			exit 1; \
		fi; \
		ACTUAL_SHA=$$(shasum -a 256 "$$PB_ZIP_PATH" | awk '{print $$1}'); \
		if [ "$$EXPECTED_SHA" != "$$ACTUAL_SHA" ]; then \
			echo "Checksum mismatch for $${PB_ZIP}"; \
			echo "Expected: $$EXPECTED_SHA"; \
			echo "Actual:   $$ACTUAL_SHA"; \
			exit 1; \
		fi; \
		unzip -o "$$PB_ZIP_PATH" -d "$(PB_DIR)"; \
		rm "$$PB_ZIP_PATH" "$(PB_DIR)/checksums.txt"; \
		chmod +x "$(PB_BIN)"; \
	fi
	@echo "Bootstrapping Admin..."
	@$(PB_BIN) superuser create $(PB_ADMIN_EMAIL) $(PB_ADMIN_PASSWORD) || echo "Admin might already exist."
	@echo "Creating pb_hooks..."
	@mkdir -p pb_hooks && touch pb_hooks/.gitkeep
	@echo "Installing pbc-gen $(PBC_GEN_VERSION)..."
	@go install github.com/mrchypark/pocketbase-client/cmd/pbc-gen@$(PBC_GEN_VERSION)
	@echo "Environment initialized."

serve:
	$(PB_BIN) serve --http 127.0.0.1:8090 --dir $(PB_DIR)/pb_data --migrationsDir $(PB_MIGRATIONS_DIR) --hooksDir $(PB_HOOKS_DIR)

migrate:
	python3 pb_migrate_ops.py apply --url $(PB_URL) --email $(PB_ADMIN_EMAIL) --password $(PB_ADMIN_PASSWORD) --schema $(SCHEMA_FILE)

dump:
	python3 pb_migrate_ops.py dump --url $(PB_URL) --email $(PB_ADMIN_EMAIL) --password $(PB_ADMIN_PASSWORD) --schema $(SCHEMA_FILE)

codegen:
	@mkdir -p $(dir $(MODELS_OUT))
	pbc-gen -schema $(SCHEMA_FILE) -path $(MODELS_OUT) -pkgname $(MODELS_PKG)

sync: dump codegen
	@echo "Synced schema from DB and generated code."

clean:
	rm -rf $(PB_DIR)
