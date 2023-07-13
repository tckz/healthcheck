.PHONY: all clean

ifeq ($(GO_CMD),)
GO_CMD:=go
endif

SRCS_OTHER=$(shell find . -type d -name vendor -prune -o -type d -name cmd -prune -o -type f -name "*.go" -print) go.mod

DIR_BIN := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))/bin
DIR_PROTOC := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))/protoc

DIST_HEALTHY_GRPC=docker/healthy-grpc/fs/healthy-grpc
DIST_HELLOCLIENT=docker/healthy-grpc/fs/helloclient
DIST_HELLOATTACKER=docker/healthy-grpc/fs/hellocattacker

TARGETS=\
	$(DIST_HELLOCLIENT) \
	$(DIST_HELLOATTACKER) \
	$(DIST_HEALTHY_GRPC)

all: $(TARGETS)
	@echo "$@ done."

clean: 
	/bin/rm -f $(TARGETS)
	@echo "$@ done."

TOOL_PROTOC = $(DIR_PROTOC)/bin/protoc
TOOL_PROTOC_GEN_GO = $(DIR_BIN)/protoc-gen-go
TOOLS = \
	$(TOOL_PROTOC) \
	$(TOOL_PROTOC_GEN_GO)

TOOLS_DEP = Makefile

.PHONY: tools
tools: $(TOOLS)
	@echo "$@ done." 1>&2

$(TOOL_PROTOC): $(TOOLS_DEP)
	curl -sLf -o protoc/protoc.zip https://github.com/protocolbuffers/protobuf/releases/download/v23.4/protoc-23.4-linux-x86_64.zip
	cd protoc && unzip -o protoc.zip
	touch $@

$(TOOL_PROTOC_GEN_GO): export GOBIN=$(DIR_BIN)
$(TOOL_PROTOC_GEN_GO): $(TOOLS_DEP)
	@echo "### `basename $@` install destination=$(GOBIN)" 1>&2
	CGO_ENABLED=0 $(GO_CMD) install github.com/golang/protobuf/protoc-gen-go@v1.5.3

.PHONY: gen
TMP_PATH := $(DIR_BIN):$(PATH)
gen: export PATH=$(TMP_PATH)
gen: $(TOOL_PROTOC) $(TOOL_PROTOC_GEN_GO)
	$(TOOL_PROTOC) -I. --proto_path=protoc/include --go_out=plugins=grpc:. --go_opt=Mapi/hello.proto=./api api/hello.proto

.PHONY: images
images: healthy-grpc
	@echo "$@ done."

.PHONY: healthy-grpc
healthy-grpc: $(DIST_HEALTHY_GRPC)
	cd docker/$@ && docker build -t $@ --pull .
	@echo "$@ done."

$(DIST_HEALTHY_GRPC): cmd/healthy-grpc/*.go $(SRCS_OTHER)
	CGO_ENABLED=0 go build -o $@ ./cmd/healthy-grpc/
	@echo "$@ done."

$(DIST_HELLOCLIENT): cmd/helloclient/*.go $(SRCS_OTHER)
	CGO_ENABLED=0 go build -o $@ ./cmd/helloclient/
	@echo "$@ done."

$(DIST_HELLOATTACKER): cmd/helloattacker/*.go $(SRCS_OTHER)
	CGO_ENABLED=0 go build -o $@ ./cmd/helloattacker/
	@echo "$@ done."
