prefix ?= /usr/local

.PHONY: all
all: build/loader build/nexd build/nexc

build/nexd: nexd/main.go proto/nex.pb.go | build
	go build -o $@ $<

build/nexc: nexc/main.go proto/nex.pb.go | build
	go build -o $@ $<

proto/nex.pb.go: proto/nex.proto
	protoc -I ./proto ./proto/nex.proto --go_out=plugins=grpc:./proto

build/loader: loader/main.go nex.go | build
	go build -o $@ $<

build:
	mkdir -p build

.PHONY: clean
clean:
	rm -rf build

.PHONY: install
install: build/loader
	mkdir -p $(prefix)/bin
	cp build/loader $(prefix)/bin/loader
