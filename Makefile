
.DEFAULT_GOAL := build

.PHONY: fmt vet build clean run snapshot ui-dev ui-build ui-clean

# --- UI ---

ui-dev:
	cd ui && npm run dev

ui-build:
	cd ui && npm run build

ui-clean:
	rm -rf jdextract/web/dist

# --- Go ---

clean: ui-clean
	go clean && rm -rf out/

fmt:
	go fmt ./...

vet: fmt
	go vet ./...

build: ui-build vet
	go build -o out/jdextractor ./cmd

run: build
	./out/jdextractor

snapshot:
	goreleaser release --snapshot --clean
