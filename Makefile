include config/client.env
include config/server.env
export

server:
	go run cmd/server/main.go

client:
	go run cmd/client/main.go

test:
	@go test ./... -v

fmt:
	@go fmt ./...

build:
	@docker-compose build

run:
	@docker-compose up	