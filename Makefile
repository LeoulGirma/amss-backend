.PHONY: dev-up dev-down migrate-up test lint seed test-integration

dev-up:
	docker compose up -d --build

dev-down:
	docker compose down -v

migrate-up:
	goose -dir migrations postgres "$$DB_URL" up

test:
	go test ./...

test-integration:
	AMSS_INTEGRATION=1 go test ./internal/infra/postgres ./internal/infra/redis

lint:
	golangci-lint run

seed:
	goose -dir migrations postgres "$$DB_URL" up
