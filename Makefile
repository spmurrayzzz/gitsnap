.PHONY: build test clean

build:
	go build -o bin/gitsnap ./cmd/gitsnap

test:
	go test ./... -cover

clean:
	rm -rf bin
