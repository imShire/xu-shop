.PHONY: run-api run-worker migrate-up migrate-down migrate-create gen-openapi sync-docs docs lint test build seed-permissions

# 默认环境变量从 .env 读取，需先 cp .env.example .env
include .env
export

run-api:
	go run ./cmd/api

run-worker:
	go run ./cmd/worker

GOOSE := $(shell go env GOPATH)/bin/goose

migrate-up:
	$(GOOSE) -dir migrations postgres "$(DB_DSN)" up

migrate-down:
	$(GOOSE) -dir migrations postgres "$(DB_DSN)" down

migrate-create:
	@read -p "name: " name; \
	$(GOOSE) -dir migrations -s create $$name sql

gen-openapi:
	# 后端：oapi-codegen 生成 server stub
	# 前端：开发者在 admin/client 里执行 openapi-typescript
	@echo "TODO: oapi-codegen -package api api/openapi.yaml -o internal/server/openapi_gen.go"

# 每次修改 api/openapi.yaml 后执行此命令，同步嵌入用的副本
sync-docs:
	@cp api/openapi.yaml internal/docs/openapi.yaml
	@echo "✓ internal/docs/openapi.yaml synced from api/openapi.yaml"

# 启动 API 服务后，打开 Swagger UI（macOS）
docs:
	@open http://localhost:8080/docs 2>/dev/null || xdg-open http://localhost:8080/docs 2>/dev/null || echo "请在浏览器打开 http://localhost:8080/docs"

lint:
	go vet ./...
	golangci-lint run

test:
	go test ./... -race -cover

build:
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/api ./cmd/api
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/worker ./cmd/worker
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/cli ./cmd/cli

seed-permissions:
	go run ./cmd/cli seed permissions
