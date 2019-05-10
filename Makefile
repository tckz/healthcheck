.PHONY: all clean
.SUFFIXES: .proto .pb.go .go

SRCS_OTHER=$(shell find . -type d -name vendor -prune -o -type d -name cmd -prune -o -type f -name "*.go" -print)

DIST_HEALTHY_OLD_GOJI=docker/healthy-old-goji/fs/healthy-old-goji

TARGETS=\
	$(DIST_HEALTHY_OLD_GOJI)

all: $(TARGETS)
	@echo "$@ done."

clean: 
	/bin/rm -f $(TARGETS)
	@echo "$@ done."

images: healthy-old-goji
	@echo "$@ done."

healthy-old-goji: $(DIST_HEALTHY_OLD_GOJI)
	cd docker/$@ && docker build -t $@ . 
	@echo "$@ done."

$(DIST_HEALTHY_OLD_GOJI): cmd/healthy-old-goji/*.go go.sum $(SRCS_OTHER)
	# link statically for alpine linux
	env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o $@ ./cmd/healthy-old-goji/
	@echo "$@ done."

