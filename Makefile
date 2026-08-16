.PHONY: up down seed api web test

up:
	docker compose -f ../docker-compose.yml --project-directory .. up -d
	../scripts/ensure-databases.sh
	@echo "waiting for postgres…"
	@sleep 3

down:
	docker compose down

seed:
	cd apps/api && DATABASE_URL=$${DATABASE_URL:-postgresql://postgres:postgres@localhost:5432/garment_ppc} go run ./cmd/seed

api:
	cd apps/api && DATABASE_URL=$${DATABASE_URL:-postgresql://postgres:postgres@localhost:5432/garment_ppc} WEB_ORIGIN=$${WEB_ORIGIN:-http://localhost:3014} go run ./cmd/api

web:
	cd apps/web && npm run dev

test:
	cd apps/api && go test ./...
