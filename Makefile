.PHONY: build run test clean

proto:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative api/gophkeeper.proto

build:
	go build -o gophkeeper ./cmd/gophkeeper
	GOOS=linux GOARCH=amd64 go build -ldflags "-X github.com/ASTeterin/gophkeeper/internal/client.version=1.0.0 -X github.com/ASTeterin/gophkeeper/internal/client.buildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o gophkeeper-client-linux ./cmd/client

test:
	go test ./...