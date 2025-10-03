GO ?= go

GOBUILD = $(GO) build

GOBINREL = bin

GOBIN = $(CURDIR)/$(GOBINREL)

db-up:
	docker-compose -f ./compose.dev.yml up -d postgres pgadmin

logs:
	docker-compose -f ./compose.dev.yml logs -f

run:
	docker-compose -f ./compose.dev.yml up -d

stop:
	docker-compose -f ./compose.dev.yml down -v

build:
	$(GOBUILD) -o $(GOBIN)/server cmd/server/main.go

run-local:
	DB_HOST=localhost \
	DB_PORT=5432 \
	DB_NAME=transfers_db \
	DB_USER=postgres \
	DB_PASSWORD=password \
	DB_SSL_MODE=disable \
	go run cmd/server/main.go

# You could add unit tests or go benchmarks too here if needed
test:
	$(MAKE) test-integration
	$(MAKE) test-concurrency

test-integration:
	TEST_DB_HOST=localhost \
	TEST_DB_PORT=5432 \
	TEST_DB_NAME=transfers_db \
	TEST_DB_USER=postgres \
	TEST_DB_PASSWORD=password \
	TEST_DB_SSL_MODE=disable \
	go test -v -count=1 ./tests -run TestCreateAccount -run TestGetAccount -run TestCreateTransaction -run TestCompleteWorkflow

test-concurrency:
	TEST_DB_HOST=localhost \
	TEST_DB_PORT=5432 \
	TEST_DB_NAME=transfers_db \
	TEST_DB_USER=postgres \
	TEST_DB_PASSWORD=password \
	TEST_DB_SSL_MODE=disable \
	go test -v -count=1 ./tests -run TestConcurrent

.PHONY: db-up db-down run build run-local logs stop test-integration test-concurrency test