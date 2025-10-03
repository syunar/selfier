.PHONY: test cover cover-html

# Run all tests with coverage
test:
	go test ./... -cover

# Run all tests and output a detailed coverage profile
cover:
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out

cover-html:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
