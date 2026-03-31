
.DEFAULT_GOAL := build

.PHONY: fmt vet build clean run snapshot ui-dev ui-build ui-clean

# --- UI ---

ui-dev:
	cd ui && npm run dev

ui-build:
	cd ui && npm run build
	find jdextract/web/dist -type f ! -name '*.gz' -delete

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
	go build -trimpath -ldflags="-s -w -X main.version=$$(git describe --tags --always --dirty 2>/dev/null || echo dev)" -o out/jdextractor ./cmd

run: build
	./out/jdextractor

snapshot:
	goreleaser release --snapshot --clean
