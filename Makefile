.PHONY: test dev wait-for-backend

# Run all tests with coverage
test:
	go test ./... -cover

# Run development environment
dev: wait-for-backend
	@echo "✅ Backend is up. Starting other processes..."
	@npx inngest-cli@latest dev --no-discovery -u http://localhost:8080/api/inngest &
	@npx openapi-typescript http://localhost:8080/openapi.json -o web/src/lib/api-types.ts &
	@cd web && npm run dev && cd ..

wait-for-backend:
	@echo "🚀 Starting Go backend with Air..."
	@air &

	@echo "⏳ Waiting for backend to start on port 8080..."
	@until curl -s http://localhost:8080/openapi.json >/dev/null; do \
		printf "."; \
		sleep 1; \
	done
	@echo "\n✅ Backend is ready!"
