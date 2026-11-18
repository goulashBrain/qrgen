BIN := bin/qrgen

.PHONY: tidy test build run docker
tidy:   ; go mod tidy
test:   ; go test ./...
build:  ; CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o $(BIN) ./cmd/qrgen
run:    ; go run ./cmd/qrgen
docker: ; docker build -t qrgen:0.1.0 .
