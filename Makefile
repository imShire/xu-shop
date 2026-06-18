.PHONY: help deps-up deps-down server-run admin-run client-h5 client-mp migrate-up api-check

help:
	@echo "deps-up      - 启动本地依赖 (postgres/redis/minio/asynqmon)"
	@echo "deps-down    - 停止本地依赖"
	@echo "server-run   - 进入 server 并 make run-api"
	@echo "admin-run    - 进入 admin 并 pnpm dev"
	@echo "client-h5    - 进入 client 并 pnpm dev:h5"
	@echo "client-mp    - 进入 client 并 pnpm dev:weapp"
	@echo "migrate-up   - 运行数据库迁移"
	@echo "api-check    - 检查 admin 生成类型与 openapi.yaml 是否同步"

deps-up:
	docker compose -f deploy/docker-compose.dev.yml up -d

deps-down:
	docker compose -f deploy/docker-compose.dev.yml down

server-run:
	$(MAKE) -C server run-api

admin-run:
	cd admin && pnpm dev

client-h5:
	cd client && pnpm dev:h5

client-mp:
	cd client && pnpm dev:weapp

migrate-up:
	$(MAKE) -C server migrate-up

.PHONY: api-check
api-check:
	cd admin && pnpm gen:api:check
	@echo "✓ admin openapi types are in sync"