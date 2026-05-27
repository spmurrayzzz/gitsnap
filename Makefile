.PHONY: build build-lib test clean

build:
	go build -o bin/gitsnap ./cmd/gitsnap

build-lib:
	@mkdir -p bin
	@case "$$(uname -s)" in \
		Darwin) ext=dylib ;; \
		Linux) ext=so ;; \
		*) ext=dll ;; \
	esac; \
	go build -buildmode=c-shared -o bin/libgitsnap.$$ext ./cmd/gitsnaplib

test:
	go test ./... -cover

clean:
	rm -rf bin
