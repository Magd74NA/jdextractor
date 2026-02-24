
.DEFAULT_GOAL := build

.PHONY: fmt vet build clean

clean:
	go clean

fmt:
	go fmt ./...

vet: fmt
	go vet ./...

build: vet
	go build -o out/jdextractor ./cmd
