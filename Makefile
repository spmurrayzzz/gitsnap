.PHONY: build test clean

build:
	go build -o bin/gitsnap ./cmd/gitsnap

test:
	go test ./...

clean:
	rm -rf bin
