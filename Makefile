BINARY   := bin/server
IMAGE    := minimule-backend
TAG      := latest
PKG      := ./...

.PHONY: all setup build run dev test lint clean \
        migrate-up migrate-down migrate-create \
        docker-build docker-push \
        dc-up dc-down \
        k8s-apply k8s-delete

all: build

## ── Local development ────────────────────────────────────────────────────────

setup:
	go mod download && go mod verify

build:
	go build -ldflags="-s -w" -o $(BINARY) ./cmd/server

run: build
	./$(BINARY)

dev:
	go run ./cmd/server

test:
	go test -v -race -coverprofile=coverage.out $(PKG)

test/cover: test
	go tool cover -html=coverage.out -o coverage.html

lint:
	golangci-lint run $(PKG)

clean:
	rm -rf bin/ coverage.out coverage.html

## ── Migrations ───────────────────────────────────────────────────────────────
# Requires golang-migrate: https://github.com/golang-migrate/migrate
# Or run: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

migrate-up:
	migrate -database "$(DATABASE_URL)" -path migrations up

migrate-down:
	migrate -database "$(DATABASE_URL)" -path migrations down 1

migrate-create:
	@read -p "Migration name: " name; \
	migrate create -ext sql -dir migrations -seq $$name

## ── Docker ───────────────────────────────────────────────────────────────────

docker-build:
	docker build -t $(IMAGE):$(TAG) .

docker-push:
	docker push $(IMAGE):$(TAG)

dc-up:
	docker-compose up -d

dc-down:
	docker-compose down

## ── Kubernetes ───────────────────────────────────────────────────────────────

k8s-apply:
	kubectl apply -f k8s/

k8s-delete:
	kubectl delete -f k8s/
