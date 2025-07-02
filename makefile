PB_VERSION := "0.28.4"
ARCH := $(shell uname -m)
UNAME_S := $(shell uname -s)


ifeq ($(ARCH), x86_64)
    # x86_64 아키텍처의 경우
    ARCH=amd64
else ifeq ($(ARCH), aarch64)
    # ARM 아키텍처의 경우
    ARCH=arm64
else ifeq ($(ARCH), arm64)
    # ARM 아키텍처의 경우
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

