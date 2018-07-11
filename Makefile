.PHONY: all clean image
.DELETE_ON_ERROR:

SHELL=/bin/bash -o pipefail
BIN_DIR=bin
PKG_CONFIG=.pkg_config
PKG_CONFIG_CONTENT=$(shell cat $(PKG_CONFIG))

IMAGE_REGISTRY ?= local
IMAGE_NAME=gitlab-slack-multiplexer
IMAGE_TAG=$(shell git rev-parse --short=7 HEAD)
IMAGE=$(IMAGE_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)

all: $(BIN_DIR)/gitlab-slack-multiplexer

clean:
	rm -rf $(BIN_DIR) $(PKG_CONFIG)

$(BIN_DIR)/gitlab-slack-multiplexer: $(PKG_CONFIG) $(wildcard gitlab-slack-multiplexer/*) $(wildcard gitlab-slack-multiplexer/**/*)
	go build -o $@ $(PKG_CONFIG_CONTENT) ./cmd/gitlab-slack-multiplexer

image:
	docker build -t $(IMAGE) .

$(PKG_CONFIG):
	scripts/pkg-config.sh > $@
