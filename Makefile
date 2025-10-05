.PHONY: test dev

# Run all tests with coverage
test:
	go test ./... -cover

# Run development environment
dev:
	air &
	npx inngest-cli@latest dev --no-discovery -u http://localhost:8080/api/inngest
