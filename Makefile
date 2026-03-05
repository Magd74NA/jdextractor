
.DEFAULT_GOAL := build

.PHONY: fmt vet build clean run snapshot

clean:
	go clean && rm -rf out/

fmt:
	go fmt ./...

vet: fmt
	go vet ./...

build: vet
	go build -o out/jdextractor ./cmd

run: build
	./out/jdextractor

snapshot:
	goreleaser release --snapshot --clean