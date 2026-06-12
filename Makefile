GO_BIN ?= /usr/local/go/bin/go
PORT ?= 9002
BASE_URL ?= http://127.0.0.1:$(PORT)

.PHONY: tidy build run restart smoke smoke-fresh cron-audit test

tidy:
	$(GO_BIN) mod tidy

build:
	$(GO_BIN) build -o ./bin/apotek ./cmd/app

run:
	PORT=$(PORT) $(GO_BIN) run ./cmd/app/main.go

restart:
	PORT=$(PORT) ./scripts/restart_local.sh

smoke:
	BASE_URL=$(BASE_URL) ./scripts/regression_inventory_smoke.py

smoke-fresh:
	PORT=$(PORT) ./scripts/fresh_clone_smoke.sh

cron-audit:
	PORT=$(PORT) ./scripts/cron_runtime_audit.sh

test:
	$(GO_BIN) test ./...
