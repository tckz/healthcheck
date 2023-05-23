.PHONY: all clean
.SUFFIXES: .proto .pb.go .go

ifeq ($(GO_CMD),)
GO_CMD:=go
endif

SRCS_OTHER=$(shell find . -type d -name vendor -prune -o -type d -name cmd -prune -o -type f -name "*.go" -print) go.mod

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

.proto.pb.go:
	protoc/bin/protoc -I. --proto_path=protoc/include --go_out=plugins=grpc:. $<

images: healthy-grpc
	@echo "$@ done."

healthy-grpc: $(DIST_HEALTHY_GRPC)
	cd docker/$@ && docker build -t $@ --pull .
	@echo "$@ done."

$(DIST_HEALTHY_GRPC): cmd/healthy-grpc/*.go api/hello.pb.go $(SRCS_OTHER)
	CGO_ENABLED=0 go build -o $@ ./cmd/healthy-grpc/
	@echo "$@ done."

$(DIST_HELLOCLIENT): cmd/helloclient/*.go api/hello.pb.go $(SRCS_OTHER)
	CGO_ENABLED=0 go build -o $@ ./cmd/helloclient/
	@echo "$@ done."

$(DIST_HELLOATTACKER): cmd/helloattacker/*.go api/hello.pb.go $(SRCS_OTHER)
	CGO_ENABLED=0 go build -o $@ ./cmd/helloattacker/
	@echo "$@ done."
