prefix ?= /usr/local

artifacts = nexd nexc nex-dhcpd
common=pkg/nex.pb.go pkg/*.go

.PHONY: all
all: $(patsubst %,build/%,$(artifacts))

build/nexd: svc/nexd/*.go $(common) | build
	$(go-build)

build/nexc: util/nexc/*.go $(common) | build
	$(go-build)

build/nex-dhcpd: svc/dhcp-server/*.go $(common) | build
	$(go-build-file)

pkg/nex.pb.go: pkg/nex.proto
	$(protoc-build)

build:
	@mkdir -p build

.PHONY: clean
clean:
	$(QUIET) rm -rf build

.PHONY: distclean
distclean: clean
	$(QUIET) find . \
		-path ./vendor \
		-prune -o -name "*.pb.go" -print \
		| xargs -n 1 rm -f
	$(QUIET) rm -rf api/portal/models
	$(QUIET) rm -rf api/portal/restapi/operations

.PHONY: install
install: build/loader
	$(QUIET) mkdir -p $(prefix)/bin
	$(QUIET) cp build/loader $(prefix)/bin/loader
	$(QUIET) cp build/nexd $(prefix)/bin/nexd
	$(QUIET) cp build/nexc $(prefix)/bin/nexc
	$(QUIET) cp build/nex-dhcpd $(prefix)/bin/nex-dhcpd
	$(QUIET) mkdir -p ${prefix}/../lib/systemd/system
	$(QUIET) cp coredns.service ${prefix}/../lib/systemd/system/coredns.service
	$(QUIET) cp nexd.service ${prefix}/../lib/systemd/system/nexd.service
	$(QUIET) cp nex-dhcpd.service ${prefix}/../lib/systemd/system/nex-dhcpd.service
	$(QUIET) mkdir -p ${prefix}/../etc/merge
	$(QUIET) cp nex-default.yml ${prefix}/../etc/merge/nex.yml

BLUE=\e[34m
GREEN=\e[32m
CYAN=\e[36m
NORMAL=\e[39m

QUIET=@
ifeq ($(V),1)
	QUIET=
endif

define build-slug
	@echo "$(BLUE)$1$(GREEN)\t $< $(CYAN)$@$(NORMAL)"
endef

define go-build
	$(call build-slug,go)
	$(QUIET) go build -o $@ $(dir $<)/*
endef

define go-build-file
	$(call build-slug,go)
	$(QUIET) go build -o $@ $<
endef

define protoc-build
	$(call build-slug,protoc)
	$(QUIET) protoc \
		-I . \
		-I $(GOPATH)/src \
		-I ./$(dir $@) \
		./$< \
		--go_out=plugins=grpc:.
endef
