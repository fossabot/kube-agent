-include .env

PKG = github.com/wodby/kube-agent
APP = kube-agent

REPO = wodby/kube-agent
NAME = kube-agent

GOOS ?= linux
GOARCH ?= amd64
VERSION ?= dev

ifneq ($(STABILITY_TAG),)
    override TAG := $(STABILITY_TAG)
else
    TAG = dev
endif

ifeq ($(GOOS),linux)
    ifeq ($(GOARCH),amd64)
        LINUX_AMD64 = 1
    endif
endif

LD_FLAGS = "-s -w -X $(PKG)/pkg/version.VERSION=$(VERSION)"

ARTIFACT := bin/$(APP)-$(GOOS)-$(GOARCH).tar.gz

default: build

.PHONY: build test dev push shell package release

build:
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) \
		go build -ldflags $(LD_FLAGS) -o bin/$(GOOS)-$(GOARCH)/$(APP) $(PKG)/cmd/$(APP)

    ifeq ($(LINUX_AMD64),1)
		make build-image
    endif

run:
	KUBE_AGENT_NODE_TOKEN=$(KUBE_AGENT_NODE_TOKEN) \
	KUBE_AGENT_NODE_UUID=$(KUBE_AGENT_NODE_UUID) \
	KUBE_AGENT_SERVER_HOST=$(KUBE_AGENT_SERVER_HOST) \
	KUBE_AGENT_SERVER_PORT=$(KUBE_AGENT_SERVER_PORT) \
	KUBE_AGENT_SKIP_VERIFY=$(KUBE_AGENT_SKIP_VERIFY) \
		./bin/$(GOOS)-$(GOARCH)/$(APP)

build-image:
	docker build -t $(REPO):$(TAG) ./

dev:
	cd ./dev && ./run.sh

push:
    ifeq ($(LINUX_AMD64),1)
		docker push $(REPO):$(TAG)
    endif

shell:
	docker run --rm --name $(NAME) $(PARAMS) -ti $(REPO):$(TAG) /bin/bash

package:
    ifeq ("$(wildcard $(ARTIFACT))","")
		tar czf $(ARTIFACT) -C bin/$(GOOS)-$(GOARCH) $(APP)
		rm -rf bin/$(GOOS)-$(GOARCH)
    endif

release: build push
