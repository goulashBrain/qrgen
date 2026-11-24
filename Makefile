BIN     := bin/qrgen
VERSION ?= $(shell git describe --tags --always --dirty)
COMMIT  ?= $(shell git rev-parse --short HEAD)
LDFLAGS := -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT)

.PHONY: tidy test build run docker snapshot release-check

tidy:          ; go mod tidy
test:          ; go test ./...
build:         ; CGO_ENABLED=0 go build -trimpath -ldflags="$(LDFLAGS)" -o $(BIN) ./cmd/qrgen
run:           ; go run ./cmd/qrgen
docker:        ; docker build --build-arg VERSION=$(VERSION) --build-arg COMMIT=$(COMMIT) -t qrgen:dev .
snapshot:      ; goreleaser release --snapshot --clean --skip=docker
release-check: ; goreleaser check