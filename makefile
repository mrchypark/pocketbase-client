PB_VERSION := "0.28.4"
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
	wget https://github.com/pocketbase/pocketbase/releases/download/v$(PB_VERSION)/pocketbase_$(PB_VERSION)_$(UNAME_S)_$(ARCH).zip \
  && unzip pocketbase_$(PB_VERSION)_$(UNAME_S)_$(ARCH).zip \
  && rm CHANGELOG.md LICENSE.md pocketbase_$(PB_VERSION)_$(UNAME_S)_$(ARCH).zip \
  && mv pocketbase ./database/pocketbase

.PHONY: pb_run
pb_run: ./database/pocketbase
	./database/pocketbase serve --dev

.PHONY: pb_clean
pb_clean:
	rm ./database/pocketbase

.PHONY: pb_snapshot pb_snap pb_ss
pb_snapshot pb_snap pb_ss: ./database/pocketbase
	./database/pocketbase migrate collections

# Version information
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build flags for version injection
LDFLAGS := -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

.PHONY: gen
gen:
	go run ./cmd/pbc-gen

.PHONY: build-pbc-gen
build-pbc-gen:
	go build -ldflags "$(LDFLAGS)" -o pbc-gen ./cmd/pbc-gen

.PHONY: install-pbc-gen
install-pbc-gen:
	go install -ldflags "$(LDFLAGS)" ./cmd/pbc-gen

