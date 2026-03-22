.PHONY: build run test lint swagger web-generate-api docker-up docker-down web-install web-dev web-build web-test web-test-e2e web-lint

build: web-build swagger
	go build -o bin/web ./cmd/web

ensure-dist:
	@mkdir -p web/dist
	@test -f web/dist/index.html || echo '<!doctype html><html><body>Run make web-build</body></html>' > web/dist/index.html

run: ensure-dist swagger
	@test -x $$(go env GOPATH)/bin/air || { echo "air not found. Install it: go install github.com/air-verse/air@latest"; exit 1; }
	$$(go env GOPATH)/bin/air

test: web-test
	go test ./...

lint: web-lint
	@test -x $$(go env GOPATH)/bin/golangci-lint || { echo "golangci-lint not found. Install it: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; exit 1; }
	$$(go env GOPATH)/bin/golangci-lint run ./...

swagger:
	@test -x $$(go env GOPATH)/bin/swag || { echo "swag not found. Install it: go install github.com/swaggo/swag/cmd/swag@latest"; exit 1; }
	$$(go env GOPATH)/bin/swag init -g cmd/web/main.go -o api/swagger
	$(MAKE) web-generate-api

web-generate-api:
	cd web && npm run generate:api

docker-up:
	docker compose up -d

docker-down:
	docker compose down

web-install:
	cd web && npm install

web-dev:
	cd web && npm run dev

web-build:
	cd web && npm run build

web-test:
	cd web && npm test

web-test-e2e:
	cd web && npm run test:e2e

web-lint:
	cd web && npm run lint
