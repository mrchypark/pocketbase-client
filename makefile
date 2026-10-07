PB_VERSION := "0.39.10"
ARCH := $(shell uname -m)
UNAME_S := $(shell uname -s)


ifeq ($(ARCH), x86_64)
    # For x86_64 architecture
    ARCH=amd64
else ifeq ($(ARCH), aarch64)
    # For ARM architecture
    ARCH=arm64
else ifeq ($(ARCH), arm64)
    # For ARM architecture
    ARCH=arm64
else
    $(error Unsupported architecture: $(ARCH))
endif

ifeq ($(UNAME_S),Linux)
    UNAME_S="linux"
else ifeq ($(UNAME_S),Darwin)
    UNAME_S="darwin"
else
    $(error Unsupported operating system)
endif

pb: ./database/pocketbase

./database/pocketbase:
	@set -eu; tmp=$$(mktemp -d); trap 'rm -rf "$$tmp"' EXIT HUP INT TERM; \
	asset=pocketbase_$(PB_VERSION)_$(UNAME_S)_$(ARCH).zip; \
	base=https://github.com/pocketbase/pocketbase/releases/download/v$(PB_VERSION); \
	curl -fsSL "$$base/$$asset" -o "$$tmp/$$asset"; \
	curl -fsSL "$$base/checksums.txt" -o "$$tmp/checksums.txt"; \
	awk -v asset="$$asset" '$$2 == asset {print; found=1} END {if (!found) exit 1}' "$$tmp/checksums.txt" > "$$tmp/checksum"; \
	(cd "$$tmp"; if command -v sha256sum >/dev/null; then sha256sum -c checksum; else shasum -a 256 -c checksum; fi); \
	unzip -q "$$tmp/$$asset" -d "$$tmp/extracted"; \
	mkdir -p database; mv "$$tmp/extracted/pocketbase" "$@"

.PHONY: pb_run
pb_run: ./database/pocketbase
	./database/pocketbase serve --dev

.PHONY: pb_clean
pb_clean:
	rm ./database/pocketbase

.PHONY: pb_snapshot pb_snap pb_ss
pb_snapshot pb_snap pb_ss: ./database/pocketbase
	./database/pocketbase migrate collections

.PHONY: gen
gen:
	go run ./cmd/pbc-gen
