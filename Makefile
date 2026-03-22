.PHONY: build run test lint swagger docker-up docker-down

build: swagger
	go build -o bin/web ./cmd/web

run: swagger
	@test -x $$(go env GOPATH)/bin/air || { echo "air not found. Install it: go install github.com/air-verse/air@latest"; exit 1; }
	$$(go env GOPATH)/bin/air

test:
	go test ./...

lint:
	$(shell go env GOPATH)/bin/golangci-lint run ./...

swagger:
	@test -x $$(go env GOPATH)/bin/swag || { echo "swag not found. Install it: go install github.com/swaggo/swag/cmd/swag@latest"; exit 1; }
	$$(go env GOPATH)/bin/swag init -g cmd/web/main.go -o api/swagger

docker-up:
	docker compose up -d

docker-down:
	docker compose down
