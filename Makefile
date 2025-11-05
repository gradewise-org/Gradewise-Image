.PHONY: build test clean docker-build docker-run

# Build the harness binary
build:
	go build -o bin/harness ./cmd/harness
	go build -o bin/sshd-setup ./cmd/sshd-setup

# Run tests
test:
	go test ./...

# Run tests with coverage
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run unit tests
test-unit:
	go test -v ./...

# Run integration tests
test-integration:
	chmod +x test/integration_test.sh
	test/integration_test.sh

# Run all tests
test-all: test-unit test-integration

# Clean build artifacts
clean:
	rm -rf bin/
	go clean

# Build Docker image
docker-build:
	docker build -t gradewise-image:latest .

# Run Docker container (for testing)
docker-run:
	docker run -it --rm \
		-e ASSIGNMENT_TITLE="Test Assignment" \
		-e AUTHENTICATION_TOKEN="test-token" \
		-e SUBMISSION_URL="http://example.com/submission.tar.gz" \
		-e SUBMIT_RESULTS_URL="http://example.com/results" \
		gradewise-image:latest

# Format code
fmt:
	go fmt ./...

# Run linter
lint:
	golangci-lint run ./...

# Install dependencies
deps:
	go mod download
	go mod tidy
